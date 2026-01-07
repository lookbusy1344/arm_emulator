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

    private let wordsToShow = 32 // Show ±128 bytes (±32 words)

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Stack (SP = \(String(format: "0x%08X", viewModel.registers.sp)))")
                    .font(.system(size: 11, weight: .semibold))
                    .padding(8)
                Spacer()
            }
            .background(Color(NSColor.controlBackgroundColor))

            Divider()

            // Stack display
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(stackData) { entry in
                        StackRowView(
                            offset: entry.offset,
                            address: entry.address,
                            value: entry.value,
                            annotation: entry.annotation,
                            isCurrent: entry.offset == 0
                        )
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(8)
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
        let sp = viewModel.registers.sp

        // Don't try to load stack if no program is loaded (SP is 0 or invalid)
        // Valid stack pointers are typically in high memory (0x00040000+)
        guard sp >= 0x1000 else {
            stackData = []
            return
        }

        let offset = UInt32(wordsToShow / 2 * 4)

        // Prevent arithmetic underflow when sp is too small
        let startAddress = sp >= offset ? sp - offset : 0

        await viewModel.loadMemory(at: startAddress, length: wordsToShow * 4)

        guard viewModel.memoryData.count >= wordsToShow * 4 else {
            stackData = []
            return
        }

        var data: [StackEntry] = []

        for i in 0 ..< wordsToShow {
            let wordOffset = i * 4
            let address = startAddress + UInt32(wordOffset)
            let offset = Int(address) - Int(sp)

            // Read 4 bytes as little-endian UInt32
            let bytes = viewModel.memoryData[wordOffset ..< wordOffset + 4]
            let value = bytes.enumerated().reduce(UInt32(0)) { result, item in
                result | (UInt32(item.element) << (item.offset * 8))
            }

            let annotation = detectAnnotation(value: value, offset: offset)
            data.append(StackEntry(offset: offset, address: address, value: value, annotation: annotation))
        }

        stackData = data
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
