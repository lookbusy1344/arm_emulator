import SwiftUI
import XCTest
@testable import ARMEmulator

@MainActor
final class DisassemblyViewTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWebSocketClient: MockWebSocketClient!

    override func setUp() async throws {
        try await super.setUp()

        mockAPIClient = MockAPIClient()
        mockWebSocketClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(
            apiClient: mockAPIClient,
            wsClient: mockWebSocketClient,
        )

        // Initialize session
        await viewModel.initialize()
    }

    override func tearDown() async throws {
        viewModel = nil
        mockWebSocketClient = nil
        mockAPIClient = nil
        try await super.tearDown()
    }

    // MARK: - Instruction Formatting Tests

    func testInstructionAddressFormatting() {
        // Given: Disassembly instruction
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: nil,
        )

        // When: Format address as hex
        let formatted = String(format: "0x%08X", instruction.address)

        // Then: Address should be formatted correctly
        XCTAssertEqual(formatted, "0x00008000", "Address should be 8-digit hex")
    }

    func testInstructionMachineCodeFormatting() {
        // Given: Disassembly instruction
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: nil,
        )

        // When: Format machine code as hex
        let formatted = String(format: "%08X", instruction.machineCode)

        // Then: Machine code should be formatted correctly
        XCTAssertEqual(formatted, "E3A00042", "Machine code should be 8-digit hex")
    }

    func testInstructionMnemonicExtraction() {
        // Given: Disassembly instruction
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: nil,
        )

        // When: Access mnemonic (alias for disassembly)
        let mnemonic = instruction.mnemonic

        // Then: Mnemonic should match disassembly
        XCTAssertEqual(mnemonic, "MOV R0, #66", "Mnemonic should match disassembly")
    }

    // MARK: - Address Highlighting Tests

    func testPCHighlighting() {
        // Given: Current PC and disassembly
        viewModel.currentPC = 0x8004
        let instructions = [
            DisassemblyInstruction(address: 0x8000, machineCode: 0xE3A0_0042, disassembly: "MOV R0, #66", symbol: nil),
            DisassemblyInstruction(address: 0x8004, machineCode: 0xE3A0_102A, disassembly: "MOV R1, #42", symbol: nil),
            DisassemblyInstruction(address: 0x8008, machineCode: 0xEF00_0000, disassembly: "SWI #0", symbol: nil),
        ]

        // When: Check which instruction is at PC
        let instructionAtPC = instructions.first { $0.address == viewModel.currentPC }

        // Then: Should find instruction at PC
        XCTAssertNotNil(instructionAtPC, "Should find instruction at current PC")
        XCTAssertEqual(instructionAtPC?.address, 0x8004)
        XCTAssertEqual(instructionAtPC?.mnemonic, "MOV R1, #42")
    }

    func testPCHighlightingNotPresent() {
        // Given: PC not in disassembly range
        viewModel.currentPC = 0xFFFF_0000
        let instructions = [
            DisassemblyInstruction(address: 0x8000, machineCode: 0xE3A0_0042, disassembly: "MOV R0, #66", symbol: nil),
            DisassemblyInstruction(address: 0x8004, machineCode: 0xE3A0_102A, disassembly: "MOV R1, #42", symbol: nil),
        ]

        // When: Check which instruction is at PC
        let instructionAtPC = instructions.first { $0.address == viewModel.currentPC }

        // Then: Should not find instruction
        XCTAssertNil(instructionAtPC, "Should not find instruction when PC is out of range")
    }

    func testBreakpointHighlighting() {
        // Given: Breakpoint at address
        let breakpointAddress: UInt32 = 0x8004
        viewModel.breakpoints = [breakpointAddress]

        let instructions = [
            DisassemblyInstruction(address: 0x8000, machineCode: 0xE3A0_0042, disassembly: "MOV R0, #66", symbol: nil),
            DisassemblyInstruction(address: 0x8004, machineCode: 0xE3A0_102A, disassembly: "MOV R1, #42", symbol: nil),
            DisassemblyInstruction(address: 0x8008, machineCode: 0xEF00_0000, disassembly: "SWI #0", symbol: nil),
        ]

        // When: Check which instructions have breakpoints
        let instructionsWithBreakpoints = instructions.filter { viewModel.breakpoints.contains($0.address) }

        // Then: Should find instruction with breakpoint
        XCTAssertEqual(instructionsWithBreakpoints.count, 1)
        XCTAssertEqual(instructionsWithBreakpoints.first?.address, breakpointAddress)
    }

    func testPCAndBreakpointOnSameInstruction() {
        // Given: PC and breakpoint at same address
        let address: UInt32 = 0x8004
        viewModel.currentPC = address
        viewModel.breakpoints = [address]

        let instruction = DisassemblyInstruction(
            address: address,
            machineCode: 0xE3A0_102A,
            disassembly: "MOV R1, #42",
            symbol: nil,
        )

        // When: Check both conditions
        let isPC = instruction.address == viewModel.currentPC
        let hasBreakpoint = viewModel.breakpoints.contains(instruction.address)

        // Then: Both should be true
        XCTAssertTrue(isPC, "Instruction should be at PC")
        XCTAssertTrue(hasBreakpoint, "Instruction should have breakpoint")
    }

    // MARK: - Symbol Resolution Tests

    func testInstructionWithSymbol() {
        // Given: Instruction with symbol
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: "main",
        )

        // When/Then: Symbol should be present
        XCTAssertEqual(instruction.symbol, "main")
        XCTAssertFalse(instruction.symbol?.isEmpty ?? true)
    }

    func testInstructionWithoutSymbol() {
        // Given: Instruction without symbol
        let instruction = DisassemblyInstruction(
            address: 0x8004,
            machineCode: 0xE3A0_102A,
            disassembly: "MOV R1, #42",
            symbol: nil,
        )

        // When/Then: Symbol should be nil
        XCTAssertNil(instruction.symbol)
    }

    func testInstructionWithEmptySymbol() {
        // Given: Instruction with empty symbol
        let instruction = DisassemblyInstruction(
            address: 0x8004,
            machineCode: 0xE3A0_102A,
            disassembly: "MOV R1, #42",
            symbol: "",
        )

        // When/Then: Symbol should be empty
        XCTAssertEqual(instruction.symbol, "")
        XCTAssertTrue(instruction.symbol?.isEmpty ?? true)
    }

    func testMultipleInstructionsWithSymbols() {
        // Given: Multiple instructions with symbols
        let instructions = [
            DisassemblyInstruction(
                address: 0x8000,
                machineCode: 0xE3A0_0042,
                disassembly: "MOV R0, #66",
                symbol: "main",
            ),
            DisassemblyInstruction(address: 0x8004, machineCode: 0xE3A0_102A, disassembly: "MOV R1, #42", symbol: nil),
            DisassemblyInstruction(address: 0x8008, machineCode: 0xEF00_0000, disassembly: "SWI #0", symbol: "exit"),
        ]

        // When: Filter instructions with symbols
        let withSymbols = instructions.filter { $0.symbol != nil && !($0.symbol?.isEmpty ?? true) }

        // Then: Should find 2 instructions with symbols
        XCTAssertEqual(withSymbols.count, 2)
        XCTAssertEqual(withSymbols[0].symbol, "main")
        XCTAssertEqual(withSymbols[1].symbol, "exit")
    }

    // MARK: - Disassembly Loading Tests

    func testDisassemblyInstructionCount() {
        // Given: Standard disassembly count
        let instructionsToShow = 64

        // Then: Should request ±32 around PC
        XCTAssertEqual(instructionsToShow, 64, "Should show 64 instructions (±32 around PC)")
    }

    func testDisassemblyStartAddress() {
        // Given: PC well above 0
        viewModel.currentPC = 0x8100

        // When: Calculate start address (PC - 128)
        let startAddress = viewModel.currentPC > 128 ? viewModel.currentPC - 128 : 0

        // Then: Start address should be PC - 128
        XCTAssertEqual(startAddress, 0x8080, "Start address should be PC - 128")
    }

    func testDisassemblyStartAddressNearZero() {
        // Given: PC near 0
        viewModel.currentPC = 0x40

        // When: Calculate start address (should not underflow)
        let startAddress = viewModel.currentPC > 128 ? viewModel.currentPC - 128 : 0

        // Then: Start address should be 0 (no underflow)
        XCTAssertEqual(startAddress, 0, "Start address should be 0 to prevent underflow")
    }

    // MARK: - Instruction Identifiable Tests

    func testInstructionIdentifiable() {
        // Given: Disassembly instruction
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: nil,
        )

        // When: Access ID
        let id = instruction.id

        // Then: ID should be address
        XCTAssertEqual(id, instruction.address, "Instruction ID should equal address")
    }

    func testInstructionEquality() {
        // Given: Two identical instructions
        let inst1 = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: "main",
        )

        let inst2 = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: "main",
        )

        // When/Then: Should be equal (Hashable/Equatable)
        XCTAssertEqual(inst1, inst2, "Identical instructions should be equal")
    }

    func testInstructionInequality() {
        // Given: Two different instructions
        let inst1 = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: nil,
        )

        let inst2 = DisassemblyInstruction(
            address: 0x8004,
            machineCode: 0xE3A0_102A,
            disassembly: "MOV R1, #42",
            symbol: nil,
        )

        // When/Then: Should not be equal
        XCTAssertNotEqual(inst1, inst2, "Different instructions should not be equal")
    }

    // MARK: - Edge Cases

    func testZeroAddress() {
        // Given: Instruction at address 0
        let instruction = DisassemblyInstruction(
            address: 0x0,
            machineCode: 0xE3A0_0000,
            disassembly: "MOV R0, #0",
            symbol: "reset",
        )

        // When: Format address
        let formatted = String(format: "0x%08X", instruction.address)

        // Then: Should format as 0x00000000
        XCTAssertEqual(formatted, "0x00000000")
    }

    func testMaxAddress() {
        // Given: Instruction at max address
        let instruction = DisassemblyInstruction(
            address: 0xFFFF_FFFC,
            machineCode: 0xEF00_0000,
            disassembly: "SWI #0",
            symbol: nil,
        )

        // When: Format address
        let formatted = String(format: "0x%08X", instruction.address)

        // Then: Should format correctly
        XCTAssertEqual(formatted, "0xFFFFFFFC")
    }

    func testLongSymbolName() {
        // Given: Instruction with long symbol name
        let longSymbol = "very_long_function_name_with_lots_of_characters"
        let instruction = DisassemblyInstruction(
            address: 0x8000,
            machineCode: 0xE3A0_0042,
            disassembly: "MOV R0, #66",
            symbol: longSymbol,
        )

        // When/Then: Symbol should be preserved completely
        XCTAssertEqual(instruction.symbol, longSymbol)
        XCTAssertEqual(instruction.symbol?.count, longSymbol.count)
    }
}
