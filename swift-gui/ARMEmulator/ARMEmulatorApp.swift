import SwiftUI

@main
struct ARMEmulatorApp: App {
    var body: some Scene {
        WindowGroup {
            MainView()
        }
        .commands {
            CommandGroup(replacing: .newItem) {}
        }
        .windowStyle(.titleBar)
        .windowToolbarStyle(.unified)
    }
}
