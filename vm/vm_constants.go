package vm

// VM Execution Limits
const (
	DefaultMaxCycles   = 1000000 // Default instruction limit
	DefaultLogCapacity = 1000    // Initial capacity for instruction log
	DefaultFDTableSize = 3       // stdin, stdout, stderr
)

// Memory overflow protection
const (
	Address32BitMax     = 0xFFFFFFFF
	Address32BitMaxSafe = 0xFFFFFFFC // Max address allowing 4-byte access
	AddressWrapBoundary = 0xFFFFFFFF // Address that would wrap on increment
)
