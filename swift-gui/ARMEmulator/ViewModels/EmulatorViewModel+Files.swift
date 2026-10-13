import Foundation

/// The answer to "save the changes before replacing them?"
enum UnsavedChangesChoice {
    case save
    case discard
    case cancel
}

// MARK: - File Operations Extension

extension EmulatorViewModel {
    /// True when the editor holds text that differs from the last opened or saved text
    var hasUnsavedChanges: Bool {
        sourceCode != savedSource
    }

    /// Returns true when the caller may replace the editor text. Asks only when there are unsaved changes;
    /// a save that fails or is cancelled stops the caller.
    func confirmDiscardingChanges(
        ask: () -> UnsavedChangesChoice,
        save: () async -> Bool,
    ) async -> Bool {
        guard hasUnsavedChanges else { return true }
        switch ask() {
        case .save: return await save()
        case .discard: return true
        case .cancel: return false
        }
    }

    /// Asks with an alert and saves through `fileService`
    func confirmDiscardingChanges(using fileService: FileService) async -> Bool {
        await confirmDiscardingChanges(ask: UnsavedChangesAlert.ask) {
            await saveProgram(using: fileService)
        }
    }

    /// Saves to the current file, or asks for one. Returns true when the text was written.
    func saveProgram(using fileService: FileService) async -> Bool {
        guard let url = fileService.currentFileURL else {
            return await saveProgramAs(using: fileService)
        }
        return await record(fileService.saveFile(content: sourceCode, url: url))
    }

    /// Asks for a file and saves to it. Returns true when the text was written.
    func saveProgramAs(using fileService: FileService) async -> Bool {
        await record(fileService.saveFileAs(content: sourceCode))
    }

    private func record(_ saved: Bool) -> Bool {
        if saved {
            markSaved()
        }
        return saved
    }
}
