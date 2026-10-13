import Foundation

/// Where the backend serves HTTP and WebSocket traffic, read from the Backend URL preference.
struct BackendEndpoint: Equatable {
    static let `default` = BackendEndpoint(baseURL: URL(string: "http://localhost:\(BackendToken.defaultPort)")!)

    private static let webSocketPath = "/api/v1/ws"
    private static let defaultPorts = ["http": 80, "https": 443]
    private static let webSocketSchemes = ["http": "ws", "https": "wss"]

    let baseURL: URL

    /// The endpoint the Backend URL preference names
    @MainActor
    static var fromSettings: BackendEndpoint {
        resolve(AppSettings.shared.backendURL)
    }

    /// The endpoint for an absolute http or https URL with a host, or the default for anything else
    static func resolve(_ setting: String) -> BackendEndpoint {
        var text = setting.trimmingCharacters(in: .whitespaces)
        while text.hasSuffix("/") {
            text.removeLast()
        }
        guard let url = URL(string: text),
              let scheme = url.scheme?.lowercased(),
              defaultPorts[scheme] != nil,
              url.host?.isEmpty == false
        else {
            return .default
        }
        return BackendEndpoint(baseURL: url)
    }

    var port: Int {
        baseURL.port ?? Self.defaultPorts[baseURL.scheme?.lowercased() ?? ""] ?? BackendToken.defaultPort
    }

    var webSocketURL: URL {
        var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false)!
        components.scheme = Self.webSocketSchemes[components.scheme?.lowercased() ?? ""]
        components.path = Self.webSocketPath
        return components.url!
    }
}
