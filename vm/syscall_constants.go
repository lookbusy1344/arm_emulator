package vm

// Syscall Return Values
const (
	SyscallSuccess      = 0
	SyscallErrorGeneral = 0xFFFFFFFF // -1 in two's complement
	SyscallNull         = 0          // NULL pointer
)

// Syscall number extraction
const (
	SWIMask = 0x00FFFFFF // Bottom 24 bits contain syscall number
)

// File operation modes
const (
	FileModeRead   = 0 // Read-only
	FileModeWrite  = 1 // Write (create/truncate)
	FileModeAppend = 2 // Append (create/read-write)
)

// File permissions (Unix-style)
const (
	FilePermDefault = 0644 // rw-r--r--
)

// Seek whence values (match io.Seek* constants)
const (
	SeekStart   = 0 // io.SeekStart
	SeekCurrent = 1 // io.SeekCurrent
	SeekEnd     = 2 // io.SeekEnd
)

// Standard file descriptors
const (
	StdIn       = 0
	StdOut      = 1
	StdErr      = 2
	FirstUserFD = 3 // First available user FD
)

// Buffer size limits
const (
	MaxStringLength    = 1024 * 1024 // 1MB for general strings
	MaxFilenameLength  = 4096        // 4KB (typical filesystem limit)
	MaxAssertMsgLen    = 1024        // 1KB for assertion messages
	MaxReadSize        = 1024 * 1024 // 1MB maximum file read
	MaxWriteSize       = 1024 * 1024 // 1MB maximum file write
	MaxFileDescriptors = 1024        // Maximum number of open FDs
	DefaultStringBuffer = 256        // Default buffer for READ_STRING
	MaxMemoryDump      = 1024        // 1KB limit for memory dumps
)

// Number bases for integer output (WRITE_INT syscall)
const (
	BaseBinary      = 2
	BaseOctal       = 8
	BaseDecimal     = 10
	BaseHexadecimal = 16
)

// ASCII character ranges
const (
	ASCIIPrintableMin = 32  // Space
	ASCIIPrintableMax = 126 // Tilde (~)
)
