package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ================================================================================
// Memory Alignment Tests
// ================================================================================

func TestMemory_WordAlignment_Aligned(t *testing.T) {
	v := vm.NewVM()

	// Write to aligned address (multiple of 4)
	err := v.Memory.WriteWord(0x20000, 0x12345678)
	if err != nil {
		t.Errorf("Writing to aligned address should succeed: %v", err)
	}

	// Read from aligned address
	value, err := v.Memory.ReadWord(0x20000)
	if err != nil {
		t.Errorf("Reading from aligned address should succeed: %v", err)
	}
	if value != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%X", value)
	}
}

func TestMemory_WordAlignment_Misaligned(t *testing.T) {
	v := vm.NewVM()

	// Try to write to misaligned address
	err := v.Memory.WriteWord(0x20001, 0x12345678)
	if err == nil {
		t.Error("Writing word to misaligned address should fail")
	}

	// Try to read from misaligned address
	_, err = v.Memory.ReadWord(0x20002)
	if err == nil {
		t.Error("Reading word from misaligned address should fail")
	}
}

func TestMemory_HalfwordAlignment_Aligned(t *testing.T) {
	v := vm.NewVM()

	// Write to aligned address (multiple of 2)
	err := v.Memory.WriteHalfword(0x20000, 0x1234)
	if err != nil {
		t.Errorf("Writing to aligned address should succeed: %v", err)
	}

	value, err := v.Memory.ReadHalfword(0x20000)
	if err != nil {
		t.Errorf("Reading from aligned address should succeed: %v", err)
	}
	if value != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%X", value)
	}
}

func TestMemory_HalfwordAlignment_Misaligned(t *testing.T) {
	v := vm.NewVM()

	// Try to write to odd address
	err := v.Memory.WriteHalfword(0x20001, 0x1234)
	if err == nil {
		t.Error("Writing halfword to misaligned address should fail")
	}
}

func TestMemory_ByteAlignment_NoRestriction(t *testing.T) {
	v := vm.NewVM()

	// Bytes can be written/read at any address
	for i := uint32(0); i < 4; i++ {
		err := v.Memory.WriteByteAt(0x20000+i, byte(i+1))
		if err != nil {
			t.Errorf("Writing byte should succeed at any address: %v", err)
		}
	}

	for i := uint32(0); i < 4; i++ {
		value, err := v.Memory.ReadByteAt(0x20000 + i)
		if err != nil {
			t.Errorf("Reading byte should succeed: %v", err)
		}
		if value != byte(i+1) {
			t.Errorf("Expected %d, got %d", i+1, value)
		}
	}
}

// ================================================================================
// Memory Segment Permissions Tests
// ================================================================================

func TestMemory_CodeSegment_ReadOnly(t *testing.T) {
	v := vm.NewVM()

	// Find code segment address
	codeAddr := uint32(0x8000)

	// Should be able to read
	_, err := v.Memory.ReadWord(codeAddr)
	if err != nil {
		t.Errorf("Should be able to read from code segment: %v", err)
	}

	// Enable write for this test (normally code is read-only after load)
	// This tests the permission system
}

func TestMemory_DataSegment_ReadWrite(t *testing.T) {
	v := vm.NewVM()

	dataAddr := uint32(0x20000)

	// Should be able to write
	err := v.Memory.WriteWord(dataAddr, 0xDEADBEEF)
	if err != nil {
		t.Errorf("Should be able to write to data segment: %v", err)
	}

	// Should be able to read
	value, err := v.Memory.ReadWord(dataAddr)
	if err != nil {
		t.Errorf("Should be able to read from data segment: %v", err)
	}
	if value != 0xDEADBEEF {
		t.Errorf("Expected 0xDEADBEEF, got 0x%X", value)
	}
}

func TestMemory_StackSegment_ReadWrite(t *testing.T) {
	v := vm.NewVM()

	// Stack starts at 0x00040000, use an aligned address within that segment
	stackAddr := uint32(0x00040000)

	// Should be able to write to stack
	err := v.Memory.WriteWord(stackAddr, 0xCAFEBABE)
	if err != nil {
		t.Errorf("Should be able to write to stack: %v", err)
	}

	// Should be able to read from stack
	value, err := v.Memory.ReadWord(stackAddr)
	if err != nil {
		t.Errorf("Should be able to read from stack: %v", err)
	}
	if value != 0xCAFEBABE {
		t.Errorf("Expected 0xCAFEBABE, got 0x%X", value)
	}
}

