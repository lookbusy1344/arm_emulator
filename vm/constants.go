package vm

// ============================================================================
// ARM2 Architecture Constants
// ============================================================================
// These values are defined by the ARM2 specification and should not be modified

const (
	// Instruction encoding
	ARMInstructionSize = 4 // bytes
	ARMPipelineOffset  = 8 // PC is instruction address + 8

	// Register counts
	ARMGeneralRegisterCount = 15 // R0-R14
	ARMTotalRegisterCount   = 16 // Including PC (R15)

	// CPSR flag bit positions (bits 31-28)
	CPSRBitN = 31 // Negative flag
	CPSRBitZ = 30 // Zero flag
	CPSRBitC = 29 // Carry flag
	CPSRBitV = 28 // Overflow flag

	// Sign bit for overflow calculations
	SignBitPos  = 31         // Position of sign bit in 32-bit word
	SignBitMask = 0x80000000 // Mask for sign bit

	// Instruction field bit positions are in arch_constants.go
	// They define the ARM instruction encoding format shared by encoder and decoder

	// Bit masks
	Mask4Bit  = 0xF
	Mask8Bit  = 0xFF
	Mask12Bit = 0xFFF
	Mask16Bit = 0xFFFF
	Mask24Bit = 0xFFFFFF
	Mask32Bit = 0xFFFFFFFF

	// Byte shift positions for endianness conversions
	ByteShift8  = 8  // Shift for byte 1 in multibyte values
	ByteShift16 = 16 // Shift for byte 2 in multibyte values
	ByteShift24 = 24 // Shift for byte 3 in multibyte values

	// Alignment constants (grouped together for discoverability)
	AlignmentWord     = 4 // 4-byte word alignment
	AlignmentHalfword = 2 // 2-byte halfword alignment
	AlignmentByte     = 1 // no alignment required

	// Computed alignment masks
	AlignMaskWord        = AlignmentWord - 1      // mask for word alignment check (address & mask == 0 means aligned)
	AlignMaskHalfword    = AlignmentHalfword - 1  // mask for halfword alignment check
	AlignRoundUpMaskWord = ^uint32(AlignMaskWord) // mask to round up to word alignment

	// Signed integer ranges (for branch offsets, etc.)
	Int24Max = 0x7FFFFF  // Maximum positive 24-bit signed value
	Int24Min = -0x800000 // Minimum negative 24-bit signed value
)

// ============================================================================
// Instruction Decoding - Bit Shift Positions
// ============================================================================
// ARM instruction field positions are in arch_constants.go
// Additional decoder-specific positions:

const (
	ShiftAmountPos = 7  // Bits 11-7: shift amount
	ShiftTypePos   = 5  // Bits 6-5: shift type
	Bit4Pos        = 4  // Bit 4: various uses
	Bit7Pos        = 7  // Bit 7: various uses
	IBitShift      = 25 // Bit 25: I (immediate/register)

	// Multiply field position
	MultiplyAShift = 21 // Bit 21: A bit (accumulate)

	// Bit ranges for multi-bit fields
	Bits27_26Shift = 26 // Bits 27-26 starting position
	Bits27_25Shift = 25 // Bits 27-25 starting position
	Bits27_23Shift = 23 // Bits 27-23 starting position
)

// ============================================================================
// Instruction Decoding - Bit Masks
// ============================================================================

