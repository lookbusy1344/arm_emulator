import SwiftUI

struct StackEntry: Identifiable {
    let id = UUID()
    let offset: Int
    let address: UInt32
    let value: UInt32
    let annotation: String
}

struct StackView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var stackData: [StackEntry] = []
    @State private var localMemoryData: [UInt8] = []
    @State private var loadCount = 0 // Debug: track load attempts

    // Stack configuration (from vm/constants.go)
    private let initialSP: UInt32 = 0x00050000 // Initial stack pointer (StackSegmentStart + StackSegmentSize)

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                let stackSize = viewModel.registers.sp < initialSP ? initialSP - viewModel.registers.sp : 0
                Text("Stack ↓ (SP = \(String(format: "0x%08X", viewModel.registers.sp))) - \(stackSize) bytes used")
                    .font(.system(size: 11, weight: .semibold))
                    .padding(8)
                Spacer()
            }
            .background(Color(NSColor.controlBackgroundColor))

            Divider()

            // Stack display
            ScrollView {
                if stackData.isEmpty {
                    Text("Stack is empty (SP at initial value)")
                        .foregroundColor(.secondary)
                        .padding()
                } else {
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(stackData) { entry in
                            StackRowView(
                                offset: entry.offset,
                                address: entry.address,
                                value: entry.value,
                                annotation: entry.annotation,
                                isCurrent: entry.offset == 0  // Highlight current SP location
                            )
                        }
                    }
                    .font(.system(size: 10, design: .monospaced))
                    .padding(8)
                }
            }
        }
        .task {
            await loadStack()
        }
        .onChange(of: viewModel.registers.sp) { _ in
            Task {
                await loadStack()
            }
        }
    }

    private func loadStack() async {
        let currentSP = viewModel.registers.sp
        DebugLog.log("loadStack() called, SP = 0x\(String(format: "%08X", currentSP))", category: "StackView")

        // Check if SP is at initial value (no stack usage yet)
        if currentSP >= initialSP {
            DebugLog.log("SP at or above initial value (0x\(String(format: "%08X", initialSP))), no stack data", category: "StackView")
            stackData = []
            localMemoryData = []
            return
        }

        // Calculate actual stack size: how many bytes have been pushed
        let stackSize = initialSP - currentSP
        DebugLog.log("Stack size: \(stackSize) bytes (from 0x\(String(format: "%08X", currentSP)) to 0x\(String(format: "%08X", initialSP - 1)))", category: "StackView")

        // Fetch ONLY the actual stack contents from currentSP to (initialSP - 1)
        let startAddress = currentSP
        let bytesToRead = Int(stackSize)

        // Sanity check: don't try to read more than 64KB (max stack size)
        guard bytesToRead <= 65536 else {
            DebugLog.error("Stack size too large: \(bytesToRead) bytes (SP may be corrupted)", category: "StackView")
            stackData = []
            localMemoryData = []
            return
        }

        DebugLog.log("Fetching actual stack contents: 0x\(String(format: "%08X", startAddress)) to 0x\(String(format: "%08X", initialSP - 1)) (\(bytesToRead) bytes)", category: "StackView")

        // Fetch memory data into local state
        do {
            localMemoryData = try await viewModel.fetchMemory(at: startAddress, length: bytesToRead)
            DebugLog.success("Fetched \(localMemoryData.count) bytes of actual stack memory", category: "StackView")
        } catch {
            DebugLog.error("Failed to fetch stack memory: \(error.localizedDescription)", category: "StackView")
            stackData = []
            localMemoryData = []
            return
        }

        // Build stack entries (word-aligned)
        var data: [StackEntry] = []
        let wordCount = localMemoryData.count / 4

        for i in 0 ..< wordCount {
            let wordOffset = i * 4
            let address = startAddress + UInt32(wordOffset)
            let offset = Int(address) - Int(currentSP)

            // Read 4 bytes as little-endian UInt32
            let bytes = localMemoryData[wordOffset ..< wordOffset + 4]
            let value = bytes.enumerated().reduce(UInt32(0)) { result, item in
                result | (UInt32(item.element) << (item.offset * 8))
            }

            let annotation = detectAnnotation(value: value, offset: offset)
            data.append(StackEntry(offset: offset, address: address, value: value, annotation: annotation))
        }

        stackData = data
        loadCount += 1
        DebugLog.success("Created \(stackData.count) stack entries (load #\(loadCount))", category: "StackView")
    }

    private func detectAnnotation(value: UInt32, offset: Int) -> String {
        // Detect likely saved registers

        // Check if it's a code address (in typical code range)
        if (0x8000 ... 0x10000).contains(value) {
            return "← Code?"
        }

        // Check if it's a stack address
        if abs(Int(value) - Int(viewModel.registers.sp)) < 4096 {
            return "← Stack?"
        }

        // Check if it matches a register value
        let registerValues: [(String, UInt32)] = [
            ("R0", viewModel.registers.r0),
            ("R1", viewModel.registers.r1),
            ("R2", viewModel.registers.r2),
            ("R3", viewModel.registers.r3),
            ("R4", viewModel.registers.r4),
            ("R5", viewModel.registers.r5),
            ("R6", viewModel.registers.r6),
            ("R7", viewModel.registers.r7),
            ("R8", viewModel.registers.r8),
            ("R9", viewModel.registers.r9),
            ("R10", viewModel.registers.r10),
            ("R11", viewModel.registers.r11),
            ("R12", viewModel.registers.r12),
        ]

        for (name, regValue) in registerValues {
            if value == regValue && regValue != 0 {
                return "← \(name)?"
            }
        }

        if value == viewModel.registers.lr && value != 0 {
            return "← LR"
        }

        return ""
    }
}

struct StackRowView: View {
    let offset: Int
    let address: UInt32
    let value: UInt32
    let annotation: String
    let isCurrent: Bool

    private var asciiRepresentation: String {
        let bytes = [
            UInt8(value & 0xFF),
            UInt8((value >> 8) & 0xFF),
            UInt8((value >> 16) & 0xFF),
            UInt8((value >> 24) & 0xFF),
        ]

        return bytes.map { byte in
            (32 ... 126).contains(byte) ? String(UnicodeScalar(byte)) : "."
        }.joined()
    }

    var body: some View {
        HStack(spacing: 4) {
            // SP offset
            Text(String(format: "SP%+d", offset))
                .foregroundColor(.secondary)
                .frame(width: 60, alignment: .leading)

            if isCurrent {
                Text("→")
                    .foregroundColor(.accentColor)
                    .fontWeight(.bold)
            } else {
                Text(" ")
            }

            Text("|")
                .foregroundColor(.secondary)

            // Address
            Text(String(format: "0x%08X", address))
                .foregroundColor(.secondary)
                .frame(width: 90, alignment: .leading)

            Text("|")
                .foregroundColor(.secondary)

            // Value
            Text(String(format: "0x%08X", value))
                .frame(width: 90, alignment: .leading)

            Text("|")
                .foregroundColor(.secondary)

            // ASCII
            Text(asciiRepresentation)
                .frame(width: 40, alignment: .leading)

            // Annotation
            if !annotation.isEmpty {
                Text(annotation)
                    .foregroundColor(.orange)
            }

            Spacer()
        }
        .padding(.vertical, 2)
        .background(isCurrent ? Color.accentColor.opacity(0.2) : Color.clear)
        .cornerRadius(2)
    }
}
