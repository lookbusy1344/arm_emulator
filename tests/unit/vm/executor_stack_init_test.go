package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVM_InitializeStack_ValidAddress(t *testing.T) {
	v := vm.NewVM()

	// Use stack end (empty stack position, per ARM convention)
	// Stack segment: [0x00040000..0x00050000], inclusive upper bound
	validStackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	err := v.InitializeStack(validStackTop)

	assert.NoError(t, err)
	assert.Equal(t, validStackTop, v.CPU.GetSP())
	assert.Equal(t, validStackTop, v.StackTop)
}

func TestVM_InitializeStack_InvalidAddress(t *testing.T) {
	v := vm.NewVM()

	tests := []struct {
		name      string
		stackTop  uint32
		expectErr string
	}{
		{"underflow", uint32(vm.StackSegmentStart - 1), "stack underflow"},
		{"overflow (one above end)", uint32(vm.StackSegmentStart + vm.StackSegmentSize + 1), "stack overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.InitializeStack(tt.stackTop)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
			assert.Contains(t, err.Error(), "failed to initialize stack")
		})
	}
}
