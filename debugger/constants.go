package debugger

// TUI Display Update Constants
const (
	// DisplayUpdateFrequency controls how often the TUI display updates during continuous execution
	// (every N cycles to keep display responsive without overwhelming the terminal)
	DisplayUpdateFrequency = 100
)

// Code View Context Constants
const (
	// CodeContextLinesBefore is the default number of lines to show before PC in the full code view
	CodeContextLinesBefore = 20

	// CodeContextLinesAfter is the default number of lines to show after PC in the full code view
	CodeContextLinesAfter = 80

	// CodeContextLinesBeforeCompact is the number of lines to show before PC in compact views
	CodeContextLinesBeforeCompact = 5

	// CodeContextLinesAfterCompact is the number of lines to show after PC in compact views
	CodeContextLinesAfterCompact = 10
)

// Memory Display Constants
const (
	// MemoryDisplayRows is the number of rows to show in the memory hex dump view
	MemoryDisplayRows = 16

	// MemoryDisplayColumns is the number of bytes per row in the memory hex dump view
	MemoryDisplayColumns = 16

	// MemoryDisplayBytesPerRow is the number of bytes displayed per row (same as columns)
	MemoryDisplayBytesPerRow = 16
)

// Stack Display Constants
const (
	// StackDisplayWords is the number of 32-bit words to show in the stack view
	StackDisplayWords = 16

	// StackDisplayBytes is the total number of bytes shown in the stack view (16 words * 4 bytes)
	StackDisplayBytes = 64

	// StackInspectionMaxOffset is the maximum byte offset when inspecting stack in debugger commands
	StackInspectionMaxOffset = 16
)

// Register Display Constants
const (
	// RegisterViewRows is the fixed height of the register view panel
	// (5 rows of registers + blank line + status line + borders)
	RegisterViewRows = 9

	// RegisterGroupSize is the number of registers displayed per row
	RegisterGroupSize = 5
)
