import XCTest
@testable import ARMEmulator

final class BackendTokenTests: XCTestCase {
    private var home: URL!

    override func setUpWithError() throws {
        home = FileManager.default.temporaryDirectory.appending(path: UUID().uuidString)
        try FileManager.default.createDirectory(at: home, withIntermediateDirectories: true)
    }

    override func tearDownWithError() throws {
        try FileManager.default.removeItem(at: home)
    }

    private func writeToken(_ text: String, port: Int) throws {
        let url = BackendToken.fileURL(port: port, home: home)
        try FileManager.default.createDirectory(at: url.deletingLastPathComponent(), withIntermediateDirectories: true)
        try Data(text.utf8).write(to: url)
    }

    func testFileURLIsPerPortInConfigDirectory() {
        XCTAssertEqual(
            BackendToken.fileURL(port: 8080, home: home).path,
            home.appending(path: ".config/arm-emu/api-token-8080").path,
        )
    }

    func testReadReturnsTokenWithoutTrailingNewline() throws {
        try writeToken("abc123\n", port: 8080)
        XCTAssertEqual(BackendToken.read(port: 8080, home: home), "abc123")
    }

    func testReadIgnoresOtherPorts() throws {
        try writeToken("abc123", port: 9090)
        XCTAssertNil(BackendToken.read(port: 8080, home: home))
    }

    func testReadReturnsNilForMissingFile() {
        XCTAssertNil(BackendToken.read(port: 8080, home: home))
    }

    func testReadReturnsNilForBlankFile() throws {
        try writeToken(" \n", port: 8080)
        XCTAssertNil(BackendToken.read(port: 8080, home: home))
    }

    func testAuthorizeSetsBearerHeader() throws {
        var request = try URLRequest(url: XCTUnwrap(URL(string: "http://localhost:8080/api/v1/session")))
        BackendToken.authorize(&request, token: "abc123")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Authorization"), "Bearer abc123")
    }

    func testAuthorizeWithoutTokenLeavesRequestAlone() throws {
        var request = try URLRequest(url: XCTUnwrap(URL(string: "http://localhost:8080/api/v1/session")))
        BackendToken.authorize(&request, token: nil)
        XCTAssertNil(request.value(forHTTPHeaderField: "Authorization"))
    }
}

final class APIClientTokenTests: XCTestCase {
    private func client(token: String?) -> APIClient {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        return APIClient(
            baseURL: URL(string: "http://localhost:8080")!,
            session: URLSession(configuration: configuration),
            token: { token },
        )
    }

    override func tearDown() {
        MockURLProtocol.requestHandler = nil
        super.tearDown()
    }

    private func authorizationHeader(sentBy client: APIClient) async throws -> String? {
        nonisolated(unsafe) var header: String?
        MockURLProtocol.requestHandler = { request in
            header = request.value(forHTTPHeaderField: "Authorization")
            let response = HTTPURLResponse(url: request.url!, statusCode: 200, httpVersion: nil, headerFields: nil)!
            return (response, Data(#"{"sessionId":"s1"}"#.utf8))
        }
        _ = try await client.createSession()
        return header
    }

    func testRequestsCarryToken() async throws {
        let header = try await authorizationHeader(sentBy: client(token: "abc123"))
        XCTAssertEqual(header, "Bearer abc123")
    }

    func testRequestsWithoutTokenHaveNoAuthorization() async throws {
        let header = try await authorizationHeader(sentBy: client(token: nil))
        XCTAssertNil(header)
    }
}

final class WebSocketClientTokenTests: XCTestCase {
    func testUpgradeRequestCarriesToken() {
        let request = WebSocketClient(token: { "abc123" }).makeRequest()
        XCTAssertEqual(request?.url?.absoluteString, "ws://localhost:8080/api/v1/ws")
        XCTAssertEqual(request?.value(forHTTPHeaderField: "Authorization"), "Bearer abc123")
    }

    func testUpgradeRequestWithoutToken() {
        let request = WebSocketClient(token: { nil }).makeRequest()
        XCTAssertNil(request?.value(forHTTPHeaderField: "Authorization"))
    }
}
