import SwiftUI

struct MainView: View {
    @EnvironmentObject var backendManager: BackendManager
    @EnvironmentObject var fileService: FileService
    @EnvironmentObject var settings: AppSettings
    @Environment(\.appDelegate) var appDelegate
    @StateObject private var viewModel = EmulatorViewModel()
    @State private var showingError = false
    @State private var showingExamplesBrowser = false
    @State private var hasLoadedStartupFile = false

    var body: some View {
        ZStack {
            if backendManager.backendStatus == .running {
                VStack(spacing: 0) {
                    if !viewModel.isConnected {
                        ConnectionView(viewModel: viewModel)
                    } else {
                        contentView
                            .toolbar {
                                toolbarContent
                            }
                    }
                }
                .frame(minWidth: 800, minHeight: 600)
            } else {
                BackendStatusView(status: backendManager.backendStatus) {
                    await backendManager.restartBackend()
                }
            }
        }
        .alert(
            "Error",
            isPresented: $showingError,
            presenting: viewModel.errorMessage,
            actions: { _ in
                Button("OK") {
                    viewModel.errorMessage = nil
                }
            },
            message: { message in
                Text(message)
            }
        )
        .onChange(of: viewModel.errorMessage) { newValue in
            showingError = newValue != nil
        }
        .onChange(of: backendManager.backendStatus) { newStatus in
            if newStatus == .running, !viewModel.isConnected {
                Task {
                    await viewModel.initialize()
                }
            }
        }
        .onChange(of: viewModel.isConnected) { isConnected in
            // Auto-load startup file if specified via command-line arguments
            if isConnected, !hasLoadedStartupFile {
                hasLoadedStartupFile = true
                if let startupPath = appDelegate?.startupFilePath {
                    Task {
                        await loadStartupFile(path: startupPath)
                    }
                }
            }
        }
        .onDisappear {
            viewModel.cleanup()
        }
        .onAppear {
            // Wire viewModel to AppDelegate for global function key handling
            appDelegate?.viewModel = viewModel
        }
        .focusedSceneValue(\.viewModel, viewModel)
        .preferredColorScheme(settings.preferredColorScheme)
        .onReceive(NotificationCenter.default.publisher(for: .showExamplesBrowser)) { _ in
            showingExamplesBrowser = true
        }
        .sheet(isPresented: $showingExamplesBrowser) {
            ExamplesBrowserView { example in
                Task {
                    if let content = try? String(contentsOf: example.url, encoding: .utf8) {
                        await viewModel.loadProgram(source: content)
                        fileService.currentFileURL = example.url
                    }
                }
                showingExamplesBrowser = false
            }
        }
    }

    private var contentView: some View {
        VSplitView {
            HSplitView {
                EditorView(text: $viewModel.sourceCode)
                    .environmentObject(viewModel)
                    .frame(minWidth: 150)

                TabView(selection: $settings.selectedTab) {
                    VStack(spacing: 0) {
                        RegistersView(registers: viewModel.registers, changedRegisters: viewModel.changedRegisters)
                            .frame(minHeight: 200)

                        Divider()

                        StatusView(
                            status: viewModel.status,
                            pc: viewModel.currentPC
                        )
                        .frame(height: 60)
                    }
                    .tabItem {
                        Label("Registers", systemImage: "cpu")
                    }
                    .tag(0)

                    MemoryView(viewModel: viewModel)
                        .tabItem {
                            Label("Memory", systemImage: "memorychip")
                        }
                        .tag(1)

                    StackView(viewModel: viewModel)
                        .tabItem {
                            Label("Stack", systemImage: "square.stack.3d.down.right")
                        }
                        .tag(2)

                    DisassemblyView(viewModel: viewModel)
                        .tabItem {
                            Label("Disassembly", systemImage: "hammer")
                        }
                        .tag(3)

                    ExpressionEvaluatorView(viewModel: viewModel)
                        .tabItem {
                            Label("Evaluator", systemImage: "function")
                        }
                        .tag(4)

                    WatchpointsView(viewModel: viewModel)
                        .tabItem {
                            Label("Watchpoints", systemImage: "eye")
                        }
                        .tag(5)

                    BreakpointsListView(viewModel: viewModel)
                        .tabItem {
                            Label("Breakpoints", systemImage: "circle.hexagongrid")
                        }
                        .tag(6)
                }
                .frame(minWidth: 250, maxWidth: 800)
            }
            .frame(minHeight: 300)

            ConsoleView(
                output: viewModel.consoleOutput,
                isWaitingForInput: viewModel.status == .waitingForInput,
                onSendInput: { input in
                    Task {
                        await viewModel.sendInput(input)
                    }
                }
            )
            .frame(minHeight: 150)
        }
    }

