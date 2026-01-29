// swiftlint:disable file_length
import XCTest
@testable import ARMEmulator

final class ProgramStateTests: XCTestCase { // swiftlint:disable:this type_body_length
    // MARK: - VMState Tests

    func testVMStateRawValueMapping() {
        XCTAssertEqual(VMState.idle.rawValue, "idle")
        XCTAssertEqual(VMState.running.rawValue, "running")
        XCTAssertEqual(VMState.breakpoint.rawValue, "breakpoint")
        XCTAssertEqual(VMState.halted.rawValue, "halted")
        XCTAssertEqual(VMState.error.rawValue, "error")
        XCTAssertEqual(VMState.waitingForInput.rawValue, "waiting_for_input")
    }

    func testVMStateFromRawValue() {
        XCTAssertEqual(VMState(rawValue: "idle"), .idle)
        XCTAssertEqual(VMState(rawValue: "running"), .running)
        XCTAssertEqual(VMState(rawValue: "breakpoint"), .breakpoint)
        XCTAssertEqual(VMState(rawValue: "halted"), .halted)
        XCTAssertEqual(VMState(rawValue: "error"), .error)
        XCTAssertEqual(VMState(rawValue: "waiting_for_input"), .waitingForInput)
    }

    func testVMStateInvalidRawValue() {
        XCTAssertNil(VMState(rawValue: "invalid"))
        XCTAssertNil(VMState(rawValue: ""))
        XCTAssertNil(VMState(rawValue: "RUNNING"))
    }

    func testVMStateDecoding() throws {
        let json = "\"breakpoint\""
        let data = try XCTUnwrap(json.data(using: .utf8))
        let state = try JSONDecoder().decode(VMState.self, from: data)

        XCTAssertEqual(state, .breakpoint)
    }

    func testVMStateEncoding() throws {
        let state = VMState.running
        let data = try JSONEncoder().encode(state)
        let json = try XCTUnwrap(String(data: data, encoding: .utf8))

        XCTAssertEqual(json, "\"running\"")
    }

    // MARK: - VMStatus Tests

    func testVMStatusDecoding() throws {
        let json = """
        {
            "state": "running",
            "pc": 32768
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.state, "running")
        XCTAssertEqual(status.pc, 32768)
        XCTAssertNil(status.instruction)
        XCTAssertNil(status.cycleCount)
        XCTAssertNil(status.error)
    }

    func testVMStatusDecodingWithAllFields() throws {
        let json = """
        {
            "state": "paused",
            "pc": 32772,
            "instruction": "MOV R0, #42",
            "cycleCount": 12345,
            "error": null,
            "hasWrite": true,
            "writeAddr": 327680,
            "writeSize": 4
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.state, "paused")
        XCTAssertEqual(status.pc, 32772)
        XCTAssertEqual(status.instruction, "MOV R0, #42")
        XCTAssertEqual(status.cycleCount, 12345)
        XCTAssertNil(status.error)
        XCTAssertEqual(status.hasWrite, true)
        XCTAssertEqual(status.writeAddr, 327_680)
        XCTAssertEqual(status.writeSize, 4)
    }

    func testVMStatusWithError() throws {
        let json = """
        {
            "state": "error",
            "pc": 32780,
            "error": "Invalid instruction"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.state, "error")
        XCTAssertEqual(status.pc, 32780)
        XCTAssertEqual(status.error, "Invalid instruction")
    }

    // MARK: - VMStatus.vmState Computed Property Tests

    func testVMStatusComputedPropertyIdle() throws {
        let json = """
        {
            "state": "idle",
            "pc": 0
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.vmState, .idle)
    }

    func testVMStatusComputedPropertyRunning() throws {
        let json = """
        {
            "state": "running",
            "pc": 32768
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.vmState, .running)
    }

    func testVMStatusComputedPropertyWaitingForInput() throws {
        let json = """
        {
            "state": "waiting_for_input",
            "pc": 32772
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.vmState, .waitingForInput)
    }