const (
	// Field extraction masks (applied after shifting)
	Mask1Bit = 0x1
	Mask2Bit = 0x3
	Mask3Bit = 0x7
	Mask5Bit = 0x1F

	// Pre-shifted masks for common patterns
	BXPatternMask     = 0x0FFFFFF0 // Mask for BX/BLX detection
	LongMultiplyMask5 = 0x1F       // 5-bit mask for long multiply detection

	// Offset masks
	Offset12BitMask    = 0xFFF      // 12-bit immediate offset
	Offset24BitMask    = 0xFFFFFF   // 24-bit branch offset
	Offset24BitSignBit = 0x800000   // Sign bit for 24-bit offset
	Offset24BitSignExt = 0xFF000000 // Sign extension mask for 24-bit offset

	// Halfword transfer field masks
	HalfwordOffsetHighMask = 0xF // High nibble of halfword offset (bits 11-8)
	HalfwordOffsetLowMask  = 0xF // Low nibble of halfword offset (bits 3-0)
	HalfwordHighShift      = 8   // Shift for high nibble
	HalfwordLowShift       = 4   // Shift for assembling halfword offset

	// Register list mask (for LDM/STM)
	RegisterListMask = 0xFFFF // Bits 0-15: register list

	// Immediate value masks
	ImmediateValueMask = 0xFF // 8-bit immediate value
	RotationMask       = 0xF  // 4-bit rotation value
	RotationShift      = 8    // Position of rotation field

	// Value truncation masks
	ByteValueMask     = 0xFF   // Mask for byte values
	HalfwordValueMask = 0xFFFF // Mask for halfword values

	// Multiply cycle calculation
	MultiplyBit2Mask = 0x3 // 2-bit mask for multiply timing
)

// ============================================================================
// Special Values and Offsets
// ============================================================================

const (
	// PC offset adjustments
	PCStoreOffset = 12 // PC+12 when storing PC in STM
	PCBranchBase  = 8  // PC+8 base for branch calculations

	// Register aliases (from arch_constants.go)
	PCRegister = ARMRegisterPC // R15 is the PC
	SPRegister = ARMRegisterSP // R13 is the SP
	LRRegister = ARMRegisterLR // R14 is the LR

	// Bit shift for word-to-byte offset conversion
	WordToByteShift = 2 // Shift left by 2 to convert word offset to byte offset

	// Thumb mode bit (for BX instruction, though not fully supported in ARM2)
	ThumbModeClearMask = 0xFFFFFFFE // Mask to clear bit 0 (Thumb mode indicator)

	// Rotate constants
	RotationMultiplier = 2  // Rotation field is multiplied by 2
	BitsInWord         = 32 // 32 bits in a word

	// Multiply timing constants
	MultiplyBaseCycles = 2  // Base cycle count for multiply
	MultiplyMaxCycles  = 16 // Maximum cycle count for multiply
	MultiplyBitPairs   = 16 // Number of 2-bit pairs to check
	MultiplyBitShift   = 2  // Shift for each iteration

	// Long multiply timing
	LongMultiplyBaseCycles       = 3 // Base cycles for UMULL/SMULL
	LongMultiplyAccumulateCycles = 4 // Cycles for UMLAL/SMLAL

	// Address alignment for LDM/STM
	MultiRegisterWordSize = 4 // 4 bytes per register in multi-register transfers
)

// ============================================================================
// Special Instruction Encoding Patterns
// ============================================================================

const (
	// BX/BLX patterns are pre-positioned (NOT shift results - the trailing 0 is part of the hex value)
	// These are the actual bit patterns used directly in instruction encoding
	// Usage: instruction := (cond << 28) | BXEncodingBase | rm
	// The lower 4 bits (0000) are reserved for the rm register operand
	BXEncodingBase  = 0x012FFF10 // BX instruction base pattern (binary: 0b0000_0001_0010_1111_1111_1111_0001_0000)
	BLXEncodingBase = 0x012FFF30 // BLX instruction base pattern (binary: 0b0000_0001_0010_1111_1111_1111_0011_0000)
	NOPEncoding     = 0xE1A00000 // MOV R0, R0 (unconditional)
)

// ============================================================================
// Instruction Detection Patterns and Masks
// ============================================================================