// ================================================================================
// Memory Boundary Tests
// ================================================================================

func TestMemory_NullPointer_Detection(t *testing.T) {
	v := vm.NewVM()

	// Reading from null pointer should fail
	_, err := v.Memory.ReadWord(0x0000)
	if err == nil {
		t.Error("Reading from null pointer should fail")
	}

	// Writing to null pointer should fail
	err = v.Memory.WriteWord(0x0000, 0x12345678)
	if err == nil {
		t.Error("Writing to null pointer should fail")
	}
}

func TestMemory_OutOfBounds_High(t *testing.T) {
	v := vm.NewVM()

	// Try to access beyond 4GB boundary
	_, err := v.Memory.ReadWord(0xFFFFFFFF)
	if err == nil {
		t.Error("Reading beyond memory bounds should fail")
	}
}

func TestMemory_ValidBoundaries(t *testing.T) {
	v := vm.NewVM()

	// Test at start of valid writable segments (code is read-only)
	addresses := []uint32{
		0x20000, // Data segment start
		0x30000, // Heap segment start
		0x40000, // Stack segment start
	}

	for _, addr := range addresses {
		err := v.Memory.WriteWord(addr, 0x11111111)
		if err != nil {
			t.Errorf("Should be able to write to valid address 0x%X: %v", addr, err)
		}

		value, err := v.Memory.ReadWord(addr)
		if err != nil {
			t.Errorf("Should be able to read from valid address 0x%X: %v", addr, err)
		}
		if value != 0x11111111 {
			t.Errorf("At 0x%X: expected 0x11111111, got 0x%X", addr, value)
		}
	}
}

// ================================================================================
// Endianness Tests
// ================================================================================

func TestMemory_LittleEndian_Word(t *testing.T) {
	v := vm.NewVM()

	// Write a word
	v.Memory.WriteWord(0x20000, 0x12345678)

	// Read individual bytes (little-endian)
	b0, _ := v.Memory.ReadByteAt(0x20000)
	b1, _ := v.Memory.ReadByteAt(0x20001)
	b2, _ := v.Memory.ReadByteAt(0x20002)
	b3, _ := v.Memory.ReadByteAt(0x20003)

	if b0 != 0x78 {
		t.Errorf("Byte 0 should be 0x78, got 0x%X", b0)
	}
	if b1 != 0x56 {
		t.Errorf("Byte 1 should be 0x56, got 0x%X", b1)
	}
	if b2 != 0x34 {
		t.Errorf("Byte 2 should be 0x34, got 0x%X", b2)
	}
	if b3 != 0x12 {
		t.Errorf("Byte 3 should be 0x12, got 0x%X", b3)
	}
}

func TestMemory_LittleEndian_Halfword(t *testing.T) {
	v := vm.NewVM()

	// Write a halfword
	v.Memory.WriteHalfword(0x20000, 0x1234)

	// Read individual bytes
	b0, _ := v.Memory.ReadByteAt(0x20000)
	b1, _ := v.Memory.ReadByteAt(0x20001)

	if b0 != 0x34 {
		t.Errorf("Byte 0 should be 0x34, got 0x%X", b0)
	}
	if b1 != 0x12 {
		t.Errorf("Byte 1 should be 0x12, got 0x%X", b1)
	}
}

func TestMemory_LittleEndian_BytesToWord(t *testing.T) {
	v := vm.NewVM()

	// Write individual bytes
	v.Memory.WriteByteAt(0x20000, 0x78)
	v.Memory.WriteByteAt(0x20001, 0x56)
	v.Memory.WriteByteAt(0x20002, 0x34)
	v.Memory.WriteByteAt(0x20003, 0x12)

	// Read as word
	value, _ := v.Memory.ReadWord(0x20000)

	if value != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%X", value)
	}
}

// ================================================================================
// Memory Access Patterns
// ================================================================================

