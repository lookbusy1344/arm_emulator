import SwiftUI

// Environment key for accessing AppDelegate
private struct AppDelegateKey: EnvironmentKey {
    static let defaultValue: AppDelegate? = nil
}

extension EnvironmentValues {
    var appDelegate: AppDelegate? {
        get { self[AppDelegateKey.self] }
        set { self[AppDelegateKey.self] = newValue }
    }
}

@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    let backendManager = BackendManager()
    let fileService = FileService.shared
    let settings = AppSettings.shared

    // Startup file path from command-line arguments (parsed once on first access)
    private(set) lazy var startupFilePath: String? = {
        let args = CommandLine.arguments

        guard args.count > 1 else {
            return nil
        }

        // Find first argument that looks like an assembly file (ends with .s)
        // This filters out Xcode debug flags like "-NSDocumentRevisionsDebugMode YES"
        guard let filePath = args.dropFirst().first(where: { $0.hasSuffix(".s") }) else {
            return nil
        }

        // Convert to absolute path if relative
        let absolutePath = (filePath as NSString).expandingTildeInPath
        let url = URL(fileURLWithPath: absolutePath)

        // Verify file exists
        guard FileManager.default.fileExists(atPath: url.path) else {
            print("Warning: Startup file not found: \(url.path)")
            return nil
        }

        return url.path
    }()
}

@main
struct ARMEmulatorApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @State private var showingAbout = false

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(appDelegate.backendManager)
                .environmentObject(appDelegate.fileService)
                .environmentObject(appDelegate.settings)
                .environment(\.appDelegate, appDelegate)
                .task {
                    await appDelegate.backendManager.ensureBackendRunning()
                }
                .sheet(isPresented: $showingAbout) {
                    AboutView()
                }
        }
        .commands {
            CommandGroup(replacing: .appInfo) {
                Button("About ARM Emulator") {
                    showingAbout = true
                }
            }
            FileCommands()
            DebugCommands()
        }
        .windowStyle(.titleBar)
        .windowToolbarStyle(.unified)

        Settings {
            PreferencesView()
                .environmentObject(appDelegate.settings)
        }
    }
}
