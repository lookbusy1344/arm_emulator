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

func TestCloseReleasesGuestFiles(t *testing.T) {
	machine := vm.NewVM()
	machine.FilesystemRoot = t.TempDir()
	svc := service.NewDebuggerService(machine)
	loadSource(t, svc, openThreeFiles)
	before := openHostDescriptors(t)

	runToHalt(t, svc)
	if opened := openHostDescriptors(t) - before; opened != guestOpens {
		t.Fatalf("setup: guest holds %d host descriptors, want %d", opened, guestOpens)
	}

	svc.Close()

	if after := openHostDescriptors(t); after != before {
		t.Errorf("host descriptors after Close = %d, want %d", after, before)
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
