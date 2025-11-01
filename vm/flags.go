package vm

// Flag calculation helpers for ARM2 CPSR flags

// UpdateFlagsNZ updates the N (negative) and Z (zero) flags based on a result
func (c *CPSR) UpdateFlagsNZ(result uint32) {
	c.N = (result & SignBitMask) != 0 // Check bit 31
	c.Z = result == 0
}

// UpdateFlagsNZC updates N, Z, and C flags
func (c *CPSR) UpdateFlagsNZC(result uint32, carry bool) {
	c.UpdateFlagsNZ(result)
	c.C = carry
}

// UpdateFlagsNZCV updates all four condition flags
func (c *CPSR) UpdateFlagsNZCV(result uint32, carry, overflow bool) {
	c.UpdateFlagsNZ(result)
	c.C = carry
	c.V = overflow
}

// CalculateAddCarry calculates the carry flag for addition
// Returns true if unsigned overflow occurred
func CalculateAddCarry(a, b, result uint32) bool {
	// Carry occurs if result < a (unsigned overflow)
	return result < a
}

// CalculateAddOverflow calculates the overflow flag for addition
// Returns true if signed overflow occurred
func CalculateAddOverflow(a, b, result uint32) bool {
	// Overflow occurs when:
	// - Adding two positive numbers yields a negative result
	// - Adding two negative numbers yields a positive result
	// This can be detected by checking if sign bits of operands match
	// but differ from the result's sign bit
	aSign := (a >> SignBitPos) & Mask1Bit
	bSign := (b >> SignBitPos) & Mask1Bit
	resultSign := (result >> SignBitPos) & Mask1Bit

	return (aSign == bSign) && (aSign != resultSign)
}

// CalculateSubCarry calculates the carry flag for subtraction
// In ARM, carry flag is set when no borrow occurs (inverted from typical x86)
// Returns true if NO borrow occurred (a >= b in unsigned arithmetic)
func CalculateSubCarry(a, b uint32) bool {
	// Carry is set if a >= b (no borrow needed)
	return a >= b
}

// CalculateSubOverflow calculates the overflow flag for subtraction
// Returns true if signed overflow occurred
func CalculateSubOverflow(a, b, result uint32) bool {
	// Overflow occurs when:
	// - Subtracting a negative from a positive yields a negative
	// - Subtracting a positive from a negative yields a positive
	aSign := (a >> SignBitPos) & Mask1Bit
	bSign := (b >> SignBitPos) & Mask1Bit
	resultSign := (result >> SignBitPos) & Mask1Bit

	return (aSign != bSign) && (aSign != resultSign)
}

// CalculateShiftCarry calculates the carry flag for shift operations
// Returns the last bit that was shifted out, or current carry if shift amount is 0
func CalculateShiftCarry(value uint32, shiftAmount int, shiftType ShiftType, currentCarry bool) bool {
	// Note: RRX is always a 1-bit shift, don't apply the shiftAmount==0 shortcut
	if shiftAmount == 0 && shiftType != ShiftRRX {
		return currentCarry
	}

	switch shiftType {
	case ShiftLSL: // Logical Shift Left
		if shiftAmount > 32 {
			return false
		}
		if shiftAmount == 32 {
			return (value & 1) != 0
		}
		return (value & (1 << (32 - shiftAmount))) != 0

	case ShiftLSR: // Logical Shift Right
		// In ARM, LSR #0 is encoded to mean LSR #32
		if shiftAmount == 0 {
			return (value & SignBitMask) != 0
		}
		if shiftAmount > 32 {
			return false
		}
		if shiftAmount == 32 {
			return (value & SignBitMask) != 0
		}
		return (value & (1 << (shiftAmount - 1))) != 0

	case ShiftASR: // Arithmetic Shift Right
		// In ARM, ASR #0 is encoded to mean ASR #32
		if shiftAmount == 0 {
			return (value & SignBitMask) != 0
		}
		if shiftAmount >= 32 {
			return (value & SignBitMask) != 0
		}
		return (value & (1 << (shiftAmount - 1))) != 0

	case ShiftROR: // Rotate Right
		shiftAmount = shiftAmount % 32
		if shiftAmount == 0 {
			return currentCarry
		}
		return (value & (1 << (shiftAmount - 1))) != 0

	case ShiftRRX: // Rotate Right Extended
		return (value & 1) != 0
	}

	return currentCarry
}

// ShiftType represents the type of shift operation
type ShiftType int

const (
	ShiftLSL ShiftType = iota // Logical Shift Left
	ShiftLSR                  // Logical Shift Right
	ShiftASR                  // Arithmetic Shift Right
	ShiftROR                  // Rotate Right
	ShiftRRX                  // Rotate Right Extended (with carry)
)

