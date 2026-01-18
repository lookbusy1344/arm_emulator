import XCTest
@testable import ARMEmulator

final class ProgramStateTests: XCTestCase {
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
        let data = json.data(using: .utf8)!
        let state = try JSONDecoder().decode(VMState.self, from: data)

        XCTAssertEqual(state, .breakpoint)
    }

    func testVMStateEncoding() throws {
        let state = VMState.running
        let data = try JSONEncoder().encode(state)
        let json = String(data: data, encoding: .utf8)!

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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
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
            symbol: nil
        )

        XCTAssertEqual(instruction.id, 32768)
    }

    func testDisassemblyInstructionMnemonicAlias() {
        let instruction = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: nil
        )

        XCTAssertEqual(instruction.mnemonic, "MOV R0, #42")
        XCTAssertEqual(instruction.mnemonic, instruction.disassembly)
    }

    func testDisassemblyInstructionHashable() {
        let instruction1 = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: "main"
        )

        let instruction2 = DisassemblyInstruction(
            address: 32768,
            machineCode: 3_758_096_384,
            disassembly: "MOV R0, #42",
            symbol: "main"
        )

        let instruction3 = DisassemblyInstruction(
            address: 32772,
            machineCode: 3_758_096_384,
            disassembly: "MOV R1, #100",
            symbol: nil
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

        let data = json.data(using: .utf8)!
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

        let data = json.data(using: .utf8)!
        let sessionInfo = try JSONDecoder().decode(SessionInfo.self, from: data)

        XCTAssertEqual(sessionInfo.sessionId, "test-session-456")
        XCTAssertNil(sessionInfo.createdAt)
    }
}
