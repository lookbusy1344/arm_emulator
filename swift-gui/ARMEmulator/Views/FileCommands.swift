import SwiftUI

struct FileCommands: Commands {
    @FocusedValue(\.viewModel) var viewModel: EmulatorViewModel?
    let backendManager: BackendManager

    private var fileService: FileService {
        FileService.shared
    }

    var body: some Commands {
        CommandGroup(replacing: .newItem) {
            Button("Open...") {
                Task {
                    guard await confirmDiscardingChanges() else { return }
                    if let content = await fileService.openFile() {
                        await viewModel?.openProgram(source: content)
                    }
                }
            }
            .keyboardShortcut("o", modifiers: .command)

            Button("Save") {
                Task {
                    _ = await viewModel?.saveProgram(using: fileService)
                }
            }
            .keyboardShortcut("s", modifiers: .command)
            .disabled(viewModel == nil)

            Button("Save As...") {
                Task {
                    _ = await viewModel?.saveProgramAs(using: fileService)
                }
            }
            .keyboardShortcut("s", modifiers: [.command, .shift])
            .disabled(viewModel == nil)

            Divider()

            Menu("Open Recent") {
                if fileService.recentFiles.isEmpty {
                    Text("No Recent Files")
                        .disabled(true)
                } else {
                    ForEach(fileService.recentFiles, id: \.self) { url in
                        Button(url.lastPathComponent) {
                            Task {
                                guard await confirmDiscardingChanges() else { return }
                                if let content = try? String(contentsOf: url, encoding: .utf8) {
                                    await viewModel?.openProgram(source: content)
                                    fileService.currentFileURL = url
                                }
                            }
                        }
                    }

                    Divider()

                    Button("Clear Menu") {
                        fileService.clearRecentFiles()
                    }
                }
            }

            Button("Open Example...") {
                // Will be handled by MainView showing sheet
                NotificationCenter.default.post(name: .showExamplesBrowser, object: nil)
            }
            .keyboardShortcut("e", modifiers: [.command, .shift])

            Divider()

            Button("Restart Backend") {
                Task {
                    await restartBackend()
                }
            }
        }
    }

    /// Restarts the backend; the view model, when there is one, reconnects and reloads its program
    private func restartBackend() async {
        let restart = {
            await backendManager.restartBackend()
            return backendManager.backendStatus == .running
        }
        if let viewModel {
            await viewModel.restartBackend(restart)
        } else {
            _ = await restart()
        }
    }

    /// True when there is no view model or its edits may be replaced
    private func confirmDiscardingChanges() async -> Bool {
        await viewModel?.confirmDiscardingChanges(using: fileService) ?? true
    }
}

/// Notification for showing examples browser
extension Notification.Name {
    static let showExamplesBrowser = Notification.Name("showExamplesBrowser")
}

/// FocusedValue key for accessing ViewModel from commands
struct ViewModelKey: FocusedValueKey {
    typealias Value = EmulatorViewModel
}

extension FocusedValues {
    var viewModel: EmulatorViewModel? {
        get { self[ViewModelKey.self] }
        set { self[ViewModelKey.self] = newValue }
    }
}
