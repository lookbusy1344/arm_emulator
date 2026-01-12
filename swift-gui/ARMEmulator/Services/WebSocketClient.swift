import Combine
import Foundation

class WebSocketClient: NSObject, ObservableObject, URLSessionWebSocketDelegate, @unchecked Sendable {
    private var webSocket: URLSessionWebSocketTask?
    private let eventSubject = PassthroughSubject<EmulatorEvent, Never>()
    private var session: URLSession?
    
    // Reconnection state
    private var currentSessionID: String?
    private var isConnected = false
    private var isReconnecting = false
    private var retryCount = 0
    private let maxRetries = 5
    
    var events: AnyPublisher<EmulatorEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }

    func connect(sessionID: String) {
        self.currentSessionID = sessionID
        self.retryCount = 0
        connectInternal()
    }
    
    private func connectInternal() {
        guard let url = URL(string: "ws://localhost:8080/api/v1/ws") else {
            return
        }

        // Clean up existing connection if any
        webSocket?.cancel(with: .goingAway, reason: nil)
        
        session = URLSession(configuration: .default, delegate: self, delegateQueue: OperationQueue())
        webSocket = session?.webSocketTask(with: url)
        webSocket?.resume()

        if let sessionID = currentSessionID {
            let subscription = SubscriptionMessage(
                type: "subscribe",
                sessionId: sessionID,
                events: ["state", "output", "event"]
            )
            send(subscription)
        }

        receiveMessage()
    }

    func disconnect() {
        currentSessionID = nil
        webSocket?.cancel(with: .goingAway, reason: nil)
        webSocket = nil
        session?.finishTasksAndInvalidate()
        session = nil
        isConnected = false
    }

    private func receiveMessage() {
        webSocket?.receive { [weak self] result in
            guard let self = self else { return }

            switch result {
            case let .success(message):
                self.isConnected = true
                self.retryCount = 0
                self.isReconnecting = false
                
                if case let .string(text) = message,
                   let data = text.data(using: .utf8)
                {
                    self.handleMessage(data: data)
                }
                self.receiveMessage()
                
            case let .failure(error):
                print("WebSocket receive error: \(error)")
                self.handleDisconnection()
            }
        }
    }
    
    private func handleDisconnection() {
        guard currentSessionID != nil else { return } // Don't reconnect if explicitly disconnected
        
        // If the error was intentional (e.g. cancellation), don't reconnect
        if webSocket?.state == .completed || webSocket?.state == .canceling {
            return
        }

        scheduleReconnect()
    }
    
    private func scheduleReconnect() {
        guard !isReconnecting, retryCount < maxRetries else {
            if retryCount >= maxRetries {
                print("WebSocket max retries reached. Giving up.")
            }
            return
        }
        
        isReconnecting = true
        retryCount += 1
        
        let delay = pow(2.0, Double(retryCount - 1)) // Exponential backoff: 1, 2, 4, 8, 16
        print("WebSocket disconnected. Retrying in \(delay) seconds (Attempt \(retryCount)/\(maxRetries))...")
        
        DispatchQueue.global().asyncAfter(deadline: .now() + delay) { [weak self] in
            guard let self = self, self.currentSessionID != nil else { return }
            self.connectInternal()
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
        isConnected = true
        isReconnecting = false
        retryCount = 0
    }

    func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
        reason: Data?
    ) {
        print("WebSocket disconnected: \(closeCode)")
        isConnected = false
        handleDisconnection()
    }
}
