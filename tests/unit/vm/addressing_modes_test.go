package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ================================================================================
// Data Processing Addressing Modes
// ================================================================================

// Mode 1: Immediate operand
func TestAddressing_DataProcessing_Immediate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, #42
	opcode := uint32(0xE3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("Expected R0=42, got %d", v.CPU.R[0])
	}
}

func TestAddressing_DataProcessing_ImmediateRotated(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, #0xFF000000 (rotated immediate)
	// rotation=4 (right by 8), immed=0xFF
	opcode := uint32(0xE3A004FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF000000 {
		t.Errorf("Expected R0=0xFF000000, got 0x%X", v.CPU.R[0])
	}
}

// Mode 2: Register operand
func TestAddressing_DataProcessing_Register(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 123
	v.CPU.PC = 0x8000

	// MOV R0, R1
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 123 {
		t.Errorf("Expected R0=123, got %d", v.CPU.R[0])
	}
}

// Mode 3: Register with LSL immediate
func TestAddressing_DataProcessing_LSL_Immediate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #2 (5 << 2 = 20)
	opcode := uint32(0xE1A00101)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 20 {
		t.Errorf("Expected R0=20, got %d", v.CPU.R[0])
	}
}

// Mode 4: Register with LSR immediate
func TestAddressing_DataProcessing_LSR_Immediate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 20
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR #2 (20 >> 2 = 5)
	opcode := uint32(0xE1A00121)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 5 {
		t.Errorf("Expected R0=5, got %d", v.CPU.R[0])
	}
}

// Mode 5: Register with ASR immediate
func TestAddressing_DataProcessing_ASR_Immediate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFC // -4
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR #1 (-4 >> 1 = -2)
	opcode := uint32(0xE1A000C1) // Fixed: shift amount in bits 11-7 = 1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFE { // -2
		t.Errorf("Expected R0=0xFFFFFFFE, got 0x%X", v.CPU.R[0])
	}
}

// Mode 6: Register with ROR immediate
func TestAddressing_DataProcessing_ROR_Immediate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000001
	v.CPU.PC = 0x8000

	// MOV R0, R1, ROR #1
	opcode := uint32(0xE1A000E1) // Fixed: shift amount in bits 11-7 = 1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xC0000000 {
		t.Errorf("Expected R0=0xC0000000, got 0x%X", v.CPU.R[0])
	}
}

// Mode 7: Register shift by register (LSL)
func TestAddressing_DataProcessing_LSL_Register(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 3
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (5 << 3 = 40)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 40 {
		t.Errorf("Expected R0=40, got %d", v.CPU.R[0])
	}
}

// Mode 8: Register shift by register (LSR)
func TestAddressing_DataProcessing_LSR_Register(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 40
	v.CPU.R[2] = 3
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR R2 (40 >> 3 = 5)
	opcode := uint32(0xE1A00231)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 5 {
		t.Errorf("Expected R0=5, got %d", v.CPU.R[0])
	}
}

// Mode 9: Register shift by register (ASR)
func TestAddressing_DataProcessing_ASR_Register(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFF0 // -16
	v.CPU.R[2] = 2
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR R2 (-16 >> 2 = -4)
	opcode := uint32(0xE1A00251)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFC { // -4
		t.Errorf("Expected R0=0xFFFFFFFC, got 0x%X", v.CPU.R[0])
	}
}

// Mode 10: Register shift by register (ROR)
func TestAddressing_DataProcessing_ROR_Register(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// MOV R0, R1, ROR R2
	opcode := uint32(0xE1A00271)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("Expected R0=0x80000000, got 0x%X", v.CPU.R[0])
	}
}

// Mode 11: RRX (Rotate Right Extended)
func TestAddressing_DataProcessing_RRX(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// MOV R0, R1, RRX
	opcode := uint32(0xE1A00061) // RRX is ROR #0
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("Expected R0=0x80000000 (carry inserted), got 0x%X", v.CPU.R[0])
	}
}

// ================================================================================
// Memory Addressing Modes
// ================================================================================

