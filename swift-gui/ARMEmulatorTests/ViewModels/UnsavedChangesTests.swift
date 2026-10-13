import XCTest
@testable import ARMEmulator

@MainActor
final class UnsavedChangesTests: XCTestCase {
    private var viewModel: EmulatorViewModel!

    override func setUp() async throws {
        viewModel = EmulatorViewModel(apiClient: MockAPIClient(), wsClient: MockWebSocketClient())
        await viewModel.initialize()
    }

    override func tearDown() async throws {
        viewModel = nil
    }

    private struct Outcome {
        let proceed: Bool
        let asked: Bool
        let saved: Bool
    }

    private func confirm(
        choice: UnsavedChangesChoice,
        saveSucceeds: Bool = true,
    ) async -> Outcome {
        var asked = false
        var saved = false
        let proceed = await viewModel.confirmDiscardingChanges(
            ask: {
                asked = true
                return choice
            },
            save: {
                saved = true
                return saveSucceeds
            },
        )
        return Outcome(proceed: proceed, asked: asked, saved: saved)
    }

    // MARK: - Unsaved state

    func testEmptyEditorHasNoUnsavedChanges() {
        XCTAssertFalse(viewModel.hasUnsavedChanges)
    }

    func testEditingMakesChangesUnsaved() {
        viewModel.sourceCode = "MOV R0, #1"

        XCTAssertTrue(viewModel.hasUnsavedChanges)
    }

    func testOpeningAProgramIsTheSavedState() async {
        await viewModel.openProgram(source: "MOV R0, #1")

        XCTAssertFalse(viewModel.hasUnsavedChanges)
    }

    func testEditAfterOpeningIsUnsaved() async {
        await viewModel.openProgram(source: "MOV R0, #1")

        viewModel.sourceCode = "MOV R0, #2"

        XCTAssertTrue(viewModel.hasUnsavedChanges)
    }

    func testUndoingAnEditBackToTheSavedTextIsSaved() async {
        await viewModel.openProgram(source: "MOV R0, #1")
        viewModel.sourceCode = "MOV R0, #2"

        viewModel.sourceCode = "MOV R0, #1"

        XCTAssertFalse(viewModel.hasUnsavedChanges)
    }

    func testMarkSavedRecordsTheCurrentText() {
        viewModel.sourceCode = "MOV R0, #1"

        viewModel.markSaved()

        XCTAssertFalse(viewModel.hasUnsavedChanges)
    }

    func testOpeningAProgramThatFailsToAssembleIsStillTheSavedState() async {
        await viewModel.openProgram(source: "NOT AN INSTRUCTION")

        XCTAssertFalse(viewModel.hasUnsavedChanges)
    }

    // MARK: - Confirmation

    func testNoChangesProceedsWithoutAsking() async {
        let result = await confirm(choice: .cancel)

        XCTAssertTrue(result.proceed)
        XCTAssertFalse(result.asked)
        XCTAssertFalse(result.saved)
    }

    func testDiscardProceedsWithoutSaving() async {
        viewModel.sourceCode = "MOV R0, #1"

        let result = await confirm(choice: .discard)

        XCTAssertTrue(result.proceed)
        XCTAssertTrue(result.asked)
        XCTAssertFalse(result.saved)
    }

    func testCancelStops() async {
        viewModel.sourceCode = "MOV R0, #1"

        let result = await confirm(choice: .cancel)

        XCTAssertFalse(result.proceed)
        XCTAssertFalse(result.saved)
        XCTAssertTrue(viewModel.hasUnsavedChanges)
    }

    func testSaveProceedsWhenTheSaveSucceeds() async {
        viewModel.sourceCode = "MOV R0, #1"

        let result = await confirm(choice: .save)

        XCTAssertTrue(result.proceed)
        XCTAssertTrue(result.saved)
    }

    func testSaveStopsWhenTheSaveFails() async {
        viewModel.sourceCode = "MOV R0, #1"

        let result = await confirm(choice: .save, saveSucceeds: false)

        XCTAssertFalse(result.proceed)
        XCTAssertTrue(result.saved)
        XCTAssertTrue(viewModel.hasUnsavedChanges)
    }
}
