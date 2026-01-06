import SwiftUI

struct MemoryView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var addressInput = "0x8000"
    @State private var memoryData: [UInt8] = []
    @State private var baseAddress: UInt32 = 0x8000
    @State private var autoScrollEnabled = true

    private let bytesPerRow = 16
    private let rowsToShow = 16
    private var totalBytes: Int { bytesPerRow * rowsToShow }

    var body: some View {
        VStack(spacing: 0) {
            // Toolbar
            HStack(spacing: 8) {
                // Address input
                TextField("Address", text: $addressInput)
                    .textFieldStyle(.roundedBorder)
                    .frame(width: 100)
                    .font(.system(.body, design: .monospaced))

                Button("Go") {
                    loadMemoryFromInput()
                }
                .keyboardShortcut(.return, modifiers: [])

                Divider()

                // Quick access buttons
                Button("PC") {
                    loadMemory(at: viewModel.currentPC)
                }
                .help("Jump to Program Counter")

                Button("SP") {
                    loadMemory(at: viewModel.registers.sp)
                }
                .help("Jump to Stack Pointer")

                // Register buttons
                Button("R0") {
                    loadMemory(at: viewModel.registers.r0)
                }
                .help("Jump to R0 value")

                Button("R1") {
                    loadMemory(at: viewModel.registers.r1)
                }
                .help("Jump to R1 value")

                Button("R2") {
                    loadMemory(at: viewModel.registers.r2)
                }
                .help("Jump to R2 value")

                Button("R3") {
                    loadMemory(at: viewModel.registers.r3)
                }
                .help("Jump to R3 value")

                Spacer()

                // Auto-scroll toggle
                Toggle("Auto-scroll", isOn: $autoScrollEnabled)
                    .help("Auto-scroll to PC when stepping")
            }
            .padding(8)
            .background(Color(NSColor.controlBackgroundColor))

            Divider()

            // Memory display
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(0 ..< rowsToShow, id: \.self) { row in
                        MemoryRowView(
                            address: baseAddress + UInt32(row * bytesPerRow),
                            bytes: Array(memoryData[row * bytesPerRow ..< min(
                                (row + 1) * bytesPerRow,
                                memoryData.count
                            )]),
                            highlightAddress: autoScrollEnabled ? viewModel.currentPC : nil,
                            lastWriteAddress: viewModel.lastMemoryWrite
                        )
                    }
                }
                .font(.system(.body, design: .monospaced))
                .padding(8)
            }
        }
        .task {
            await loadMemory(at: baseAddress)
        }
        .onChange(of: viewModel.currentPC) { _ in
            if autoScrollEnabled {
                Task {
                    await loadMemory(at: viewModel.currentPC)
                }
            }
        }
    }

    private func loadMemoryFromInput() {
        var input = addressInput.trimmingCharacters(in: .whitespaces)

        // Remove 0x prefix if present
        if input.lowercased().hasPrefix("0x") {
            input = String(input.dropFirst(2))
        }

        guard let address = UInt32(input, radix: 16) else {
            return
        }

        loadMemory(at: address)
    }

    private func loadMemory(at address: UInt32) {
        Task {
            await viewModel.loadMemory(at: address, length: totalBytes)
            memoryData = viewModel.memoryData
            baseAddress = viewModel.memoryAddress
            addressInput = String(format: "0x%X", baseAddress)
        }
    }
}

struct MemoryRowView: View {
    let address: UInt32
    let bytes: [UInt8]
    let highlightAddress: UInt32?
    let lastWriteAddress: UInt32?

    private var isHighlighted: Bool {
        guard let highlight = highlightAddress else { return false }
        return address <= highlight && highlight < address + UInt32(bytes.count)
    }

    private func isRecentWrite(byteIndex: Int) -> Bool {
        guard let writeAddr = lastWriteAddress else { return false }
        let byteAddr = address + UInt32(byteIndex)
        // Highlight the written byte and the next 3 bytes (for 32-bit writes)
        return writeAddr <= byteAddr && byteAddr < writeAddr + 4
    }

    var body: some View {
        HStack(spacing: 4) {
            // Address
            Text(String(format: "0x%08X", address))
                .foregroundColor(.secondary)
                .frame(width: 90, alignment: .leading)

            Text("|")
                .foregroundColor(.secondary)

            // Hex bytes
            HStack(spacing: 2) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        Text(String(format: "%02X", bytes[i]))
                            .frame(width: 20)
                            .foregroundColor(isRecentWrite(byteIndex: i) ? .green : .primary)
                    } else {
                        Text("  ")
                            .frame(width: 20)
                    }
                }
            }

            Text("|")
                .foregroundColor(.secondary)

            // ASCII representation
            HStack(spacing: 0) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        let char = bytes[i]
                        let displayChar = (32 ... 126).contains(char) ? String(UnicodeScalar(char)) : "."
                        Text(displayChar)
                            .frame(width: 10)
                    } else {
                        Text(" ")
                            .frame(width: 10)
                    }
                }
            }

            Spacer()
        }
        .padding(.vertical, 2)
        .background(isHighlighted ? Color.accentColor.opacity(0.2) : Color.clear)
        .cornerRadius(2)
    }
}
