import AppKit
import Foundation

enum BackendError: Error, LocalizedError {
    case binaryNotFound
    case startupTimeout
    case startupFailed(String)
    case alreadyRunning

    var errorDescription: String? {
        switch self {
        case .binaryNotFound:
            return "ARM Emulator backend binary not found. Please rebuild the app."
        case .startupTimeout:
            return "Backend failed to start within timeout period."
        case let .startupFailed(message):
            return "Failed to start backend: \(message)"
        case .alreadyRunning:
            return "Backend is already running."
        }
    }
}

@MainActor
class BackendManager: ObservableObject {
    @Published var backendStatus: BackendStatus = .unknown

    private var backendProcess: Process?
    private var didStartBackend = false
    private let baseURL = URL(string: "http://localhost:8080")!
    private var processMonitorTask: Task<Void, Never>?

    enum BackendStatus: Equatable {
        case unknown
        case starting
        case running
        case stopped
        case error(String)
    }

    init() {
        NotificationCenter.default.addObserver(
            forName: NSApplication.willTerminateNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor in
                self?.shutdownSync()
            }
        }
    }

    deinit {
        processMonitorTask?.cancel()
        NotificationCenter.default.removeObserver(self)
    }

    private func shutdownSync() {
        let semaphore = DispatchSemaphore(value: 0)
        Task {
            await shutdown()
            semaphore.signal()
        }
        semaphore.wait()
    }

    func ensureBackendRunning() async {
        // Check if backend is already running
        if await checkBackendHealth() {
            backendStatus = .running
            didStartBackend = false
            return
        }

        // Backend not running, start it
        // Note: No need to kill orphaned backends anymore - the Go backend
        // uses parent process monitoring to auto-terminate when the Swift app dies
        do {
            try await startBackend()
        } catch {
            backendStatus = .error(error.localizedDescription)
        }
    }

    func checkBackendHealth() async -> Bool {
        let healthURL = baseURL.appendingPathComponent("/api/v1/session")

        var request = URLRequest(url: healthURL, timeoutInterval: 0.5)
        request.httpMethod = "GET"

        do {
            let (_, response) = try await URLSession.shared.data(for: request)
            if let httpResponse = response as? HTTPURLResponse {
                return (200 ... 499).contains(httpResponse.statusCode)
            }
            return false
        } catch {
            return false
        }
    }

    private func startBackend() async throws {
        backendStatus = .starting

        guard let binaryPath = findBinaryPath() else {
            throw BackendError.binaryNotFound
        }

        let process = Process()
        process.executableURL = binaryPath
        process.arguments = ["--api-server", "--port", "8080"]

        // Enable backend debug logging in DEBUG builds
        #if DEBUG
            var environment = ProcessInfo.processInfo.environment
            environment["ARM_EMULATOR_DEBUG"] = "1"
            process.environment = environment
            DebugLog.log("Starting backend with debug logging enabled", category: "BackendManager")
        #endif

        let outputPipe = Pipe()
        process.standardOutput = outputPipe
        process.standardError = outputPipe

        // In DEBUG builds, log backend output to Xcode console
        #if DEBUG
            outputPipe.fileHandleForReading.readabilityHandler = { fileHandle in
                let data = fileHandle.availableData
                if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                    DebugLog.log(
                        "Backend: \(output.trimmingCharacters(in: .whitespacesAndNewlines))",
                        category: "Backend"
                    )
                }
            }
        #endif

        process.terminationHandler = { [weak self] process in
            Task { @MainActor [weak self] in
                guard let self = self else { return }
                if self.didStartBackend, self.backendStatus == .running {
                    self
                        .backendStatus =
                        .error("Backend process terminated unexpectedly (exit code: \(process.terminationStatus))")
                }
            }
        }

        do {
            try process.run()
        } catch {
            throw BackendError.startupFailed(error.localizedDescription)
        }

        backendProcess = process
        didStartBackend = true

        try await waitForBackendReady(timeout: 10.0)
        backendStatus = .running
    }

    private func waitForBackendReady(timeout: TimeInterval) async throws {
        let deadline = Date().addingTimeInterval(timeout)

        while Date() < deadline {
            if await checkBackendHealth() {
                return
            }
            try await Task.sleep(nanoseconds: 200_000_000)
        }

        throw BackendError.startupTimeout
    }

    private func findBinaryPath() -> URL? {
        if let resourcePath = Bundle.main.resourceURL?.appendingPathComponent("arm-emulator"),
           FileManager.default.isExecutableFile(atPath: resourcePath.path)
        {
            return resourcePath
        }

        let projectRoot = FileManager.default.currentDirectoryPath
        let devPath = URL(fileURLWithPath: projectRoot).appendingPathComponent("arm-emulator")
        if FileManager.default.isExecutableFile(atPath: devPath.path) {
            return devPath
        }

        let parentDevPath = URL(fileURLWithPath: projectRoot)
            .deletingLastPathComponent()
            .appendingPathComponent("arm-emulator")
        if FileManager.default.isExecutableFile(atPath: parentDevPath.path) {
            return parentDevPath
        }

        return nil
    }

    func shutdown() async {
        processMonitorTask?.cancel()
        processMonitorTask = nil

        guard let process = backendProcess else {
            return
        }

        guard didStartBackend else {
            return
        }

        if process.isRunning {
            process.terminate()

            for _ in 0 ..< 15 {
                if !process.isRunning {
                    break
                }
                try? await Task.sleep(nanoseconds: 200_000_000)
            }

            if process.isRunning {
                kill(process.processIdentifier, SIGKILL)
            }
        }

        backendProcess = nil
        didStartBackend = false
        backendStatus = .stopped
    }

    func restartBackend() async {
        await shutdown()
        try? await Task.sleep(nanoseconds: 500_000_000)
        await ensureBackendRunning()
    }
}
