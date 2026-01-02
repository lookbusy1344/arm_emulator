import SwiftUI

struct BackendStatusView: View {
    let status: BackendManager.BackendStatus
    let onRetry: () async -> Void

    var body: some View {
        VStack(spacing: 20) {
            switch status {
            case .unknown, .starting:
                ProgressView()
                Text("Starting backend...")
                    .foregroundColor(.secondary)

            case .running:
                EmptyView()

            case .stopped:
                Image(systemName: "exclamationmark.triangle")
                    .font(.system(size: 48))
                    .foregroundColor(.orange)
                Text("Backend Stopped")
                    .font(.headline)
                Text("The ARM Emulator backend is not running")
                    .foregroundColor(.secondary)
                Button("Start Backend") {
                    Task {
                        await onRetry()
                    }
                }
                .buttonStyle(.borderedProminent)

            case let .error(message):
                Image(systemName: "xmark.circle")
                    .font(.system(size: 48))
                    .foregroundColor(.red)
                Text("Backend Error")
                    .font(.headline)
                Text(message)
                    .foregroundColor(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
                Button("Retry") {
                    Task {
                        await onRetry()
                    }
                }
                .buttonStyle(.borderedProminent)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(NSColor.windowBackgroundColor))
    }
}

struct BackendStatusView_Previews: PreviewProvider {
    static var previews: some View {
        Group {
            BackendStatusView(status: .starting) {}
                .previewDisplayName("Starting")

            BackendStatusView(status: .error("Port 8080 is already in use")) {}
                .previewDisplayName("Error")

            BackendStatusView(status: .stopped) {}
                .previewDisplayName("Stopped")

            BackendStatusView(status: .unknown) {}
                .previewDisplayName("Unknown")
        }
        .frame(width: 400, height: 300)
    }
}