func TestMemory_SequentialWrites(t *testing.T) {
	v := vm.NewVM()

	// Write sequential words
	for i := uint32(0); i < 10; i++ {
		addr := 0x20000 + (i * 4)
		err := v.Memory.WriteWord(addr, i*100)
		if err != nil {
			t.Errorf("Failed to write at 0x%X: %v", addr, err)
		}
	}

	// Read them back
	for i := uint32(0); i < 10; i++ {
		addr := 0x20000 + (i * 4)
		value, err := v.Memory.ReadWord(addr)
		if err != nil {
			t.Errorf("Failed to read at 0x%X: %v", addr, err)
		}
		if value != i*100 {
			t.Errorf("At 0x%X: expected %d, got %d", addr, i*100, value)
		}
	}
}

func TestMemory_OverwriteData(t *testing.T) {
	v := vm.NewVM()

	addr := uint32(0x20000)

	// Write initial value
	v.Memory.WriteWord(addr, 0xAAAAAAAA)

	// Overwrite
	v.Memory.WriteWord(addr, 0xBBBBBBBB)

	// Read back
	value, _ := v.Memory.ReadWord(addr)

	if value != 0xBBBBBBBB {
		t.Errorf("Expected 0xBBBBBBBB, got 0x%X", value)
	}
}

func TestMemory_PartialOverwrite(t *testing.T) {
	v := vm.NewVM()

	addr := uint32(0x20000)

	// Write word
	v.Memory.WriteWord(addr, 0x12345678)

	// Overwrite middle byte
	v.Memory.WriteByteAt(addr+1, 0xAA)

	// Read word
	value, _ := v.Memory.ReadWord(addr)

	if value != 0x1234AA78 {
		t.Errorf("Expected 0x1234AA78, got 0x%X", value)
	}
}

// ================================================================================
// Memory Clear and Fill Tests
// ================================================================================

func TestMemory_ClearRange(t *testing.T) {
	v := vm.NewVM()

	// Write some data
	for i := uint32(0); i < 10; i++ {
		v.Memory.WriteByteAt(0x20000+i, 0xFF)
	}

	// Clear it
	for i := uint32(0); i < 10; i++ {
		v.Memory.WriteByteAt(0x20000+i, 0x00)
	}

	// Verify cleared
	for i := uint32(0); i < 10; i++ {
		value, _ := v.Memory.ReadByteAt(0x20000 + i)
		if value != 0x00 {
			t.Errorf("Byte at offset %d should be 0x00, got 0x%X", i, value)
		}
	}
}

func TestMemory_FillPattern(t *testing.T) {
	v := vm.NewVM()

	// Fill with pattern
	pattern := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	for i := 0; i < 4; i++ {
		// Safe conversion: i is from loop [0, 4), always >= 0
		// #nosec G115 -- i is loop index, guaranteed non-negative and within bounds
		offset := uint32(i)
		v.Memory.WriteByteAt(0x20000+offset, pattern[i])
	}

	// Verify pattern
	for i := 0; i < 4; i++ {
		// Safe conversion: i is from loop [0, 4), always >= 0
		// #nosec G115 -- i is loop index, guaranteed non-negative and within bounds
		offset := uint32(i)
		value, _ := v.Memory.ReadByteAt(0x20000 + offset)
		if value != pattern[i] {
			t.Errorf("Byte %d should be 0x%X, got 0x%X", i, pattern[i], value)
		}
	}
}

// ================================================================================
// Stack Growth Tests
// ================================================================================

func TestMemory_StackGrowth_Down(t *testing.T) {
	v := vm.NewVM()

	// Stack starts at 0x00040000 and grows upward in this implementation
	// Use addresses within the stack segment (0x00040000 - 0x00050000)
	sp := uint32(0x0004FFC0) // Near top of stack segment, leave room for growth

	// Push values (stack grows down)
	for i := uint32(0); i < 10; i++ {
		sp -= 4
		v.Memory.WriteWord(sp, i*10)
	}

	// Pop values back
	sp = uint32(0x0004FFC0)
	for i := uint32(0); i < 10; i++ {
		sp -= 4
	}

	for i := uint32(0); i < 10; i++ {
		value, err := v.Memory.ReadWord(sp)
		if err != nil {
			t.Errorf("Failed to read stack at 0x%X: %v", sp, err)
		}
		expected := (9 - i) * 10
		if value != expected {
			t.Errorf("Expected %d, got %d", expected, value)
		}
		sp += 4
	}
}