// Mode 1: Immediate offset
func TestAddressing_Memory_ImmediateOffset(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0xDEADBEEF)

	// LDR R0, [R1, #8]
	opcode := uint32(0xE5910008)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xDEADBEEF {
		t.Errorf("Expected R0=0xDEADBEEF, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20000 {
		t.Error("R1 should not be modified with offset addressing")
	}
}

// Mode 2: Register offset
func TestAddressing_Memory_RegisterOffset(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0xCAFEBABE)

	// LDR R0, [R1, R2]
	opcode := uint32(0xE7910002)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xCAFEBABE {
		t.Errorf("Expected R0=0xCAFEBABE, got 0x%X", v.CPU.R[0])
	}
}

// Mode 3: Scaled register offset (register with LSL)
func TestAddressing_Memory_ScaledRegisterOffset_LSL(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 2 // Will be shifted left by 2 (2 << 2 = 8)
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0x12345678)

	// LDR R0, [R1, R2, LSL #2]
	opcode := uint32(0xE7910102)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("Expected R0=0x12345678, got 0x%X", v.CPU.R[0])
	}
}

// Mode 4: Pre-indexed
func TestAddressing_Memory_PreIndexed(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0xABCDEF01)

	// LDR R0, [R1, #8]! (pre-indexed with writeback)
	opcode := uint32(0xE5B10008)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xABCDEF01 {
		t.Errorf("Expected R0=0xABCDEF01, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("R1 should be updated to 0x20008, got 0x%X", v.CPU.R[1])
	}
}

// Mode 5: Post-indexed
func TestAddressing_Memory_PostIndexed(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11223344)

	// LDR R0, [R1], #8 (post-indexed)
	opcode := uint32(0xE4910008)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x11223344 {
		t.Errorf("Expected R0=0x11223344, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("R1 should be updated to 0x20008 after load, got 0x%X", v.CPU.R[1])
	}
}

// Mode 6: Negative offset
func TestAddressing_Memory_NegativeOffset(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20008
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x55667788)

	// LDR R0, [R1, #-8]
	opcode := uint32(0xE5110008)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x55667788 {
		t.Errorf("Expected R0=0x55667788, got 0x%X", v.CPU.R[0])
	}
}

// ================================================================================
// Load/Store Multiple Addressing Modes
// ================================================================================