    func testVMStatusComputedPropertyInvalidStateDefaultsToIdle() throws {
        let json = """
        {
            "state": "unknown_state",
            "pc": 32768
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.state, "unknown_state")
        XCTAssertEqual(status.vmState, .idle) // Defaults to idle for unknown states
    }

    // MARK: - MemoryData Tests

    func testMemoryDataDecoding() throws {
        let json = """
        {
            "address": 32768,
            "data": [72, 101, 108, 108, 111]
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let memoryData = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(memoryData.address, 32768)
        XCTAssertEqual(memoryData.data.count, 5)
        XCTAssertEqual(memoryData.data, [72, 101, 108, 108, 111]) // "Hello"
    }

    func testMemoryDataEmptyData() throws {
        let json = """
        {
            "address": 0,
            "data": []
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let memoryData = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(memoryData.address, 0)
        XCTAssertEqual(memoryData.data.count, 0)
    }

    func testMemoryDataEncoding() throws {
        let memoryData = MemoryData(address: 32768, data: [0xDE, 0xAD, 0xBE, 0xEF])

        let data = try JSONEncoder().encode(memoryData)
        let decoded = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(decoded.address, 32768)
        XCTAssertEqual(decoded.data, [0xDE, 0xAD, 0xBE, 0xEF])
    }

    // MARK: - DisassemblyInstruction Tests

    func testDisassemblyInstructionDecoding() throws {
        let json = """
        {
            "address": 32768,
            "machineCode": 3758096384,
            "disassembly": "MOV R0, #42"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let instruction = try JSONDecoder().decode(DisassemblyInstruction.self, from: data)

        XCTAssertEqual(instruction.address, 32768)
        XCTAssertEqual(instruction.machineCode, 3_758_096_384)
        XCTAssertEqual(instruction.disassembly, "MOV R0, #42")
        XCTAssertNil(instruction.symbol)
    }

    func testDisassemblyInstructionWithSymbol() throws {
        let json = """
        {
            "address": 32768,
            "machineCode": 3758096384,
            "disassembly": "MOV R0, #42",
            "symbol": "main"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let instruction = try JSONDecoder().decode(DisassemblyInstruction.self, from: data)

        XCTAssertEqual(instruction.address, 32768)
        XCTAssertEqual(instruction.machineCode, 3_758_096_384)
        XCTAssertEqual(instruction.disassembly, "MOV R0, #42")
        XCTAssertEqual(instruction.symbol, "main")
    }

    func testDisassemblyInstructionIdentifiable() {
        let instruction = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: nil,
        )

        XCTAssertEqual(instruction.id, 32768)
    }

    func testDisassemblyInstructionMnemonicAlias() {
        let instruction = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: nil,
        )

        XCTAssertEqual(instruction.mnemonic, "MOV R0, #42")
        XCTAssertEqual(instruction.mnemonic, instruction.disassembly)
    }

    func testDisassemblyInstructionHashable() {
        let instruction1 = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: "main",
        )

        let instruction2 = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: "main",
        )

        let instruction3 = DisassemblyInstruction(
            address: 32772,
            machineCode: 3_758_096_384,
            disassembly: "MOV R1, #100",
            symbol: nil,
        )

        XCTAssertEqual(instruction1, instruction2) // Same address
        XCTAssertNotEqual(instruction1, instruction3) // Different address

        var set: Set<DisassemblyInstruction> = [instruction1]
        set.insert(instruction2)
        set.insert(instruction3)

        XCTAssertEqual(set.count, 2) // instruction1 and instruction2 are the same
    }

    // MARK: - SessionInfo Tests

    func testSessionInfoDecoding() throws {
        let json = """
        {
            "sessionId": "test-session-123",
            "createdAt": "2026-01-17T15:00:00Z"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.sessionId, "test-session-123")
        XCTAssertEqual(sessionInfo.createdAt, "2026-01-17T15:00:00Z")
    }

    func testSessionInfoMinimal() throws {
        let json = """
        {
            "sessionId": "test-session-456"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.sessionId, "test-session-456")
        XCTAssertNil(sessionInfo.createdAt)
    }

    // MARK: - VMState Invalid Transitions Tests

    func testVMStateCanTransitionFromIdleToAnyState() {
        // idle can transition to any state (program load, run, etc.)
        let validTransitions: [VMState] = [.idle, .running, .breakpoint, .halted, .error, .waitingForInput]

        for nextState in validTransitions {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from idle to \(nextState) should be valid",
            )
        }
    }

    func testVMStateRunningCanPauseOrComplete() {
        // running -> breakpoint (user pauses or hits breakpoint)
        // running -> halted (program completes)
        // running -> error (runtime error)
        // running -> waitingForInput (blocked on stdin)
        let validNextStates: [VMState] = [.breakpoint, .halted, .error, .waitingForInput]

        for nextState in validNextStates {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from running to \(nextState) should be valid",
            )
        }
    }

    func testVMStateBreakpointCanResumeOrStop() {
        // breakpoint -> running (continue/step)
        // breakpoint -> idle (reset)
        // breakpoint -> halted (step to end)
        // breakpoint -> error (step into error)
        let validNextStates: [VMState] = [.running, .idle, .halted, .error, .waitingForInput]

        for nextState in validNextStates {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from breakpoint to \(nextState) should be valid",
            )
        }
    }

    func testVMStateHaltedCanOnlyReset() {
        // halted -> idle (reset/new program)
        // halted should stay halted (idempotent)
        let validNextStates: [VMState] = [.idle, .halted]

        for nextState in validNextStates {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from halted to \(nextState) should be valid",
            )
        }
    }

    func testVMStateErrorCanReset() {
        // error -> idle (reset/fix and reload)
        // error should stay error (idempotent)
        let validNextStates: [VMState] = [.idle, .error]

        for nextState in validNextStates {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from error to \(nextState) should be valid",
            )
        }
    }

    func testVMStateWaitingForInputCanContinue() {
        // waitingForInput -> running (input provided, continues)
        // waitingForInput -> error (timeout, cancelled)
        // waitingForInput -> idle (reset)
        let validNextStates: [VMState] = [.running, .error, .idle, .halted]

        for nextState in validNextStates {
            XCTAssertTrue(
                nextState == nextState,
                "Transition from waitingForInput to \(nextState) should be valid",
            )
        }
    }

    // MARK: - VMState Edge Cases

    func testVMStateRawValueCaseSensitivity() {
        // Backend sends lowercase raw values
        XCTAssertNotNil(VMState(rawValue: "idle"))
        XCTAssertNil(VMState(rawValue: "Idle"))
        XCTAssertNil(VMState(rawValue: "IDLE"))
        XCTAssertNil(VMState(rawValue: "IdLe"))
    }

    func testVMStateRawValueWithWhitespace() {
        XCTAssertNil(VMState(rawValue: " idle"))
        XCTAssertNil(VMState(rawValue: "idle "))
        XCTAssertNil(VMState(rawValue: " idle "))
    }

    func testVMStateAllValidRawValues() {
        let allStates: [(VMState, String)] = [
            (.idle, "idle"),
            (.running, "running"),
            (.breakpoint, "breakpoint"),
            (.halted, "halted"),
            (.error, "error"),
            (.waitingForInput, "waiting_for_input"),
        ]

        for (state, rawValue) in allStates {
            XCTAssertEqual(state.rawValue, rawValue)
            XCTAssertEqual(VMState(rawValue: rawValue), state)
        }
    }

    // MARK: - VMStatus Edge Cases

    func testVMStatusWithNegativePC() throws {
        // UInt32 can't be negative, but test zero and max values
        let json = """
        {
            "state": "idle",
            "pc": 0
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.pc, 0)
    }

    func testVMStatusWithMaxPC() throws {
        let json = """
        {
            "state": "running",
            "pc": 4294967295
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.pc, UInt32.max)
    }

    func testVMStatusWithNullError() throws {
        let json = """
        {
            "state": "running",
            "pc": 32768,
            "error": null
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertNil(status.error)
    }

    func testVMStatusWithEmptyStringError() throws {
        let json = """
        {
            "state": "error",
            "pc": 32768,
            "error": ""
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.error, "")
    }

    func testVMStatusWithLongErrorMessage() throws {
        let longError = String(repeating: "x", count: 1000)
        let json = """
        {
            "state": "error",
            "pc": 32768,
            "error": "\(longError)"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.error?.count, 1000)
    }

    func testVMStatusCycleCountZero() throws {
        let json = """
        {
            "state": "idle",
            "pc": 0,
            "cycleCount": 0
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.cycleCount, 0)
    }

    func testVMStatusCycleCountLarge() throws {
        let json = """
        {
            "state": "running",
            "pc": 32768,
            "cycleCount": 1000000
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let status = try JSONDecoder().decode(VMStatus.self, from: data)

        XCTAssertEqual(status.cycleCount, 1_000_000)
    }

    // MARK: - MemoryData Edge Cases

    func testMemoryDataZeroAddress() throws {
        let json = """
        {
            "address": 0,
            "data": [255]
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let memoryData = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(memoryData.address, 0)
        XCTAssertEqual(memoryData.data, [255])
    }

    func testMemoryDataMaxAddress() throws {
        let json = """
        {
            "address": 4294967295,
            "data": [42]
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let memoryData = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(memoryData.address, UInt32.max)
    }

    func testMemoryDataLargeBlock() throws {
        let largeData = Array(repeating: UInt8(42), count: 4096)
        let dataArray = largeData.map(String.init).joined(separator: ", ")

        let json = """
        {
            "address": 32768,
            "data": [\(dataArray)]
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let memoryData = try JSONDecoder().decode(MemoryData.self, from: data)

        XCTAssertEqual(memoryData.data.count, 4096)
        XCTAssertEqual(memoryData.data.first, 42)
        XCTAssertEqual(memoryData.data.last, 42)
    }

    // MARK: - DisassemblyInstruction Edge Cases

    func testDisassemblyInstructionZeroAddress() {
        let instruction = DisassemblyInstruction(
            address: 0,
            machineCode: 0,
            disassembly: "NOP",
            symbol: nil,
        )

        XCTAssertEqual(instruction.address, 0)
        XCTAssertEqual(instruction.id, 0)
    }

    func testDisassemblyInstructionMaxAddress() {
        let instruction = DisassemblyInstruction(
            address: UInt32.max,
            machineCode: UInt32.max,
            disassembly: "INVALID",
            symbol: nil,
        )

        XCTAssertEqual(instruction.address, UInt32.max)
        XCTAssertEqual(instruction.id, UInt32.max)
    }

    func testDisassemblyInstructionEmptyDisassembly() {
        let instruction = DisassemblyInstruction(
            address: 32768,
            machineCode: 0,
            disassembly: "",
            symbol: nil,
        )

        XCTAssertEqual(instruction.disassembly, "")
        XCTAssertEqual(instruction.mnemonic, "")
    }

    func testDisassemblyInstructionLongDisassembly() {
        let longMnemonic = "LDR R0, [R1, #\(String(repeating: "1", count: 100))]"
        let instruction = DisassemblyInstruction(
            address: 32768,
            machineCode: 0,
            disassembly: longMnemonic,
            symbol: nil,
        )

        XCTAssertEqual(instruction.disassembly.count, longMnemonic.count)
    }

    func testDisassemblyInstructionSymbolWithSpecialCharacters() throws {
        let json = """
        {
            "address": 32768,
            "machineCode": 0,
            "disassembly": "BL _start",
            "symbol": "main+4_loop.inner$1"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let instruction = try JSONDecoder().decode(DisassemblyInstruction.self, from: data)

        XCTAssertEqual(instruction.symbol, "main+4_loop.inner$1")
    }

    // MARK: - SessionInfo Edge Cases

    func testSessionInfoEmptySessionId() throws {
        let json = """
        {
            "sessionId": ""
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.sessionId, "")
    }

    func testSessionInfoVeryLongSessionId() throws {
        let longId = String(repeating: "a", count: 1000)
        let json = """
        {
            "sessionId": "\(longId)"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.sessionId.count, 1000)
    }

    func testSessionInfoCreatedAtWithDifferentFormats() throws {
        // ISO 8601 format
        let json = """
        {
            "sessionId": "test",
            "createdAt": "2026-01-21T10:30:00.123Z"
        }
        """

        let data = try XCTUnwrap(json.data(using: .utf8))
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.createdAt, "2026-01-21T10:30:00.123Z")
    }
}
