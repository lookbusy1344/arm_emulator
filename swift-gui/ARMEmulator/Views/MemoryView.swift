import SwiftUI

struct MemoryView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var addressInput = "0x8000"
    @State private var memoryData: [UInt8] = []
    @State private var baseAddress: UInt32 = 0x8000
    @State private var autoScrollEnabled = true
    @State private var highlightedWriteAddress: UInt32?
    @State private var refreshID = UUID()
    @State private var scrollToRow: Int?

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
            .fixedSize(horizontal: false, vertical: true)

            Divider()

            // Memory display
            MemoryDisplayView(
                rowsToShow: rowsToShow,
                bytesPerRow: bytesPerRow,
                baseAddress: baseAddress,
                memoryData: memoryData,
                viewModel: viewModel,
                highlightedWriteAddress: highlightedWriteAddress,
                scrollToRow: scrollToRow,
                refreshID: refreshID
            )
        }
        .task {
            await loadMemoryAsync(at: baseAddress)
        }
        .onChange(of: viewModel.lastMemoryWrite) {
            // Handle memory write highlighting and auto-scroll
            // Note: We intentionally DON'T refresh on PC change - memory content
            // only changes when there's a write, not when PC advances.
            guard let writeAddr = viewModel.lastMemoryWrite else {
                // No write this step - keep previous highlight visible
                return
            }

            Task {
                // Check if write is within currently visible range
                let visibleEnd = baseAddress + UInt32(totalBytes)
                let isVisible = writeAddr >= baseAddress && writeAddr < visibleEnd

                if autoScrollEnabled && !isVisible {
                    // Write is outside visible range - scroll to it (aligned to 16-byte boundary)
                    let alignedAddress = writeAddr & ~UInt32(0xF)
                    await loadMemoryAsync(at: alignedAddress)
                } else {
                    // Write is visible (or auto-scroll disabled) - just refresh data in place
                    await refreshMemoryAsync()
                }

                // Set highlighting AFTER memory is loaded (critical for correct display)
                highlightedWriteAddress = writeAddr

                // Trigger scroll to the row containing the write (if auto-scroll enabled)
                if autoScrollEnabled {
                    let rowOffset = Int((writeAddr - baseAddress) / UInt32(bytesPerRow))
                    if rowOffset >= 0 && rowOffset < rowsToShow {
                        scrollToRow = rowOffset
                    }
                }

                // Force view refresh
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
        await viewModel.loadMemory(at: address, length: totalBytes)

        // Copy data from viewModel to local state
        memoryData = viewModel.memoryData
        baseAddress = viewModel.memoryAddress
        addressInput = String(format: "0x%X", baseAddress)
    }

    /// Refreshes memory at the current base address without changing navigation
    private func refreshMemoryAsync() async {
        await viewModel.loadMemory(at: baseAddress, length: totalBytes)
        memoryData = viewModel.memoryData
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
        // Compute highlighting once per row render
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
    }
}

/// Separate view for the memory display with ScrollViewReader
/// This separation avoids SwiftUI type inference issues with nested modifiers
struct MemoryDisplayView: View {
    let rowsToShow: Int
    let bytesPerRow: Int
    let baseAddress: UInt32
    let memoryData: [UInt8]
    @ObservedObject var viewModel: EmulatorViewModel
    let highlightedWriteAddress: UInt32?
    let scrollToRow: Int?
    let refreshID: UUID

    var body: some View {
        ScrollView([.vertical, .horizontal]) {
            ScrollViewReader { proxy in
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(0 ..< rowsToShow, id: \.self) { row in
                        MemoryRowView(
                            address: baseAddress + UInt32(row * bytesPerRow),
                            bytes: bytesForRow(row),
                            highlightAddress: viewModel.currentPC,
                            lastWriteAddress: highlightedWriteAddress
                        )
                        .id("row_\(row)")
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(8)
                .task(id: scrollToRow) {
                    if let row = scrollToRow {
                        try? await Task.sleep(nanoseconds: 10_000_000) // 10ms delay
                        withAnimation(.easeInOut(duration: 0.3)) {
                            proxy.scrollTo("row_\(row)", anchor: .center)
                        }
                    }
                }
            }
        }
        .id(refreshID)
    }

    private func bytesForRow(_ row: Int) -> [UInt8] {
        let startIndex = row * bytesPerRow
        let endIndex = min((row + 1) * bytesPerRow, memoryData.count)

        guard startIndex < memoryData.count else {
            return []
        }

        return Array(memoryData[startIndex ..< endIndex])
    }
}
