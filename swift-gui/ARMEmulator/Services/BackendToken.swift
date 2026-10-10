import Foundation

/// The per-launch token the Go backend writes for its port. Every request except
/// `/health` must carry it as a bearer token.
enum BackendToken {
    static let defaultPort = 8080

    static func fileURL(port: Int, home: URL = FileManager.default.homeDirectoryForCurrentUser) -> URL {
        home.appending(path: ".config/arm-emu/api-token-\(port)")
    }

    /// Reads the token afresh, so a restarted backend's new token is picked up.
    static func read(port: Int, home: URL = FileManager.default.homeDirectoryForCurrentUser) -> String? {
        guard let text = try? String(contentsOf: fileURL(port: port, home: home), encoding: .utf8) else {
            return nil
        }
        let token = text.trimmingCharacters(in: .whitespacesAndNewlines)
        return token.isEmpty ? nil : token
    }

    static func authorize(_ request: inout URLRequest, token: String?) {
        if let token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
    }
}
