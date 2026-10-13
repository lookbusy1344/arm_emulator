import XCTest
@testable import ARMEmulator

final class SourceTextLineTests: XCTestCase {
    private let text = "_start:\n  MOV R0, #1\n  SWI #0\n"

    func testOffsetZeroIsLineOne() {
        XCTAssertEqual(SourceText.line(atOffset: 0, in: text), 1)
    }

    func testOffsetOnTheNewlineBelongsToThatLine() {
        XCTAssertEqual(SourceText.line(atOffset: 7, in: text), 1)
    }

    func testOffsetAfterTheNewlineStartsTheNextLine() {
        XCTAssertEqual(SourceText.line(atOffset: 8, in: text), 2)
    }

    func testOffsetAtTheEndAfterATrailingNewlineIsTheEmptyLastLine() {
        XCTAssertEqual(SourceText.line(atOffset: text.utf16.count, in: text), 4)
    }

    func testEmptyTextIsLineOne() {
        XCTAssertEqual(SourceText.line(atOffset: 0, in: ""), 1)
    }

    func testOffsetPastTheEndClampsToTheLastLine() {
        XCTAssertEqual(SourceText.line(atOffset: 1000, in: "a\nb"), 2)
    }

    func testOffsetsCountUTF16UnitsAsNSTextViewDoes() {
        // "é" is one UTF-16 unit, "😀" is two
        let text = "😀é\nX"

        XCTAssertEqual(SourceText.line(atOffset: 3, in: text), 1)
        XCTAssertEqual(SourceText.line(atOffset: 4, in: text), 2)
    }
}

@MainActor
final class CaretBreakpointTests: XCTestCase {
    private var viewModel: EmulatorViewModel!
    private var mockAPIClient: MockAPIClient!

    private let program = "_start:\n  MOV R0, #1\n  MOV R1, #2\n  SWI #0\n"

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockAPIClient.mockSourceMap = [
            SourceMapEntry(address: 0x8000, lineNumber: 2, line: ""),
            SourceMapEntry(address: 0x8004, lineNumber: 3, line: ""),
            SourceMapEntry(address: 0x8008, lineNumber: 4, line: ""),
        ]
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: MockWebSocketClient())
        viewModel.sessionID = "test-session"
    }

    override func tearDown() async throws {
        viewModel = nil
        mockAPIClient = nil
    }

    func testCaretStartsOnLineOne() {
        XCTAssertEqual(viewModel.caretLine, 1)
    }

    func testTogglesTheBreakpointOnTheCaretLine() async {
        await viewModel.openProgram(source: program)
        viewModel.caretLine = 3

        await viewModel.toggleBreakpointAtCaret()

        XCTAssertEqual(mockAPIClient.addedBreakpoints, [0x8004])
        XCTAssertEqual(viewModel.breakpoints, [0x8004])
    }

    func testSecondToggleRemovesTheBreakpoint() async {
        await viewModel.openProgram(source: program)
        viewModel.caretLine = 4
        await viewModel.toggleBreakpointAtCaret()

        await viewModel.toggleBreakpointAtCaret()

        XCTAssertEqual(mockAPIClient.removedBreakpoints, [0x8008])
        XCTAssertEqual(viewModel.breakpoints, [])
    }

    func testAssemblesEditedSourceFirst() async {
        viewModel.sourceCode = program
        viewModel.caretLine = 2

        await viewModel.toggleBreakpointAtCaret()

        XCTAssertEqual(mockAPIClient.callSequence, ["loadProgram", "addBreakpoint"])
        XCTAssertEqual(viewModel.breakpoints, [0x8000])
    }

    func testCaretOnALineWithoutAnInstructionReportsTheLine() async {
        await viewModel.openProgram(source: program)
        viewModel.caretLine = 1

        await viewModel.toggleBreakpointAtCaret()

        XCTAssertEqual(viewModel.errorMessage, "Line 1 has no instruction for a breakpoint")
        XCTAssertEqual(mockAPIClient.addedBreakpoints, [])
    }

    func testFailedAssemblyTogglesNothing() async {
        mockAPIClient.shouldFailLoadProgram = true
        viewModel.sourceCode = program
        viewModel.caretLine = 2

        await viewModel.toggleBreakpointAtCaret()

        XCTAssertEqual(viewModel.errorMessage, "Failed to assemble program: Mock load program failed")
        XCTAssertEqual(mockAPIClient.addedBreakpoints, [])
    }
}
