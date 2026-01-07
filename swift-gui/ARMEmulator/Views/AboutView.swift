import SwiftUI

struct AboutView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var backendVersion: BackendVersion?
    @State private var isLoading = false
    @State private var errorMessage: String?
    private let apiClient = APIClient()

    // Cache version info to avoid repeated API calls
    private static var cachedVersion: BackendVersion?

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "cpu")
                .font(.system(size: 60))
                .foregroundColor(.accentColor)

            Text("ARM Emulator")
                .font(.title)
                .fontWeight(.bold)

            if let version = backendVersion {
                VStack(spacing: 8) {
                    Text("Backend Version: \(version.version)")
                        .font(.headline)

                    Text("Commit: \(formatCommitHash(version.commit))")
                        .font(.subheadline)
                        .foregroundColor(.secondary)

                    Text("Build Date: \(formatDate(version.date))")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                }
            } else if isLoading {
                ProgressView()
                    .progressViewStyle(.circular)
            } else if let error = errorMessage {
                Text(error)
                    .font(.subheadline)
                    .foregroundColor(.red)
            }

            Divider()
                .padding(.horizontal)

            Text("An ARMv2 emulator with debugger")
                .font(.subheadline)
                .foregroundColor(.secondary)

            Text("© 2024-2026")
                .font(.caption)
                .foregroundColor(.secondary)

            Button("OK") {
                dismiss()
            }
            .keyboardShortcut(.defaultAction)
            .padding(.top, 8)
        }
        .padding(30)
        .frame(width: 400)
        .onAppear {
            loadVersionSync()
        }
    }

    private func loadVersionSync() {
        // Check cache first - instant display
        if let cached = Self.cachedVersion {
            backendVersion = cached
            DebugLog.log("Using cached version info", category: "AboutView")
            return
        }

        // Not cached - fetch asynchronously
        Task {
            await loadVersionAsync()
        }
    }

    private func loadVersionAsync() async { isLoading = true
        errorMessage = nil

        do {
            let version = try await apiClient.getVersion()
            backendVersion = version
            Self.cachedVersion = version // Cache for future use
            DebugLog.log("Fetched and cached version info", category: "AboutView")
        } catch {
            errorMessage = "Failed to load backend version"
            DebugLog.error("Failed to load version: \(error)", category: "AboutView")
        }

        isLoading = false
    }

    private func formatCommitHash(_ hash: String) -> String {
        if hash == "unknown" {
            return hash
        }
        return String(hash.prefix(8))
    }

    private func formatDate(_ dateString: String) -> String {
        if dateString == "unknown" {
            return dateString
        }
        return dateString
    }
}

#Preview {
    AboutView()
}
