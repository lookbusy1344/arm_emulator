import SwiftUI

struct BreakpointsListView: View {
    @ObservedObject var viewModel: EmulatorViewModel

    var body: some View {
        VStack(spacing: 0) {
            if viewModel.breakpoints.isEmpty && viewModel.watchpoints.isEmpty {
                emptyStateView
            } else {
                List {
                    if !viewModel.breakpoints.isEmpty {
                        Section(header: Text("Breakpoints").font(.system(size: 11, weight: .semibold))) {
                            ForEach(Array(viewModel.breakpoints).sorted(), id: \.self) { address in
                                breakpointRow(address: address)
                            }
                        }
                    }

                    if !viewModel.watchpoints.isEmpty {
                        Section(header: Text("Watchpoints").font(.system(size: 11, weight: .semibold))) {
                            ForEach(viewModel.watchpoints) { watchpoint in
                                watchpointRow(watchpoint: watchpoint)
                            }
                        }
                    }
                }
                .listStyle(.inset)
            }
        }
    }

    private var emptyStateView: some View {
        VStack(spacing: 12) {
            Image(systemName: "circle.hexagongrid")
                .font(.system(size: 48))
                .foregroundColor(.secondary)

            Text("No Breakpoints or Watchpoints")
                .font(.system(size: 11, weight: .semibold))
                .foregroundColor(.secondary)

            Text("Set breakpoints in the editor or toggle them in the disassembly view")
                .font(.caption)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func breakpointRow(address: UInt32) -> some View {
        HStack {
            Image(systemName: "circle.fill")
                .foregroundColor(.red)
                .font(.system(size: 11))

            Text(String(format: "0x%08X", address))
                .font(.system(size: 10, design: .monospaced))

            Spacer()

            Button(action: {
                Task {
                    await viewModel.toggleBreakpoint(at: address)
                }
            }) {
                Image(systemName: "trash")
                    .foregroundColor(.red)
            }
            .buttonStyle(.plain)
            .help("Remove breakpoint")
        }
        .padding(.vertical, 4)
    }

    private func watchpointRow(watchpoint: Watchpoint) -> some View {
        HStack {
            Image(systemName: "eye.fill")
                .foregroundColor(.blue)
                .font(.system(size: 11))

            VStack(alignment: .leading, spacing: 2) {
                Text(String(format: "0x%08X", watchpoint.address))
                    .font(.system(size: 10, design: .monospaced))

                Text(watchpoint.type.capitalized)
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Button(action: {
                Task {
                    await viewModel.removeWatchpoint(id: watchpoint.id)
                }
            }) {
                Image(systemName: "trash")
                    .foregroundColor(.red)
            }
            .buttonStyle(.plain)
            .help("Remove watchpoint")
        }
        .padding(.vertical, 4)
    }
}
