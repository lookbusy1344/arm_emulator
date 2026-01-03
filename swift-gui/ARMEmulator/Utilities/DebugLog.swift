import Foundation

/// Debug logging utility with conditional compilation
/// Only logs in DEBUG builds, completely removed in RELEASE builds
enum DebugLog {
    /// Enable/disable debug logging even in DEBUG builds
    /// Set to false to silence all debug logs without rebuilding
    static var enabled = true

    /// Log a debug message (only in DEBUG builds)
    static func log(_ message: String, category: String = "App") {
        #if DEBUG
            if enabled {
                print("🔵 [\(category)] \(message)")
            }
        #endif
    }

    /// Log a success message (only in DEBUG builds)
    static func success(_ message: String, category: String = "App") {
        #if DEBUG
            if enabled {
                print("✅ [\(category)] \(message)")
            }
        #endif
    }

    /// Log an error message (only in DEBUG builds)
    static func error(_ message: String, category: String = "App") {
        #if DEBUG
            if enabled {
                print("❌ [\(category)] \(message)")
            }
        #endif
    }

    /// Log a warning message (only in DEBUG builds)
    static func warning(_ message: String, category: String = "App") {
        #if DEBUG
            if enabled {
                print("⚠️ [\(category)] \(message)")
            }
        #endif
    }

    /// Log a network request (only in DEBUG builds)
    static func network(_ message: String) {
        #if DEBUG
            if enabled {
                print("🌐 [Network] \(message)")
            }
        #endif
    }

    /// Log a UI event (only in DEBUG builds)
    static func ui(_ message: String) {
        #if DEBUG
            if enabled {
                print("🟢 [UI] \(message)")
            }
        #endif
    }
}
