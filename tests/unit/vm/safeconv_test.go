package vm_test

import (
	"math"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestSafeInt32ToUint32(t *testing.T) {
	tests := []struct {
		input     int32
		expected  uint32
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxInt32, math.MaxInt32, false},
		{-1, 0, true},
		{-100, 0, true},
		{math.MinInt32, 0, true},
	}

	for _, tt := range tests {
		result, err := vm.SafeInt32ToUint32(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeInt32ToUint32(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeInt32ToUint32(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeInt32ToUint32(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeIntToUint32(t *testing.T) {
	tests := []struct {
		input     int
		expected  uint32
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxUint32, math.MaxUint32, false},
		{-1, 0, true},
		{-100, 0, true},
	}

	// Add test for overflow on 64-bit systems
	if math.MaxInt > math.MaxUint32 {
		tests = append(tests, struct {
			input     int
			expected  uint32
			shouldErr bool
		}{math.MaxUint32 + 1, 0, true})
	}

	for _, tt := range tests {
		result, err := vm.SafeIntToUint32(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeIntToUint32(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeIntToUint32(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeIntToUint32(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeInt64ToUint32(t *testing.T) {
	tests := []struct {
		input     int64
		expected  uint32
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxUint32, math.MaxUint32, false},
		{-1, 0, true},
		{-100, 0, true},
		{math.MaxUint32 + 1, 0, true},
		{math.MaxInt64, 0, true},
	}

	for _, tt := range tests {
		result, err := vm.SafeInt64ToUint32(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeInt64ToUint32(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeInt64ToUint32(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeInt64ToUint32(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeUint32ToUint16(t *testing.T) {
	tests := []struct {
		input     uint32
		expected  uint16
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxUint16, math.MaxUint16, false},
		{math.MaxUint16 + 1, 0, true},
		{math.MaxUint32, 0, true},
	}

	for _, tt := range tests {
		result, err := vm.SafeUint32ToUint16(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeUint32ToUint16(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeUint32ToUint16(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeUint32ToUint16(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeUint32ToUint8(t *testing.T) {
	tests := []struct {
		input     uint32
		expected  uint8
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxUint8, math.MaxUint8, false},
		{math.MaxUint8 + 1, 0, true},
		{math.MaxUint32, 0, true},
	}

	for _, tt := range tests {
		result, err := vm.SafeUint32ToUint8(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeUint32ToUint8(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeUint32ToUint8(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeUint32ToUint8(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestSafeUintToUint32(t *testing.T) {
	tests := []struct {
		input     uint
		expected  uint32
		shouldErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{math.MaxUint32, math.MaxUint32, false},
	}

	// Add test for overflow on 64-bit systems
	if ^uint(0) > math.MaxUint32 {
		tests = append(tests, struct {
			input     uint
			expected  uint32
			shouldErr bool
		}{math.MaxUint32 + 1, 0, true})
	}

	for _, tt := range tests {
		result, err := vm.SafeUintToUint32(tt.input)
		if tt.shouldErr {
			if err == nil {
				t.Errorf("vm.SafeUintToUint32(%d) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("vm.SafeUintToUint32(%d) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("vm.SafeUintToUint32(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestAsInt32(t *testing.T) {
	tests := []struct {
		input    uint32
		expected int32
	}{
		{0, 0},
		{1, 1},
		{0x7FFFFFFF, 0x7FFFFFFF},
		{0x80000000, -2147483648}, // Most significant bit set
		{0xFFFFFFFF, -1},
	}

	for _, tt := range tests {
		result := vm.AsInt32(tt.input)
		if result != tt.expected {
			t.Errorf("vm.AsInt32(0x%X) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