    private func loadStartupFile(path: String) async {
        let url = URL(fileURLWithPath: path)

        // Validate file exists
        guard FileManager.default.fileExists(atPath: path) else {
            viewModel.errorMessage = "Could not load '\(url.lastPathComponent)': File not found"
            return
        }

        // Validate file extension
        guard url.pathExtension == "s" else {
            viewModel.errorMessage = "Could not load '\(url.lastPathComponent)': Not an assembly file (.s)"
            return
        }

        // Try to read file
        do {
            let content = try String(contentsOf: url, encoding: .utf8)
            await viewModel.loadProgram(source: content)
            fileService.currentFileURL = url
            fileService.addToRecentFiles(url)
        } catch {
            viewModel.errorMessage = "Could not load '\(url.lastPathComponent)': \(error.localizedDescription)"
        }
    }

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .automatic) {
            // Status indicator with icon
            HStack(spacing: 4) {
                Image(systemName: statusIcon)
                    .foregroundColor(statusColor)
                    .font(.system(size: 11, weight: .semibold))
                Text(statusText)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            .padding(.horizontal, 4)

            Divider()

            Button(
                action: { Task { await viewModel.loadProgram(source: viewModel.sourceCode) } },
                label: { Label("Load", systemImage: "doc.text") }
            )
            .help("Load program (⌘L)")
            .keyboardShortcut("l", modifiers: .command)

            Divider()

            Button(
                action: {
                    DebugLog.ui("Run/Continue button clicked")
                    Task { await viewModel.run() }
                },
                label: {
                    Label(
                        viewModel.status == .paused ? "Continue" : "Run",
                        systemImage: viewModel.status == .paused ? "play.circle.fill" : "play.fill"
                    )
                }
            )
            .help(viewModel.status == .paused ? "Continue execution (⌘R)" : "Run program (⌘R)")
            .keyboardShortcut("r", modifiers: .command)
            .disabled(viewModel.status == .running || viewModel.status == .waitingForInput)

            Button(
                action: { Task { await viewModel.stop() } },
                label: { Label("Stop", systemImage: "stop.fill") }
            )
            .help("Stop execution (⌘.)")
            .keyboardShortcut(".", modifiers: .command)
            .disabled(viewModel.status != .running && viewModel.status != .waitingForInput)

            Button(
                action: { Task { await viewModel.step() } },
                label: { Label("Step", systemImage: "forward.frame") }
            )
            .help("Step one instruction (⌘T)")
            .keyboardShortcut("t", modifiers: .command)
            .disabled(viewModel.status == .running || viewModel.status == .waitingForInput)

            Button(
                action: { Task { await viewModel.stepOver() } },
                label: { Label("Step Over", systemImage: "arrow.right.to.line") }
            )
            .help("Step over function calls (⌘⇧T)")
            .keyboardShortcut("t", modifiers: [.command, .shift])
            .disabled(viewModel.status == .running || viewModel.status == .waitingForInput)

            Button(
                action: { Task { await viewModel.stepOut() } },
                label: { Label("Step Out", systemImage: "arrow.up.left") }
            )
            .help("Step out of current function (⌘⌥T)")
            .keyboardShortcut("t", modifiers: [.command, .option])
            .disabled(viewModel.status == .running || viewModel.status == .waitingForInput)

            Button(
                action: { Task { await viewModel.reset() } },
                label: { Label("Reset", systemImage: "arrow.counterclockwise") }
            )
            .help("Reset VM (⌘⇧R)")
            .keyboardShortcut("r", modifiers: [.command, .shift])

            Divider()

            Button(
                action: { viewModel.scrollToCurrentPC?() },
                label: { Label("Show PC", systemImage: "arrow.down.to.line") }
            )
            .help("Scroll to current PC (⌘J)")
            .keyboardShortcut("j", modifiers: .command)
            .disabled(viewModel.currentPC == 0)
        }
    }

    private var statusIcon: String {
        switch viewModel.status {
        case .running:
            return "play.fill" // Green arrow (running)
        case .paused:
            return "pause.fill" // Pause symbol (stepping/paused)
        case .halted, .idle:
            return "stop.fill" // Red square (stopped/not executing)
        case .waitingForInput:
            return "keyboard.fill" // Keyboard icon (waiting for input)
        case .error:
            return "exclamationmark.triangle.fill"
        }
    }

    private var statusColor: Color {
        switch viewModel.status {
        case .running:
            return .green
        case .paused:
            return .orange
        case .halted, .idle:
            return .red
        case .waitingForInput:
            return .orange
        case .error:
            return .red
        }
    }

    private var statusText: String {
        switch viewModel.status {
        case .running:
            return "Running"
        case .paused:
            return "Paused"
        case .halted:
            return "Halted"
        case .idle:
            return "Idle"
        case .waitingForInput:
            return "Waiting for Input"
        case .error:
            return "Error"
        }
    }
}

struct ConnectionView: View {
    @ObservedObject var viewModel: EmulatorViewModel

    var body: some View {
        VStack(spacing: 20) {
            ProgressView()

            Text("Connecting...")
                .foregroundColor(.secondary)

            Button("Retry") {
                Task {
                    await viewModel.initialize()
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .task {
            await viewModel.initialize()
        }
    }
}

struct StatusView: View {
    let status: VMState
    let pc: UInt32

    var body: some View {
        HStack {
            HStack(spacing: 6) {
                Circle()
                    .fill(statusColor)
                    .frame(width: 10, height: 10)

                Text(status.rawValue.capitalized)
                    .font(.system(size: 10, design: .monospaced))
                    .fontWeight(.medium)
            }
            .padding(.horizontal)

            Divider()

            HStack(spacing: 4) {
                Text("PC:")
                    .font(.system(size: 10, design: .monospaced))
                    .foregroundColor(.secondary)

                Text(String(format: "0x%08X", pc))
                    .font(.system(size: 10, design: .monospaced))
            }
            .padding(.horizontal)

            Spacer()
        }
        .padding(.vertical, 8)
        .background(Color(NSColor.controlBackgroundColor))
    }

    private var statusColor: Color {
        switch status {
        case .idle:
            return .gray
        case .running:
            return .green
        case .paused:
            return .orange
        case .halted:
            return .blue
        case .waitingForInput:
            return .orange
        case .error:
            return .red
        }
    }
}

struct MainView_Previews: PreviewProvider {
    static var previews: some View {
        MainView()
    }
}
