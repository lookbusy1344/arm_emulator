import SwiftUI

struct DebugCommands: Commands {
    @FocusedValue(\.viewModel) var viewModel: EmulatorViewModel?

    var body: some Commands {
        CommandMenu("Debug") {
            Button("Run/Continue") {
                Task {
                    await viewModel?.run()
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF5FunctionKey)!)), modifiers: [.function])
            .disabled(viewModel?.status == .running)

            Button("Stop") {
                Task {
                    await viewModel?.stop()
                }
            }
            .disabled(viewModel?.status != .running)

            Divider()

            Button("Step") {
                Task {
                    await viewModel?.step()
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF11FunctionKey)!)), modifiers: [.function])
            .disabled(viewModel?.status == .running)

            Button("Step Over") {
                Task {
                    await viewModel?.stepOver()
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF10FunctionKey)!)), modifiers: [.function])
            .disabled(viewModel?.status == .running)

            Button("Step Out") {
                Task {
                    await viewModel?.stepOut()
                }
            }
            .disabled(viewModel?.status == .running)

            Divider()

            Button("Toggle Breakpoint") {
                Task {
                    guard let vm = viewModel else { return }
                    await vm.toggleBreakpoint(at: vm.currentPC)
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF9FunctionKey)!)), modifiers: [.function])
            .disabled(viewModel == nil)

            Divider()

            Button("Reset") {
                Task {
                    await viewModel?.reset()
                }
            }
        }
    }
}
