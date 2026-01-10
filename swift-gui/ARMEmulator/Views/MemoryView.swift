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
            #if DEBUG
            // Diagnostic overlay - shows live state for debugging
            VStack(alignment: .leading, spacing: 2) {
                Text("baseAddress: 0x\(String(format: "%08X", baseAddress))")
                Text("highlightAddr: \(highlightedWriteAddress.map { String(format: "0x%08X", $0) } ?? "nil")")
                Text("memoryData.count: \(memoryData.count)")
                Text("refreshID: \(refreshID.uuidString.prefix(8))...")
                if memoryData.count > 0 {
                    let visibleRange = "0x\(String(format: "%08X", baseAddress)) - 0x\(String(format: "%08X", baseAddress + UInt32(memoryData.count)))"
                    Text("Visible range: \(visibleRange)")
                }
                if let highlight = highlightedWriteAddress {
                    let inRange = highlight >= baseAddress && highlight < baseAddress + UInt32(memoryData.count)
                    Text("Highlight in range: \(inRange ? "YES" : "NO")")
                    if inRange {
                        let offset = Int(highlight - baseAddress)
                        let byteValue = offset < memoryData.count ? String(format: "0x%02X", memoryData[offset]) : "?"
                        Text("Byte at highlight: \(byteValue)")
                    }
                }
            }
            .font(.system(size: 9, design: .monospaced))
            .padding(4)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Color.yellow.opacity(0.3))
            #endif

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
            .fixedSize(horizontal: false, vertical: true)

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

                    // Check if write is within currently visible range
                    let visibleEnd = baseAddress + UInt32(totalBytes)
                    let isVisible = writeAddr >= baseAddress && writeAddr < visibleEnd

                    DebugLog.log(
                        "Visibility check: writeAddr=0x\(String(format: "%08X", writeAddr)), visible=[0x\(String(format: "%08X", baseAddress))-0x\(String(format: "%08X", visibleEnd))), isVisible=\(isVisible)",
                        category: "MemoryView"
                    )

                    if autoScrollEnabled && !isVisible {
                        // Write is outside visible range - scroll to it (aligned to 16-byte boundary)
                        let alignedAddress = writeAddr & ~UInt32(0xF)  // Align to 16-byte boundary
                        DebugLog.log(
                            "Scrolling to aligned address 0x\(String(format: "%08X", alignedAddress))",
                            category: "MemoryView"
                        )
                        await loadMemoryAsync(at: alignedAddress)
                    } else {
                        // Write is visible (or auto-scroll disabled) - just refresh data in place
                        DebugLog.log("Write is visible, refreshing in place", category: "MemoryView")
                        await refreshMemoryAsync()
                    }

                    // Set highlighting AFTER memory is loaded
                    highlightedWriteAddress = writeAddr
                    DebugLog.log(
                        "Highlight set to 0x\(String(format: "%08X", writeAddr)), baseAddress=0x\(String(format: "%08X", baseAddress)), memoryData.count=\(memoryData.count)",
                        category: "MemoryView"
                    )

                    // Force view refresh
                    DebugLog.log("Refreshing view with new UUID", category: "MemoryView")
                    refreshID = UUID()
                } else {
                    // No write this step - DON'T clear highlighting
                    // Keep the previous highlight visible until the next write occurs
                    DebugLog.log("No write this step, keeping previous highlight", category: "MemoryView")
                }
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
        // Compute highlighting once per row render for debugging
        let highlightOffset: Int? = {
            guard let writeAddr = lastWriteAddress else { return nil }
            if writeAddr >= address && writeAddr < address + 16 {
                return Int(writeAddr - address)
            }
            return nil
        }()

        HStack(spacing: 4) {
            // Address
            Text(String(format: "0x%08X", address))
                .foregroundColor(.secondary)
                .frame(width: 90, alignment: .leading)

            // Debug: show if this row should have highlighting
            #if DEBUG
            if let offset = highlightOffset {
                Text("[\(offset)]")
                    .foregroundColor(.red)
                    .font(.system(size: 8))
            }
            #endif

            Text("|")
                .foregroundColor(.secondary)

            // Hex bytes
            HStack(spacing: 2) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        // Use pre-computed offset for highlighting
                        let shouldHighlight: Bool = {
                            guard let offset = highlightOffset else { return false }
                            return i >= offset && i < offset + 4
                        }()
                        Text(String(format: "%02X", bytes[i]))
                            .frame(width: 20)
                            .foregroundColor(shouldHighlight ? .white : .primary)
                            .fontWeight(shouldHighlight ? .bold : .regular)
                            .background(shouldHighlight ? Color.green : Color.clear)
                            .cornerRadius(2)
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
                        let shouldHighlight: Bool = {
                            guard let offset = highlightOffset else { return false }
                            return i >= offset && i < offset + 4
                        }()
                        Text(displayChar)
                            .frame(width: 10)
                            .foregroundColor(shouldHighlight ? .white : .primary)
                            .fontWeight(shouldHighlight ? .bold : .regular)
                            .background(shouldHighlight ? Color.green : Color.clear)
                            .cornerRadius(2)
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
        #if DEBUG
        .onAppear {
            // Check if this row should have highlighting
            let hasHighlight = lastWriteAddress != nil
            let highlightInThisRow: Bool = {
                guard let writeAddr = lastWriteAddress else { return false }
                return writeAddr >= address && writeAddr < address + 16
            }()
            DebugLog.log(
                "RowView APPEAR: 0x\(String(format: "%08X", address)), bytes=\(bytes.count), lastWriteAddress=\(lastWriteAddress.map { String(format: "0x%08X", $0) } ?? "nil"), highlightInRow=\(highlightInThisRow)",
                category: "MemoryRow"
            )
        }
        .onChange(of: bytes) { newBytes in
            DebugLog.log(
                "RowView BYTES CHANGED: 0x\(String(format: "%08X", address)) -> [\(newBytes.prefix(4).map { String(format: "%02X", $0) }.joined(separator: " "))]",
                category: "MemoryRow"
            )
        }
        .onChange(of: lastWriteAddress) { newHighlight in
            let highlightInThisRow: Bool = {
                guard let writeAddr = newHighlight else { return false }
                return writeAddr >= address && writeAddr < address + 16
            }()
            if highlightInThisRow {
                DebugLog.log(
                    "RowView HIGHLIGHT CHANGED: 0x\(String(format: "%08X", address)) should highlight 0x\(String(format: "%08X", newHighlight!))",
                    category: "MemoryRow"
                )
            }
        }
        #endif
    }
}
