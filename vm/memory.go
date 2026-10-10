package vm

import (
	"cmp"
	"fmt"
	"slices"
)

// Memory access permissions
type MemoryPermission byte

const (
	PermNone    MemoryPermission = 0
	PermRead    MemoryPermission = 1 << 0
	PermWrite   MemoryPermission = 1 << 1
	PermExecute MemoryPermission = 1 << 2
)

// MemorySegment represents a region of memory with permissions
type MemorySegment struct {
	Start       uint32
	Size        uint32
	Data        []byte
	Permissions MemoryPermission
	Name        string
}

// Memory represents the ARM2 virtual memory system
type Memory struct {
	Segments        []*MemorySegment
	LittleEndian    bool
	StrictAlign     bool
	AccessCount     uint64
	ReadCount       uint64
	WriteCount      uint64
	HeapAllocations map[uint32]*HeapAllocation
	NextHeapAddress uint32

	// freeBlocks holds freed heap blocks below NextHeapAddress, sorted by address.
	// Adjacent blocks are merged, and a block that reaches NextHeapAddress is
	// returned to the unallocated top.
	freeBlocks []HeapAllocation
}

// NewMemory creates and initializes a new Memory instance
func NewMemory() *Memory {
	m := &Memory{
		Segments:        make([]*MemorySegment, 0),
		LittleEndian:    true,
		StrictAlign:     true,
		HeapAllocations: make(map[uint32]*HeapAllocation),
		NextHeapAddress: HeapSegmentStart,
	}

	for _, seg := range standardSegments {
		m.AddSegment(seg.name, seg.start, seg.size, seg.permissions)
	}
	return m
}

// standardSegments is the layout every Memory starts with and returns to on Reset.
// The code segment is writable to match ARM2 hardware, which had no memory protection;
// many programs embed writable data in it with .space and .word.
var standardSegments = []struct {
	name        string
	start, size uint32
	permissions MemoryPermission
}{
	{"code", CodeSegmentStart, CodeSegmentSize, PermRead | PermWrite | PermExecute},
	{"data", DataSegmentStart, DataSegmentSize, PermRead | PermWrite},
	{"heap", HeapSegmentStart, HeapSegmentSize, PermRead | PermWrite},
	{"stack", StackSegmentStart, StackSegmentSize, PermRead | PermWrite},
}

// AddSegment adds a new memory segment
func (m *Memory) AddSegment(name string, start, size uint32, permissions MemoryPermission) {
	segment := &MemorySegment{
		Start:       start,
		Size:        size,
		Data:        make([]byte, size),
		Permissions: permissions,
		Name:        name,
	}
	m.Segments = append(m.Segments, segment)
}

// findSegment finds the memory segment containing the given address
func (m *Memory) findSegment(address uint32) (*MemorySegment, uint32, error) {
	for _, seg := range m.Segments {
		// Check if address is in segment range using explicit bounds checking
		// Step 1: Verify address >= seg.Start (protects against wraparound attacks)
		// Step 2: Calculate offset and verify offset < seg.Size
		//
		// This two-step approach correctly handles all edge cases:
		// - If address < seg.Start: First check fails, never calculates offset
		// - If address >= seg.Start but beyond segment: offset >= seg.Size, second check fails
		// - Segments near 0xFFFFFFFF: Addresses that wrap to low memory fail the first check
		//
		// Example: Segment at 0xFFFF0000, size 0x00020000
		//   - Access to 0xFFFF0000-0xFFFFFFFF: Both checks pass (valid)
		//   - Access to 0x00000100: First check fails (0x00000100 < 0xFFFF0000)
		//   - No wraparound vulnerability exists in this implementation
		if address >= seg.Start {
			offset := address - seg.Start
			if offset < seg.Size {
				return seg, offset, nil
			}
		}
	}
	return nil, 0, fmt.Errorf("memory access violation: address 0x%08X is not mapped", address)
}

// checkAlignment checks if an address is properly aligned
func (m *Memory) checkAlignment(address uint32, size int) error {
	if !m.StrictAlign {
		return nil
	}

	switch size {
	case AlignmentWord: // Word access
		if address&AlignMaskWord != 0 {
			return fmt.Errorf("unaligned word access at 0x%08X (must be 4-byte aligned)", address)
		}
	case AlignmentHalfword: // Halfword access
		if address&AlignMaskHalfword != 0 {
			return fmt.Errorf("unaligned halfword access at 0x%08X (must be 2-byte aligned)", address)
		}
	case AlignmentByte: // Byte access - no alignment required
	default:
		return fmt.Errorf("invalid memory access size: %d", size)
	}
	return nil
}

