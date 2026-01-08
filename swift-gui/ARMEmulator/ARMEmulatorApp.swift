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

        // Take the first argument (ignore any extras)
        let filePath = args[1]

        // Convert to absolute path if relative
        let absolutePath = (filePath as NSString).expandingTildeInPath
        let url = URL(fileURLWithPath: absolutePath)
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
