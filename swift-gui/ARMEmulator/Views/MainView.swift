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
                                MainViewToolbar(viewModel: viewModel)
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

                VStack(spacing: 0) {
                    // Compact popup picker navigation
                    HStack {
                        Text("View:")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)

                        Picker("", selection: $settings.selectedTab) {
                            Label("Registers", systemImage: "cpu").tag(0)
                            Label("Memory", systemImage: "memorychip").tag(1)
                            Label("Stack", systemImage: "square.stack.3d.down.right").tag(2)
                            Label("Disassembly", systemImage: "hammer").tag(3)
                            Label("Evaluator", systemImage: "function").tag(4)
                            Label("Watchpoints", systemImage: "eye").tag(5)
                            Label("Breakpoints", systemImage: "circle.hexagongrid").tag(6)
                        }
                        .pickerStyle(.menu)
                        .frame(width: 150)

                        Spacer()
                    }
                    .padding(8)
                    .background(Color(NSColor.controlBackgroundColor))

                    Divider()

                    // Content based on selected tab
                    Group {
                        switch settings.selectedTab {
                        case 0:
                            VStack(spacing: 0) {
                                RegistersView(
                                    registers: viewModel.registers,
                                    changedRegisters: viewModel.changedRegisters
                                )
                                .frame(minHeight: 200)

                                Divider()

                                StatusView(
                                    status: viewModel.status,
                                    pc: viewModel.currentPC
                                )
                                .frame(height: 60)
                            }
                        case 1:
                            MemoryView(viewModel: viewModel)
                        case 2:
                            StackView(viewModel: viewModel)
                        case 3:
                            DisassemblyView(viewModel: viewModel)
                        case 4:
                            ExpressionEvaluatorView(viewModel: viewModel)
                        case 5:
                            WatchpointsView(viewModel: viewModel)
                        case 6:
                            BreakpointsListView(viewModel: viewModel)
                        default:
                            VStack(spacing: 0) {
                                RegistersView(
                                    registers: viewModel.registers,
                                    changedRegisters: viewModel.changedRegisters
                                )
                                .frame(minHeight: 200)

                                Divider()

                                StatusView(
                                    status: viewModel.status,
                                    pc: viewModel.currentPC
                                )
                                .frame(height: 60)
                            }
                        }
                    }
                }
                .frame(idealWidth: 665, maxWidth: 800)
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
