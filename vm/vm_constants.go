package vm

// VM Execution Limits
const (
	DefaultMaxCycles   = 1000000 // Default instruction limit
	DefaultLogCapacity = 1000    // Initial capacity for instruction log
	DefaultFDTableSize = 3       // Initial FD table size (FDs 0-2: stdin, stdout, stderr)
)

// Memory overflow protection
const (
	Address32BitMax     = ^uint32(0)
	Address32BitMaxSafe = 0xFFFFFFFC // Max address allowing 4-byte access without overflow
)
