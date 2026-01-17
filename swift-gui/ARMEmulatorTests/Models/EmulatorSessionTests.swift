import XCTest
@testable import ARMEmulator

final class EmulatorSessionTests: XCTestCase {
    // MARK: - EmulatorEvent Tests

    func testEmulatorEventDecoding() throws {
        let json = """
        {
            "type": "state",
            "sessionId": "test-session-123"
        }
        """

        let data = json.data(using: .utf8)!
        let event = try JSONDecoder().decode(EmulatorEvent.self, from: data)

        XCTAssertEqual(event.type, "state")
        XCTAssertEqual(event.sessionId, "test-session-123")
        XCTAssertNil(event.data)
    }

    func testEmulatorEventWithDataDecoding() throws {
        let json = """
        {
            "type": "state",
            "sessionId": "test-session-123",
            "data": {
                "status": "running",
                "pc": 32768
            }
        }
        """

        let data = json.data(using: .utf8)!
        let event = try JSONDecoder().decode(EmulatorEvent.self, from: data)

        XCTAssertEqual(event.type, "state")
        XCTAssertEqual(event.sessionId, "test-session-123")
        XCTAssertNotNil(event.data)

        if case let .state(stateUpdate) = event.data {
            XCTAssertEqual(stateUpdate.status, "running")
            XCTAssertEqual(stateUpdate.pc, 32768)
        } else {
            XCTFail("Expected state event data")
        }
    }

    // MARK: - EventData State Variant Tests

