package vm

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Global stdin reader for consistent input handling across syscalls
var stdinReader = bufio.NewReader(os.Stdin)

// ResetStdinReader resets the global stdin reader to read from os.Stdin
// This is useful for testing when os.Stdin has been redirected
func ResetStdinReader() {
	stdinReader = bufio.NewReader(os.Stdin)
}

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
	// Extract the syscall number from the immediate value (bottom 24 bits)
	// ARM2 traditional convention: SWI #num
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
	case SWI_READ_STRING:
		return handleReadString(vm)
	case SWI_READ_INT:
		return handleReadInt(vm)
	case SWI_WRITE_NEWLINE:
		fmt.Println()
		vm.CPU.IncrementPC()
		return nil

	// File Operations
	case SWI_OPEN:
		return handleOpen(vm)
	case SWI_CLOSE:
		return handleClose(vm)
	case SWI_READ:
		return handleRead(vm)
	case SWI_WRITE:
		return handleWrite(vm)
	case SWI_SEEK:
		return handleSeek(vm)
	case SWI_TELL:
		return handleTell(vm)
	case SWI_FILE_SIZE:
		return handleFileSize(vm)

	// Memory Operations
	case SWI_ALLOCATE:
		return handleAllocate(vm)
	case SWI_FREE:
		return handleFree(vm)
	case SWI_REALLOCATE:
		return handleReallocate(vm)

	// System Information
	case SWI_GET_TIME:
		return handleGetTime(vm)
	case SWI_GET_RANDOM:
		return handleGetRandom(vm)
	case SWI_GET_ARGUMENTS:
		return handleGetArguments(vm)
	case SWI_GET_ENVIRONMENT:
		return handleGetEnvironment(vm)

	// Error Handling
	case SWI_GET_ERROR:
		return handleGetError(vm)
	case SWI_SET_ERROR:
		return handleSetError(vm)
	case SWI_PRINT_ERROR:
		return handlePrintError(vm)

	// Debugging Support
	case SWI_DEBUG_PRINT:
		return handleDebugPrint(vm)
	case SWI_BREAKPOINT:
		return handleBreakpoint(vm)
	case SWI_DUMP_REGISTERS:
		return handleDumpRegisters(vm)
	case SWI_DUMP_MEMORY:
		return handleDumpMemory(vm)
	case SWI_ASSERT:
		return handleAssert(vm)

	default:
		return fmt.Errorf("unimplemented SWI: 0x%06X", swiNum)
	}
}

// Console I/O handlers
func handleExit(vm *VM) error {
	exitCode := vm.CPU.GetRegister(0)
	// Intentional conversion - exit codes are typically signed
	//nolint:gosec // G115: Exit code conversion uint32->int32
	vm.ExitCode = int32(exitCode)
	vm.State = StateHalted
	return fmt.Errorf("program exited with code %d", exitCode)
}

func handleWriteChar(vm *VM) error {
	char := vm.CPU.GetRegister(0)
	fmt.Printf("%c", char)
	_ = os.Stdout.Sync() // Ignore sync errors
	vm.CPU.IncrementPC()
	return nil
}

func handleWriteString(vm *VM) error {
	addr := vm.CPU.GetRegister(0)

	// Read null-terminated string from memory
	var str []byte
	for {
		b, err := vm.Memory.ReadByteAt(addr)
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
	_ = os.Stdout.Sync() // Ignore sync errors
	vm.CPU.IncrementPC()
	return nil
}

func handleWriteInt(vm *VM) error {
	value := vm.CPU.GetRegister(0)
	base := vm.CPU.GetRegister(1)

	// Validate base and default to 10 for invalid values
	// This handles cases where R1 wasn't explicitly set (e.g., contains error flags from previous syscalls)
	if base == 0 || base == 0xFFFFFFFF || (base != 2 && base != 8 && base != 10 && base != 16) {
		base = 10 // Default to decimal
	}

	switch base {
	case 2:
		fmt.Printf("%b", value)
	case 8:
		fmt.Printf("%o", value)
	case 10:
		fmt.Printf("%d", AsInt32(value))
	case 16:
		fmt.Printf("%x", value)
	default:
		// This should never happen due to validation above, but keep for safety
		fmt.Printf("%d", AsInt32(value))
	}

	_ = os.Stdout.Sync() // Ignore sync errors
	vm.CPU.IncrementPC()
	return nil
}

func handleReadChar(vm *VM) error {
	// Skip any leading whitespace (newlines, spaces, tabs)
	for {
		char, err := stdinReader.ReadByte()
		if err != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF) // Return -1 on error
			vm.CPU.IncrementPC()
			return nil
		}
		// If it's not whitespace, we found our character
		if char != '\n' && char != '\r' && char != ' ' && char != '\t' {
			vm.CPU.SetRegister(0, uint32(char))
			vm.CPU.IncrementPC()
			return nil
		}
	}
}

