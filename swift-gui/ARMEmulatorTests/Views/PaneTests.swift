import XCTest
@testable import ARMEmulator

final class PaneTests: XCTestCase {
    func testForwardFromNoFocusIsTheEditor() {
        XCTAssertEqual(Pane.next(after: nil, forward: true), .editor)
    }

    func testBackwardFromNoFocusIsTheConsole() {
        XCTAssertEqual(Pane.next(after: nil, forward: false), .console)
    }

    func testForwardVisitsEditorInspectorConsoleInTurn() {
        XCTAssertEqual(Pane.next(after: .editor, forward: true), .inspector)
        XCTAssertEqual(Pane.next(after: .inspector, forward: true), .console)
    }

    func testForwardFromTheConsoleWrapsToTheEditor() {
        XCTAssertEqual(Pane.next(after: .console, forward: true), .editor)
    }

    func testBackwardFromTheEditorWrapsToTheConsole() {
        XCTAssertEqual(Pane.next(after: .editor, forward: false), .console)
    }

    func testBackwardFromTheConsoleIsTheInspector() {
        XCTAssertEqual(Pane.next(after: .console, forward: false), .inspector)
    }
}

final class InspectorPanelCycleTests: XCTestCase {
    func testNextPanelFollowsTheSelectedOne() {
        XCTAssertEqual(InspectorPanel.cycle(from: 0, forward: true), 1)
    }

    func testNextPanelAfterTheLastWrapsToTheFirst() {
        XCTAssertEqual(InspectorPanel.cycle(from: InspectorPanel.count - 1, forward: true), 0)
    }

    func testPreviousPanelBeforeTheFirstWrapsToTheLast() {
        XCTAssertEqual(InspectorPanel.cycle(from: 0, forward: false), InspectorPanel.count - 1)
    }

    func testOutOfRangeSelectionStartsFromTheFirstPanel() {
        XCTAssertEqual(InspectorPanel.cycle(from: 42, forward: true), 1)
    }

    func testThereAreSevenPanels() {
        XCTAssertEqual(InspectorPanel.count, 7)
    }
}
