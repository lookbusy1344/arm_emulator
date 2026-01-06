import Foundation

enum APIError: Error, LocalizedError {
    case invalidURL
    case networkError(Error)
    case invalidResponse
    case serverError(Int, String)
    case decodingError(Error)
    case encodingError(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case let .networkError(error):
            return "Network error: \(error.localizedDescription)"
        case .invalidResponse:
            return "Invalid server response"
        case let .serverError(code, message):
            return "Server error (\(code)): \(message)"
        case let .decodingError(error):
            return "Failed to decode response: \(error.localizedDescription)"
        case let .encodingError(error):
            return "Failed to encode request: \(error.localizedDescription)"
        }
    }
}

class APIClient: ObservableObject {
    private let baseURL: URL
    private let session: URLSession

    init(baseURL: URL = URL(string: "http://localhost:8080")!) {
        self.baseURL = baseURL
        session = URLSession.shared
    }

    // MARK: - Session Management

    func createSession() async throws -> String {
        struct CreateSessionResponse: Codable {
            let sessionId: String
        }

        let url = baseURL.appendingPathComponent("/api/v1/session")
        let response: CreateSessionResponse = try await post(url: url, body: EmptyBody())
        return response.sessionId
    }

    func destroySession(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)")
        try await delete(url: url)
    }

    func getStatus(sessionID: String) async throws -> VMStatus {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)")
        return try await get(url: url)
    }

    // MARK: - Program Management

    func loadProgram(sessionID: String, source: String) async throws -> LoadProgramResponse {
        struct LoadProgramRequest: Codable {
            let source: String
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/load")
        DebugLog.network("POST \(url.absoluteString) - \(source.count) chars")
        let response: LoadProgramResponse = try await post(url: url, body: LoadProgramRequest(source: source))
        DebugLog.success("Load response - success: \(response.success)", category: "Network")
        return response
    }

    // MARK: - Execution Control

    func run(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/run")
        DebugLog.network("POST \(url.absoluteString)")
        try await post(url: url, body: EmptyBody())
        DebugLog.success("POST \(url.absoluteString) - success", category: "Network")
    }

    func stop(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/stop")
        try await post(url: url, body: EmptyBody())
    }

    func step(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/step")
        try await post(url: url, body: EmptyBody())
    }

    func stepOver(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/step-over")
        try await post(url: url, body: EmptyBody())
    }

    func stepOut(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/step-out")
        try await post(url: url, body: EmptyBody())
    }

    func reset(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/reset")
        try await post(url: url, body: EmptyBody())
    }

    func sendStdin(sessionID: String, data: String) async throws {
        struct StdinRequest: Codable {
            let data: String
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/stdin")
        try await post(url: url, body: StdinRequest(data: data))
    }

    // MARK: - State Inspection

    func getRegisters(sessionID: String) async throws -> RegisterState {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/registers")
        return try await get(url: url)
    }

    func getMemory(sessionID: String, address: UInt32, length: UInt32) async throws -> MemoryData {
        var components = URLComponents(
            url: baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/memory"),
            resolvingAgainstBaseURL: false
        )!
        components.queryItems = [
            URLQueryItem(name: "address", value: String(address)),
            URLQueryItem(name: "length", value: String(length)),
        ]

        guard let url = components.url else {
            throw APIError.invalidURL
        }

        return try await get(url: url)
    }

    func getDisassembly(sessionID: String, address: UInt32, count: UInt32) async throws -> [DisassemblyInstruction] {
        struct DisassemblyResponse: Codable {
            let instructions: [DisassemblyInstruction]
        }

        var components = URLComponents(
            url: baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/disassembly"),
            resolvingAgainstBaseURL: false
        )!
        components.queryItems = [
            URLQueryItem(name: "address", value: String(address)),
            URLQueryItem(name: "count", value: String(count)),
        ]

        guard let url = components.url else {
            throw APIError.invalidURL
        }

        let response: DisassemblyResponse = try await get(url: url)
        return response.instructions
    }

    // MARK: - Debugging

    func addBreakpoint(sessionID: String, address: UInt32) async throws {
        struct BreakpointRequest: Codable {
            let address: UInt32
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/breakpoint")
        try await post(url: url, body: BreakpointRequest(address: address))
    }

    func removeBreakpoint(sessionID: String, address: UInt32) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/breakpoint/\(address)")
        try await delete(url: url)
    }

    func getBreakpoints(sessionID: String) async throws -> [UInt32] {
        struct BreakpointsResponse: Codable {
            let breakpoints: [UInt32]
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/breakpoints")
        let response: BreakpointsResponse = try await get(url: url)
        return response.breakpoints
    }

    func evaluateExpression(sessionID: String, expression: String) async throws -> UInt32 {
        struct EvaluateRequest: Codable {
            let expression: String
        }

        struct EvaluateResponse: Codable {
            let result: UInt32
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/evaluate")
        let response: EvaluateResponse = try await post(url: url, body: EvaluateRequest(expression: expression))
        return response.result
    }

    // MARK: - Watchpoints

    func addWatchpoint(sessionID: String, address: UInt32, type: String) async throws -> Watchpoint {
        struct WatchpointRequest: Codable {
            let address: UInt32
            let type: String
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/watchpoint")
        let watchpoint: Watchpoint = try await post(url: url, body: WatchpointRequest(address: address, type: type))
        return watchpoint
    }

    func removeWatchpoint(sessionID: String, watchpointID: Int) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/watchpoint/\(watchpointID)")
        try await delete(url: url)
    }

    func getWatchpoints(sessionID: String) async throws -> [Watchpoint] {
        struct WatchpointsResponse: Codable {
            let watchpoints: [Watchpoint]
        }

        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/watchpoints")
        let response: WatchpointsResponse = try await get(url: url)
        return response.watchpoints
    }

    // MARK: - Generic HTTP Methods

    private func get<T: Decodable>(url: URL) async throws -> T {
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        return try await performRequest(request: request)
    }

    private func post<T: Encodable, R: Decodable>(url: URL, body: T) async throws -> R {
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        do {
            request.httpBody = try JSONEncoder().encode(body)
        } catch {
            throw APIError.encodingError(error)
        }

        return try await performRequest(request: request)
    }

    private func post<T: Encodable>(url: URL, body: T) async throws {
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        do {
            request.httpBody = try JSONEncoder().encode(body)
        } catch {
            throw APIError.encodingError(error)
        }

        let _: EmptyResponse = try await performRequest(request: request)
    }

    private func delete(url: URL) async throws {
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"

        let _: EmptyResponse = try await performRequest(request: request)
    }

    // MARK: - Memory Operations

    func getMemory(sessionID: String, address: UInt32, length: Int) async throws -> [UInt8] {
        struct MemoryResponse: Codable {
            let address: UInt32
            let length: Int
            let data: String // Base64 encoded
        }

        var components = URLComponents(
            url: baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/memory"),
            resolvingAgainstBaseURL: false
        )!
        components.queryItems = [
            URLQueryItem(name: "address", value: String(format: "0x%X", address)),
            URLQueryItem(name: "length", value: String(length)),
        ]

        guard let url = components.url else {
            throw APIError.invalidURL
        }

        let response: MemoryResponse = try await get(url: url)

        // Decode base64 data
        guard let data = Data(base64Encoded: response.data) else {
            throw APIError.decodingError(NSError(
                domain: "APIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Failed to decode base64 memory data"]
            ))
        }

        return Array(data)
    }

    func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassembledInstruction] {
        struct DisassemblyResponse: Codable {
            let instructions: [DisassembledInstruction]
        }

        var components = URLComponents(
            url: baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/disassembly"),
            resolvingAgainstBaseURL: false
        )!
        components.queryItems = [
            URLQueryItem(name: "addr", value: String(format: "0x%X", address)),
            URLQueryItem(name: "count", value: String(count)),
        ]

        guard let url = components.url else {
            throw APIError.invalidURL
        }

        let response: DisassemblyResponse = try await get(url: url)
        return response.instructions
    }

    private func performRequest<T: Decodable>(request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard (200 ... 299).contains(httpResponse.statusCode) else {
            let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
            throw APIError.serverError(httpResponse.statusCode, errorMessage)
        }

        if T.self == EmptyResponse.self, let empty = EmptyResponse() as? T {
            return empty
        }

        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw APIError.decodingError(error)
        }
    }
}

// MARK: - Models

struct DisassembledInstruction: Codable, Identifiable, Hashable {
    let address: UInt32
    let machineCode: UInt32
    let mnemonic: String
    let symbol: String?

    var id: UInt32 { address }

    func hash(into hasher: inout Hasher) {
        hasher.combine(address)
    }

    static func == (lhs: DisassembledInstruction, rhs: DisassembledInstruction) -> Bool {
        lhs.address == rhs.address
    }
}

// MARK: - Helper Types

private struct EmptyBody: Codable {}
private struct EmptyResponse: Codable {}

struct LoadProgramResponse: Codable {
    let success: Bool
    let errors: [String]?
    let symbols: [String: UInt32]?
}
