package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Stack segments use exclusive upper bounds: [Start, Start+Size)
// For a stack at 0x00040000 with size 0x10000 (64KB):
// - Valid addresses: [0x00040000..0x0004FFFF]
// - First invalid address: 0x00050000

func TestCPU_SetSP_ValidRange(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack start (minimum)", vm.StackSegmentStart},
		{"stack middle", vm.StackSegmentStart + vm.StackSegmentSize/2},
		{"stack end minus 4 (last valid word)", vm.StackSegmentStart + vm.StackSegmentSize - 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			assert.NoError(t, err, "Valid SP value should not error")
			assert.Equal(t, tt.value, cpu.GetSP(), "SP should be set to requested value")
		})
	}
}

func TestCPU_SetSP_Underflow(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"one below minimum", vm.StackSegmentStart - 1},
		{"far below minimum", vm.StackSegmentStart - 0x1000},
		{"zero address", 0x00000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			require.Error(t, err, "SP below stack segment should error")
			assert.Contains(t, err.Error(), "stack underflow", "Error should mention underflow")
		})
	}
}

func TestCPU_SetSP_Overflow(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack end (one past)", vm.StackSegmentStart + vm.StackSegmentSize},
		{"far above maximum", vm.StackSegmentStart + vm.StackSegmentSize + 0x1000},
		{"max address", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			require.Error(t, err, "SP above stack segment should error")
			assert.Contains(t, err.Error(), "stack overflow", "Error should mention overflow")
		})
	}
}
