import SwiftUI

@MainActor
class AppSettings: ObservableObject {
    @AppStorage("backendURL") var backendURL = "http://localhost:8080"
    @AppStorage("editorFontSize") var editorFontSize = 14
    @AppStorage("colorScheme") var colorScheme = "auto" // "light", "dark", "auto"
    @AppStorage("maxRecentFiles") var maxRecentFiles = 10
    @AppStorage("selectedTab") var selectedTab = 0

    static let shared = AppSettings()

    var preferredColorScheme: ColorScheme? {
        switch colorScheme {
        case "light":
            return .light
        case "dark":
            return .dark
        default:
            return nil // Auto (use system)
        }
    }
}