// ReadByteAt reads a single byte from memory at the specified address
func (m *Memory) ReadByteAt(address uint32) (byte, error) {
	seg, offset, err := m.findSegment(address)
	if err != nil {
		return 0, err
	}

	if seg.Permissions&PermRead == 0 {
		return 0, fmt.Errorf("read permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	m.AccessCount++
	m.ReadCount++
	return seg.Data[offset], nil
}

// WriteByteAt writes a single byte to memory at the specified address
func (m *Memory) WriteByteAt(address uint32, value byte) error {
	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermWrite == 0 {
		return fmt.Errorf("write permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	m.AccessCount++
	m.WriteCount++
	seg.Data[offset] = value
	return nil
}

// ReadHalfword reads a 16-bit halfword from memory
func (m *Memory) ReadHalfword(address uint32) (uint16, error) {
	if err := m.checkAlignment(address, AlignmentHalfword); err != nil {
		return 0, err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return 0, err
	}

	if seg.Permissions&PermRead == 0 {
		return 0, fmt.Errorf("read permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset+1 >= segLen {
		return 0, fmt.Errorf("halfword read exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.ReadCount++

	var value uint16
	if m.LittleEndian {
		value = uint16(seg.Data[offset]) | uint16(seg.Data[offset+1])<<ByteShift8
	} else {
		value = uint16(seg.Data[offset])<<ByteShift8 | uint16(seg.Data[offset+1])
	}
	return value, nil
}

// WriteHalfword writes a 16-bit halfword to memory
func (m *Memory) WriteHalfword(address uint32, value uint16) error {
	if err := m.checkAlignment(address, AlignmentHalfword); err != nil {
		return err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermWrite == 0 {
		return fmt.Errorf("write permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset+1 >= segLen {
		return fmt.Errorf("halfword write exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.WriteCount++

	if m.LittleEndian {
		seg.Data[offset] = byte(value)        // #nosec G115 -- intentional byte extraction from uint16
		seg.Data[offset+1] = byte(value >> 8) // #nosec G115 -- intentional byte extraction from uint16
	} else {
		seg.Data[offset] = byte(value >> 8) // #nosec G115 -- intentional byte extraction from uint16
		seg.Data[offset+1] = byte(value)    // #nosec G115 -- intentional byte extraction from uint16
	}
	return nil
}

// ReadWord reads a 32-bit word from memory
func (m *Memory) ReadWord(address uint32) (uint32, error) {
	if err := m.checkAlignment(address, AlignmentWord); err != nil {
		return 0, err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return 0, err
	}

	if seg.Permissions&PermRead == 0 {
		return 0, fmt.Errorf("read permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset+3 >= segLen {
		return 0, fmt.Errorf("word read exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.ReadCount++

	var value uint32
	if m.LittleEndian {
		value = uint32(seg.Data[offset]) |
			uint32(seg.Data[offset+1])<<ByteShift8 |
			uint32(seg.Data[offset+2])<<ByteShift16 |
			uint32(seg.Data[offset+3])<<ByteShift24
	} else {
		value = uint32(seg.Data[offset])<<ByteShift24 |
			uint32(seg.Data[offset+1])<<ByteShift16 |
			uint32(seg.Data[offset+2])<<ByteShift8 |
			uint32(seg.Data[offset+3])
	}
	return value, nil
}

// WriteWord writes a 32-bit word to memory
func (m *Memory) WriteWord(address uint32, value uint32) error {
	if err := m.checkAlignment(address, AlignmentWord); err != nil {
		return err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermWrite == 0 {
		return fmt.Errorf("write permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset+3 >= segLen {
		return fmt.Errorf("word write exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.WriteCount++

	if m.LittleEndian {
		seg.Data[offset] = byte(value)         // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+1] = byte(value >> 8)  // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+2] = byte(value >> 16) // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+3] = byte(value >> 24) // #nosec G115 -- intentional byte extraction from uint32
	} else {
		seg.Data[offset] = byte(value >> 24)   // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+1] = byte(value >> 16) // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+2] = byte(value >> 8)  // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+3] = byte(value)       // #nosec G115 -- intentional byte extraction from uint32
	}
	return nil
}

// LoadBytes loads a byte array into memory at the specified address
func (m *Memory) LoadBytes(address uint32, data []byte) error {
	for i, b := range data {
		offset, err := SafeIntToUint32(i)
		if err != nil {
			return fmt.Errorf("offset too large at index %d: %w", i, err)
		}
		if err := m.WriteByteAt(address+offset, b); err != nil {
			return fmt.Errorf("failed to load byte at offset %d: %w", i, err)
		}
	}
	return nil
}

// LoadBytesUnsafe loads a byte array into memory bypassing permission checks
// This should be used for initial program loading where we need to write to
// read-only segments (e.g., loading .word data into the code segment)
func (m *Memory) LoadBytesUnsafe(address uint32, data []byte) error {
	for i, b := range data {
		offset, err := SafeIntToUint32(i)
		if err != nil {
			return fmt.Errorf("offset too large at index %d: %w", i, err)
		}
		if err := m.WriteByteUnsafe(address+offset, b); err != nil {
			return fmt.Errorf("failed to load byte at offset %d: %w", i, err)
		}
	}
	return nil
}

// WriteByteUnsafe writes a byte to memory bypassing permission checks (for program loading)
func (m *Memory) WriteByteUnsafe(address uint32, value byte) error {
	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset >= segLen {
		return fmt.Errorf("write beyond segment bounds at 0x%08X", address)
	}

	seg.Data[offset] = value
	return nil
}

// WriteWordUnsafe writes a 32-bit word to memory bypassing permission checks (for program loading)
func (m *Memory) WriteWordUnsafe(address uint32, value uint32) error {
	if err := m.checkAlignment(address, 4); err != nil {
		return err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	segLen, err := SafeIntToUint32(len(seg.Data))
	if err != nil || offset+3 >= segLen {
		return fmt.Errorf("write beyond segment bounds at 0x%08X", address)
	}

	m.WriteCount++

	// Write word in appropriate endianness
	if m.LittleEndian {
		seg.Data[offset] = byte(value)         // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+1] = byte(value >> 8)  // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+2] = byte(value >> 16) // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+3] = byte(value >> 24) // #nosec G115 -- intentional byte extraction from uint32
	} else {
		seg.Data[offset] = byte(value >> 24)   // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+1] = byte(value >> 16) // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+2] = byte(value >> 8)  // #nosec G115 -- intentional byte extraction from uint32
		seg.Data[offset+3] = byte(value)       // #nosec G115 -- intentional byte extraction from uint32
	}
	return nil
}

// GetBytes retrieves a byte array from memory
func (m *Memory) GetBytes(address uint32, length uint32) ([]byte, error) {
	result := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b, err := m.ReadByteAt(address + i)
		if err != nil {
			return nil, fmt.Errorf("failed to read byte at offset %d: %w", i, err)
		}
		result[i] = b
	}
	return result, nil
}

// Reset returns memory to the standard layout: it drops added segments, restores
// segment permissions and zeroes every byte.
func (m *Memory) Reset() {
	n := len(standardSegments)
	clear(m.Segments[n:])
	m.Segments = m.Segments[:n]
	for i, seg := range m.Segments {
		clear(seg.Data)
		seg.Permissions = standardSegments[i].permissions
	}
	m.AccessCount = 0
	m.ReadCount = 0
	m.WriteCount = 0
	m.ResetHeap()
}

// CheckExecutePermission checks if an address has execute permission
func (m *Memory) CheckExecutePermission(address uint32) error {
	seg, _, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermExecute == 0 {
		return fmt.Errorf("execute permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}
	return nil
}

// MakeCodeReadOnly locks the code segment to prevent writes after loading
func (m *Memory) MakeCodeReadOnly() {
	for _, seg := range m.Segments {
		if seg.Name == "code" {
			seg.Permissions = PermRead | PermExecute
		}
	}
}

// Heap allocation tracking
type HeapAllocation struct {
	Address uint32
	Size    uint32
}

// Allocate allocates memory from the heap
func (m *Memory) Allocate(size uint32) (uint32, error) {
	if size == 0 {
		return 0, fmt.Errorf("cannot allocate 0 bytes")
	}

	// Check for overflow BEFORE alignment (alignment can overflow too)
	// If size > Address32BitMaxSafe, alignment will overflow
	if size > Address32BitMaxSafe {
		return 0, fmt.Errorf("allocation size too large (would overflow during alignment)")
	}

	// Align to 4-byte boundary (round up) with overflow check
	if size&AlignMaskWord != 0 {
		aligned := (size + AlignMaskWord) & AlignRoundUpMaskWord
		// Check for wraparound: if aligned < size, overflow occurred
		// This happens when size+3 exceeds 0xFFFFFFFF
		if aligned < size {
			return 0, fmt.Errorf("allocation size causes overflow during alignment")
		}
		size = aligned
	}

	addr, ok := m.takeFreeBlock(size)
	if !ok {
		// Check for overflow in m.NextHeapAddress + size
		if size > Address32BitMax-m.NextHeapAddress {
			return 0, fmt.Errorf("allocation size causes address overflow")
		}
		if m.NextHeapAddress+size > HeapSegmentStart+HeapSegmentSize {
			return 0, fmt.Errorf("out of heap memory")
		}
		addr = m.NextHeapAddress
		m.NextHeapAddress += size
	}

	// Track allocation
	m.HeapAllocations[addr] = &HeapAllocation{
		Address: addr,
		Size:    size,
	}

	// Zero the allocated memory
	for i := uint32(0); i < size; i++ {
		_ = m.WriteByteAt(addr+i, 0) // Ignore error - address is guaranteed valid
	}

	return addr, nil
}

// Free frees previously allocated memory
func (m *Memory) Free(address uint32) error {
	if address == 0 {
		return nil // Freeing NULL is a no-op
	}

	alloc, ok := m.HeapAllocations[address]
	if !ok {
		return fmt.Errorf("invalid free: address 0x%08X was not allocated", address)
	}

	// Remove from tracking
	delete(m.HeapAllocations, address)

	// Zero the freed memory (helps catch use-after-free)
	for i := uint32(0); i < alloc.Size; i++ {
		_ = m.WriteByteAt(address+i, 0) // Ignore error - address is guaranteed valid
	}

	m.releaseBlock(*alloc)
	return nil
}

// takeFreeBlock carves size bytes from the first free block large enough.
func (m *Memory) takeFreeBlock(size uint32) (uint32, bool) {
	i := slices.IndexFunc(m.freeBlocks, func(b HeapAllocation) bool { return b.Size >= size })
	if i < 0 {
		return 0, false
	}
	block := &m.freeBlocks[i]
	addr := block.Address
	if block.Size == size {
		m.freeBlocks = slices.Delete(m.freeBlocks, i, i+1)
	} else {
		block.Address += size
		block.Size -= size
	}
	return addr, true
}

// releaseBlock adds a freed block to the free list, merging it with its neighbours
// and with the unallocated top.
func (m *Memory) releaseBlock(block HeapAllocation) {
	i, _ := slices.BinarySearchFunc(m.freeBlocks, block.Address, func(b HeapAllocation, addr uint32) int {
		return cmp.Compare(b.Address, addr)
	})
	m.freeBlocks = slices.Insert(m.freeBlocks, i, block)

	if i+1 < len(m.freeBlocks) && adjacent(m.freeBlocks[i], m.freeBlocks[i+1]) {
		m.freeBlocks[i].Size += m.freeBlocks[i+1].Size
		m.freeBlocks = slices.Delete(m.freeBlocks, i+1, i+2)
	}
	if i > 0 && adjacent(m.freeBlocks[i-1], m.freeBlocks[i]) {
		m.freeBlocks[i-1].Size += m.freeBlocks[i].Size
		m.freeBlocks = slices.Delete(m.freeBlocks, i, i+1)
	}

	if last := len(m.freeBlocks) - 1; last >= 0 && m.freeBlocks[last].Address+m.freeBlocks[last].Size == m.NextHeapAddress {
		m.NextHeapAddress = m.freeBlocks[last].Address
		m.freeBlocks = m.freeBlocks[:last]
	}
}

// adjacent reports whether block b starts where block a ends.
func adjacent(a, b HeapAllocation) bool {
	return a.Address+a.Size == b.Address
}

// ResetHeap resets the heap allocator
func (m *Memory) ResetHeap() {
	m.HeapAllocations = make(map[uint32]*HeapAllocation)
	m.NextHeapAddress = HeapSegmentStart
	m.freeBlocks = nil
}
