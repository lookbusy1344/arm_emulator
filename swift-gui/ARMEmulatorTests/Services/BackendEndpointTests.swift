import XCTest
@testable import ARMEmulator

final class BackendEndpointTests: XCTestCase {
    func testDefaultIsLocalhost8080() {
        let endpoint = BackendEndpoint.default

        XCTAssertEqual(endpoint.baseURL.absoluteString, "http://localhost:8080")
        XCTAssertEqual(endpoint.port, 8080)
        XCTAssertEqual(endpoint.webSocketURL.absoluteString, "ws://localhost:8080/api/v1/ws")
    }

    func testResolvesHttpURLWithPort() {
        let endpoint = BackendEndpoint.resolve("http://127.0.0.1:9090")

        XCTAssertEqual(endpoint.baseURL.absoluteString, "http://127.0.0.1:9090")
        XCTAssertEqual(endpoint.port, 9090)
        XCTAssertEqual(endpoint.webSocketURL.absoluteString, "ws://127.0.0.1:9090/api/v1/ws")
    }

    func testHttpsUsesSecureWebSocket() {
        let endpoint = BackendEndpoint.resolve("https://emu.example:8443")

        XCTAssertEqual(endpoint.webSocketURL.absoluteString, "wss://emu.example:8443/api/v1/ws")
    }

    func testPortDefaultsFromScheme() {
        XCTAssertEqual(BackendEndpoint.resolve("http://emu.example").port, 80)
        XCTAssertEqual(BackendEndpoint.resolve("https://emu.example").port, 443)
    }

    func testTrimsWhitespaceAndTrailingSlash() {
        let endpoint = BackendEndpoint.resolve("  http://localhost:9090/  ")

        XCTAssertEqual(endpoint.baseURL.absoluteString, "http://localhost:9090")
        XCTAssertEqual(endpoint.webSocketURL.absoluteString, "ws://localhost:9090/api/v1/ws")
    }

    func testInvalidSettingsFallBackToDefault() {
        for setting in [
            "",
            "   ",
            "not a url",
            "localhost:9090",
            "ftp://localhost:9090",
            "http://",
            "ws://localhost:9090",
        ] {
            XCTAssertEqual(BackendEndpoint.resolve(setting), .default, "'\(setting)' should fall back to the default")
        }
    }
}

final class BackendEndpointClientTests: XCTestCase {
    func testWebSocketClientConnectsToTheEndpoint() {
        let client = WebSocketClient(endpoint: .resolve("http://localhost:9090"), token: { nil })

        XCTAssertEqual(client.makeRequest()?.url?.absoluteString, "ws://localhost:9090/api/v1/ws")
    }

    @MainActor
    func testBackendManagerLaunchesOnTheEndpointPort() {
        XCTAssertEqual(
            BackendManager.serverArguments(for: .resolve("http://localhost:9090")),
            ["--api-server", "--port", "9090"],
        )
    }
}
