import Combine
import Foundation

class WebSocketClient: NSObject, ObservableObject, URLSessionWebSocketDelegate {
    private var webSocket: URLSessionWebSocketTask?
    private let eventSubject = PassthroughSubject<EmulatorEvent, Never>()
    private var session: URLSession?

    var events: AnyPublisher<EmulatorEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }

    func connect(sessionID: String) {
        guard let url = URL(string: "ws://localhost:8080/api/v1/ws") else {
            return
        }

        session = URLSession(configuration: .default, delegate: self, delegateQueue: OperationQueue())
        webSocket = session?.webSocketTask(with: url)
        webSocket?.resume()

        let subscription = SubscriptionMessage(
            type: "subscribe",
            sessionId: sessionID,
            events: ["state", "output", "event"]
        )
        send(subscription)

        receiveMessage()
    }

    func disconnect() {
        webSocket?.cancel(with: .goingAway, reason: nil)
        webSocket = nil
        session?.finishTasksAndInvalidate()
        session = nil
    }

    private func receiveMessage() {
        webSocket?.receive { [weak self] result in
            guard let self = self else { return }

            switch result {
            case let .success(message):
                if case let .string(text) = message,
                   let data = text.data(using: .utf8)
                {
                    self.handleMessage(data: data)
                }
                self.receiveMessage()
            case let .failure(error):
                print("WebSocket receive error: \(error)")
            }
        }
    }

    private func handleMessage(data: Data) {
        do {
            let event = try JSONDecoder().decode(EmulatorEvent.self, from: data)
            eventSubject.send(event)
        } catch {
            print("Failed to decode WebSocket message: \(error)")
        }
    }

    private func send<T: Encodable>(_ message: T) {
        guard let data = try? JSONEncoder().encode(message),
              let string = String(data: data, encoding: .utf8)
        else {
            return
        }

        webSocket?.send(.string(string)) { error in
            if let error = error {
                print("WebSocket send error: \(error)")
            }
        }
    }

    func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didOpenWithProtocol protocol: String?
    ) {
        print("WebSocket connected")
    }

    func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
        reason: Data?
    ) {
        print("WebSocket disconnected: \(closeCode)")
    }
}
