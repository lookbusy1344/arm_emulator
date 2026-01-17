import SwiftUI

struct RegistersView: View {
    let registers: RegisterState
    let registerHighlights: [String: UUID]

    init(registers: RegisterState, registerHighlights: [String: UUID] = [:]) {
        self.registers = registers
        self.registerHighlights = registerHighlights
    }

    // Determine grid columns based on available width (1-3 columns)
    private func gridColumns(for width: CGFloat) -> [GridItem] {
        let columnCount: Int
        if width < 500 {
            columnCount = 1
        } else if width < 700 {
            columnCount = 2
        } else {
            columnCount = 3
        }
        return Array(repeating: GridItem(.flexible(), spacing: 4), count: columnCount)
    }

    var body: some View {
        GeometryReader { geometry in
            VStack(alignment: .leading, spacing: 0) {
                Text("Registers")
                    .font(.system(size: 11, weight: .semibold))
                    .padding(.horizontal)
                    .padding(.vertical, 8)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color(NSColor.controlBackgroundColor))

                ScrollView {
                    VStack(alignment: .leading, spacing: 4) {
                        // General-purpose registers (R0-R12) in responsive grid
                        LazyVGrid(columns: gridColumns(for: geometry.size.width), alignment: .leading, spacing: 4) {
                            RegisterRow(name: "R0", value: registers.r0, highlightID: registerHighlights["R0"])
                            RegisterRow(name: "R1", value: registers.r1, highlightID: registerHighlights["R1"])
                            RegisterRow(name: "R2", value: registers.r2, highlightID: registerHighlights["R2"])
                            RegisterRow(name: "R3", value: registers.r3, highlightID: registerHighlights["R3"])
                            RegisterRow(name: "R4", value: registers.r4, highlightID: registerHighlights["R4"])
                            RegisterRow(name: "R5", value: registers.r5, highlightID: registerHighlights["R5"])
                            RegisterRow(name: "R6", value: registers.r6, highlightID: registerHighlights["R6"])
                            RegisterRow(name: "R7", value: registers.r7, highlightID: registerHighlights["R7"])
                            RegisterRow(name: "R8", value: registers.r8, highlightID: registerHighlights["R8"])
                            RegisterRow(name: "R9", value: registers.r9, highlightID: registerHighlights["R9"])
                            RegisterRow(name: "R10", value: registers.r10, highlightID: registerHighlights["R10"])
                            RegisterRow(name: "R11", value: registers.r11, highlightID: registerHighlights["R11"])
                            RegisterRow(name: "R12", value: registers.r12, highlightID: registerHighlights["R12"])
                        }
                        .padding(.horizontal, 8)

                        Divider()
                            .padding(.vertical, 4)

                        // Special registers in responsive grid
                        LazyVGrid(columns: gridColumns(for: geometry.size.width), alignment: .leading, spacing: 4) {
                            RegisterRow(name: "SP", value: registers.sp, highlightID: registerHighlights["SP"])
                            RegisterRow(name: "LR", value: registers.lr, highlightID: registerHighlights["LR"])
                            RegisterRow(name: "PC", value: registers.pc, highlightID: registerHighlights["PC"])
                        }
                        .padding(.horizontal, 8)

                        Divider()
                            .padding(.vertical, 4)

                        HStack {
                            Text("CPSR:")
                                .font(.system(size: 10, design: .monospaced))
                                .fontWeight(.bold)
                                .frame(width: 60, alignment: .leading)

                            Text(registers.cpsr.displayString)
                                .font(.system(size: 10, design: .monospaced))
                                .foregroundColor(registerHighlights["CPSR"] != nil ? .green : .primary)
                        }
                        .padding(.horizontal)
                        .padding(.vertical, 2)
                    }
                    .padding(.vertical, 8)
                }
                .background(Color(NSColor.textBackgroundColor))
            }
        }
    }
}

struct RegisterRow: View {
    let name: String
    let value: UInt32
    let highlightID: UUID? // Changed from isChanged: Bool

    init(name: String, value: UInt32, highlightID: UUID? = nil) {
        self.name = name
        self.value = value
        self.highlightID = highlightID
    }

    var body: some View {
        HStack {
            Text("\(name):")
                .font(.system(size: 10, design: .monospaced))
                .fontWeight(.bold)
                .frame(width: 60, alignment: .leading)
                .foregroundColor(highlightID != nil ? .green : .primary)
                .animation(.easeOut(duration: 1.5), value: highlightID)

            Text(String(format: "0x%08X", value))
                .font(.system(size: 10, design: .monospaced))
                .foregroundColor(highlightID != nil ? .green : .primary)
                .animation(.easeOut(duration: 1.5), value: highlightID)

            Spacer()

            Text(String(value))
                .font(.system(size: 10, design: .monospaced))
                .foregroundColor(highlightID != nil ? .green : .secondary)
                .animation(.easeOut(duration: 1.5), value: highlightID)
        }
        .padding(.horizontal)
        .padding(.vertical, 2)
    }
}

struct RegistersView_Previews: PreviewProvider {
    static var previews: some View {
        RegistersView(
            registers: RegisterState(
                r0: 0x0000_0042, r1: 0x0000_0001, r2: 0x0000_0002, r3: 0x0000_0003,
                r4: 0, r5: 0, r6: 0, r7: 0,
                r8: 0, r9: 0, r10: 0, r11: 0,
                r12: 0, sp: 0x0005_0000, lr: 0, pc: 0x0000_8004,
                cpsr: CPSRFlags(n: false, z: false, c: true, v: false)
            ),
            registerHighlights: ["R0": UUID(), "PC": UUID()] // Show R0 and PC highlighted
        )
        .frame(width: 300, height: 500)
    }
}
