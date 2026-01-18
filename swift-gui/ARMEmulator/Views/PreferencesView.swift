import SwiftUI

struct PreferencesView: View {
    @EnvironmentObject var settings: AppSettings

    var body: some View {
        TabView {
            GeneralPreferences(settings: settings)
                .tabItem {
                    Label("General", systemImage: "gearshape")
                }

            EditorPreferences(settings: settings)
                .tabItem {
                    Label("Editor", systemImage: "doc.text")
                }
        }
        .frame(width: 500, height: 300)
    }
}

struct GeneralPreferences: View {
    @ObservedObject var settings: AppSettings

    var body: some View {
        Form {
            Section("Backend") {
                TextField("Backend URL", text: $settings.backendURL)
                    .textFieldStyle(.roundedBorder)
                Text("Default: http://localhost:8080")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Section("Appearance") {
                Picker("Color Scheme", selection: $settings.colorScheme) {
                    Text("Auto (System)").tag("auto")
                    Text("Light").tag("light")
                    Text("Dark").tag("dark")
                }
                .pickerStyle(.segmented)
            }

            Section("Files") {
                Stepper(
                    "Recent Files: \(settings.maxRecentFiles)",
                    value: $settings.maxRecentFiles,
                    in: 5 ... 20,
                )
            }
        }
        .padding()
    }
}

struct EditorPreferences: View {
    @ObservedObject var settings: AppSettings

    var body: some View {
        Form {
            Section("Font") {
                Stepper(
                    "Font Size: \(settings.editorFontSize)",
                    value: $settings.editorFontSize,
                    in: 10 ... 24,
                )
                Text("Current size: \(settings.editorFontSize) pt")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Section("Preview") {
                Text("MOV R0, #42  ; Example assembly code")
                    .font(.system(size: CGFloat(settings.editorFontSize), design: .monospaced))
                    .padding()
                    .background(Color(NSColor.textBackgroundColor))
                    .cornerRadius(4)
            }
        }
        .padding()
    }
}
