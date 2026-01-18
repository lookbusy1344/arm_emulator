import SwiftUI

// Environment key for accessing AppDelegate
private struct AppDelegateKey: EnvironmentKey {
    static let defaultValue: AppDelegate? = nil
}

extension EnvironmentValues {
    var appDelegate: AppDelegate? {
        get { self[AppDelegateKey.self] }
        set { self[AppDelegateKey.self] = newValue }
    }
}

@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    let backendManager = BackendManager()
    let fileService = FileService.shared
    let settings = AppSettings.shared
    weak var viewModel: EmulatorViewModel?

    // Startup file path from command-line arguments (parsed once on first access)
    private(set) lazy var startupFilePath: String? = {
        let args = CommandLine.arguments

        guard args.count > 1 else {
            return nil
        }

        // Find first argument that looks like an assembly file (ends with .s)
        // This filters out Xcode debug flags like "-NSDocumentRevisionsDebugMode YES"
        guard let filePath = args.dropFirst().first(where: { $0.hasSuffix(".s") }) else {
            return nil
        }

        // Convert to absolute path if relative
        let absolutePath = (filePath as NSString).expandingTildeInPath
        let url = URL(fileURLWithPath: absolutePath)

        // Verify file exists
        guard FileManager.default.fileExists(atPath: url.path) else {
            print("Warning: Startup file not found: \(url.path)")
            return nil
        }

        return url.path
    }()

    func applicationDidFinishLaunching(_: Notification) {
        // Install global key event monitor to handle function keys even when focus is lost
        NSEvent.addLocalMonitorForEvents(matching: .keyDown) { [weak self] event in
            if self?.handleFunctionKey(event) == true {
                return nil // Consume the event
            }
            return event // Pass through
        }
    }

    private func handleFunctionKey(_ event: NSEvent) -> Bool {
        guard let viewModel else { return false }

        // Handle both actual function keys (NSFxFunctionKey) AND their media key equivalents.
        // Modern Mac keyboards send different keyCodes depending on whether Fn is pressed:
        // - Without Fn: Media key codes (96, 100, 101, 103, 109)
        // - With Fn: Function key codes (NSF5FunctionKey=63240, NSF8FunctionKey=63243, etc.)
        // This ensures shortcuts work regardless of keyboard settings or focus state.
        let keyCode = Int(event.keyCode)

        switch keyCode {
        case NSF5FunctionKey, 96: // F5 (Run)
            Task { await viewModel.run() }
            return true
        case NSF8FunctionKey, 100, // F8 (Step Over)
             NSF10FunctionKey, 109: // F10 (Step Over)
            Task { await viewModel.stepOver() }
            return true
        case NSF11FunctionKey, 103: // F11 (Step)
            Task { await viewModel.step() }
            return true
        case NSF9FunctionKey, 101: // F9 (Toggle Breakpoint)
            Task { await viewModel.toggleBreakpoint(at: viewModel.currentPC) }
            return true
        default:
            return false
        }
    }
}

@main
struct ARMEmulatorApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @State private var showingAbout = false

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(appDelegate.backendManager)
                .environmentObject(appDelegate.fileService)
                .environmentObject(appDelegate.settings)
                .environment(\.appDelegate, appDelegate)
                .task {
                    await appDelegate.backendManager.ensureBackendRunning()
                }
                .sheet(isPresented: $showingAbout) {
                    AboutView()
                }
        }
        .commands {
            CommandGroup(replacing: .appInfo) {
                Button("About ARM Emulator") {
                    showingAbout = true
                }
            }
            FileCommands()
            DebugCommands()
        }
        .windowStyle(.titleBar)
        .windowToolbarStyle(.unified)

        Settings {
            PreferencesView()
                .environmentObject(appDelegate.settings)
        }
    }
}