    func testEventDataStateVariantDecoding() throws {
        let json = """
        {
            "status": "paused",
            "pc": 32772
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .state(stateUpdate) = eventData {
            XCTAssertEqual(stateUpdate.status, "paused")
            XCTAssertEqual(stateUpdate.pc, 32772)
            XCTAssertNil(stateUpdate.registers)
            XCTAssertNil(stateUpdate.flags)
        } else {
            XCTFail("Expected state variant")
        }
    }

    func testEventDataStateVariantWithRegisters() throws {
        let json = """
        {
            "status": "running",
            "pc": 32768,
            "registers": {
                "r0": 42, "r1": 100, "r2": 0, "r3": 0,
                "r4": 0, "r5": 0, "r6": 0, "r7": 0,
                "r8": 0, "r9": 0, "r10": 0, "r11": 0,
                "r12": 0, "sp": 327680, "lr": 0, "pc": 32768,
                "cpsr": {"n": false, "z": true, "c": false, "v": false}
            },
            "flags": {"n": false, "z": true, "c": false, "v": false}
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .state(stateUpdate) = eventData {
            XCTAssertEqual(stateUpdate.status, "running")
            XCTAssertEqual(stateUpdate.pc, 32768)
            XCTAssertNotNil(stateUpdate.registers)
            XCTAssertEqual(stateUpdate.registers?.r0, 42)
            XCTAssertEqual(stateUpdate.registers?.r1, 100)
            XCTAssertNotNil(stateUpdate.flags)
            XCTAssertTrue(stateUpdate.flags?.z ?? false)
        } else {
            XCTFail("Expected state variant")
        }
    }

    // MARK: - EventData Output Variant Tests

    func testEventDataOutputVariantDecoding() throws {
        let json = """
        {
            "stream": "stdout",
            "content": "Hello, World!"
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .output(outputUpdate) = eventData {
            XCTAssertEqual(outputUpdate.stream, "stdout")
            XCTAssertEqual(outputUpdate.content, "Hello, World!")
        } else {
            XCTFail("Expected output variant")
        }
    }

    func testEventDataOutputVariantWithStderr() throws {
        let json = """
        {
            "stream": "stderr",
            "content": "Error: File not found"
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .output(outputUpdate) = eventData {
            XCTAssertEqual(outputUpdate.stream, "stderr")
            XCTAssertEqual(outputUpdate.content, "Error: File not found")
        } else {
            XCTFail("Expected output variant")
        }
    }

    // MARK: - EventData Event Variant Tests

    func testEventDataEventVariantDecoding() throws {
        let json = """
        {
            "event": "breakpoint_hit",
            "address": 32776,
            "symbol": "main",
            "message": "Breakpoint hit at main"
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .event(executionEvent) = eventData {
            XCTAssertEqual(executionEvent.event, "breakpoint_hit")
            XCTAssertEqual(executionEvent.address, 32776)
            XCTAssertEqual(executionEvent.symbol, "main")
            XCTAssertEqual(executionEvent.message, "Breakpoint hit at main")
        } else {
            XCTFail("Expected event variant")
        }
    }

    func testEventDataEventVariantMinimal() throws {
        let json = """
        {
            "event": "program_exit"
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .event(executionEvent) = eventData {
            XCTAssertEqual(executionEvent.event, "program_exit")
            XCTAssertNil(executionEvent.address)
            XCTAssertNil(executionEvent.symbol)
            XCTAssertNil(executionEvent.message)
        } else {
            XCTFail("Expected event variant")
        }
    }

    // MARK: - EventData Error Handling

    func testEventDataInvalidJSONThrowsError() {
        let json = """
        {
            "unknown_field": "value"
        }
        """

        let data = json.data(using: .utf8)!

        XCTAssertThrowsError(try JSONDecoder().decode(EventData.self, from: data)) { error in
            guard case DecodingError.dataCorrupted = error else {
                XCTFail("Expected dataCorrupted error, got \(error)")
                return
            }
        }
    }

    // MARK: - StateUpdate Partial Updates

    func testStateUpdatePartialPC() throws {
        let json = """
        {
            "status": "running",
            "pc": 32768
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .state(stateUpdate) = eventData {
            XCTAssertEqual(stateUpdate.status, "running")
            XCTAssertEqual(stateUpdate.pc, 32768)
            XCTAssertNil(stateUpdate.registers)
            XCTAssertNil(stateUpdate.flags)
        } else {
            XCTFail("Expected state variant")
        }
    }

    func testStateUpdatePartialFlags() throws {
        let json = """
        {
            "status": "paused",
            "flags": {"n": true, "z": false, "c": true, "v": false}
        }
        """

        let data = json.data(using: .utf8)!
        let eventData = try JSONDecoder().decode(EventData.self, from: data)

        if case let .state(stateUpdate) = eventData {
            XCTAssertEqual(stateUpdate.status, "paused")
            XCTAssertNil(stateUpdate.pc)
            XCTAssertNil(stateUpdate.registers)
            XCTAssertNotNil(stateUpdate.flags)
            XCTAssertTrue(stateUpdate.flags?.n ?? false)
            XCTAssertTrue(stateUpdate.flags?.c ?? false)
        } else {
            XCTFail("Expected state variant")
        }
    }

    // MARK: - SubscriptionMessage Tests

    func testSubscriptionMessageEncoding() throws {
        let subscription = SubscriptionMessage(
            type: "subscribe",
            sessionId: "test-session-456",
            events: ["state", "output", "event"]
        )

        let data = try JSONEncoder().encode(subscription)
        let json = String(data: data, encoding: .utf8)!

        XCTAssertTrue(json.contains("\"type\":\"subscribe\""))
        XCTAssertTrue(json.contains("\"sessionId\":\"test-session-456\""))
        XCTAssertTrue(json.contains("\"events\""))
    }

    func testSubscriptionMessageDecoding() throws {
        let json = """
        {
            "type": "subscribe",
            "sessionId": "test-session-789",
            "events": ["state", "output"]
        }
        """

        let data = json.data(using: .utf8)!
        let subscription = try JSONDecoder().decode(SubscriptionMessage.self, from: data)

        XCTAssertEqual(subscription.type, "subscribe")
        XCTAssertEqual(subscription.sessionId, "test-session-789")
        XCTAssertEqual(subscription.events.count, 2)
        XCTAssertTrue(subscription.events.contains("state"))
        XCTAssertTrue(subscription.events.contains("output"))
    }

    // MARK: - EventData Encoding Tests

    func testEventDataStateEncoding() throws {
        let stateUpdate = StateUpdate(status: "running", pc: 32768, registers: nil, flags: nil)
        let eventData = EventData.state(stateUpdate)

        let data = try JSONEncoder().encode(eventData)
        let decoded = try JSONDecoder().decode(EventData.self, from: data)

        if case let .state(decodedUpdate) = decoded {
            XCTAssertEqual(decodedUpdate.status, "running")
            XCTAssertEqual(decodedUpdate.pc, 32768)
        } else {
            XCTFail("Expected state variant after encoding/decoding")
        }
    }

    func testEventDataOutputEncoding() throws {
        let outputUpdate = OutputUpdate(stream: "stdout", content: "Test output")
        let eventData = EventData.output(outputUpdate)

        let data = try JSONEncoder().encode(eventData)
        let decoded = try JSONDecoder().decode(EventData.self, from: data)

        if case let .output(decodedUpdate) = decoded {
            XCTAssertEqual(decodedUpdate.stream, "stdout")
            XCTAssertEqual(decodedUpdate.content, "Test output")
        } else {
            XCTFail("Expected output variant after encoding/decoding")
        }
    }

    func testEventDataEventEncoding() throws {
        let executionEvent = ExecutionEvent(event: "step_complete", address: 32780, symbol: "loop", message: nil)
        let eventData = EventData.event(executionEvent)

        let data = try JSONEncoder().encode(eventData)
        let decoded = try JSONDecoder().decode(EventData.self, from: data)

        if case let .event(decodedEvent) = decoded {
            XCTAssertEqual(decodedEvent.event, "step_complete")
            XCTAssertEqual(decodedEvent.address, 32780)
            XCTAssertEqual(decodedEvent.symbol, "loop")
        } else {
            XCTFail("Expected event variant after encoding/decoding")
        }
    }
}
