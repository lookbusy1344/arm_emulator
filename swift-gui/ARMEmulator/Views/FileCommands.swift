import SwiftUI

struct FileCommands: Commands {
    @EnvironmentObject var fileService: FileService
    @FocusedValue(\.viewModel) var viewModel: EmulatorViewModel?

    var body: some Commands {
        CommandGroup(replacing: .newItem) {
            Button("Open...") {
                Task {
                    if let content = await fileService.openFile() {
                        await viewModel?.loadProgram(source: content)
                    }
                }
            }
            .keyboardShortcut("o", modifiers: .command)

            Button("Save") {
                Task {
                    guard let vm = viewModel else { return }
                    if let url = fileService.currentFileURL {
                        _ = await fileService.saveFile(content: vm.sourceCode, url: url)
                    } else {
                        _ = await fileService.saveFileAs(content: vm.sourceCode)
                    }
                }
            }
            .keyboardShortcut("s", modifiers: .command)
            .disabled(viewModel == nil)

            Button("Save As...") {
                Task {
                    guard let vm = viewModel else { return }
                    _ = await fileService.saveFileAs(content: vm.sourceCode)
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
                                if let content = try? String(contentsOf: url, encoding: .utf8) {
                                    await viewModel?.loadProgram(source: content)
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
        }
    }
}

// Notification for showing examples browser
extension Notification.Name {
    static let showExamplesBrowser = Notification.Name("showExamplesBrowser")
}

// FocusedValue key for accessing ViewModel from commands
struct ViewModelKey: FocusedValueKey {
    typealias Value = EmulatorViewModel
}

extension FocusedValues {
    var viewModel: EmulatorViewModel? {
        get { self[ViewModelKey.self] }
        set { self[ViewModelKey.self] = newValue }
    }
}
