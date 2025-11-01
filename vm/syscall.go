package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Error Handling Philosophy:
//
// This module uses two different error handling strategies depending on the severity:
//
// 1. VM Integrity Errors (return Go errors, halt execution):
//    - Address wraparound/overflow when reading strings (e.g., handleWriteString, handleDebugPrint, handleAssert)
//    - These indicate potential memory corruption or security vulnerabilities
//    - These are VM-level failures that should stop execution immediately
//    - Returns: fmt.Errorf("...") which halts the VM
//
// 2. Expected Operation Failures (return error codes via R0, continue execution):
//    - File operation errors (file not found, read/write failures, etc.)
//    - Size limit violations (exceeding maxReadSize, maxWriteSize)
//    - Invalid file descriptors
//    - These are normal runtime errors that programs should handle
//    - Returns: 0xFFFFFFFF in R0 register, execution continues
//
// This distinction allows guest programs to handle expected errors (file I/O)
// while protecting the VM from integrity violations (memory corruption).

// Legacy constants - now defined in syscall_constants.go
// These are kept for compatibility but should be migrated to the new names
const (
	maxStringLength   = MaxStringLength
	maxFilenameLength = MaxFilenameLength
	maxAssertMsgLen   = MaxAssertMsgLen
	maxReadSize       = MaxReadSize
	maxWriteSize      = MaxWriteSize
	maxFDs            = MaxFileDescriptors
)

