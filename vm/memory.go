package vm

import (
	"fmt"
)

// Memory segments
const (
	CodeSegmentStart  = 0x00008000 // 32KB offset
	CodeSegmentSize   = 0x00010000 // 64KB
	DataSegmentStart  = 0x00020000
	DataSegmentSize   = 0x00010000 // 64KB
	HeapSegmentStart  = 0x00030000
	HeapSegmentSize   = 0x00010000 // 64KB
	StackSegmentStart = 0x00040000
	StackSegmentSize  = 0x00010000 // 64KB
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
	Segments     []*MemorySegment
	LittleEndian bool
	StrictAlign  bool
	AccessCount  uint64
	ReadCount    uint64
	WriteCount   uint64
}

// NewMemory creates and initializes a new Memory instance
func NewMemory() *Memory {
	m := &Memory{
		Segments:     make([]*MemorySegment, 0),
		LittleEndian: true,
		StrictAlign:  true,
	}

	// Initialize standard memory segments
	m.AddSegment("code", CodeSegmentStart, CodeSegmentSize, PermRead|PermExecute)
	m.AddSegment("data", DataSegmentStart, DataSegmentSize, PermRead|PermWrite)
	m.AddSegment("heap", HeapSegmentStart, HeapSegmentSize, PermRead|PermWrite)
	m.AddSegment("stack", StackSegmentStart, StackSegmentSize, PermRead|PermWrite)

	return m
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
		if address >= seg.Start && address < seg.Start+seg.Size {
			offset := address - seg.Start
			return seg, offset, nil
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
	case 4: // Word access
		if address&0x3 != 0 {
			return fmt.Errorf("unaligned word access at 0x%08X (must be 4-byte aligned)", address)
		}
	case 2: // Halfword access
		if address&0x1 != 0 {
			return fmt.Errorf("unaligned halfword access at 0x%08X (must be 2-byte aligned)", address)
		}
	case 1: // Byte access - no alignment required
	default:
		return fmt.Errorf("invalid memory access size: %d", size)
	}
	return nil
}

// ReadByte reads a single byte from memory
func (m *Memory) ReadByte(address uint32) (byte, error) {
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

// WriteByte writes a single byte to memory
func (m *Memory) WriteByte(address uint32, value byte) error {
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
	if err := m.checkAlignment(address, 2); err != nil {
		return 0, err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return 0, err
	}

	if seg.Permissions&PermRead == 0 {
		return 0, fmt.Errorf("read permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	if offset+1 >= uint32(len(seg.Data)) {
		return 0, fmt.Errorf("halfword read exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.ReadCount++

	var value uint16
	if m.LittleEndian {
		value = uint16(seg.Data[offset]) | uint16(seg.Data[offset+1])<<8
	} else {
		value = uint16(seg.Data[offset])<<8 | uint16(seg.Data[offset+1])
	}
	return value, nil
}

// WriteHalfword writes a 16-bit halfword to memory
func (m *Memory) WriteHalfword(address uint32, value uint16) error {
	if err := m.checkAlignment(address, 2); err != nil {
		return err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermWrite == 0 {
		return fmt.Errorf("write permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	if offset+1 >= uint32(len(seg.Data)) {
		return fmt.Errorf("halfword write exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.WriteCount++

	if m.LittleEndian {
		seg.Data[offset] = byte(value)
		seg.Data[offset+1] = byte(value >> 8)
	} else {
		seg.Data[offset] = byte(value >> 8)
		seg.Data[offset+1] = byte(value)
	}
	return nil
}

// ReadWord reads a 32-bit word from memory
func (m *Memory) ReadWord(address uint32) (uint32, error) {
	if err := m.checkAlignment(address, 4); err != nil {
		return 0, err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return 0, err
	}

	if seg.Permissions&PermRead == 0 {
		return 0, fmt.Errorf("read permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	if offset+3 >= uint32(len(seg.Data)) {
		return 0, fmt.Errorf("word read exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.ReadCount++

	var value uint32
	if m.LittleEndian {
		value = uint32(seg.Data[offset]) |
			uint32(seg.Data[offset+1])<<8 |
			uint32(seg.Data[offset+2])<<16 |
			uint32(seg.Data[offset+3])<<24
	} else {
		value = uint32(seg.Data[offset])<<24 |
			uint32(seg.Data[offset+1])<<16 |
			uint32(seg.Data[offset+2])<<8 |
			uint32(seg.Data[offset+3])
	}
	return value, nil
}

// WriteWord writes a 32-bit word to memory
func (m *Memory) WriteWord(address uint32, value uint32) error {
	if err := m.checkAlignment(address, 4); err != nil {
		return err
	}

	seg, offset, err := m.findSegment(address)
	if err != nil {
		return err
	}

	if seg.Permissions&PermWrite == 0 {
		return fmt.Errorf("write permission denied for segment '%s' at 0x%08X", seg.Name, address)
	}

	if offset+3 >= uint32(len(seg.Data)) {
		return fmt.Errorf("word write exceeds segment bounds at 0x%08X", address)
	}

	m.AccessCount++
	m.WriteCount++

	if m.LittleEndian {
		seg.Data[offset] = byte(value)
		seg.Data[offset+1] = byte(value >> 8)
		seg.Data[offset+2] = byte(value >> 16)
		seg.Data[offset+3] = byte(value >> 24)
	} else {
		seg.Data[offset] = byte(value >> 24)
		seg.Data[offset+1] = byte(value >> 16)
		seg.Data[offset+2] = byte(value >> 8)
		seg.Data[offset+3] = byte(value)
	}
	return nil
}

// LoadBytes loads a byte array into memory at the specified address
func (m *Memory) LoadBytes(address uint32, data []byte) error {
	for i, b := range data {
		if err := m.WriteByte(address+uint32(i), b); err != nil {
			return fmt.Errorf("failed to load byte at offset %d: %w", i, err)
		}
	}
	return nil
}

// GetBytes retrieves a byte array from memory
func (m *Memory) GetBytes(address uint32, length uint32) ([]byte, error) {
	result := make([]byte, length)
	for i := uint32(0); i < length; i++ {
		b, err := m.ReadByte(address + i)
		if err != nil {
			return nil, fmt.Errorf("failed to read byte at offset %d: %w", i, err)
		}
		result[i] = b
	}
	return result, nil
}

// Reset clears all memory segments
func (m *Memory) Reset() {
	for _, seg := range m.Segments {
		for i := range seg.Data {
			seg.Data[i] = 0
		}
	}
	m.AccessCount = 0
	m.ReadCount = 0
	m.WriteCount = 0
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