// PerformShift performs a shift operation and returns the result
func PerformShift(value uint32, shiftAmount int, shiftType ShiftType, carry bool) uint32 {
	// Note: In ARM encoding, shift amount 0 has special meanings:
	// - LSR #0 means LSR #32
	// - ASR #0 means ASR #32
	// - ROR #0 means RRX (handled in data_processing.go)
	// - LSL #0 means no shift
	if shiftAmount == 0 && shiftType == ShiftLSL {
		return value
	}

	switch shiftType {
	case ShiftLSL:
		if shiftAmount >= 32 {
			return 0
		}
		return value << shiftAmount

	case ShiftLSR:
		// In ARM, LSR #0 is encoded to mean LSR #32
		if shiftAmount == 0 {
			return 0
		}
		if shiftAmount >= 32 {
			return 0
		}
		return value >> shiftAmount

	case ShiftASR:
		// In ARM, ASR #0 is encoded to mean ASR #32
		if shiftAmount == 0 {
			shiftAmount = 32
		}
		if shiftAmount >= 32 {
			// Arithmetic shift preserves sign bit
			if (value & SignBitMask) != 0 {
				return Mask32Bit
			}
			return 0
		}
		// Perform arithmetic shift (sign extension)
		result := value >> shiftAmount
		if (value & SignBitMask) != 0 {
			// Fill with 1s from the left
			mask := uint32(Mask32Bit << (BitsInWord - shiftAmount))
			result |= mask
		}
		return result

	case ShiftROR:
		shiftAmount = shiftAmount % BitsInWord
		if shiftAmount == 0 {
			return value
		}
		return (value >> shiftAmount) | (value << (BitsInWord - shiftAmount))

	case ShiftRRX:
		// Rotate right by 1 with carry
		result := value >> 1
		if carry {
			result |= SignBitMask
		}
		return result
	}

	return value
}

// EvaluateCondition checks if a condition code is satisfied
func (c *CPSR) EvaluateCondition(cond ConditionCode) bool {
	switch cond {
	case CondEQ: // Equal (Z == 1)
		return c.Z
	case CondNE: // Not Equal (Z == 0)
		return !c.Z
	case CondCS: // Carry Set / Unsigned Higher or Same (C == 1)
		return c.C
	case CondCC: // Carry Clear / Unsigned Lower (C == 0)
		return !c.C
	case CondMI: // Minus / Negative (N == 1)
		return c.N
	case CondPL: // Plus / Positive or Zero (N == 0)
		return !c.N
	case CondVS: // Overflow Set (V == 1)
		return c.V
	case CondVC: // Overflow Clear (V == 0)
		return !c.V
	case CondHI: // Unsigned Higher (C == 1 AND Z == 0)
		return c.C && !c.Z
	case CondLS: // Unsigned Lower or Same (C == 0 OR Z == 1)
		return !c.C || c.Z
	case CondGE: // Signed Greater or Equal (N == V)
		return c.N == c.V
	case CondLT: // Signed Less Than (N != V)
		return c.N != c.V
	case CondGT: // Signed Greater Than (Z == 0 AND N == V)
		return !c.Z && (c.N == c.V)
	case CondLE: // Signed Less or Equal (Z == 1 OR N != V)
		return c.Z || (c.N != c.V)
	case CondAL: // Always
		return true
	case CondNV: // Never (deprecated in ARM2, should warn)
		return false
	}
	return false
}

// ConditionCode represents ARM condition codes
type ConditionCode int

const (
	CondEQ ConditionCode = iota // 0000 - Equal (Z == 1)
	CondNE                      // 0001 - Not Equal (Z == 0)
	CondCS                      // 0010 - Carry Set / HS (Unsigned Higher or Same)
	CondCC                      // 0011 - Carry Clear / LO (Unsigned Lower)
	CondMI                      // 0100 - Minus / Negative
	CondPL                      // 0101 - Plus / Positive or Zero
	CondVS                      // 0110 - Overflow Set
	CondVC                      // 0111 - Overflow Clear
	CondHI                      // 1000 - Unsigned Higher
	CondLS                      // 1001 - Unsigned Lower or Same
	CondGE                      // 1010 - Signed Greater or Equal
	CondLT                      // 1011 - Signed Less Than
	CondGT                      // 1100 - Signed Greater Than
	CondLE                      // 1101 - Signed Less or Equal
	CondAL                      // 1110 - Always (unconditional)
	CondNV                      // 1111 - Never (deprecated)
)

// String returns the string representation of a condition code
func (cc ConditionCode) String() string {
	names := []string{
		"EQ", "NE", "CS", "CC", "MI", "PL", "VS", "VC",
		"HI", "LS", "GE", "LT", "GT", "LE", "AL", "NV",
	}
	if cc >= 0 && int(cc) < len(names) {
		return names[cc]
	}
	return "??"
}

// ParseConditionCode parses a condition code string
func ParseConditionCode(s string) (ConditionCode, bool) {
	conditions := map[string]ConditionCode{
		"EQ": CondEQ, "NE": CondNE,
		"CS": CondCS, "HS": CondCS, // HS is alias for CS
		"CC": CondCC, "LO": CondCC, // LO is alias for CC
		"MI": CondMI, "PL": CondPL,
		"VS": CondVS, "VC": CondVC,
		"HI": CondHI, "LS": CondLS,
		"GE": CondGE, "LT": CondLT,
		"GT": CondGT, "LE": CondLE,
		"AL": CondAL, "": CondAL, // Empty string defaults to AL
		"NV": CondNV,
	}
	cond, ok := conditions[s]
	return cond, ok
}
