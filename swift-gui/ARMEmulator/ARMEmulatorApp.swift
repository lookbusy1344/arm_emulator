import SwiftUI

@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    let backendManager = BackendManager()
    private var isShuttingDown = false

    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        guard !isShuttingDown else {
            return .terminateNow
        }

        isShuttingDown = true

        Task {
            await backendManager.shutdown()
            NSApp.reply(toApplicationShouldTerminate: true)
        }

        return .terminateLater
    }
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
