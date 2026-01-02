import SwiftUI

struct MainView: View {
    @EnvironmentObject var backendManager: BackendManager
    @EnvironmentObject var fileService: FileService
    @EnvironmentObject var settings: AppSettings
    @StateObject private var viewModel = EmulatorViewModel()
    @State private var showingError = false
    @State private var showingExamplesBrowser = false

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
        .onDisappear {
            viewModel.cleanup()
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
                    .frame(minWidth: 300)

                TabView(selection: $settings.selectedTab) {
                    VStack(spacing: 0) {
                        RegistersView(registers: viewModel.registers)
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
                }
                .frame(minWidth: 250, maxWidth: 500)
            }
            .frame(minHeight: 300)

            ConsoleView(
                output: viewModel.consoleOutput,
                onSendInput: { input in
                    Task {
                        await viewModel.sendInput(input)
                    }
                }
            )
            .frame(minHeight: 150)
        }
    }

    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItemGroup(placement: .automatic) {
            Button(
                action: { Task { await viewModel.loadProgram(source: viewModel.sourceCode) } },
                label: { Label("Load", systemImage: "doc.text") }
            )
            .help("Load program (⌘L)")
            .keyboardShortcut("l", modifiers: .command)

            Divider()

            Button(
                action: { Task { await viewModel.run() } },
                label: { Label("Run", systemImage: "play.fill") }
            )
            .help("Run program (⌘R)")
            .keyboardShortcut("r", modifiers: .command)
            .disabled(viewModel.status == .running)

            Button(
                action: { Task { await viewModel.stop() } },
                label: { Label("Stop", systemImage: "stop.fill") }
            )
            .help("Stop execution (⌘.)")
            .keyboardShortcut(".", modifiers: .command)
            .disabled(viewModel.status != .running)

            Button(
                action: { Task { await viewModel.step() } },
                label: { Label("Step", systemImage: "forward.frame") }
            )
            .help("Step one instruction (⌘T)")
            .keyboardShortcut("t", modifiers: .command)
            .disabled(viewModel.status == .running)

            Button(
                action: { Task { await viewModel.reset() } },
                label: { Label("Reset", systemImage: "arrow.counterclockwise") }
            )
            .help("Reset VM (⌘⇧R)")
            .keyboardShortcut("r", modifiers: [.command, .shift])
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
                    .font(.system(.body, design: .monospaced))
                    .fontWeight(.medium)
            }
            .padding(.horizontal)

            Divider()

            HStack(spacing: 4) {
                Text("PC:")
                    .font(.system(.caption, design: .monospaced))
                    .foregroundColor(.secondary)

                Text(String(format: "0x%08X", pc))
                    .font(.system(.body, design: .monospaced))
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
