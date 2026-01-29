import SwiftUI

struct RegistersView: View {
    let registers: RegisterState
    let registerHighlights: [String: UUID]

    init(registers: RegisterState, registerHighlights: [String: UUID] = [:]) {
        self.registers = registers
        self.registerHighlights = registerHighlights
    }

    /// Helper struct to hold register data for column layout
    private struct RegisterData: Identifiable {
        let id = UUID()
        let name: String
        let value: UInt32
        let highlightID: UUID?
    }

    /// Split registers into vertical columns based on available width
    private func verticalColumns(for width: CGFloat) -> [[RegisterData]] {
        // Create array of all general-purpose registers (R0-R12)
        let allRegisters = [
            RegisterData(name: "R0", value: registers.r0, highlightID: registerHighlights["R0"]),
            RegisterData(name: "R1", value: registers.r1, highlightID: registerHighlights["R1"]),
            RegisterData(name: "R2", value: registers.r2, highlightID: registerHighlights["R2"]),
            RegisterData(name: "R3", value: registers.r3, highlightID: registerHighlights["R3"]),
            RegisterData(name: "R4", value: registers.r4, highlightID: registerHighlights["R4"]),
            RegisterData(name: "R5", value: registers.r5, highlightID: registerHighlights["R5"]),
            RegisterData(name: "R6", value: registers.r6, highlightID: registerHighlights["R6"]),
            RegisterData(name: "R7", value: registers.r7, highlightID: registerHighlights["R7"]),
            RegisterData(name: "R8", value: registers.r8, highlightID: registerHighlights["R8"]),
            RegisterData(name: "R9", value: registers.r9, highlightID: registerHighlights["R9"]),
            RegisterData(name: "R10", value: registers.r10, highlightID: registerHighlights["R10"]),
            RegisterData(name: "R11", value: registers.r11, highlightID: registerHighlights["R11"]),
            RegisterData(name: "R12", value: registers.r12, highlightID: registerHighlights["R12"]),
        ]

        // Determine number of columns based on width
        let columnCount = if width < 500 {
            1
        } else if width < 700 {
            2
        } else {
            3
        }

        // Calculate registers per column (distribute evenly, with extra in first columns)
        let registersPerColumn = (allRegisters.count + columnCount - 1) / columnCount

        // Split into columns vertically
        var columns: [[RegisterData]] = []
        for columnIndex in 0 ..< columnCount {
            let start = columnIndex * registersPerColumn
            let end = min(start + registersPerColumn, allRegisters.count)
            if start < allRegisters.count {
                columns.append(Array(allRegisters[start ..< end]))
            }
        }

        return columns
    }

    /// Create special registers column array
    private func specialRegistersColumns(for width: CGFloat) -> [[RegisterData]] {
        let specialRegisters = [
            RegisterData(name: "SP", value: registers.sp, highlightID: registerHighlights["SP"]),
            RegisterData(name: "LR", value: registers.lr, highlightID: registerHighlights["LR"]),
            RegisterData(name: "PC", value: registers.pc, highlightID: registerHighlights["PC"]),
        ]

        // Determine number of columns (same logic as general registers)
        let columnCount = if width < 500 {
            1
        } else if width < 700 {
            2
        } else {
            3
        }

        // For only 3 special registers, use fewer columns if needed
        let actualColumnCount = min(columnCount, specialRegisters.count)
        let registersPerColumn = (specialRegisters.count + actualColumnCount - 1) / actualColumnCount

        var columns: [[RegisterData]] = []
        for columnIndex in 0 ..< actualColumnCount {
            let start = columnIndex * registersPerColumn
            let end = min(start + registersPerColumn, specialRegisters.count)
            if start < specialRegisters.count {
                columns.append(Array(specialRegisters[start ..< end]))
            }
        }

        return columns
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
                        // General-purpose registers (R0-R12) in vertical columns
                        let columns = verticalColumns(for: geometry.size.width)
                        HStack(alignment: .top, spacing: 8) {
                            ForEach(Array(columns.enumerated()), id: \.offset) { _, column in
                                VStack(alignment: .leading, spacing: 4) {
                                    ForEach(column) { regData in
                                        RegisterRow(
                                            name: regData.name,
                                            value: regData.value,
                                            highlightID: regData.highlightID,
                                        )
                                    }
                                }
                                .frame(maxWidth: .infinity, alignment: .leading)
                            }
                        }
                        .padding(.horizontal, 8)

                        Divider()
                            .padding(.vertical, 4)

                        // Special registers in vertical columns
                        let specialColumns = specialRegistersColumns(for: geometry.size.width)
                        HStack(alignment: .top, spacing: 8) {
                            ForEach(Array(specialColumns.enumerated()), id: \.offset) { _, column in
                                VStack(alignment: .leading, spacing: 4) {
                                    ForEach(column) { regData in
                                        RegisterRow(
                                            name: regData.name,
                                            value: regData.value,
                                            highlightID: regData.highlightID,
                                        )
                                    }
                                }
                                .frame(maxWidth: .infinity, alignment: .leading)
                            }
                        }
                        .padding(.horizontal, 8)
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
                cpsr: CPSRFlags(n: false, z: false, c: true, v: false),
            ),
            registerHighlights: ["R0": UUID(), "PC": UUID()], // Show R0 and PC highlighted
        )
        .frame(width: 300, height: 500)
    }
}
