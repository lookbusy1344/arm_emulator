package vm_test

import (
	"errors"
	"fmt"
	"strings"
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

const swiRead = 0xEF000012

func TestReadFromStdinDescriptorUsesStdinReader(t *testing.T) {
	const (
		entry   = 0x8000
		bufAddr = 0x20000
		bufLen  = 16
	)
	v := vm.NewVM()
	v.SetStdinReader(strings.NewReader("abc"))
	setupDataWrite(v)
	v.CPU.PC = entry
	v.CPU.R[0] = vm.StdIn
	v.CPU.R[1] = bufAddr
	v.CPU.R[2] = bufLen
	stepSWI(t, v, swiRead)

	if v.CPU.R[0] != 3 {
		t.Fatalf("expected 3 bytes read, got 0x%08X", v.CPU.R[0])
	}
	got, err := v.Memory.GetBytes(bufAddr, 3)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "abc" {
		t.Errorf("expected %q in buffer, got %q", "abc", got)
	}
}

func TestReadFromStdinDescriptorInterrupted(t *testing.T) {
	const (
		entry   = 0x8000
		bufAddr = 0x20000
	)
	v := vm.NewVM()
	v.SetStdinReader(interruptedReader{})
	setupCodeWrite(v)
	if err := v.Memory.WriteWord(entry, swiRead); err != nil {
		t.Fatal(err)
	}
	v.CPU.PC = entry
	v.CPU.R[0] = vm.StdIn
	v.CPU.R[1] = bufAddr
	v.CPU.R[2] = 16
	v.State = vm.StateRunning

	if err := v.Step(); !errors.Is(err, vm.ErrInputInterrupted) {
		t.Fatalf("expected ErrInputInterrupted, got %v", err)
	}
	if v.CPU.PC != entry || v.CPU.R[0] != vm.StdIn {
		t.Errorf("expected SWI not to complete, got PC=0x%08X R0=0x%08X", v.CPU.PC, v.CPU.R[0])
	}
	if v.State != vm.StateRunning {
		t.Errorf("expected state restored to running, got %v", v.State)
	}
}
