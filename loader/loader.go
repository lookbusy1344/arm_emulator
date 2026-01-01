package loader

import (
	"fmt"
	"os"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// LoadProgramIntoVM loads a parsed assembly program into the VM's memory.
// It creates necessary memory segments, processes data directives, encodes instructions,
// and sets up the entry point.
func LoadProgramIntoVM(machine *vm.VM, program *parser.Program, entryPoint uint32) error {
	// Ensure memory segment exists for the entry point
	// Check if entry point falls outside standard segments
	if entryPoint < vm.CodeSegmentStart {
		// Create a low memory segment for programs using .org 0x0000 or similar
		segmentSize := uint32(vm.CodeSegmentStart) // Cover 0x0000 to 0x8000
		machine.Memory.AddSegment("low-memory", 0, segmentSize, vm.PermRead|vm.PermWrite|vm.PermExecute)
	}

	// Create encoder
	enc := encoder.NewEncoder(program.SymbolTable)

	// Track the maximum address used for literal pool placement
	maxAddr := entryPoint

	// Build address map for instructions using parser-calculated addresses
	// The parser has already correctly calculated addresses accounting for
	// the interleaved layout of instructions and directives
	addressMap := make(map[*parser.Instruction]uint32)

	for _, inst := range program.Instructions {
		addressMap[inst] = inst.Address
		instEnd := inst.Address + 4
		if instEnd > maxAddr {
			maxAddr = instEnd
		}
	}

	// Process data directives using parser-calculated addresses
	for _, directive := range program.Directives {
		dataAddr := directive.Address

		switch directive.Name {
		case ".org":
			// .org directive is handled at parse time, skip it here
			continue

		case ".align":
			// Alignment is already handled by parser in directive.Address
			continue

		case ".balign":
			// Alignment is already handled by parser in directive.Address
			continue

		case ".word":
			// Write 32-bit words
			for _, arg := range directive.Args {
				var value uint32
				// Try to parse as a number first
				if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						// Not a number, try to look up as a symbol (label)
						symValue, symErr := program.SymbolTable.Get(arg)
						if symErr != nil {
							return fmt.Errorf("invalid .word value %q: %w", arg, symErr)
						}
						value = symValue
					}
				}
				if err := machine.Memory.WriteWordUnsafe(dataAddr, value); err != nil {
					return err
				}
				dataAddr += 4
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".byte":
			// Write bytes
			for _, arg := range directive.Args {
				var value uint32
				// Check for character literal: 'A', '\n', '\x41', '\123'
				if len(arg) >= 3 && arg[0] == '\'' && arg[len(arg)-1] == '\'' {
					charContent := arg[1 : len(arg)-1] // Content between quotes
					if len(charContent) == 1 {
						// Simple character: 'A'
						value = uint32(charContent[0])
					} else if len(charContent) >= 2 && charContent[0] == '\\' {
						// Escape sequence: '\n', '\x41', '\123'
						b, _, err := parser.ParseEscapeChar(charContent)
						if err != nil {
							return fmt.Errorf("invalid .byte escape sequence: %s", arg)
						}
						value = uint32(b)
					} else {
						return fmt.Errorf("invalid .byte character literal: %s", arg)
					}
				} else if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						return fmt.Errorf("invalid .byte value: %s", arg)
					}
				}
				if err := machine.Memory.WriteByteUnsafe(dataAddr, byte(value)); err != nil {
					return err
				}
				dataAddr++
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".ascii":
			// Write string without null terminator
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes (parser may have already removed them)
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Process escape sequences
				processedStr := parser.ProcessEscapeSequences(str)
				// Write string bytes
				for i := 0; i < len(processedStr); i++ {
					if err := machine.Memory.WriteByteUnsafe(dataAddr, processedStr[i]); err != nil {
						return fmt.Errorf(".ascii write failed at 0x%08X: %w", dataAddr, err)
					}
					dataAddr++
				}
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".asciz", ".string":
			// Write null-terminated string
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Process escape sequences
				processedStr := parser.ProcessEscapeSequences(str)
				// Write string bytes
				for i := 0; i < len(processedStr); i++ {
					if err := machine.Memory.WriteByteUnsafe(dataAddr, processedStr[i]); err != nil {
						return err
					}
					dataAddr++
				}
				// Write null terminator
				if err := machine.Memory.WriteByteUnsafe(dataAddr, 0); err != nil {
					return err
				}
				dataAddr++
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".space", ".skip":
			// Space is reserved but not written - just track the address
			if len(directive.Args) > 0 {
				var size uint32
				if _, err := fmt.Sscanf(directive.Args[0], "0x%x", &size); err != nil {
					if _, err := fmt.Sscanf(directive.Args[0], "%d", &size); err == nil {
						// Successfully parsed
					}
				}
				endAddr := dataAddr + size
				if endAddr > maxAddr {
					maxAddr = endAddr
				}
			}

		case ".ltorg":
			// Literal pool directive - space will be reserved during encoding
			// The parser has already recorded this location in program.LiteralPoolLocs
			// We don't know yet how many literals will be placed here, so we can't
			// reserve space now. This will be handled after encoding.
			continue
		}
	}

	// Set literal pool start address to after all data
	// Align to 4-byte boundary
	// This is used as a fallback if no .ltorg directives are specified
	literalPoolStart := (maxAddr + 3) & ^uint32(3)
	enc.LiteralPoolStart = literalPoolStart

	// Second pass: encode and write instructions
	for _, inst := range program.Instructions {
		addr := addressMap[inst]

		// Encode instruction
		opcode, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			return fmt.Errorf("failed to encode instruction at 0x%08X (%s): %w", addr, inst.Mnemonic, err)
		}

		// Write to memory
		if err := machine.Memory.WriteWordUnsafe(addr, opcode); err != nil {
			return fmt.Errorf("failed to write instruction at 0x%08X: %w", addr, err)
		}
	}

	// Write any literal pool values generated during encoding
	for addr, value := range enc.LiteralPool {
		if err := machine.Memory.WriteWordUnsafe(addr, value); err != nil {
			return fmt.Errorf("failed to write literal at 0x%08X: %w", addr, err)
		}
	}

	// Validate literal pool capacity and collect warnings
	enc.ValidatePoolCapacity()
	if enc.HasPoolWarnings() && os.Getenv("ARM_WARN_POOLS") != "" {
		for _, warning := range enc.GetPoolWarnings() {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
	}

	// Set PC to entry point and save entry point for debugger resets
	machine.CPU.PC = entryPoint
	machine.EntryPoint = entryPoint

	return nil
}