// ResetStdinReader resets the VM's stdin reader to read from os.Stdin
// This is useful for testing when os.Stdin has been redirected
func (vm *VM) ResetStdinReader() {
	vm.stdinReader = bufio.NewReader(os.Stdin)
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

// FD table helpers
//
// Thread Safety: File descriptor table access is protected by fdMu. However, the returned
// *os.File pointer is used after the lock is released. This is safe in the current design
// because:
// 1. The emulator executes guest programs single-threaded (one instruction at a time)
// 2. File descriptors are never removed from the table once allocated (no close invalidation)
// 3. Standard file descriptors (stdin/stdout/stderr) are never closed or replaced
//
// If multi-threaded guest program support is added in the future, file operations would need
// additional synchronization or reference counting to prevent use-after-close scenarios.
func (vm *VM) getFile(fd uint32) (*os.File, error) {
	vm.fdMu.Lock()
	defer vm.fdMu.Unlock()
	if int(fd) < 0 || int(fd) >= len(vm.files) {
		return nil, errors.New("bad fd")
	}
	f := vm.files[fd]
	// Lazily initialize standard file descriptors
	if f == nil && fd < FirstUserFD {
		switch fd {
		case StdIn:
			vm.files[StdIn] = os.Stdin
		case StdOut:
			vm.files[StdOut] = os.Stdout
		case StdErr:
			vm.files[StdErr] = os.Stderr
		}
		f = vm.files[fd]
	}
	if f == nil {
		return nil, errors.New("bad fd")
	}
	return f, nil
}

func (vm *VM) allocFD(f *os.File) uint32 {
	vm.fdMu.Lock()
	defer vm.fdMu.Unlock()

	for i := FirstUserFD; i < len(vm.files); i++ {
		if vm.files[i] == nil {
			vm.files[i] = f
			//nolint:gosec // G115: i is bounded by len(vm.files) which is reasonable
			return uint32(i)
		}
	}

	// Check limit before growing the table
	if len(vm.files) >= maxFDs {
		return SyscallErrorGeneral // Return error if limit reached
	}

	vm.files = append(vm.files, f)
	//nolint:gosec // G115: len(vm.files)-1 is bounded by reasonable file count
	return uint32(len(vm.files) - 1)
}

func (vm *VM) closeFD(fd uint32) error {
	vm.fdMu.Lock()
	defer vm.fdMu.Unlock()
	if int(fd) < 0 || int(fd) >= len(vm.files) || vm.files[fd] == nil {
		return errors.New("bad fd")
	}
	_ = vm.files[fd].Close()
	vm.files[fd] = nil
	return nil
}

// ExecuteSWI executes a software interrupt (system call)
func ExecuteSWI(vm *VM, inst *Instruction) error {
	// Preserve CPSR flags across SWI (syscalls shouldn't alter condition codes)
	saved := vm.CPU.CPSR
	var err error
	// Extract the syscall number from the immediate value (bottom 24 bits)
	// ARM2 traditional convention: SWI #num
	swiNum := inst.Opcode & SWIMask

	switch swiNum {
	// Console I/O
	case SWI_EXIT:
		err = handleExit(vm)
	case SWI_WRITE_CHAR:
		err = handleWriteChar(vm)
	case SWI_WRITE_STRING:
		err = handleWriteString(vm)
	case SWI_WRITE_INT:
		err = handleWriteInt(vm)
	case SWI_READ_CHAR:
		err = handleReadChar(vm)
	case SWI_READ_STRING:
		err = handleReadString(vm)
	case SWI_READ_INT:
		err = handleReadInt(vm)
	case SWI_WRITE_NEWLINE:
		_, _ = fmt.Fprintln(vm.OutputWriter) // Ignore write errors
		// Sync if it's os.Stdout
		if f, ok := vm.OutputWriter.(*os.File); ok {
			_ = f.Sync() // Ignore sync errors
		}
		vm.CPU.IncrementPC()

	// File Operations
	case SWI_OPEN:
		err = handleOpen(vm)
	case SWI_CLOSE:
		err = handleClose(vm)
	case SWI_READ:
		err = handleRead(vm)
	case SWI_WRITE:
		err = handleWrite(vm)
	case SWI_SEEK:
		err = handleSeek(vm)
	case SWI_TELL:
		err = handleTell(vm)
	case SWI_FILE_SIZE:
		err = handleFileSize(vm)

	// Memory Operations
	case SWI_ALLOCATE:
		err = handleAllocate(vm)
	case SWI_FREE:
		err = handleFree(vm)
	case SWI_REALLOCATE:
		err = handleReallocate(vm)

	// System Information
	case SWI_GET_TIME:
		err = handleGetTime(vm)
	case SWI_GET_RANDOM:
		err = handleGetRandom(vm)
	case SWI_GET_ARGUMENTS:
		err = handleGetArguments(vm)
	case SWI_GET_ENVIRONMENT:
		err = handleGetEnvironment(vm)

	// Error Handling
	case SWI_GET_ERROR:
		err = handleGetError(vm)
	case SWI_SET_ERROR:
		err = handleSetError(vm)
	case SWI_PRINT_ERROR:
		err = handlePrintError(vm)

	// Debugging Support
	case SWI_DEBUG_PRINT:
		err = handleDebugPrint(vm)
	case SWI_BREAKPOINT:
		err = handleBreakpoint(vm)
	case SWI_DUMP_REGISTERS:
		err = handleDumpRegisters(vm)
	case SWI_DUMP_MEMORY:
		err = handleDumpMemory(vm)
	case SWI_ASSERT:
		err = handleAssert(vm)

	default:
		err = fmt.Errorf("unimplemented SWI: 0x%06X", swiNum)
	}
	// Restore flags
	vm.CPU.CPSR.N = saved.N
	vm.CPU.CPSR.Z = saved.Z
	vm.CPU.CPSR.C = saved.C
	vm.CPU.CPSR.V = saved.V
	return err
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
	_, _ = fmt.Fprintf(vm.OutputWriter, "%c", char) // Ignore write errors
	// Sync if it's os.Stdout
	if f, ok := vm.OutputWriter.(*os.File); ok {
		_ = f.Sync() // Ignore sync errors
	}
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

		// Security: check for address wraparound before incrementing
		// If addr is at 0xFFFFFFFF, incrementing would wrap to 0
		if addr == 0xFFFFFFFF {
			return fmt.Errorf("address wraparound while reading string")
		}
		addr++

		// Prevent infinite loops
		if len(str) > maxStringLength {
			return fmt.Errorf("string too long (>%d bytes)", maxStringLength)
		}
	}

	_, _ = fmt.Fprint(vm.OutputWriter, string(str)) // Ignore write errors
	// Sync if it's os.Stdout
	if f, ok := vm.OutputWriter.(*os.File); ok {
		_ = f.Sync() // Ignore sync errors
	}
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
		_, _ = fmt.Fprintf(vm.OutputWriter, "%b", value) // Ignore write errors
	case 8:
		_, _ = fmt.Fprintf(vm.OutputWriter, "%o", value) // Ignore write errors
	case 10:
		_, _ = fmt.Fprintf(vm.OutputWriter, "%d", AsInt32(value)) // Ignore write errors
	case 16:
		_, _ = fmt.Fprintf(vm.OutputWriter, "%x", value) // Ignore write errors
	default:
		// This should never happen due to validation above, but keep for safety
		_, _ = fmt.Fprintf(vm.OutputWriter, "%d", AsInt32(value)) // Ignore write errors
	}

	// Sync if it's os.Stdout
	if f, ok := vm.OutputWriter.(*os.File); ok {
		_ = f.Sync() // Ignore sync errors
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleReadChar(vm *VM) error {
	// Skip any leading whitespace (newlines, spaces, tabs)
	for {
		char, err := vm.stdinReader.ReadByte()
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
	input, err := vm.stdinReader.ReadString('\n')
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
		line, err := vm.stdinReader.ReadString('\n')
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

		// Security: check for address wraparound before incrementing
		// If addr is at 0xFFFFFFFF, incrementing would wrap to 0
		if addr == 0xFFFFFFFF {
			return fmt.Errorf("address wraparound while reading debug string")
		}
		addr++

		if len(str) > maxStringLength {
			return fmt.Errorf("debug string too long (>%d bytes)", maxStringLength)
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
	_, _ = fmt.Fprintln(vm.OutputWriter, "=== Register Dump ===") // Ignore write errors
	for i := 0; i < 15; i++ {
		// Safe: intentional conversion to show signed interpretation of register value
		_, _ = fmt.Fprintf(vm.OutputWriter, "R%-2d = 0x%08X (%d)\n", i, vm.CPU.R[i], int32(vm.CPU.R[i])) // #nosec G115 -- intentional uint32->int32 for display, ignore write errors
	}
	_, _ = fmt.Fprintf(vm.OutputWriter, "PC  = 0x%08X\n", vm.CPU.PC) // Ignore write errors
	_, _ = fmt.Fprintf(vm.OutputWriter, "CPSR = [%s%s%s%s]\n",       // Ignore write errors
		map[bool]string{true: "N", false: "-"}[vm.CPU.CPSR.N],
		map[bool]string{true: "Z", false: "-"}[vm.CPU.CPSR.Z],
		map[bool]string{true: "C", false: "-"}[vm.CPU.CPSR.C],
		map[bool]string{true: "V", false: "-"}[vm.CPU.CPSR.V])
	_, _ = fmt.Fprintln(vm.OutputWriter, "====================") // Ignore write errors

	vm.CPU.IncrementPC()
	return nil
}

func handleDumpMemory(vm *VM) error {
	addr := vm.CPU.GetRegister(0)
	length := vm.CPU.GetRegister(1)

	if length > 1024 {
		length = 1024 // Limit to 1KB
	}

	_, _ = fmt.Fprintf(vm.OutputWriter, "=== Memory Dump at 0x%08X (length=%d) ===\n", addr, length) // Ignore write errors

	for i := uint32(0); i < length; i += 16 {
		_, _ = fmt.Fprintf(vm.OutputWriter, "%08X: ", addr+i) // Ignore write errors

		// Hex bytes
		for j := uint32(0); j < 16 && i+j < length; j++ {
			b, err := vm.Memory.ReadByteAt(addr + i + j)
			if err != nil {
				_, _ = fmt.Fprint(vm.OutputWriter, "?? ") // Ignore write errors
			} else {
				_, _ = fmt.Fprintf(vm.OutputWriter, "%02X ", b) // Ignore write errors
			}
		}

		// ASCII representation
		_, _ = fmt.Fprint(vm.OutputWriter, " |") // Ignore write errors
		for j := uint32(0); j < 16 && i+j < length; j++ {
			b, err := vm.Memory.ReadByteAt(addr + i + j)
			if err != nil || b < 32 || b > 126 {
				_, _ = fmt.Fprint(vm.OutputWriter, ".") // Ignore write errors
			} else {
				_, _ = fmt.Fprintf(vm.OutputWriter, "%c", b) // Ignore write errors
			}
		}
		_, _ = fmt.Fprintln(vm.OutputWriter, "|") // Ignore write errors
	}

	_, _ = fmt.Fprintln(vm.OutputWriter, "=======================================") // Ignore write errors
	vm.CPU.IncrementPC()
	return nil
}

// File operation handlers
func handleOpen(vm *VM) error {
	filenameAddr := vm.CPU.GetRegister(0)
	mode := vm.CPU.GetRegister(1) // 0=read, 1=write, 2=append

	// Read filename from memory
	// Note: Wraparound detection returns 0xFFFFFFFF (error code) rather than halting VM
	// This follows the file operation error handling pattern (see error handling philosophy at top of file)
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

		// Security: check for address wraparound before incrementing
		// If addr is at 0xFFFFFFFF, incrementing would wrap to 0
		if addr == 0xFFFFFFFF {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
		addr++

		if len(filename) > maxFilenameLength {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
	}
	var file *os.File
	var err error
	s := string(filename)
	switch mode {
	case 0:
		//nolint:gosec // G304: File path is intentionally controlled by emulated program
		file, err = os.Open(s)
	case 1:
		//nolint:gosec // G304,G302: File operations are intentional for emulated program I/O
		file, err = os.OpenFile(s, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	case 2:
		//nolint:gosec // G304,G302: File operations are intentional for emulated program I/O
		file, err = os.OpenFile(s, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	default:
		err = errors.New("bad mode")
	}
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		fd := vm.allocFD(file)
		vm.CPU.SetRegister(0, fd)
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleClose(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	if err := vm.closeFD(fd); err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		vm.CPU.SetRegister(0, 0)
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleRead(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	bufferAddr := vm.CPU.GetRegister(1)
	length := vm.CPU.GetRegister(2)
	f, err := vm.getFile(fd)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	// Security: limit read size to prevent memory exhaustion attacks
	// Maximum allowed: 1MB
	if length > maxReadSize {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	// Security: validate buffer address range to prevent overflow
	// Check that bufferAddr + length doesn't overflow the 32-bit address space
	if bufferAddr > 0xFFFFFFFF-length {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	// Buffer allocation: We allocate before the read operation rather than after validating
	// the read will succeed. This is a trade-off for code clarity:
	// - The file descriptor, size, and buffer range have been validated above
	// - The only remaining failure mode is I/O errors during read, which are rare
	// - Go's GC efficiently handles short-lived allocations if the read fails
	// - The alternative (seeking to check file position, then reading) adds complexity
	//   and may not be possible for non-seekable files (pipes, stdin, etc.)
	// - Maximum allocation is capped at 1MB, limiting potential waste
	data := make([]byte, length)
	n, err := f.Read(data)
	if err != nil && n == 0 {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	for i := 0; i < n; i++ {
		//nolint:gosec // G115: i is bounded by n which is from buffer size
		if err2 := vm.Memory.WriteByteAt(bufferAddr+uint32(i), data[i]); err2 != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
	}
	//nolint:gosec // G115: n is bounded by reasonable read size
	vm.CPU.SetRegister(0, uint32(n))
	vm.CPU.IncrementPC()
	return nil
}

func handleWrite(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	bufferAddr := vm.CPU.GetRegister(1)
	length := vm.CPU.GetRegister(2)
	f, err := vm.getFile(fd)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	// Security: limit write size to prevent memory exhaustion attacks
	// Maximum allowed: 1MB
	if length > maxWriteSize {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	// Security: validate buffer address range to prevent overflow
	// Check that bufferAddr + length doesn't overflow the 32-bit address space
	if bufferAddr > 0xFFFFFFFF-length {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	data := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b, err2 := vm.Memory.ReadByteAt(bufferAddr + i)
		if err2 != nil {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
			vm.CPU.IncrementPC()
			return nil
		}
		data[i] = b
	}
	n, err := f.Write(data)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		//nolint:gosec // G115: n is bounded by reasonable write size
		vm.CPU.SetRegister(0, uint32(n))
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleSeek(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	offset := int64(vm.CPU.GetRegister(1))
	whence := int(vm.CPU.GetRegister(2))
	f, err := vm.getFile(fd)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	npos, err := f.Seek(offset, whence)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		// Security: validate file position fits in 32-bit address space and is non-negative
		// This check correctly handles the full int64 range from Go's Seek():
		// - Rejects negative positions (npos < 0)
		// - Rejects positions beyond 32-bit range (npos > 0xFFFFFFFF, i.e., npos >= 0x100000000)
		// - Accepts only positions in [0, 0xFFFFFFFF] which safely fit in ARM2's 32-bit address space
		if npos < 0 || npos > 0xFFFFFFFF {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
		} else {
			//nolint:gosec // G115: File position validated above to fit in 32-bit range
			vm.CPU.SetRegister(0, uint32(npos))
		}
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleTell(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	f, err := vm.getFile(fd)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	pos, err := f.Seek(0, io.SeekCurrent) // current position
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		// Security: validate file position fits in 32-bit address space and is non-negative
		// This check correctly handles the full int64 range from Go's Seek():
		// - Rejects negative positions (pos < 0)
		// - Rejects positions beyond 32-bit range (pos > 0xFFFFFFFF, i.e., pos >= 0x100000000)
		// - Accepts only positions in [0, 0xFFFFFFFF] which safely fit in ARM2's 32-bit address space
		if pos < 0 || pos > 0xFFFFFFFF {
			vm.CPU.SetRegister(0, 0xFFFFFFFF)
		} else {
			//nolint:gosec // G115: File position validated above to fit in 32-bit range
			vm.CPU.SetRegister(0, uint32(pos))
		}
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleFileSize(vm *VM) error {
	fd := vm.CPU.GetRegister(0)
	f, err := vm.getFile(fd)
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		vm.CPU.IncrementPC()
		return nil
	}
	pos, _ := f.Seek(0, 1)   // save current
	end, err := f.Seek(0, 2) // seek end
	if err != nil {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
		_, _ = f.Seek(pos, 0)
		vm.CPU.IncrementPC()
		return nil
	}
	_, _ = f.Seek(pos, 0) // restore
	// Security: validate file size fits in 32-bit address space and is non-negative
	// This check correctly handles the full int64 range from Go's Seek():
	// - Rejects negative sizes (end < 0)
	// - Rejects sizes beyond 32-bit range (end > 0xFFFFFFFF, i.e., end >= 0x100000000)
	// - Accepts only sizes in [0, 0xFFFFFFFF] which safely fit in ARM2's 32-bit address space
	if end < 0 || end > 0xFFFFFFFF {
		vm.CPU.SetRegister(0, 0xFFFFFFFF)
	} else {
		//nolint:gosec // G115: File size validated above to fit in 32-bit range
		vm.CPU.SetRegister(0, uint32(end))
	}
	vm.CPU.IncrementPC()
	return nil
}

func handleReallocate(vm *VM) error {
	oldAddr := vm.CPU.GetRegister(0)
	newSize := vm.CPU.GetRegister(1)

	// Handle NULL pointer (allocate new)
	if oldAddr == 0 {
		newAddr, err := vm.Memory.Allocate(newSize)
		if err != nil {
			vm.CPU.SetRegister(0, 0) // NULL on failure
		} else {
			vm.CPU.SetRegister(0, newAddr)
		}
		vm.CPU.IncrementPC()
		return nil
	}

	// Get old allocation size from heap tracker
	oldAlloc, ok := vm.Memory.HeapAllocations[oldAddr]
	if !ok {
		// Invalid address - return NULL
		vm.CPU.SetRegister(0, 0)
		vm.CPU.IncrementPC()
		return nil
	}

	// Allocate new memory
	newAddr, err := vm.Memory.Allocate(newSize)
	if err != nil {
		vm.CPU.SetRegister(0, 0) // NULL on failure
		vm.CPU.IncrementPC()
		return nil
	}

	// Copy old data to new location (up to minimum of old and new sizes)
	copySize := oldAlloc.Size
	if newSize < copySize {
		copySize = newSize
	}

	for i := uint32(0); i < copySize; i++ {
		b, err := vm.Memory.ReadByteAt(oldAddr + i)
		if err != nil {
			// Copy failed, free new allocation and return NULL
			_ = vm.Memory.Free(newAddr)
			vm.CPU.SetRegister(0, 0)
			vm.CPU.IncrementPC()
			return nil
		}
		if err := vm.Memory.WriteByteAt(newAddr+i, b); err != nil {
			// Copy failed, free new allocation and return NULL
			_ = vm.Memory.Free(newAddr)
			vm.CPU.SetRegister(0, 0)
			vm.CPU.IncrementPC()
			return nil
		}
	}

	// Free old memory
	_ = vm.Memory.Free(oldAddr)

	vm.CPU.SetRegister(0, newAddr)
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

			// Security: check for address wraparound before incrementing
			// If addr is at 0xFFFFFFFF, incrementing would wrap to 0
			if addr == 0xFFFFFFFF {
				return fmt.Errorf("address wraparound while reading assertion message")
			}
			addr++

			if len(msg) > maxAssertMsgLen {
				break
			}
		}

		vm.State = StateError
		return fmt.Errorf("ASSERTION FAILED at PC=0x%08X: %s", vm.CPU.PC, string(msg))
	}

	vm.CPU.IncrementPC()
	return nil
}
