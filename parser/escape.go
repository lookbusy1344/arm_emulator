// Package parser provides assembly parsing functionality for the ARM emulator.
package parser

import (
	"fmt"
	"strconv"
)

// isOctalDigit returns true if the byte is an octal digit (0-7)
func isOctalDigit(b byte) bool {
	return b >= '0' && b <= '7'
}

// ProcessEscapeSequences converts escape sequences in a string to their actual byte values.
// Supports standard C-style escape sequences plus hex escapes (\xNN).
//
// Supported escapes:
//   - \n  newline
//   - \t  tab
//   - \r  carriage return
//   - \\  backslash
//   - \0  null byte (also works as octal 0)
//   - \"  double quote
//   - \'  single quote
//   - \a  alert/bell
//   - \b  backspace
//   - \f  form feed
//   - \v  vertical tab
//   - \xNN  hex byte value (exactly 2 hex digits required)
//   - \NNN octal byte value (1-3 octal digits, value 0-377)
//
// Unknown escape sequences are preserved as-is.
//
// Note: Incomplete hex escapes (e.g., "\x" or "\x1" without full 2 digits)
// are treated as unknown escapes and preserved as-is. Use ParseEscapeChar
// for validation if strict error handling is required.
func ProcessEscapeSequences(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			consumed, b, ok := parseEscapeAt(s, i)
			if ok {
				result = append(result, b...)
				i += consumed
			} else {
				// Unknown escape, keep as-is
				result = append(result, s[i], s[i+1])
				i += 2
			}
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

// ParseEscapeChar parses a single escape sequence and returns the character value.
// This is useful for parsing character literals like '\n' or '\x0A'.
// Returns the character value, the number of characters consumed, and any error.
//
// The input should start with the backslash (e.g., "\\n" or "\\x0A").
func ParseEscapeChar(escape string) (byte, int, error) {
	if len(escape) < 2 || escape[0] != '\\' {
		return 0, 0, fmt.Errorf("invalid escape sequence: %s", escape)
	}

	consumed, bytes, ok := parseEscapeAt(escape, 0)
	if !ok {
		return 0, 0, fmt.Errorf("unknown escape sequence: %s", escape)
	}

	if len(bytes) != 1 {
		return 0, 0, fmt.Errorf("escape sequence must produce single byte: %s", escape)
	}

	return bytes[0], consumed, nil
}

// parseEscapeAt parses an escape sequence starting at position i in string s.
// Returns the number of characters consumed, the resulting byte(s), and success status.
func parseEscapeAt(s string, i int) (int, []byte, bool) {
	if i+1 >= len(s) || s[i] != '\\' {
		return 0, nil, false
	}

	switch s[i+1] {
	case 'n':
		return 2, []byte{'\n'}, true
	case 't':
		return 2, []byte{'\t'}, true
	case 'r':
		return 2, []byte{'\r'}, true
	case '\\':
		return 2, []byte{'\\'}, true
	case '"':
		return 2, []byte{'"'}, true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// Octal escape: \N, \NN, or \NNN (1-3 octal digits)
		return parseOctalEscape(s, i)
	case '\'':
		return 2, []byte{'\''}, true
	case 'a':
		return 2, []byte{'\a'}, true
	case 'b':
		return 2, []byte{'\b'}, true
	case 'f':
		return 2, []byte{'\f'}, true
	case 'v':
		return 2, []byte{'\v'}, true
	case 'x':
		// Hex escape: \xNN (exactly 2 hex digits required)
		// Incomplete hex escapes (e.g., "\x" or "\x1") return false
		// and are preserved as-is by the caller
		if i+3 >= len(s) {
			return 0, nil, false // Not enough characters for \xNN
		}
		hexStr := s[i+2 : i+4]
		val, err := strconv.ParseUint(hexStr, 16, 8)
		if err != nil {
			return 0, nil, false
		}
		return 4, []byte{byte(val)}, true
	default:
		return 0, nil, false
	}
}

// parseOctalEscape parses an octal escape sequence starting at position i.
// Octal sequences are \N, \NN, or \NNN where N is an octal digit (0-7).
// The value must be <= 255 (octal 377).
// Returns the number of characters consumed, the resulting byte, and success status.
func parseOctalEscape(s string, i int) (int, []byte, bool) {
	// We know s[i] == '\\' and s[i+1] is an octal digit
	start := i + 1 // Start after backslash
	end := start

	// Consume up to 3 octal digits
	maxEnd := start + 3
	if maxEnd > len(s) {
		maxEnd = len(s)
	}

	for end < maxEnd && isOctalDigit(s[end]) {
		end++
	}

	// Parse the octal value
	octalStr := s[start:end]
	val, err := strconv.ParseUint(octalStr, 8, 9) // 9 bits to detect overflow > 255
	if err != nil || val > 255 {
		return 0, nil, false
	}

	consumed := 1 + (end - start) // 1 for backslash + octal digits
	return consumed, []byte{byte(val)}, true
}
