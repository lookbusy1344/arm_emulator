package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestMRS_Basic(t *testing.T) {
	// MRS R0, CPSR - Read CPSR into R0
	v := vm.NewVM()
	// Set some flags
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = false
	v.CPU.PC = 0x8000

	// MRS R0, CPSR (E10F0000)
	// Bits: cond=1110, 00010, PSR=0, 00, 1111, Rd=0000, 0000 0000 0000
	opcode := uint32(0xE10F0000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected CPSR value: N=1 (bit 31), Z=0 (bit 30), C=1 (bit 29), V=0 (bit 28)
	// 0xA0000000
	expectedCPSR := uint32(0xA0000000)
	if v.CPU.R[0] != expectedCPSR {
		t.Errorf("expected R0=0x%08X, got R0=0x%08X", expectedCPSR, v.CPU.R[0])
	}
}

func TestMRS_AllFlagsSet(t *testing.T) {
	// MRS R1, CPSR - Read CPSR with all flags set
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// MRS R1, CPSR (E10F1000)
	opcode := uint32(0xE10F1000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// All flags set: 0xF0000000
	expectedCPSR := uint32(0xF0000000)
	if v.CPU.R[1] != expectedCPSR {
		t.Errorf("expected R1=0x%08X, got R1=0x%08X", expectedCPSR, v.CPU.R[1])
	}
}

func TestMRS_NoFlagsSet(t *testing.T) {
	// MRS R2, CPSR - Read CPSR with no flags set
	v := vm.NewVM()
	// All flags default to false
	v.CPU.PC = 0x8000

	// MRS R2, CPSR (E10F2000)
	opcode := uint32(0xE10F2000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// No flags set: 0x00000000
	if v.CPU.R[2] != 0 {
		t.Errorf("expected R2=0, got R2=0x%08X", v.CPU.R[2])
	}
}

func TestMRS_InvalidDestination(t *testing.T) {
	// MRS R15, CPSR - R15 (PC) cannot be used as destination
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MRS R15, CPSR (E10FF000)
	opcode := uint32(0xE10FF000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when using R15 as destination")
	}
}

func TestMSR_Register(t *testing.T) {
	// MSR CPSR, R0 - Write R0 to CPSR
	v := vm.NewVM()
	// Set R0 to have N and C flags set
	v.CPU.R[0] = 0xA0000000 // N=1, Z=0, C=1, V=0
	v.CPU.PC = 0x8000

	// MSR CPSR_f, R0 (E129F000)
	// Bits: cond=1110, 00010, PSR=0, 10, mask=1001, 1111, 0000 0000, Rm=0000
	opcode := uint32(0xE129F000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Check flags were updated
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set")
	}
	if v.CPU.CPSR.V {
		t.Error("expected V flag to be clear")
	}
}

func TestMSR_AllFlags(t *testing.T) {
	// MSR CPSR, R1 - Write all flags from R1
	v := vm.NewVM()
	v.CPU.R[1] = 0xF0000000 // All flags set
	v.CPU.PC = 0x8000

	// MSR CPSR_f, R1 (E129F001)
	opcode := uint32(0xE129F001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// All flags should be set
	if !v.CPU.CPSR.N || !v.CPU.CPSR.Z || !v.CPU.CPSR.C || !v.CPU.CPSR.V {
		t.Errorf("expected all flags to be set, got N=%v Z=%v C=%v V=%v",
			v.CPU.CPSR.N, v.CPU.CPSR.Z, v.CPU.CPSR.C, v.CPU.CPSR.V)
	}
}

func TestMSR_ClearFlags(t *testing.T) {
	// MSR CPSR, R2 - Clear all flags
	v := vm.NewVM()
	// Set all flags initially
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.R[2] = 0x00000000 // Clear all flags
	v.CPU.PC = 0x8000

	// MSR CPSR_f, R2 (E129F002)
	opcode := uint32(0xE129F002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// All flags should be clear
	if v.CPU.CPSR.N || v.CPU.CPSR.Z || v.CPU.CPSR.C || v.CPU.CPSR.V {
		t.Errorf("expected all flags to be clear, got N=%v Z=%v C=%v V=%v",
			v.CPU.CPSR.N, v.CPU.CPSR.Z, v.CPU.CPSR.C, v.CPU.CPSR.V)
	}
}

func TestMSR_InvalidSource(t *testing.T) {
	// MSR CPSR, R15 - R15 (PC) cannot be used as source
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MSR CPSR_f, R15 (E129F00F)
	opcode := uint32(0xE129F00F)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when using R15 as source")
	}
}

func TestMRS_MSR_RoundTrip(t *testing.T) {
	// Test reading and writing CPSR
	v := vm.NewVM()
	// Set initial flags
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// MRS R0, CPSR (E10F0000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xE10F0000)
	v.Step()

	// Save the value
	savedCPSR := v.CPU.R[0]

	// Clear all flags
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = false

	// MSR CPSR, R0 (E129F000)
	v.Memory.WriteWord(0x8004, 0xE129F000)
	v.Step()

	// Flags should be restored
	if v.CPU.CPSR.N != false || v.CPU.CPSR.Z != true || v.CPU.CPSR.C != false || v.CPU.CPSR.V != true {
		t.Errorf("expected flags N=false Z=true C=false V=true, got N=%v Z=%v C=%v V=%v",
			v.CPU.CPSR.N, v.CPU.CPSR.Z, v.CPU.CPSR.C, v.CPU.CPSR.V)
	}

	// Read again and compare
	v.Memory.WriteWord(0x8008, 0xE10F1000) // MRS R1, CPSR
	v.Step()

	if v.CPU.R[1] != savedCPSR {
		t.Errorf("expected R1=0x%08X (saved CPSR), got R1=0x%08X", savedCPSR, v.CPU.R[1])
	}
}

func TestMSR_Immediate(t *testing.T) {
	// MSR CPSR_f, #0xF0000000 - Write immediate to CPSR (if supported)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MSR CPSR_f, #immediate
	// To get 0xF0000000: immediate=0x0F, rotate_field=2 (2*2=4, ROR 4 bits)
	// 0x0F ROR 4 = 0xF0000000
	// Bits: cond=1110, 00110, PSR=0, 10, mask=1001, 1111, rotate=0010, immediate=0000 1111
	// Pattern: E329 F20F
	opcode := uint32(0xE329F20F)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// All flags should be set
	if !v.CPU.CPSR.N || !v.CPU.CPSR.Z || !v.CPU.CPSR.C || !v.CPU.CPSR.V {
		t.Errorf("expected all flags to be set from immediate, got N=%v Z=%v C=%v V=%v",
			v.CPU.CPSR.N, v.CPU.CPSR.Z, v.CPU.CPSR.C, v.CPU.CPSR.V)
	}
}

func TestCPSR_ToUint32(t *testing.T) {
	// Test CPSR.ToUint32() conversion
	cpsr := vm.CPSR{N: true, Z: false, C: true, V: false}
	value := cpsr.ToUint32()

	expected := uint32(0xA0000000) // N=bit31, C=bit29
	if value != expected {
		t.Errorf("expected CPSR.ToUint32()=0x%08X, got 0x%08X", expected, value)
	}
}

func TestCPSR_FromUint32(t *testing.T) {
	// Test CPSR.FromUint32() conversion
	cpsr := vm.CPSR{}
	cpsr.FromUint32(0xF0000000) // All flags set

	if !cpsr.N || !cpsr.Z || !cpsr.C || !cpsr.V {
		t.Errorf("expected all flags set from 0xF0000000, got N=%v Z=%v C=%v V=%v",
			cpsr.N, cpsr.Z, cpsr.C, cpsr.V)
	}
}

func TestCPSR_RoundTripConversion(t *testing.T) {
	// Test that ToUint32 and FromUint32 are inverse operations
	original := vm.CPSR{N: false, Z: true, C: false, V: true}
	value := original.ToUint32()

	restored := vm.CPSR{}
	restored.FromUint32(value)

	if restored.N != original.N || restored.Z != original.Z ||
		restored.C != original.C || restored.V != original.V {
		t.Errorf("round trip failed: original=%+v, restored=%+v", original, restored)
	}
}
