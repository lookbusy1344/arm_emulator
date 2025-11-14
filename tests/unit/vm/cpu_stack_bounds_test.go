package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/stretchr/testify/assert"
)

// SetSP and SetSPWithTrace behave like real ARM hardware - they allow SP to be set to any value.
// This enables advanced use cases like cooperative multitasking with multiple stacks.
// Memory protection occurs when memory is accessed, not when SP is set.
// Stack overflow/underflow detection is provided by StackTrace monitoring when enabled.

func TestCPU_SetSP_AllowsAnyValue(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack segment start", vm.StackSegmentStart},
		{"stack segment middle", vm.StackSegmentStart + vm.StackSegmentSize/2},
		{"stack segment end", vm.StackSegmentStart + vm.StackSegmentSize},
		{"data segment (multi-stack use case)", vm.DataSegmentStart + 0x100},
		{"code segment (multi-stack use case)", vm.CodeSegmentStart + 0x200},
		{"low address", 0x00001000},
		{"high address", 0xFFFF0000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			assert.NoError(t, err, "SetSP should allow any value like real ARM hardware")
			assert.Equal(t, tt.value, cpu.GetSP(), "SP should be set to requested value")
		})
	}
}

func TestCPU_SetSPWithTrace_AllowsAnyValue(t *testing.T) {
	v := vm.NewVM()
	v.StackTrace = vm.NewStackTrace(nil, vm.StackSegmentStart+vm.StackSegmentSize, vm.StackSegmentStart)
	pc := uint32(0x00008000)

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack segment start", vm.StackSegmentStart},
		{"stack segment middle", vm.StackSegmentStart + vm.StackSegmentSize/2},
		{"stack segment end", vm.StackSegmentStart + vm.StackSegmentSize},
		{"data segment (multi-stack use case)", vm.DataSegmentStart + 0x100},
		{"code segment (multi-stack use case)", vm.CodeSegmentStart + 0x200},
		{"heap segment", vm.HeapSegmentStart + 0x300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.CPU.SetSPWithTrace(v, tt.value, pc)
			assert.NoError(t, err, "SetSPWithTrace should allow any value like real ARM hardware")
			assert.Equal(t, tt.value, v.CPU.GetSP(), "SP should be set to requested value")
		})
	}
}
