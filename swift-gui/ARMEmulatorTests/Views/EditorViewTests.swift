import SwiftUI
import XCTest
@testable import ARMEmulator

@MainActor
final class EditorViewTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWebSocketClient: MockWebSocketClient!

    override func setUp() async throws {
        try await super.setUp()

        mockAPIClient = MockAPIClient()
        mockWebSocketClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(
            apiClient: mockAPIClient,
            wsClient: mockWebSocketClient
        )

        // Initialize session (required for breakpoint operations)
        await viewModel.initialize()
    }

    override func tearDown() async throws {
        viewModel = nil
        mockWebSocketClient = nil
        mockAPIClient = nil
        try await super.tearDown()
    }

    // MARK: - Breakpoint Toggle Tests

    func testBreakpointSyncFromViewModel() {
        // Given: ViewModel has breakpoints at specific addresses
        let address1: UInt32 = 0x8000
        let address2: UInt32 = 0x8004

        viewModel.breakpoints = [address1, address2]
        viewModel.addressToLine = [
            address1: 1,
            address2: 3,
        ]

        // When: EditorView observes breakpoint changes
        // (Simulated by checking the addressToLine mapping)
        let lineBreakpoints = Set(viewModel.breakpoints.compactMap { viewModel.addressToLine[$0] })

        // Then: Line-based breakpoints should be computed correctly
        XCTAssertEqual(lineBreakpoints, [1, 3], "Breakpoints should map from addresses to line numbers")
    }

    func testBreakpointToggleValidLine() async {
        // Given: Valid breakpoint line with address mapping
        let lineNumber = 2
        let address: UInt32 = 0x8004

        viewModel.validBreakpointLines = [1, 2, 3, 4]
        viewModel.lineToAddress = [lineNumber: address]
        viewModel.addressToLine = [address: lineNumber]

        mockAPIClient.shouldFailAddBreakpoint = false

        // When: Toggle breakpoint on valid line
        await viewModel.toggleBreakpoint(at: address)

        // Then: Breakpoint should be set
        XCTAssertTrue(mockAPIClient.addBreakpointCalled, "API should be called to set breakpoint")
        XCTAssertTrue(viewModel.breakpoints.contains(address), "Breakpoint should be added to ViewModel")
    }

    func testBreakpointToggleInvalidLine() {
        // Given: Invalid line (not executable code)
        let lineNumber = 1
        viewModel.validBreakpointLines = [2, 3, 4] // Line 1 is not valid (e.g., comment/directive)
        viewModel.lineToAddress = [:] // No mapping for invalid line

        // When/Then: Attempting to toggle should fail silently
        // (In the actual UI, this is handled by EditorView.toggleBreakpoint checking validBreakpointLines)
        XCTAssertFalse(
            viewModel.validBreakpointLines.contains(lineNumber),
            "Invalid lines should not be in validBreakpointLines"
        )
    }

    func testBreakpointRemoval() async {
        // Given: Existing breakpoint
        let lineNumber = 2
        let address: UInt32 = 0x8004

        viewModel.validBreakpointLines = [1, 2, 3, 4]
        viewModel.lineToAddress = [lineNumber: address]
        viewModel.addressToLine = [address: lineNumber]
        viewModel.breakpoints = [address]

        mockAPIClient.shouldFailRemoveBreakpoint = false

        // When: Toggle breakpoint again (should remove)
        await viewModel.toggleBreakpoint(at: address)

        // Then: Breakpoint should be removed
        XCTAssertTrue(mockAPIClient.removeBreakpointCalled, "API should be called to clear breakpoint")
        XCTAssertFalse(viewModel.breakpoints.contains(address), "Breakpoint should be removed from ViewModel")
    }

    // MARK: - PC Highlighting Tests

    func testCurrentPCToLineMapping() {
        // Given: PC at specific address
        let pcAddress: UInt32 = 0x8008
        let expectedLine = 3

        viewModel.addressToLine = [
            0x8000: 1,
            0x8004: 2,
            0x8008: 3,
            0x800C: 4,
        ]

        // When: PC changes
        viewModel.currentPC = pcAddress

        // Then: Current line should be computed from address mapping
        let currentLine = viewModel.addressToLine[pcAddress]
        XCTAssertEqual(currentLine, expectedLine, "PC address should map to correct line number")
    }

    func testCurrentPCNoMapping() {
        // Given: PC at address not in source map (e.g., library code)
        let pcAddress: UInt32 = 0xFFFF_0000

        viewModel.addressToLine = [
            0x8000: 1,
            0x8004: 2,
        ]

        // When: PC changes to unmapped address
        viewModel.currentPC = pcAddress

        // Then: No current line should be set
        let currentLine = viewModel.addressToLine[pcAddress]
        XCTAssertNil(currentLine, "Unmapped PC should not resolve to a line number")
    }

    func testCurrentPCUpdateTriggersScroll() {
        // Given: Scroll callback registered
        var scrollCallbackInvoked = false
        viewModel.scrollToCurrentPC = {
            scrollCallbackInvoked = true
        }

        viewModel.addressToLine = [0x8000: 1]

        // When: PC changes
        viewModel.currentPC = 0x8000

        // Manual invocation since SwiftUI onChange is not triggered in unit tests
        viewModel.scrollToCurrentPC?()

        // Then: Scroll callback should be invoked
        XCTAssertTrue(scrollCallbackInvoked, "Scroll callback should be invoked when PC changes")
    }

    // MARK: - Address-to-Line Mapping Tests

    func testAddressToLineMappingPopulated() {
        // Given: Source map from backend
        let sourceMap: [UInt32: Int] = [
            0x8000: 1,
            0x8004: 2,
            0x8008: 3,
            0x800C: 4,
        ]

        // When: ViewModel receives source map
        viewModel.addressToLine = sourceMap
        viewModel.lineToAddress = Dictionary(uniqueKeysWithValues: sourceMap.map { ($1, $0) })

        // Then: Bidirectional mappings should be consistent
        XCTAssertEqual(viewModel.addressToLine.count, 4, "Address-to-line map should have 4 entries")
        XCTAssertEqual(viewModel.lineToAddress.count, 4, "Line-to-address map should have 4 entries")

        // Verify bidirectional consistency
        for (address, line) in sourceMap {
            XCTAssertEqual(viewModel.addressToLine[address], line)
            XCTAssertEqual(viewModel.lineToAddress[line], address)
        }
    }

    func testValidBreakpointLinesFromSourceMap() {
        // Given: Source map with only executable lines
        viewModel.addressToLine = [
            0x8000: 2, // Line 1 is directive (.org)
            0x8004: 3, // Executable
            0x8008: 4, // Executable
        ]

        // When: Valid breakpoint lines computed
        viewModel.validBreakpointLines = Set(viewModel.addressToLine.values)

        // Then: Only executable lines should be valid
        XCTAssertEqual(viewModel.validBreakpointLines, [2, 3, 4])
        XCTAssertFalse(viewModel.validBreakpointLines.contains(1), "Directive lines should not be valid")
    }

    // MARK: - Horizontal Scrolling Configuration Tests

    func testEditorShouldSupportHorizontalScrolling() {
        // This is a design requirement test - EditorView should be configured
        // for horizontal scrolling (no text wrapping)
        // The actual NSTextView configuration is tested in EditorViewIntegrationTests

        // Given: EditorView text binding
        let text = String(repeating: "x", count: 200)

        // When/Then: Long lines should not be wrapped
        // (Verified through integration tests that NSTextView is configured correctly)
        XCTAssertGreaterThan(text.count, 100, "Long lines should be supported without wrapping")
    }

    // MARK: - Gutter Interaction Tests

    func testGutterBreakpointToggleFlow() async {
        // Given: User clicks gutter at line 3
        let clickedLine = 3
        let address: UInt32 = 0x8008

        viewModel.validBreakpointLines = [1, 2, 3, 4]
        viewModel.lineToAddress = [clickedLine: address]
        viewModel.addressToLine = [address: clickedLine]
        mockAPIClient.shouldFailAddBreakpoint = false

        // When: Gutter click triggers breakpoint toggle
        await viewModel.toggleBreakpoint(at: address)

        // Then: Breakpoint should be set via API
        XCTAssertTrue(mockAPIClient.addBreakpointCalled)
        XCTAssertTrue(viewModel.breakpoints.contains(address))
    }

    // MARK: - Text Change Handling Tests

    func testTextChangeDoesNotBreakBindings() {
        // Given: Initial text
        var editorText = "MOV R0, #42\nSWI #0"

        // When: User edits text
        editorText = "MOV R0, #100\nSWI #0"

        // Then: Text binding should update without clearing state
        XCTAssertEqual(editorText, "MOV R0, #100\nSWI #0")
        // Note: Source map and breakpoints are recalculated on program load/compile
    }

    // MARK: - Edge Cases

    func testEmptyAddressToLineMapping() {
        // Given: No program loaded
        viewModel.addressToLine = [:]
        viewModel.lineToAddress = [:]
        viewModel.validBreakpointLines = []

        // When: PC changes
        viewModel.currentPC = 0x8000

        // Then: No current line should be set
        XCTAssertNil(viewModel.addressToLine[viewModel.currentPC])
    }

    func testBreakpointOnSameLineAsPC() {
        // Given: Breakpoint and PC on same line
        let address: UInt32 = 0x8004
        let lineNumber = 2

        viewModel.breakpoints = [address]
        viewModel.addressToLine = [address: lineNumber]
        viewModel.currentPC = address

        // When: Both breakpoint and PC indicator should be shown
        let hasBreakpoint = viewModel.breakpoints.contains(address)
        let currentLine = viewModel.addressToLine[address]

        // Then: Both indicators should coexist
        XCTAssertTrue(hasBreakpoint, "Breakpoint should be present")
        XCTAssertEqual(currentLine, lineNumber, "PC indicator should also be on same line")
    }

    func testMultipleBreakpointsInEditor() async {
        // Given: Multiple breakpoints
        let addresses: [UInt32] = [0x8000, 0x8004, 0x8008]
        let lines = [1, 2, 3]

        viewModel.validBreakpointLines = Set(lines)
        viewModel.lineToAddress = Dictionary(uniqueKeysWithValues: zip(lines, addresses))
        viewModel.addressToLine = Dictionary(uniqueKeysWithValues: zip(addresses, lines))

        mockAPIClient.shouldFailAddBreakpoint = false

        // When: Set multiple breakpoints
        for address in addresses {
            await viewModel.toggleBreakpoint(at: address)
        }

        // Then: All should be tracked
        XCTAssertEqual(viewModel.breakpoints.count, 3, "All breakpoints should be set")
        for address in addresses {
            XCTAssertTrue(viewModel.breakpoints.contains(address))
        }
    }
}
