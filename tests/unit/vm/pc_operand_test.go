package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Instructions run at vm.CodeSegmentStart. Reading R15 gives the instruction address
// plus 8, or plus 12 when the instruction also shifts by a register. STM of R15
// stores the address plus 12.
const (
	pcPlus8  = vm.CodeSegmentStart + 8
	pcPlus12 = vm.CodeSegmentStart + 12
)

func TestPCAsOperand(t *testing.T) {
	tests := []struct {
		name   string
		opcode uint32
		setup  func(t *testing.T, v *vm.VM)
		want   uint32 // R0
	}{
		{"MOV R0, PC", 0xE1A0000F, nil, pcPlus8},
		{"ADD R0, PC, #4", 0xE28F0004, nil, pcPlus8 + 4},
		{"ADD R0, PC, R1", 0xE08F0001, func(t *testing.T, v *vm.VM) { v.CPU.R[1] = 1 }, pcPlus8 + 1},
		{"ADD R0, R1, PC, LSL #0", 0xE081000F, func(t *testing.T, v *vm.VM) { v.CPU.R[1] = 1 }, pcPlus8 + 1},
		{"MOV R0, PC, LSL R2 (register shift)", 0xE1A0021F, func(t *testing.T, v *vm.VM) { v.CPU.R[2] = 0 }, pcPlus12},
		{"ADD R0, PC, R1, LSL R2 (Rn is PC)", 0xE08F0211, func(t *testing.T, v *vm.VM) { v.CPU.R[1], v.CPU.R[2] = 1, 0 }, pcPlus12 + 1},
		{"LDR R0, [PC, #-8] reads itself", 0xE51F0008, nil, 0xE51F0008},
		{"LDR R0, [PC]", 0xE59F0000, func(t *testing.T, v *vm.VM) { mustWriteWord(t, v, pcPlus8, 0xCAFE) }, 0xCAFE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			if tt.setup != nil {
				tt.setup(t, v)
			}
			stepOne(t, v, tt.opcode)
			if v.CPU.R[0] != tt.want {
				t.Errorf("R0 = 0x%08X, want 0x%08X", v.CPU.R[0], tt.want)
			}
			if v.CPU.PC != vm.CodeSegmentStart+4 {
				t.Errorf("PC = 0x%08X, want next instruction", v.CPU.PC)
			}
		})
	}
}

func TestStorePC(t *testing.T) {
	const buf = vm.DataSegmentStart
	t.Run("STMIA with PC", func(t *testing.T) {
		v := vm.NewVM()
		v.CPU.R[1] = buf
		stepOne(t, v, 0xE8818000) // STMIA R1, {PC}
		if got, _ := v.Memory.ReadWord(buf); got != pcPlus12 {
			t.Errorf("stored 0x%08X, want PC+12 0x%08X", got, uint32(pcPlus12))
		}
	})
}

func TestWritePC(t *testing.T) {
	const target = vm.CodeSegmentStart + 0x100
	tests := []struct {
		name     string
		opcode   uint32
		setup    func(t *testing.T, v *vm.VM)
		wantPC   uint32
		wantCPSR *vm.CPSR
	}{
		{"MOV PC, R1", 0xE1A0F001, func(t *testing.T, v *vm.VM) { v.CPU.R[1] = target }, target, nil},
		{"ADD PC, PC, #4", 0xE28FF004, nil, pcPlus8 + 4, nil},
		{"MOV PC, LR", 0xE1A0F00E, func(t *testing.T, v *vm.VM) { v.CPU.R[vm.LR] = target }, target, nil},
		{"LDR PC, [R1]", 0xE591F000, func(t *testing.T, v *vm.VM) {
			v.CPU.R[1] = vm.DataSegmentStart
			mustWriteWord(t, v, vm.DataSegmentStart, target)
		}, target, nil},
		{"LDMIA R1, {PC}", 0xE8918000, func(t *testing.T, v *vm.VM) {
			v.CPU.R[1] = vm.DataSegmentStart
			mustWriteWord(t, v, vm.DataSegmentStart, target)
		}, target, nil},
		// With S set and PC as destination, CPSR is restored from SPSR (exception return).
		{"MOVS PC, LR restores CPSR", 0xE1B0F00E, func(t *testing.T, v *vm.VM) {
			v.CPU.R[vm.LR] = target
			v.CPU.CPSR = vm.CPSR{N: true, C: true}
			v.CPU.SPSR = vm.CPSR{Z: true, V: true}
		}, target, &vm.CPSR{Z: true, V: true}},
		{"SUBS PC, LR, #4 restores CPSR", 0xE25EF004, func(t *testing.T, v *vm.VM) {
			v.CPU.R[vm.LR] = target + 4
			v.CPU.SPSR = vm.CPSR{N: true, V: true}
		}, target, &vm.CPSR{N: true, V: true}},
		{"MOVNE PC skipped", 0x11A0F001, func(t *testing.T, v *vm.VM) {
			v.CPU.R[1] = target
			v.CPU.CPSR.Z = true
		}, vm.CodeSegmentStart + 4, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			if tt.setup != nil {
				tt.setup(t, v)
			}
			stepOne(t, v, tt.opcode)
			if v.CPU.PC != tt.wantPC {
				t.Errorf("PC = 0x%08X, want 0x%08X", v.CPU.PC, tt.wantPC)
			}
			if tt.wantCPSR != nil && v.CPU.CPSR != *tt.wantCPSR {
				t.Errorf("CPSR = %+v, want %+v", v.CPU.CPSR, *tt.wantCPSR)
			}
		})
	}
}

func TestBranchExchangeToPC(t *testing.T) {
	v := vm.NewVM()
	stepOne(t, v, 0xE12FFF1F) // BX PC
	if v.CPU.PC != pcPlus8 {
		t.Errorf("PC = 0x%08X, want 0x%08X", v.CPU.PC, uint32(pcPlus8))
	}
}
