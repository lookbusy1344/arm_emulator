import SwiftUI

@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    let backendManager = BackendManager()
    let fileService = FileService.shared
    let settings = AppSettings.shared
}

@main
struct ARMEmulatorApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(appDelegate.backendManager)
                .environmentObject(appDelegate.fileService)
                .environmentObject(appDelegate.settings)
                .task {
                    await appDelegate.backendManager.ensureBackendRunning()
                }
        }
        .commands {
            FileCommands()
        }
        .windowStyle(.titleBar)
        .windowToolbarStyle(.unified)

        Settings {
            PreferencesView()
                .environmentObject(appDelegate.settings)
        }
    }
}
