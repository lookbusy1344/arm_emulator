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
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF5FunctionKey)!)), modifiers: [])
            .disabled(viewModel?.status == .running)

            Button("Pause") {
                Task {
                    await viewModel?.pause()
                }
            }
            .disabled(!(viewModel?.canPause ?? false))

            Divider()

            Button("Step") {
                Task {
                    await viewModel?.step()
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF11FunctionKey)!)), modifiers: [])
            .disabled(!(viewModel?.canStep ?? false))

            Button("Step Over") {
                Task {
                    await viewModel?.stepOver()
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF10FunctionKey)!)), modifiers: [])
            .disabled(!(viewModel?.canStep ?? false))

            Button("Step Out") {
                Task {
                    await viewModel?.stepOut()
                }
            }
            .disabled(!(viewModel?.canStep ?? false))

            Divider()

            Button("Toggle Breakpoint") {
                Task {
                    guard let vm = viewModel else { return }
                    await vm.toggleBreakpoint(at: vm.currentPC)
                }
            }
            .keyboardShortcut(KeyEquivalent(Character(UnicodeScalar(NSF9FunctionKey)!)), modifiers: [])
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