func handleReadString(vm *VM) error {
	addr := vm.CPU.GetRegister(0)
	maxLen := vm.CPU.GetRegister(1)

	if maxLen == 0 {
		maxLen = 256 // Default max length
	}

	// Read string from stdin (up to newline)
	input, err := stdinReader.ReadString('\n')
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF) // Return -1 on error
		vm.CPU.IncrementPC()
		return nil
	}

	// Remove trailing newline
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")

	// Write string to memory (up to maxLen-1 chars + null terminator)
	// Safe: input is from reader, length bounded by buffer size and maxLen check below
	bytesToWrite := uint32(len(input)) // #nosec G115 -- bounded by maxLen
	if bytesToWrite >= maxLen {
		bytesToWrite = maxLen - 1
	}

	for i := uint32(0); i < bytesToWrite; i++ {
		if err := vm.Memory.WriteByteAt(addr+i, input[i]); err != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF) // Return -1 on error
			vm.CPU.IncrementPC()
			return nil
		}
	}

	// Write null terminator
	if err := vm.Memory.WriteByteAt(addr+bytesToWrite, 0); err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}

	vm.CPU.SetRegister(0, bytesToWrite) // Return number of bytes written (excluding null)
	vm.CPU.IncrementPC()
	return nil
}

func handleReadInt(vm *VM) error {
	// Read lines until we get a non-empty one or hit EOF
	for {
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			vm.CPU.SetRegister(0, 0)
			vm.CPU.IncrementPC()
			return nil
		}

		// Parse the integer from the line
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip empty lines
			continue
		}

		value, err := strconv.ParseInt(line, 10, 32)
		if err != nil {
			// On error, return 0 in R0
			vm.CPU.SetRegister(0, 0)
		} else {
			// Safe: value from ParseInt with bitSize 32, fits in int32 range [-2^31, 2^31-1]
			// which maps correctly to uint32 via two's complement
			vm.CPU.SetRegister(0, uint32(int32(value))) // #nosec G115 -- int32 to uint32, intentional
		}
		vm.CPU.IncrementPC()
		return nil
	}
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
	// Safe: masking with 0xFFFFFFFF before conversion ensures result fits in uint32
	vm.CPU.SetRegister(0, uint32(millis&0xFFFFFFFF)) // #nosec G115 -- masked to 32 bits
	vm.CPU.IncrementPC()
	return nil
}

func handleGetRandom(vm *VM) error {
	// Return a random 32-bit number (non-cryptographic use)
	vm.CPU.SetRegister(0, rand.Uint32()) // #nosec G404 -- pseudo-random for emulator, not crypto
	vm.CPU.IncrementPC()
	return nil
}

