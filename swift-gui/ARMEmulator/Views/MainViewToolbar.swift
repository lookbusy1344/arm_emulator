import SwiftUI

struct MainViewToolbar: ToolbarContent {
    @ObservedObject var viewModel: EmulatorViewModel

    var body: some ToolbarContent {
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
            .padding(.leading, 12)
            .padding(.trailing, 4)

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
                        viewModel.status == .breakpoint ? "Continue" : "Run",
                        systemImage: viewModel.status == .breakpoint ? "play.circle.fill" : "play.fill"
                    )
                }
            )
            .help(viewModel.status == .breakpoint ? "Continue execution (⌘R)" : "Run program (⌘R)")
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
        case .breakpoint:
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
        case .breakpoint:
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
        case .breakpoint:
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
