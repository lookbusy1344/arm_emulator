package vm

import (
	"fmt"
	"math"
)

// SafeInt32ToUint32 safely converts int32 to uint32
// Returns error if value is negative
func SafeInt32ToUint32(v int32) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("cannot convert negative int32 %d to uint32", v)
	}
	return uint32(v), nil
}

// SafeIntToUint32 safely converts int to uint32
// Returns error if value is negative or exceeds uint32 range
func SafeIntToUint32(v int) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("cannot convert negative int %d to uint32", v)
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("int value %d exceeds uint32 maximum", v)
	}
	return uint32(v), nil
}

// SafeInt64ToUint32 safely converts int64 to uint32
// Returns error if value is negative or exceeds uint32 range
func SafeInt64ToUint32(v int64) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("cannot convert negative int64 %d to uint32", v)
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("int64 value %d exceeds uint32 maximum", v)
	}
	return uint32(v), nil
}

// SafeUint32ToUint16 safely converts uint32 to uint16
// Returns error if value exceeds uint16 range
func SafeUint32ToUint16(v uint32) (uint16, error) {
	if v > math.MaxUint16 {
		return 0, fmt.Errorf("uint32 value 0x%X exceeds uint16 maximum", v)
	}
	return uint16(v), nil
}

// SafeUint32ToUint8 safely converts uint32 to uint8
// Returns error if value exceeds uint8 range
func SafeUint32ToUint8(v uint32) (uint8, error) {
	if v > math.MaxUint8 {
		return 0, fmt.Errorf("uint32 value 0x%X exceeds uint8 maximum", v)
	}
	return uint8(v), nil
}

// SafeUintToUint32 safely converts uint to uint32
// Returns error if value exceeds uint32 range
func SafeUintToUint32(v uint) (uint32, error) {
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("uint value %d exceeds uint32 maximum", v)
	}
	return uint32(v), nil
}

// AsInt32 converts uint32 to int32 for display purposes
// This is intentional for showing the signed interpretation of a uint32 value
// No error checking as the bit pattern is preserved
func AsInt32(v uint32) int32 {
	//nolint:gosec // G115: Intentional conversion for signed display
	return int32(v)
}
