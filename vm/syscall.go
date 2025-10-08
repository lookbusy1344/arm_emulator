package vm

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// SWI (Software Interrupt) syscall numbers
const (
	// Console I/O
	SWI_EXIT          = 0x00
	SWI_WRITE_CHAR    = 0x01
	SWI_WRITE_STRING  = 0x02
	SWI_WRITE_INT     = 0x03
	SWI_READ_CHAR     = 0x04
	SWI_READ_STRING   = 0x05
	SWI_READ_INT      = 0x06
	SWI_WRITE_NEWLINE = 0x07

	// File Operations
	SWI_OPEN      = 0x10
	SWI_CLOSE     = 0x11
	SWI_READ      = 0x12
	SWI_WRITE     = 0x13
	SWI_SEEK      = 0x14
	SWI_TELL      = 0x15
	SWI_FILE_SIZE = 0x16

	// Memory Operations
	SWI_ALLOCATE   = 0x20
	SWI_FREE       = 0x21
	SWI_REALLOCATE = 0x22

	// System Information
	SWI_GET_TIME        = 0x30
	SWI_GET_RANDOM      = 0x31
	SWI_GET_ARGUMENTS   = 0x32
	SWI_GET_ENVIRONMENT = 0x33

	// Error Handling
	SWI_GET_ERROR   = 0x40
	SWI_SET_ERROR   = 0x41
	SWI_PRINT_ERROR = 0x42

	// Debugging Support
	SWI_DEBUG_PRINT    = 0xF0
	SWI_BREAKPOINT     = 0xF1
	SWI_DUMP_REGISTERS = 0xF2
	SWI_DUMP_MEMORY    = 0xF3
	SWI_ASSERT         = 0xF4
)

// ExecuteSWI executes a software interrupt (system call)
func ExecuteSWI(vm *VM, inst *Instruction) error {
	// Extract the SWI number (bottom 24 bits)
	swiNum := inst.Opcode & 0x00FFFFFF

	switch swiNum {
	// Console I/O
	case SWI_EXIT:
		return handleExit(vm)
	case SWI_WRITE_CHAR:
		return handleWriteChar(vm)
	case SWI_WRITE_STRING:
		return handleWriteString(vm)
	case SWI_WRITE_INT:
		return handleWriteInt(vm)
	case SWI_READ_CHAR:
		return handleReadChar(vm)
	case SWI_WRITE_NEWLINE:
		fmt.Println()
		vm.CPU.IncrementPC()
		return nil

	// Memory Operations
	case SWI_ALLOCATE:
		return handleAllocate(vm)
	case SWI_FREE:
		return handleFree(vm)

	// System Information
	case SWI_GET_TIME:
		return handleGetTime(vm)
	case SWI_GET_RANDOM:
		return handleGetRandom(vm)

	// Debugging Support
	case SWI_DEBUG_PRINT:
		return handleDebugPrint(vm)
	case SWI_BREAKPOINT:
		return handleBreakpoint(vm)
	case SWI_DUMP_REGISTERS:
		return handleDumpRegisters(vm)
	case SWI_DUMP_MEMORY:
		return handleDumpMemory(vm)

	default:
		return fmt.Errorf("unimplemented SWI: 0x%06X", swiNum)
	}
}

// Console I/O handlers
func handleExit(vm *VM) error {
	exitCode := vm.CPU.GetRegister(0)
	vm.State = StateHalted
	return fmt.Errorf("program exited with code %d", exitCode)
}

func handleWriteChar(vm *VM) error {
	char := vm.CPU.GetRegister(0)
	fmt.Printf("%c", char)
	vm.CPU.IncrementPC()
	return nil
}

func handleWriteString(vm *VM) error {
	addr := vm.CPU.GetRegister(0)

	// Read null-terminated string from memory
	var str []byte
	for {
		b, err := vm.Memory.ReadByte(addr)
		if err != nil {
			return fmt.Errorf("failed to read string at 0x%08X: %w", addr, err)
		}
		if b == 0 {
			break
		}
		str = append(str, b)
		addr++

		// Prevent infinite loops
		if len(str) > 1024*1024 {
			return fmt.Errorf("string too long (>1MB)")
		}
	}

	fmt.Print(string(str))
	vm.CPU.IncrementPC()
	return nil
}

func handleWriteInt(vm *VM) error {
	value := vm.CPU.GetRegister(0)
	base := vm.CPU.GetRegister(1)

	if base == 0 {
		base = 10 // Default to decimal
	}

	switch base {
	case 2:
		fmt.Printf("%b", value)
	case 8:
		fmt.Printf("%o", value)
	case 10:
		fmt.Printf("%d", int32(value))
	case 16:
		fmt.Printf("%x", value)
	default:
		return fmt.Errorf("unsupported base: %d", base)
	}

	vm.CPU.IncrementPC()
	return nil
}