// Mode 1: IA (Increment After)
func TestAddressing_LDM_IA(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11111111)
	v.Memory.WriteWord(0x20004, 0x22222222)
	v.Memory.WriteWord(0x20008, 0x33333333)

	// LDMIA R0, {R1, R2, R3}
	opcode := uint32(0xE890000E) // bits 1-3 set for R1-R3
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x11111111 {
		t.Errorf("Expected R1=0x11111111, got 0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x22222222 {
		t.Errorf("Expected R2=0x22222222, got 0x%X", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0x33333333 {
		t.Errorf("Expected R3=0x33333333, got 0x%X", v.CPU.R[3])
	}
}

// Mode 2: IB (Increment Before)
func TestAddressing_LDM_IB(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0xAAAAAAAA)
	v.Memory.WriteWord(0x20008, 0xBBBBBBBB)

	// LDMIB R0, {R1, R2}
	opcode := uint32(0xE9900006)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0xAAAAAAAA {
		t.Errorf("Expected R1=0xAAAAAAAA, got 0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0xBBBBBBBB {
		t.Errorf("Expected R2=0xBBBBBBBB, got 0x%X", v.CPU.R[2])
	}
}

// Mode 3: DA (Decrement After)
func TestAddressing_LDM_DA(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x20008
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0xCCCCCCCC)
	v.Memory.WriteWord(0x20004, 0xDDDDDDDD)
	v.Memory.WriteWord(0x20008, 0xEEEEEEEE)

	// LDMDA R0, {R1, R2, R3}
	opcode := uint32(0xE810000E)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0xCCCCCCCC {
		t.Errorf("Expected R1=0xCCCCCCCC, got 0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0xDDDDDDDD {
		t.Errorf("Expected R2=0xDDDDDDDD, got 0x%X", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0xEEEEEEEE {
		t.Errorf("Expected R3=0xEEEEEEEE, got 0x%X", v.CPU.R[3])
	}
}

// Mode 4: DB (Decrement Before)
func TestAddressing_LDM_DB(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x20008
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x12341234)
	v.Memory.WriteWord(0x20004, 0x56785678)

	// LDMDB R0, {R1, R2}
	opcode := uint32(0xE9100006)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x12341234 {
		t.Errorf("Expected R1=0x12341234, got 0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x56785678 {
		t.Errorf("Expected R2=0x56785678, got 0x%X", v.CPU.R[2])
	}
}

// ================================================================================
// Stack Addressing Modes
// ================================================================================

// Full Descending Stack (LDMFD/STMFD)
func TestAddressing_Stack_FD(t *testing.T) {
	v := vm.NewVM()
	initialSP := uint32(vm.StackSegmentStart + 0x3000) // 0x00043000
	v.CPU.R[13] = initialSP                            // SP
	v.CPU.R[1] = 0xAAAA
	v.CPU.R[2] = 0xBBBB
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// STMFD SP!, {R1, R2} (push)
	opcode := uint32(0xE92D0006)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	expectedSPAfterPush := initialSP - 8
	if v.CPU.R[13] != expectedSPAfterPush {
		t.Errorf("SP should be 0x%X after push, got 0x%X", expectedSPAfterPush, v.CPU.R[13])
	}

	// Clear R1, R2
	v.CPU.R[1] = 0
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8004

	// LDMFD SP!, {R1, R2} (pop)
	opcode = uint32(0xE8BD0006)
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	if v.CPU.R[1] != 0xAAAA {
		t.Errorf("Expected R1=0xAAAA after pop, got 0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0xBBBB {
		t.Errorf("Expected R2=0xBBBB after pop, got 0x%X", v.CPU.R[2])
	}
	if v.CPU.R[13] != initialSP {
		t.Errorf("SP should be restored to 0x30000, got 0x%X", v.CPU.R[13])
	}
}

// ================================================================================
// Complex Addressing Mode Tests
// ================================================================================

func TestAddressing_Complex_ShiftedRegisterInDataProcessing(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 5
	v.CPU.R[2] = 2
	v.CPU.PC = 0x8000

	// ADD R3, R0, R1, LSL R2 (10 + (5 << 2) = 10 + 20 = 30)
	opcode := uint32(0xE0803211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[3] != 30 {
		t.Errorf("Expected R3=30, got %d", v.CPU.R[3])
	}
}

func TestAddressing_Complex_NegativePreIndexed(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20010
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0xFEEDFACE)

	// LDR R0, [R1, #-16]! (pre-indexed with negative offset)
	opcode := uint32(0xE5310010)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFEEDFACE {
		t.Errorf("Expected R0=0xFEEDFACE, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20000 {
		t.Errorf("R1 should be 0x20000, got 0x%X", v.CPU.R[1])
	}
}

func TestAddressing_Complex_PostIndexedNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20010
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20010, 0x99887766)

	// LDR R0, [R1], #-16 (post-indexed with negative offset)
	opcode := uint32(0xE4110010)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x99887766 {
		t.Errorf("Expected R0=0x99887766, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20000 {
		t.Errorf("R1 should be 0x20000 after post-index, got 0x%X", v.CPU.R[1])
	}
}

// ================================================================================
// Boundary and Edge Cases
// ================================================================================

func TestAddressing_ZeroShift(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #0 (no shift)
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("Expected R0=42, got %d", v.CPU.R[0])
	}
}

func TestAddressing_MaxShift(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #31
	opcode := uint32(0xE1A00F81)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("Expected R0=0x80000000, got 0x%X", v.CPU.R[0])
	}
}

func TestAddressing_ZeroOffset(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11111111)

	// LDR R0, [R1, #0] (zero offset)
	opcode := uint32(0xE5910000)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x11111111 {
		t.Errorf("Expected R0=0x11111111, got 0x%X", v.CPU.R[0])
	}
}
