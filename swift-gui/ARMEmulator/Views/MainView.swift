import SwiftUI

struct MainView: View {
    @EnvironmentObject var backendManager: BackendManager
    @EnvironmentObject var fileService: FileService
    @EnvironmentObject var settings: AppSettings
    @Environment(\.appDelegate) var appDelegate
    @StateObject private var viewModel = EmulatorViewModel(endpoint: .fromSettings)
    @State private var showingError = false
    @State private var showingExamplesBrowser = false
    @State private var hasLoadedStartupFile = false
    @State private var editorTextView: NSTextView?
    @FocusState private var focusedPane: Pane?

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
            },
        )
        .onChange(of: viewModel.errorMessage) { _, newValue in
            showingError = newValue != nil
        }
        .onChange(of: backendManager.backendStatus) { _, newStatus in
            if newStatus == .running, !viewModel.isConnected {
                Task {
                    await viewModel.initialize()
                }
            }
        }
        .onChange(of: viewModel.isConnected) { _, isConnected in
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
        .navigationTitle(WindowTitle.format(
            fileURL: fileService.currentFileURL,
            isDirty: viewModel.hasUnsavedChanges,
        ))
        .background(MainWindowSetup(isDocumentEdited: viewModel.hasUnsavedChanges) { [viewModel, fileService] in
            WindowCloseGuard(
                hasUnsavedChanges: { viewModel.hasUnsavedChanges },
                confirm: { await viewModel.confirmDiscardingChanges(using: fileService) },
            )
        })
        .preferredColorScheme(settings.preferredColorScheme)
        .onReceive(NotificationCenter.default.publisher(for: .showExamplesBrowser)) { _ in
            showingExamplesBrowser = true
        }
        .sheet(isPresented: $showingExamplesBrowser) {
            ExamplesBrowserView { example in
                Task {
                    guard await viewModel.confirmDiscardingChanges(using: fileService) else { return }
                    if let content = try? String(contentsOf: example.url, encoding: .utf8) {
                        await viewModel.openProgram(source: content)
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
                EditorView(text: $viewModel.sourceCode, fontSize: CGFloat(settings.editorFontSize)) { textView in
                    editorTextView = textView
                }
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
                    .frame(height: 40)
                    .padding(.horizontal, 8)
                    .background(Color(NSColor.controlBackgroundColor))
                    // Focusable whatever the keyboard navigation setting; arrows switch panels
                    .focusable()
                    .focused($focusedPane, equals: .inspector)
                    .onKeyPress(.upArrow) { selectPanel(forward: false) }
                    .onKeyPress(.downArrow) { selectPanel(forward: true) }

                    Divider()

                    // Content based on selected tab
                    Group {
                        switch settings.selectedTab {
                        case 0:
                            VStack(spacing: 0) {
                                RegistersView(
                                    registers: viewModel.registers,
                                    registerHighlights: viewModel.registerHighlights,
                                )
                                .frame(minHeight: 200)

                                Divider()

                                StatusView(
                                    status: viewModel.status,
                                    cpsr: viewModel.registers.cpsr,
                                    cpsrHighlightID: viewModel.registerHighlights["CPSR"],
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
                                    registerHighlights: viewModel.registerHighlights,
                                )
                                .frame(minHeight: 200)

                                Divider()

                                StatusView(
                                    status: viewModel.status,
                                    cpsr: viewModel.registers.cpsr,
                                    cpsrHighlightID: viewModel.registerHighlights["CPSR"],
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
                },
                paneFocus: $focusedPane,
            )
            .frame(minHeight: 150)
        }
        .focusedSceneValue(\.paneNavigator, PaneNavigator(move: movePane))
    }

    /// Focuses the pane after (or before) the focused one. The editor is AppKit, outside SwiftUI focus state.
    private func movePane(forward: Bool) {
        let editorFocused = editorTextView.map { $0.window?.firstResponder === $0 } ?? false
        switch Pane.next(after: editorFocused ? .editor : focusedPane, forward: forward) {
        case .editor:
            focusedPane = nil
            if let editorTextView {
                editorTextView.window?.makeFirstResponder(editorTextView)
            }
        case let pane:
            focusedPane = pane
        }
    }

    private func selectPanel(forward: Bool) -> KeyPress.Result {
        settings.selectedTab = InspectorPanel.cycle(from: settings.selectedTab, forward: forward)
        return .handled
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
            await viewModel.openProgram(source: content)
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
    let cpsr: CPSRFlags
    let cpsrHighlightID: UUID?

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
                Text("CPSR:")
                    .font(.system(size: 10, design: .monospaced))
                    .foregroundColor(cpsrHighlightID != nil ? .green : .secondary)
                    .animation(.easeOut(duration: 1.5), value: cpsrHighlightID)

                ForEach(cpsr.items, id: \.name) { flag in
                    FlagPill(flag: flag)
                }
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
            .gray
        case .running:
            .green
        case .breakpoint:
            .orange
        case .halted:
            .blue
        case .waitingForInput:
            .orange
        case .error:
            .red
        }
    }
}

/// One CPSR flag: filled when set, outlined when clear
struct FlagPill: View {
    let flag: CPSRFlag

    var body: some View {
        Text(flag.name)
            .font(.system(size: 10, weight: .semibold, design: .monospaced))
            .foregroundColor(flag.isSet ? .white : .secondary)
            .padding(.horizontal, 6)
            .padding(.vertical, 1)
            .background(Capsule().fill(flag.isSet ? Color.accentColor : .clear))
            .overlay(Capsule().strokeBorder(flag.isSet ? Color.accentColor : Color.secondary.opacity(0.5)))
            .accessibilityLabel("\(flag.name) flag \(flag.isSet ? "set" : "clear")")
    }
}

struct MainView_Previews: PreviewProvider {
    static var previews: some View {
        MainView()
    }
}
