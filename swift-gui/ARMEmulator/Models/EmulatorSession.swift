import Foundation

struct EmulatorEvent: Codable {
    var type: String
    var sessionId: String
    var data: EventData?
}

enum EventData: Codable {
    case state(StateUpdate)
    case output(OutputUpdate)
    case event(ExecutionEvent)

    enum CodingKeys: String, CodingKey {
        case type
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()

        if let stateUpdate = try? container.decode(StateUpdate.self) {
            self = .state(stateUpdate)
        } else if let outputUpdate = try? container.decode(OutputUpdate.self) {
            self = .output(outputUpdate)
        } else if let executionEvent = try? container.decode(ExecutionEvent.self) {
            self = .event(executionEvent)
        } else {
            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Unable to decode EventData"
            )
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case let .state(update):
            try container.encode(update)
        case let .output(update):
            try container.encode(update)
        case let .event(event):
            try container.encode(event)
        }
    }
}

struct StateUpdate: Codable {
    var status: String
    var pc: UInt32?
    var registers: RegisterState?
    var flags: CPSRFlags?
}

struct OutputUpdate: Codable {
    var stream: String
    var content: String
}

struct ExecutionEvent: Codable {
    var event: String
    var address: UInt32?
    var symbol: String?
    var message: String?
}

struct SubscriptionMessage: Codable {
    var type: String
    var sessionId: String
    var events: [String]
}
