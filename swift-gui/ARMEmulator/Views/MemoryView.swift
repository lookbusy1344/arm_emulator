import SwiftUI

struct MemoryView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var addressInput = "0x8000"
    @State private var memoryData: [UInt8] = []
    @State private var baseAddress: UInt32 = 0x8000
    @State private var autoScrollEnabled = true
    @State private var highlightedWriteAddress: UInt32?
    @State private var refreshID = UUID()

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
                    .font(.system(size: 10, design: .monospaced))

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
                    .help("Auto-scroll to memory writes when stepping")
            }
            .padding(8)
            .background(Color(NSColor.controlBackgroundColor))

            Divider()

            // Memory display
            ScrollView([.vertical, .horizontal]) {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(0 ..< rowsToShow, id: \.self) { row in
                        MemoryRowView(
                            address: baseAddress + UInt32(row * bytesPerRow),
                            bytes: bytesForRow(row),
                            highlightAddress: viewModel.currentPC,
                            lastWriteAddress: highlightedWriteAddress
                        )
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(8)
            }
            .id(refreshID)
        }
        .task {
            await loadMemoryAsync(at: baseAddress)
        }
        .onChange(of: viewModel.lastMemoryWrite) { newWriteAddress in
            // Handle memory write highlighting and auto-scroll
            // Note: We intentionally DON'T refresh on PC change - memory content
            // only changes when there's a write, not when PC advances.
            DebugLog.log(
                "onChange(lastMemoryWrite): newWriteAddress=\(newWriteAddress.map { String(format: "0x%08X", $0) } ?? "nil")",
                category: "MemoryView"
            )

            Task {
                DebugLog.log("onChange Task starting, autoScrollEnabled=\(autoScrollEnabled)", category: "MemoryView")

                if let writeAddr = newWriteAddress {
                    DebugLog.log(
                        "Processing write at 0x\(String(format: "%08X", writeAddr))",
                        category: "MemoryView"
                    )

                    if autoScrollEnabled {
                        // Scroll to write address and highlight
                        DebugLog.log("Calling loadMemoryAsync...", category: "MemoryView")
                        await loadMemoryAsync(at: writeAddr)
                        DebugLog.log("loadMemoryAsync completed", category: "MemoryView")
                    } else {
                        // Just refresh current view in case write is visible
                        await refreshMemoryAsync()
                    }

                    // Set highlighting AFTER memory is loaded
                    highlightedWriteAddress = writeAddr
                    DebugLog.log(
                        "Highlight set to 0x\(String(format: "%08X", writeAddr)), baseAddress=0x\(String(format: "%08X", baseAddress)), memoryData.count=\(memoryData.count)",
                        category: "MemoryView"
                    )
                } else {
                    // No write this step - clear highlighting
                    DebugLog.log("Clearing highlight (no write)", category: "MemoryView")
                    highlightedWriteAddress = nil
                }
                // Force view refresh
                DebugLog.log("Refreshing view with new UUID", category: "MemoryView")
                refreshID = UUID()
            }
        }
    }

    private func bytesForRow(_ row: Int) -> [UInt8] {
        let startIndex = row * bytesPerRow
        let endIndex = min((row + 1) * bytesPerRow, memoryData.count)

        // Return empty array if start is beyond available data
        guard startIndex < memoryData.count else {
            return []
        }

        return Array(memoryData[startIndex ..< endIndex])
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

        // Clear any existing highlight when manually navigating
        highlightedWriteAddress = nil
        Task {
            await loadMemoryAsync(at: address)
            refreshID = UUID()
        }
    }

    /// Loads memory at the specified address and updates local state
    /// This is a proper async function that can be awaited
    private func loadMemoryAsync(at address: UInt32) async {
        DebugLog.log("loadMemoryAsync: Requesting address 0x\(String(format: "%08X", address))", category: "MemoryView")
        await viewModel.loadMemory(at: address, length: totalBytes)

        // Copy data from viewModel to local state
        let loadedData = viewModel.memoryData
        let loadedAddress = viewModel.memoryAddress

        DebugLog.log(
            "loadMemoryAsync: viewModel returned \(loadedData.count) bytes at 0x\(String(format: "%08X", loadedAddress))",
            category: "MemoryView"
        )

        memoryData = loadedData
        baseAddress = loadedAddress
        addressInput = String(format: "0x%X", baseAddress)

        DebugLog.log(
            "loadMemoryAsync: Local state updated - baseAddress=0x\(String(format: "%08X", baseAddress)), memoryData.count=\(memoryData.count)",
            category: "MemoryView"
        )
    }

    /// Refreshes memory at the current base address without changing navigation
    private func refreshMemoryAsync() async {
        await viewModel.loadMemory(at: baseAddress, length: totalBytes)
        memoryData = viewModel.memoryData
        DebugLog.log(
            "Memory refreshed at 0x\(String(format: "%08X", baseAddress)), \(memoryData.count) bytes",
            category: "MemoryView"
        )
    }

    /// Synchronous wrapper for button actions that start a Task
    private func loadMemory(at address: UInt32) {
        highlightedWriteAddress = nil
        Task {
            await loadMemoryAsync(at: address)
            refreshID = UUID()
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
                            .fontWeight(isRecentWrite(byteIndex: i) ? .bold : .regular)
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
                            .foregroundColor(isRecentWrite(byteIndex: i) ? .green : .primary)
                            .fontWeight(isRecentWrite(byteIndex: i) ? .bold : .regular)
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
