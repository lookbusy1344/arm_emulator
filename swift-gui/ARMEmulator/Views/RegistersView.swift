import SwiftUI

struct RegistersView: View {
    let registers: RegisterState
    let changedRegisters: Set<String>

    init(registers: RegisterState, changedRegisters: Set<String> = []) {
        self.registers = registers
        self.changedRegisters = changedRegisters
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Registers")
                .font(.headline)
                .padding(.horizontal)
                .padding(.vertical, 8)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(NSColor.controlBackgroundColor))

            ScrollView {
                VStack(alignment: .leading, spacing: 4) {
                    RegisterRow(name: "R0", value: registers.r0, isChanged: changedRegisters.contains("R0"))
                    RegisterRow(name: "R1", value: registers.r1, isChanged: changedRegisters.contains("R1"))
                    RegisterRow(name: "R2", value: registers.r2, isChanged: changedRegisters.contains("R2"))
                    RegisterRow(name: "R3", value: registers.r3, isChanged: changedRegisters.contains("R3"))
                    RegisterRow(name: "R4", value: registers.r4, isChanged: changedRegisters.contains("R4"))
                    RegisterRow(name: "R5", value: registers.r5, isChanged: changedRegisters.contains("R5"))
                    RegisterRow(name: "R6", value: registers.r6, isChanged: changedRegisters.contains("R6"))
                    RegisterRow(name: "R7", value: registers.r7, isChanged: changedRegisters.contains("R7"))
                    RegisterRow(name: "R8", value: registers.r8, isChanged: changedRegisters.contains("R8"))
                    RegisterRow(name: "R9", value: registers.r9, isChanged: changedRegisters.contains("R9"))
                    RegisterRow(name: "R10", value: registers.r10, isChanged: changedRegisters.contains("R10"))
                    RegisterRow(name: "R11", value: registers.r11, isChanged: changedRegisters.contains("R11"))
                    RegisterRow(name: "R12", value: registers.r12, isChanged: changedRegisters.contains("R12"))

                    Divider()
                        .padding(.vertical, 4)

                    RegisterRow(name: "SP", value: registers.sp, isChanged: changedRegisters.contains("SP"))
                    RegisterRow(name: "LR", value: registers.lr, isChanged: changedRegisters.contains("LR"))
                    RegisterRow(name: "PC", value: registers.pc, isChanged: changedRegisters.contains("PC"))

                    Divider()
                        .padding(.vertical, 4)

                    HStack {
                        Text("CPSR:")
                            .font(.system(.body, design: .monospaced))
                            .fontWeight(.bold)
                            .frame(width: 60, alignment: .leading)

                        Text(registers.cpsr.displayString)
                            .font(.system(.body, design: .monospaced))
                            .foregroundColor(changedRegisters.contains("CPSR") ? .green : .primary)
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

struct RegisterRow: View {
    let name: String
    let value: UInt32
    let isChanged: Bool

    init(name: String, value: UInt32, isChanged: Bool = false) {
        self.name = name
        self.value = value
        self.isChanged = isChanged
    }

    var body: some View {
        HStack {
            Text("\(name):")
                .font(.system(.body, design: .monospaced))
                .fontWeight(.bold)
                .frame(width: 60, alignment: .leading)
                .foregroundColor(isChanged ? .green : .primary)

            Text(String(format: "0x%08X", value))
                .font(.system(.body, design: .monospaced))
                .foregroundColor(isChanged ? .green : .primary)

            Spacer()

            Text(String(value))
                .font(.system(.caption, design: .monospaced))
                .foregroundColor(isChanged ? .green : .secondary)
        }
        .padding(.horizontal)
        .padding(.vertical, 2)
    }
}

struct RegistersView_Previews: PreviewProvider {
    static var previews: some View {
        RegistersView(registers: RegisterState(
            r0: 0x0000_0042, r1: 0x0000_0001, r2: 0x0000_0002, r3: 0x0000_0003,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 0x0005_0000, lr: 0, pc: 0x0000_8004,
            cpsr: CPSRFlags(n: false, z: false, c: true, v: false)
        ))
        .frame(width: 300, height: 500)
    }
}