func handleReadChar(vm *VM) error {
	var char byte
	_, err := fmt.Scanf("%c", &char)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF) // Return -1 on error
	} else {
		vm.CPU.SetRegister(0, uint32(char))
	}
	vm.CPU.IncrementPC()
	return nil
}

// Memory operation handlers
func handleAllocate(vm *VM) error {
	size := vm.CPU.GetRegister(0)

	// Allocate memory from heap
	addr, err := vm.Memory.Allocate(size)
	if err != nil {
		vm.CPU.SetRegister(0, 0) // Return NULL on failure
	} else {
		vm.CPU.SetRegister(0, addr)
	}

	vm.CPU.IncrementPC()
	return nil
}

func handleFree(vm *VM) error {
	addr := vm.CPU.GetRegister(0)

	err := vm.Memory.Free(addr)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF) // Return -1 on error
	} else {
		vm.CPU.SetRegister(0, 0) // Return 0 on success
	}

	vm.CPU.IncrementPC()
	return nil
}

// System information handlers
func handleGetTime(vm *VM) error {
	// Return time in milliseconds since Unix epoch
	millis := time.Now().UnixMilli()
	vm.CPU.SetRegister(0, uint32(millis&0xFFFFFFFF))
	vm.CPU.IncrementPC()
	return nil
}

func handleGetRandom(vm *VM) error {
	// Return a random 32-bit number
	vm.CPU.SetRegister(0, rand.Uint32())
	vm.CPU.IncrementPC()
	return nil
}

// Debugging handlers
func handleDebugPrint(vm *VM) error {
	addr := vm.CPU.GetRegister(0)

	// Read null-terminated string from memory
	var str []byte
	for {
		b, err := vm.Memory.ReadByte(addr)
		if err != nil {
			return fmt.Errorf("failed to read debug string at 0x%08X: %w", addr, err)
		}
		if b == 0 {
			break
		}
		str = append(str, b)
		addr++

		if len(str) > 1024*1024 {
			return fmt.Errorf("debug string too long (>1MB)")
		}
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", string(str))
	vm.CPU.IncrementPC()
	return nil
}

func handleBreakpoint(vm *VM) error {
	vm.State = StateBreakpoint
	return fmt.Errorf("breakpoint hit at PC=0x%08X", vm.CPU.PC)
}

func handleDumpRegisters(vm *VM) error {
	fmt.Println("=== Register Dump ===")
	for i := 0; i < 15; i++ {
		fmt.Printf("R%-2d = 0x%08X (%d)\n", i, vm.CPU.R[i], int32(vm.CPU.R[i]))
	}
	fmt.Printf("PC  = 0x%08X\n", vm.CPU.PC)
	fmt.Printf("CPSR = [%s%s%s%s]\n",
		map[bool]string{true: "N", false: "-"}[vm.CPU.CPSR.N],
		map[bool]string{true: "Z", false: "-"}[vm.CPU.CPSR.Z],
		map[bool]string{true: "C", false: "-"}[vm.CPU.CPSR.C],
		map[bool]string{true: "V", false: "-"}[vm.CPU.CPSR.V])
	fmt.Println("====================")

	vm.CPU.IncrementPC()
	return nil
}

func handleDumpMemory(vm *VM) error {
	addr := vm.CPU.GetRegister(0)
	length := vm.CPU.GetRegister(1)

	if length > 1024 {
		length = 1024 // Limit to 1KB
	}

	fmt.Printf("=== Memory Dump at 0x%08X (length=%d) ===\n", addr, length)

	for i := uint32(0); i < length; i += 16 {
		fmt.Printf("%08X: ", addr+i)

		// Hex bytes
		for j := uint32(0); j < 16 && i+j < length; j++ {
			b, err := vm.Memory.ReadByte(addr + i + j)
			if err != nil {
				fmt.Printf("?? ")
			} else {
				fmt.Printf("%02X ", b)
			}
		}

		// ASCII representation
		fmt.Print(" |")
		for j := uint32(0); j < 16 && i+j < length; j++ {
			b, err := vm.Memory.ReadByte(addr + i + j)
			if err != nil || b < 32 || b > 126 {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", b)
			}
		}
		fmt.Println("|")
	}

	fmt.Println("=======================================")
	vm.CPU.IncrementPC()
	return nil
}