// Debugging handlers
func handleDebugPrint(vm *VM) error {
	addr := vm.CPU.GetRegister(0)

	// Read null-terminated string from memory
	var str []byte
	for {
		b, err := vm.Memory.ReadByteAt(addr)
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
		// Safe: intentional conversion to show signed interpretation of register value
		fmt.Printf("R%-2d = 0x%08X (%d)\n", i, vm.CPU.R[i], int32(vm.CPU.R[i])) // #nosec G115 -- intentional uint32->int32 for display
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
			b, err := vm.Memory.ReadByteAt(addr + i + j)
			if err != nil {
				fmt.Printf("?? ")
			} else {
				fmt.Printf("%02X ", b)
			}
		}

		// ASCII representation
		fmt.Print(" |")
		for j := uint32(0); j < 16 && i+j < length; j++ {
			b, err := vm.Memory.ReadByteAt(addr + i + j)
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

// File operation handlers
func handleOpen(vm *VM) error {
	filenameAddr := vm.CPU.GetRegister(0)
	mode := vm.CPU.GetRegister(1) // 0=read, 1=write, 2=append

	// Read filename from memory
	var filename []byte
	addr := filenameAddr
	for {
		b, err := vm.Memory.ReadByteAt(addr)
		if err != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
		if b == 0 {
			break
		}
		filename = append(filename, b)
		addr++
		if len(filename) > 1024 {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
	}

	var file *os.File
	var err error

	switch mode {
	case 0: // Read
		file, err = os.Open(string(filename))
	case 1: // Write
		file, err = os.Create(string(filename))
	case 2: // Append
		file, err = os.OpenFile(string(filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	default:
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}

	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		// Store file descriptor (using file pointer as fd for simplicity)
		// In a real implementation, we'd maintain a file descriptor table
		fd := uint32(file.Fd())
		vm.CPU.SetRegister(0, fd)
	}

	vm.CPU.IncrementPC()
	return nil
}

func handleClose(vm *VM) error {
	// Note: This is a simplified implementation
	// A full implementation would maintain a file descriptor table
	vm.CPU.SetRegister(0, 0) // Success
	vm.CPU.IncrementPC()
	return nil
}

func handleRead(vm *VM) error {
	// fd := vm.CPU.GetRegister(0)
	bufferAddr := vm.CPU.GetRegister(1)
	length := vm.CPU.GetRegister(2)

	// Simplified: read from stdin
	data := make([]byte, length)
	n, err := os.Stdin.Read(data)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}

	// Write to memory
	for i := 0; i < n; i++ {
		// Validate offset won't overflow
		if i < 0 || i > int(^uint32(0)) {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
		if err := vm.Memory.WriteByteAt(bufferAddr+uint32(i), data[i]); err != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
	}

	// Validate n fits in uint32
	if n < 0 || n > int(^uint32(0)) {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	vm.CPU.SetRegister(0, uint32(n))
	vm.CPU.IncrementPC()
	return nil
}

func handleWrite(vm *VM) error {
	// fd := vm.CPU.GetRegister(0)
	bufferAddr := vm.CPU.GetRegister(1)
	length := vm.CPU.GetRegister(2)

	// Read data from memory
	data := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b, err := vm.Memory.ReadByteAt(bufferAddr + i)
		if err != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
		data[i] = b
	}

	// Simplified: write to stdout
	n, err := os.Stdout.Write(data)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		// Validate n fits in uint32
		if n < 0 || n > int(^uint32(0)) {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
		} else {
			vm.CPU.SetRegister(0, uint32(n))
		}
	}
	_ = os.Stdout.Sync() // Ignore sync errors

	vm.CPU.IncrementPC()
	return nil
}

func handleSeek(vm *VM) error {
	// Simplified implementation - not fully functional
	vm.CPU.SetRegister(0, 0) // Return 0 for success
	vm.CPU.IncrementPC()
	return nil
}

func handleTell(vm *VM) error {
	// Simplified implementation - not fully functional
	vm.CPU.SetRegister(0, 0) // Return position 0
	vm.CPU.IncrementPC()
	return nil
}

func handleFileSize(vm *VM) error {
	// Simplified implementation - not fully functional
	vm.CPU.SetRegister(0, 0) // Return size 0
	vm.CPU.IncrementPC()
	return nil
}

func handleReallocate(vm *VM) error {
	oldAddr := vm.CPU.GetRegister(0)
	newSize := vm.CPU.GetRegister(1)

	// Simplified: allocate new memory and copy
	// A full implementation would track allocations
	newAddr, err := vm.Memory.Allocate(newSize)
	if err != nil {
		vm.CPU.SetRegister(0, 0) // NULL on failure
	} else {
		// Would copy old data here in full implementation
		_ = vm.Memory.Free(oldAddr) // Free old memory (ignore error if invalid)
		vm.CPU.SetRegister(0, newAddr)
	}

	vm.CPU.IncrementPC()
	return nil
}

// System information handlers (extended)
func handleGetArguments(vm *VM) error {
	// Return number of arguments in R0, pointer to argv in R1
	argLen := len(vm.ProgramArguments)
	// Validate length fits in uint32
	if argLen < 0 || argLen > int(^uint32(0)) {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	argc := uint32(argLen)
	vm.CPU.SetRegister(0, argc)

	// In a full implementation, we would:
	// 1. Allocate memory for argv array
	// 2. Copy argument strings to memory
	// 3. Return pointer to argv array in R1
	// For now, return 0 for argv pointer
	vm.CPU.SetRegister(1, 0)

	vm.CPU.IncrementPC()
	return nil
}

func handleGetEnvironment(vm *VM) error {
	// Return pointer to environment variables
	// Simplified: return NULL
	vm.CPU.SetRegister(0, 0)
	vm.CPU.IncrementPC()
	return nil
}

// Error handling handlers
func handleGetError(vm *VM) error {
	// Return last error code
	// Simplified: return 0 (no error)
	vm.CPU.SetRegister(0, 0)
	vm.CPU.IncrementPC()
	return nil
}

func handleSetError(vm *VM) error {
	// Set error code
	// errorCode := vm.CPU.GetRegister(0)
	// Would store this in VM state in full implementation
	vm.CPU.IncrementPC()
	return nil
}

func handlePrintError(vm *VM) error {
	errorCode := vm.CPU.GetRegister(0)
	fmt.Fprintf(os.Stderr, "Error code: %d\n", errorCode)
	vm.CPU.IncrementPC()
	return nil
}

func handleAssert(vm *VM) error {
	condition := vm.CPU.GetRegister(0)
	msgAddr := vm.CPU.GetRegister(1)

	if condition == 0 {
		// Assertion failed
		var msg []byte
		addr := msgAddr
		for {
			b, err := vm.Memory.ReadByteAt(addr)
			if err != nil || b == 0 {
				break
			}
			msg = append(msg, b)
			addr++
			if len(msg) > 1024 {
				break
			}
		}

		vm.State = StateError
		return fmt.Errorf("ASSERTION FAILED at PC=0x%08X: %s", vm.CPU.PC, string(msg))
	}

	vm.CPU.IncrementPC()
	return nil
}
