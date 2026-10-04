package vm_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

type interruptedReader struct{}

func (interruptedReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("stdin closed: %w", vm.ErrInputInterrupted)
}

func TestConsoleReadInterrupted(t *testing.T) {
	const (
		swiReadChar   = 0xEF000004
		swiReadString = 0xEF000005
		swiReadInt    = 0xEF000006
		entry         = 0x8000
		sentinelR0    = 0x10000
	)

	for _, tc := range []struct {
		name   string
		opcode uint32
	}{
		{"read char", swiReadChar},
		{"read string", swiReadString},
		{"read int", swiReadInt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := vm.NewVM()
			v.SetStdinReader(interruptedReader{})
			setupCodeWrite(v)
			if err := v.Memory.WriteWord(entry, tc.opcode); err != nil {
				t.Fatal(err)
			}
			v.CPU.PC = entry
			v.CPU.R[0] = sentinelR0
			v.CPU.R[1] = 16
			v.State = vm.StateRunning

			err := v.Step()
			if !errors.Is(err, vm.ErrInputInterrupted) {
				t.Fatalf("expected ErrInputInterrupted, got %v", err)
			}
			if v.CPU.PC != entry {
				t.Errorf("expected PC to stay on the SWI, got 0x%08X", v.CPU.PC)
			}
			if v.CPU.R[0] != sentinelR0 {
				t.Errorf("expected R0 unchanged, got 0x%08X", v.CPU.R[0])
			}
			if v.State != vm.StateRunning {
				t.Errorf("expected state restored to running, got %v", v.State)
			}
			if v.LastError != nil {
				t.Errorf("expected no LastError, got %v", v.LastError)
			}
		})
	}
}
