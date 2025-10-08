package vm_test

import "github.com/lookbusy1344/arm-emulator/vm"

// Helper function to enable write permissions on code segment
func setupCodeWrite(v *vm.VM) {
	for _, seg := range v.Memory.Segments {
		if seg.Name == "code" {
			seg.Permissions = vm.PermRead | vm.PermWrite | vm.PermExecute
		}
	}
}

// Helper function to enable write permissions on data segment
func setupDataWrite(v *vm.VM) {
	for _, seg := range v.Memory.Segments {
		if seg.Name == "data" {
			seg.Permissions = vm.PermRead | vm.PermWrite
		}
	}
}
