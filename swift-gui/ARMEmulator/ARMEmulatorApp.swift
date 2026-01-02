import SwiftUI

@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    let backendManager = BackendManager()
}

@main
struct ARMEmulatorApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(appDelegate.backendManager)
                .task {
                    await appDelegate.backendManager.ensureBackendRunning()
                }
        }
        .commands {
            CommandGroup(replacing: .newItem) {}
        }
        .windowStyle(.titleBar)
        .windowToolbarStyle(.unified)
    }
}
