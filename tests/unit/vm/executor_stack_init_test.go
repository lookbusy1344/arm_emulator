package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/stretchr/testify/assert"
)

func TestVM_InitializeStack_AllowsAnyAddress(t *testing.T) {
	v := vm.NewVM()

	tests := []struct {
		name     string
		stackTop uint32
	}{
		{"stack segment start", vm.StackSegmentStart},
		{"stack segment end", vm.StackSegmentStart + vm.StackSegmentSize},
		{"data segment (multi-stack)", vm.DataSegmentStart + 0x100},
		{"code segment (multi-stack)", vm.CodeSegmentStart + 0x200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.InitializeStack(tt.stackTop)
			assert.NoError(t, err, "InitializeStack should allow any address like real ARM hardware")
			assert.Equal(t, tt.stackTop, v.CPU.GetSP())
			assert.Equal(t, tt.stackTop, v.StackTop)
		})
	}
}
