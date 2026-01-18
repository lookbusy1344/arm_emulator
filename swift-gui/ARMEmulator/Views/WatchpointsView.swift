import SwiftUI

struct WatchpointsView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var addressInput = ""
    @State private var selectedType = "readwrite"

    let watchpointTypes = [
        ("read", "Read"),
        ("write", "Write"),
        ("readwrite", "Read/Write"),
    ]

    var body: some View {
        VStack(spacing: 0) {
            // Add watchpoint form
            VStack(spacing: 12) {
                HStack {
                    TextField("Address (e.g., 0x8000)", text: $addressInput)
                        .textFieldStyle(.roundedBorder)
                        .font(.system(size: 10, design: .monospaced))

                    Picker("Type", selection: $selectedType) {
                        ForEach(watchpointTypes, id: \.0) { type in
                            Text(type.1).tag(type.0)
                        }
                    }
                    .pickerStyle(.segmented)
                    .frame(width: 200)

                    Button("Add") {
                        Task { await addWatchpoint() }
                    }
                    .disabled(addressInput.isEmpty)
                }
            }
            .padding()

            Divider()

            // Watchpoints list
            if viewModel.watchpoints.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "eye.slash")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary)
                    Text("No watchpoints set")
                        .foregroundColor(.secondary)
                    Text("Watchpoints trigger when memory is accessed")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    LazyVStack(spacing: 8) {
                        ForEach(viewModel.watchpoints) { watchpoint in
                            watchpointRow(watchpoint)
                        }
                    }
                    .padding()
                }
            }
        }
        .navigationTitle("Watchpoints")
        .task {
            await viewModel.refreshWatchpoints()
        }
    }

    @ViewBuilder
    private func watchpointRow(_ watchpoint: Watchpoint) -> some View {
        HStack {
            Image(systemName: watchpointIcon(for: watchpoint.type))
                .foregroundColor(.blue)
                .frame(width: 24)

            VStack(alignment: .leading, spacing: 2) {
                Text(String(format: "0x%08X", watchpoint.address))
                    .font(.system(size: 10, design: .monospaced))
                    .fontWeight(.medium)

                Text(watchpointTypeLabel(watchpoint.type))
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Button {
                Task { await viewModel.removeWatchpoint(id: watchpoint.id) }
            } label: {
                Image(systemName: "trash")
                    .foregroundColor(.red)
            }
            .buttonStyle(.plain)
            .help("Remove watchpoint")
        }
        .padding()
        .background(Color.secondary.opacity(0.1))
        .cornerRadius(8)
    }

    private func watchpointIcon(for type: String) -> String {
        switch type {
        case "read": "eye"
        case "write": "pencil"
        case "readwrite": "eye.fill"
        default: "eye"
        }
    }

    private func watchpointTypeLabel(_ type: String) -> String {
        switch type {
        case "read": "Read"
        case "write": "Write"
        case "readwrite": "Read/Write"
        default: type
        }
    }

    private func addWatchpoint() async {
        // Parse address
        let trimmed = addressInput.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }

        let address: UInt32
        if trimmed.hasPrefix("0x") {
            guard let parsed = UInt32(trimmed.dropFirst(2), radix: 16) else {
                viewModel.errorMessage = "Invalid hexadecimal address"
                return
            }
            address = parsed
        } else {
            guard let parsed = UInt32(trimmed) else {
                viewModel.errorMessage = "Invalid address"
                return
            }
            address = parsed
        }

        await viewModel.addWatchpoint(at: address, type: selectedType)
        addressInput = ""
    }
}
