package main

import (
	"fmt"
	"os"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func main() {
	fmt.Println("ARM2 Emulator - Phase 1: Core VM Foundation")
	fmt.Println("============================================")
	fmt.Println()

	// Create a new VM instance
	machine := vm.NewVM()

	fmt.Printf("VM initialized successfully\n")
	fmt.Printf("Initial state: %s\n", machine.DumpState())
	fmt.Println()

	// Display memory configuration
	fmt.Println("Memory Configuration:")
	fmt.Printf("  Code Segment:  0x%08X - 0x%08X (%d KB)\n",
		vm.CodeSegmentStart,
		vm.CodeSegmentStart+vm.CodeSegmentSize,
		vm.CodeSegmentSize/1024)
	fmt.Printf("  Data Segment:  0x%08X - 0x%08X (%d KB)\n",
		vm.DataSegmentStart,
		vm.DataSegmentStart+vm.DataSegmentSize,
		vm.DataSegmentSize/1024)
	fmt.Printf("  Heap Segment:  0x%08X - 0x%08X (%d KB)\n",
		vm.HeapSegmentStart,
		vm.HeapSegmentStart+vm.HeapSegmentSize,
		vm.HeapSegmentSize/1024)
	fmt.Printf("  Stack Segment: 0x%08X - 0x%08X (%d KB)\n",
		vm.StackSegmentStart,
		vm.StackSegmentStart+vm.StackSegmentSize,
		vm.StackSegmentSize/1024)
	fmt.Println()

	// Test basic memory operations
	fmt.Println("Testing basic memory operations...")
	testAddress := uint32(vm.DataSegmentStart)

	// Write and read a word
	testValue := uint32(0x12345678)
	if err := machine.Memory.WriteWord(testAddress, testValue); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to memory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Wrote 0x%08X to address 0x%08X\n", testValue, testAddress)

	readValue, err := machine.Memory.ReadWord(testAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from memory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Read  0x%08X from address 0x%08X\n", readValue, testAddress)

	if readValue == testValue {
		fmt.Println("  ✓ Memory read/write test passed")
	} else {
		fmt.Println("  ✗ Memory read/write test failed")
		os.Exit(1)
	}
	fmt.Println()

	// Test register operations
	fmt.Println("Testing register operations...")
	machine.CPU.SetRegister(0, 42)
	machine.CPU.SetRegister(1, 100)
	fmt.Printf("  Set R0 = %d\n", machine.CPU.GetRegister(0))
	fmt.Printf("  Set R1 = %d\n", machine.CPU.GetRegister(1))

	// Test stack pointer
	stackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	machine.InitializeStack(stackTop)
	fmt.Printf("  Initialized SP = 0x%08X\n", machine.CPU.GetSP())
	fmt.Println()

	// Test flag operations
	fmt.Println("Testing CPSR flag operations...")
	machine.CPU.CPSR.UpdateFlagsNZ(0)
	fmt.Printf("  Result=0: N=%v Z=%v (expected N=false Z=true)\n",
		machine.CPU.CPSR.N, machine.CPU.CPSR.Z)

	machine.CPU.CPSR.UpdateFlagsNZ(0x80000000)
	fmt.Printf("  Result=0x80000000: N=%v Z=%v (expected N=true Z=false)\n",
		machine.CPU.CPSR.N, machine.CPU.CPSR.Z)
	fmt.Println()

	// Test condition evaluation
	fmt.Println("Testing condition code evaluation...")
	machine.CPU.CPSR.Z = true
	if machine.CPU.CPSR.EvaluateCondition(vm.CondEQ) {
		fmt.Println("  ✓ CondEQ correctly evaluates to true when Z=1")
	}
	machine.CPU.CPSR.Z = false
	if !machine.CPU.CPSR.EvaluateCondition(vm.CondEQ) {
		fmt.Println("  ✓ CondEQ correctly evaluates to false when Z=0")
	}
	if machine.CPU.CPSR.EvaluateCondition(vm.CondAL) {
		fmt.Println("  ✓ CondAL always evaluates to true")
	}
	fmt.Println()

	fmt.Println("Phase 1 foundation complete!")
	fmt.Println("Next steps:")
	fmt.Println("  - Implement parser (Phase 2)")
	fmt.Println("  - Implement instruction set (Phase 3)")
	fmt.Println("  - Implement debugger (Phase 5)")
	fmt.Println("  - Implement TUI (Phase 6)")
}