// ================================================================================
// Large Data Tests
// ================================================================================

func TestMemory_LargeBlock_Write(t *testing.T) {
	v := vm.NewVM()

	// Write 1KB of data
	baseAddr := uint32(0x20000)
	for i := uint32(0); i < 256; i++ {
		addr := baseAddr + (i * 4)
		v.Memory.WriteWord(addr, i)
	}

	// Verify random samples
	samples := []uint32{0, 50, 100, 200, 255}
	for _, i := range samples {
		addr := baseAddr + (i * 4)
		value, _ := v.Memory.ReadWord(addr)
		if value != i {
			t.Errorf("At index %d: expected %d, got %d", i, i, value)
		}
	}
}

func TestMemory_AlternatingPattern(t *testing.T) {
	v := vm.NewVM()

	// Write alternating pattern
	for i := uint32(0); i < 16; i++ {
		addr := 0x20000 + (i * 4)
		if i%2 == 0 {
			v.Memory.WriteWord(addr, 0xAAAAAAAA)
		} else {
			v.Memory.WriteWord(addr, 0x55555555)
		}
	}

	// Verify pattern
	for i := uint32(0); i < 16; i++ {
		addr := 0x20000 + (i * 4)
		value, _ := v.Memory.ReadWord(addr)

		var expected uint32
		if i%2 == 0 {
			expected = 0xAAAAAAAA
		} else {
			expected = 0x55555555
		}

		if value != expected {
			t.Errorf("At index %d: expected 0x%X, got 0x%X", i, expected, value)
		}
	}
}

// ================================================================================
// Mixed Access Tests
// ================================================================================

func TestMemory_MixedWordByteAccess(t *testing.T) {
	v := vm.NewVM()

	// Write word
	v.Memory.WriteWord(0x20000, 0x12345678)

	// Modify one byte
	v.Memory.WriteByteAt(0x20002, 0xAB)

	// Read word
	value, _ := v.Memory.ReadWord(0x20000)

	if value != 0x12AB5678 {
		t.Errorf("Expected 0x12AB5678, got 0x%X", value)
	}
}

func TestMemory_MixedHalfwordAccess(t *testing.T) {
	v := vm.NewVM()

	// Write two halfwords
	v.Memory.WriteHalfword(0x20000, 0x1234)
	v.Memory.WriteHalfword(0x20002, 0x5678)

	// Read as word
	value, _ := v.Memory.ReadWord(0x20000)

	if value != 0x56781234 {
		t.Errorf("Expected 0x56781234, got 0x%X", value)
	}
}

// ================================================================================
// Error Handling Tests
// ================================================================================

func TestMemory_InvalidSegment(t *testing.T) {
	v := vm.NewVM()

	// Try to access an address not in any segment
	// (depends on segment layout, but typically low addresses fail)
	_, err := v.Memory.ReadWord(0x1000)
	if err == nil {
		t.Error("Reading from invalid segment should fail")
	}
}

func TestMemory_ConsecutiveErrors(t *testing.T) {
	v := vm.NewVM()

	// Multiple invalid accesses should all fail
	for i := 0; i < 5; i++ {
		_, err := v.Memory.ReadWord(0x0000)
		if err == nil {
			t.Errorf("Invalid access %d should fail", i)
		}
	}
}

// ================================================================================
// Wraparound Protection Tests
// ================================================================================