const (
	// Multiply instruction patterns
	MultiplyPattern     = 0x00000090 // MUL/MLA pattern: bits [7:4] = 0b1001
	MultiplyMask        = 0x0FC000F0 // Mask to detect multiply instructions
	LongMultiplyPattern = 0x00800090 // UMULL/UMLAL/SMULL/SMLAL pattern
	LongMultiplyMask    = 0x0F8000F0 // Mask to detect long multiply instructions

	// PSR transfer instruction patterns
	MRSPattern    = 0x010F0000 // MRS instruction pattern
	MRSMask       = 0x0FBF0FFF // Mask to detect MRS instruction
	MSRRegPattern = 0x01200000 // MSR register form pattern
	MSRRegMask    = 0x0FB000F0 // Mask to detect MSR register
	MSRImmPattern = 0x03200000 // MSR immediate form pattern
	MSRImmMask    = 0x0FB00000 // Mask to detect MSR immediate

	// Branch detection patterns
	BranchBitMask     = 0x02000000 // Bit 25 set indicates branch in bits27-26=10 case
	BranchLinkPattern = 0x0B000000 // BL pattern: bits [27:24] = 0b1011
	BranchLinkMask    = 0x0F000000 // Mask to detect BL instruction
	SWIPattern        = 0x0F000000 // SWI pattern: bits [27:24] = 0b1111
	SWIDetectMask     = 0x0F000000 // Mask to detect SWI instruction (different from syscall extraction mask)

	// Link register initialization
	LRInitValue = 0xFFFFFFFF // Initial LR value for exception detection
)

// ============================================================================
// Memory Layout Constants
// ============================================================================

const (
	CodeSegmentStart  = 0x00008000 // 32KB offset - code begins at 32KB
	CodeSegmentSize   = 0x00010000 // 64KB - code segment size
	DataSegmentStart  = 0x00020000 // 128KB - data segment start
	DataSegmentSize   = 0x00010000 // 64KB - data segment size
	HeapSegmentStart  = 0x00030000 // 192KB - heap segment start
	HeapSegmentSize   = 0x00010000 // 64KB - heap segment size
	StackSegmentStart = 0x00040000 // 256KB - stack segment start
	StackSegmentSize  = 0x00010000 // 64KB - stack segment size
)

// ============================================================================
// VM Execution Limits
// ============================================================================

const (
	DefaultMaxCycles   = 1000000 // Default instruction limit
	DefaultLogCapacity = 1000    // Initial capacity for instruction log
	DefaultFDTableSize = 3       // Initial FD table size (FDs 0-2: stdin, stdout, stderr)
)

// ============================================================================
// Memory Overflow Protection
// ============================================================================

const (
	Address32BitMax     = ^uint32(0)
	Address32BitMaxSafe = 0xFFFFFFFC // Max address allowing 4-byte access without overflow
)

// ============================================================================
// Syscall Constants
// ============================================================================

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

// Note: For seek operations, use io.SeekStart, io.SeekCurrent, io.SeekEnd from the standard library

// Standard file descriptors
const (
	StdIn       = 0
	StdOut      = 1
	StdErr      = 2
	FirstUserFD = 3 // First available user FD
)

// Buffer size limits
const (
	MaxStringLength     = 1024 * 1024 // 1MB for general strings
	MaxFilenameLength   = 4096        // 4KB (typical filesystem limit)
	MaxAssertMsgLen     = 1024        // 1KB for assertion messages
	MaxReadSize         = 1024 * 1024 // 1MB maximum file read
	MaxWriteSize        = 1024 * 1024 // 1MB maximum file write
	MaxFileDescriptors  = 1024        // Maximum number of open FDs
	DefaultStringBuffer = 256         // Default buffer for READ_STRING
	MaxMemoryDump       = 1024        // 1KB limit for memory dumps
	MaxStdinInputSize   = 4096        // 4KB maximum stdin input per read (DoS protection)
)

// Note: Number bases (2, 8, 10, 16) are used directly as literals - they are self-documenting

// ASCII character ranges
const (
	ASCIIPrintableMin = 32  // Space
	ASCIIPrintableMax = 126 // Tilde (~)
)

// Statistics and Reporting Constants
const (
	// DefaultTopItemsCount is the default number of top items to show in statistics reports
	// (e.g., top instructions, hot paths, functions)
	DefaultTopItemsCount = 20

	// CompactTopItemsCount is the number of top items to show in compact statistics views
	CompactTopItemsCount = 10
)
