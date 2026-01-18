import SwiftUI

struct DisassemblyView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var disassembly: [DisassemblyInstruction] = []

    private let instructionsToShow = 64 // ±32 around PC

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Disassembly (PC = \(String(format: "0x%08X", viewModel.currentPC)))")
                    .font(.system(size: 11, weight: .semibold))
                    .padding(8)
                Spacer()
                Button("Refresh") {
                    loadDisassembly()
                }
                .padding(.trailing, 8)
            }
            .background(Color(NSColor.controlBackgroundColor))

            Divider()

            // Disassembly display
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(disassembly) { instruction in
                        DisassemblyRowView(
                            instruction: instruction,
                            isPC: instruction.address == viewModel.currentPC,
                            hasBreakpoint: viewModel.breakpoints.contains(instruction.address),
                        )
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(8)
            }
        }
        .task {
            loadDisassembly()
        }
        .onChange(of: viewModel.currentPC) {
            loadDisassembly()
        }
    }

    private func loadDisassembly() {
        Task {
            let startAddress = viewModel.currentPC > 128 ? viewModel.currentPC - 128 : 0
            await viewModel.loadDisassembly(around: startAddress, count: instructionsToShow)
            disassembly = viewModel.disassembly
        }
    }
}

struct DisassemblyRowView: View {
    let instruction: DisassemblyInstruction
    let isPC: Bool
    let hasBreakpoint: Bool

    var body: some View {
        HStack(spacing: 4) {
            // Breakpoint indicator
            if hasBreakpoint {
                Text("●")
                    .foregroundColor(.red)
                    .fontWeight(.bold)
            } else {
                Text(" ")
            }

            // PC indicator
            if isPC {
                Text("►")
                    .foregroundColor(.accentColor)
                    .fontWeight(.bold)
            } else {
                Text(" ")
            }

            // Address
            Text(String(format: "0x%08X", instruction.address))
                .foregroundColor(.secondary)
                .frame(width: 90, alignment: .leading)

            Text("|")
                .foregroundColor(.secondary)

            // Machine code
            Text(String(format: "%08X", instruction.machineCode))
                .frame(width: 70, alignment: .leading)

            Text("|")
                .foregroundColor(.secondary)

            // Mnemonic
            Text(instruction.mnemonic)
                .frame(minWidth: 150, alignment: .leading)

            // Symbol (if present)
            if let symbol = instruction.symbol, !symbol.isEmpty {
                Text("|")
                    .foregroundColor(.secondary)
                Text(symbol)
                    .foregroundColor(.secondary)
            }

            Spacer()
        }
        .padding(.vertical, 2)
        .background(isPC ? Color.accentColor.opacity(0.2) : Color.clear)
        .cornerRadius(2)
    }
}