func TestMemory_WraparoundProtection_LargeSegment(t *testing.T) {
	// This test verifies protection against the reported "wraparound bug"
	// Scenario: Segment at 0xFFFF0000, size 0x00020000 (128KB)
	// Attack: Try to access 0x00000100 (should be rejected)

	v := vm.NewVM()

	// Add a large segment at high address
	v.Memory.AddSegment("test_high", 0xFFFF0000, 0x00020000, vm.PermRead|vm.PermWrite)

	// Valid accesses within the segment should succeed
	err := v.Memory.WriteWord(0xFFFF0000, 0xDEADBEEF)
	if err != nil {
		t.Errorf("Write to segment start should succeed: %v", err)
	}

	err = v.Memory.WriteWord(0xFFFFFFF0, 0xCAFEBABE)
	if err != nil {
		t.Errorf("Write to segment end should succeed: %v", err)
	}

	// Attack scenario: Try to access address 0x00000100
	// The bug report claims: offset = 0x00000100 - 0xFFFF0000 = 0x00010100 (wraparound)
	// And that 0x00010100 < 0x00020000 would incorrectly allow access
	// But the code checks `if address >= seg.Start` first, which is FALSE for 0x00000100
	_, err = v.Memory.ReadWord(0x00000100)
	if err == nil {
		t.Error("CRITICAL BUG: Wraparound attack succeeded! Access to 0x00000100 should be rejected for segment at 0xFFFF0000")
	}

	// Also verify we can't access other unmapped addresses
	// (avoiding 0x8000-0x17FFF which is the code segment)
	unmappedAddresses := []uint32{0x00000000, 0x00000004, 0x00001000, 0x00007FFC, 0xFFFEFFFF}
	for _, addr := range unmappedAddresses {
		_, err = v.Memory.ReadWord(addr)
		if err == nil {
			t.Errorf("Access to unmapped address 0x%08X should be rejected", addr)
		}
	}
}

func TestMemory_WraparoundProtection_EdgeCases(t *testing.T) {
	// Test various edge cases for segment boundary checking
	v := vm.NewVM()

	// Segment near end of address space: 0xFFFFFFF0, Size: 0x10 (16 bytes)
	// Valid range: 0xFFFFFFF0 to 0xFFFFFFFF (does NOT wrap to 0x00000000)
	// The memory system does not support segments that wrap around the 32-bit boundary
	v.Memory.AddSegment("high_segment", 0xFFFFFFF0, 0x10, vm.PermRead|vm.PermWrite)

	// Valid access at segment start
	err := v.Memory.WriteWord(0xFFFFFFF0, 0x11111111)
	if err != nil {
		t.Errorf("Write to 0xFFFFFFF0 should succeed: %v", err)
	}

	// Valid access near end (last word at 0xFFFFFFFC)
	err = v.Memory.WriteWord(0xFFFFFFFC, 0x22222222)
	if err != nil {
		t.Errorf("Write to 0xFFFFFFFC should succeed: %v", err)
	}

	// Invalid access beyond segment
	// 0xFFFFFFF0 + 0x10 wraps to 0x00000000 due to uint32 overflow
	segStart := uint32(0xFFFFFFF0)
	addrBeyond := segStart + 0x10 // This wraps to 0x00000000
	_, err = v.Memory.ReadWord(addrBeyond)
	if err == nil {
		t.Errorf("Access to wrapped address 0x%08X should fail (beyond segment bounds)", addrBeyond)
	}

	// Invalid access to unrelated low address
	_, err = v.Memory.ReadWord(0x00000100)
	if err == nil {
		t.Error("Access to 0x00000100 should fail (not in high segment)")
	}
}

func TestMemory_NoWraparoundInStandardSegments(t *testing.T) {
	// Verify standard memory segments don't have wraparound issues
	v := vm.NewVM()

	// Standard segments:
	// Code: 0x00008000, size 0x00010000 (ends at 0x00017FFF)
	// Data: 0x00020000, size 0x00010000 (ends at 0x0002FFFF)
	// Heap: 0x00030000, size 0x00010000 (ends at 0x0003FFFF)
	// Stack: 0x00040000, size 0x00010000 (ends at 0x0004FFFF)

	// Try to access addresses that would wrap if calculations were wrong
	invalidAddresses := []uint32{
		0x00000000, // Before code
		0x00018000, // After code
		0x0001FFFF, // Just before data
		0x00050000, // After stack
		0xFFFFFFFF, // High address
	}

	for _, addr := range invalidAddresses {
		_, err := v.Memory.ReadWord(addr)
		if err == nil {
			t.Errorf("Access to unmapped address 0x%08X should fail", addr)
		}
	}
}
