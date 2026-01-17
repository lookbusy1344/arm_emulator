import SwiftUI

struct MemoryView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var addressInput = "0x8000"
    @State private var memoryData: [UInt8] = []
    @State private var baseAddress: UInt32 = 0x8000
    @State private var autoScrollEnabled = true
    @State private var refreshID = UUID()
    @State private var scrollToRow: Int?
    @State private var lastScrolledRow: Int?

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
                scrollToRow: scrollToRow,
                refreshID: refreshID
            )
        }
        .task {
            await loadMemoryAsync(at: baseAddress)
        }
        .onChange(of: viewModel.lastMemoryWrite) {
            // Handle memory write highlighting and auto-scroll
            guard let writeAddr = viewModel.lastMemoryWrite else {
                return
            }

            Task {
                // Highlight the written address (triggers 1.5s fade)
                viewModel.highlightMemoryAddress(writeAddr, size: viewModel.lastMemoryWriteSize)

                // Check if write is within currently visible range
                let visibleEnd = baseAddress + UInt32(totalBytes)
                let isVisible = writeAddr >= baseAddress && writeAddr < visibleEnd

                if autoScrollEnabled && !isVisible {
                    // Write is outside visible range - scroll to it
                    let alignedAddress = writeAddr & ~UInt32(0xF)
                    await loadMemoryAsync(at: alignedAddress)
                } else {
                    // Write is visible - just refresh data in place
                    await refreshMemoryAsync()
                }

                // Trigger scroll to the row containing the write (if auto-scroll enabled)
                if autoScrollEnabled {
                    let rowOffset = Int((writeAddr - baseAddress) / UInt32(bytesPerRow))
                    if rowOffset >= 0 && rowOffset < rowsToShow && rowOffset != lastScrolledRow {
                        scrollToRow = rowOffset
                        lastScrolledRow = rowOffset
                    }
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

        // Clear scroll tracking when manually navigating
        lastScrolledRow = nil
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
        lastScrolledRow = nil
        Task {
            await loadMemoryAsync(at: address)
            refreshID = UUID()
        }
    }
}

struct MemoryRowView: View {
    let address: UInt32
    let bytes: [UInt8]
    let highlightAddress: UInt32?  // PC highlight (keep for now)
    let memoryHighlights: [UInt32: UUID]  // Changed from lastWriteAddress
    let lastWriteSize: UInt32  // Keep for compatibility

    private var isHighlighted: Bool {
        guard let highlight = highlightAddress else { return false }
        return address <= highlight && highlight < address + UInt32(bytes.count)
    }

    private func highlightID(for byteIndex: Int) -> UUID? {
        let byteAddr = address + UInt32(byteIndex)
        return memoryHighlights[byteAddr]
    }

    var body: some View {
        HStack(spacing: 4) {
            // Address
            Text(String(format: "0x%08X", address))
                .foregroundColor(Color(red: 0.5, green: 0.6, blue: 0.7))
                .frame(width: 75, alignment: .leading)

            // Hex bytes
            HStack(spacing: 2) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        let highlight = highlightID(for: i)
                        Text(String(format: "%02X", bytes[i]))
                            .frame(width: 20)
                            .foregroundColor(highlight != nil ? .white : .primary)
                            .fontWeight(highlight != nil ? .bold : .regular)
                            .background(highlight != nil ? Color.green : Color.clear)
                            .cornerRadius(2)
                            .animation(.easeOut(duration: 1.5), value: highlight)
                    } else {
                        Text("  ")
                            .frame(width: 20)
                    }
                }
            }

            // ASCII representation
            HStack(spacing: 0) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        let char = bytes[i]
                        let displayChar = (32 ... 126).contains(char) ? String(UnicodeScalar(char)) : "."
                        let highlight = highlightID(for: i)
                        Text(displayChar)
                            .frame(width: 10)
                            .foregroundColor(highlight != nil ? .white : .primary)
                            .fontWeight(highlight != nil ? .bold : .regular)
                            .background(highlight != nil ? Color.green : Color.clear)
                            .cornerRadius(2)
                            .animation(.easeOut(duration: 1.5), value: highlight)
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
                            memoryHighlights: viewModel.memoryHighlights,  // Pass highlights map
                            lastWriteSize: viewModel.lastMemoryWriteSize
                        )
                        .id("row_\(row)")
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(.vertical, 8)
                .padding(.horizontal, 4)
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
