package service_test

import (
	"os"
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// openThreeFiles opens out.txt for writing three times and exits without closing.
const openThreeFiles = `.org 0x8000
_start:
	LDR R0, =name
	MOV R1, #1
	SWI #0x10
	LDR R0, =name
	MOV R1, #1
	SWI #0x10
	LDR R0, =name
	MOV R1, #1
	SWI #0x10
	MOV R0, #0
	SWI #0x00
name:
	.asciz "out.txt"
`

const guestOpens = 3

func TestCloseReleasesGuestFilesAndRoot(t *testing.T) {
	baseline := openHostDescriptors(t)
	machine := vm.NewVM()
	if err := machine.OpenFilesystemRoot(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(machine.Close)
	svc := service.NewDebuggerService(machine)
	loadSource(t, svc, openThreeFiles)
	before := openHostDescriptors(t)

	runToHalt(t, svc)
	if opened := openHostDescriptors(t) - before; opened != guestOpens {
		t.Fatalf("setup: guest holds %d host descriptors, want %d", opened, guestOpens)
	}

	svc.Close()

	if after := openHostDescriptors(t); after != baseline {
		t.Errorf("host descriptors after Close = %d, want %d", after, baseline)
	}
}

// openHostDescriptors counts this process's open file descriptors.
func openHostDescriptors(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("cannot list open descriptors: %v", err)
	}
	return len(entries)
}
