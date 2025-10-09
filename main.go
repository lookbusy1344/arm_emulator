package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	version = "1.0.0"
)

func main() {
	// Command-line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		debugMode   = flag.Bool("debug", false, "Start in debugger mode")
		tuiMode     = flag.Bool("tui", false, "Use TUI (Text User Interface) debugger")
		maxCycles   = flag.Uint64("max-cycles", 1000000, "Maximum CPU cycles before halt")
		stackSize   = flag.Uint("stack-size", vm.StackSegmentSize, "Stack size in bytes")
		entryPoint  = flag.String("entry", "0x8000", "Entry point address (hex or decimal)")
		verboseMode = flag.Bool("verbose", false, "Verbose output")
	)

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("ARM2 Emulator v%s\n", version)
		os.Exit(0)
	}

	// Show help
	if *showHelp || flag.NArg() == 0 {
		printHelp()
		os.Exit(0)
	}

	// Get assembly file from arguments
	asmFile := flag.Arg(0)
	if _, err := os.Stat(asmFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", asmFile)
		os.Exit(1)
	}

	// Read assembly file
	if *verboseMode {
		fmt.Printf("Loading assembly file: %s\n", asmFile)
	}

	input, err := os.ReadFile(asmFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse assembly
	if *verboseMode {
		fmt.Println("Parsing assembly...")
	}

	p := parser.NewParser(string(input), filepath.Base(asmFile))
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error:\n%v\n", err)
		os.Exit(1)
	}

	if *verboseMode {
		fmt.Printf("Parsed %d instructions, %d directives\n",
			len(program.Instructions), len(program.Directives))
	}

	// Create VM instance
	machine := vm.NewVM()
	machine.CycleLimit = *maxCycles

	// Initialize stack
	stackTop := uint32(vm.StackSegmentStart + *stackSize)
	machine.InitializeStack(stackTop)

	// Parse entry point
	var entryAddr uint32
	if _, err := fmt.Sscanf(*entryPoint, "0x%x", &entryAddr); err != nil {
		if _, err := fmt.Sscanf(*entryPoint, "%d", &entryAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid entry point: %s\n", *entryPoint)
			os.Exit(1)
		}
	}

	// Load program into memory
	if *verboseMode {
		fmt.Println("Loading program into memory...")
	}

	err = loadProgramIntoVM(machine, program, entryAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
		os.Exit(1)
	}

	// Create symbol table for debugger
	symbols := make(map[string]uint32)
	sourceMap := make(map[uint32]string)

	for name, symbol := range program.SymbolTable.GetAllSymbols() {
		if symbol.Type == parser.SymbolLabel {
			symbols[name] = symbol.Value
		}
	}

	// Build source map (address -> source line)
	for _, inst := range program.Instructions {
		if inst.Label != "" {
			if addr, exists := symbols[inst.Label]; exists {
				sourceMap[addr] = inst.RawLine
			}
		}
	}

	if *verboseMode {
		fmt.Printf("Entry point: 0x%08X\n", entryAddr)
		fmt.Printf("Stack: 0x%08X - 0x%08X (%d bytes)\n",
			vm.StackSegmentStart, stackTop, *stackSize)
		fmt.Printf("Symbols: %d labels defined\n", len(symbols))
	}

	// Run in appropriate mode
	if *debugMode || *tuiMode {
		// Start debugger
		dbg := debugger.NewDebugger(machine)
		dbg.LoadSymbols(symbols)
		dbg.LoadSourceMap(sourceMap)

		if *tuiMode {
			// Start TUI interface
			if err := debugger.RunTUI(dbg); err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Start command-line debugger
			fmt.Println("ARM2 Debugger - Type 'help' for commands")
			fmt.Printf("Program loaded: %s\n", asmFile)
			fmt.Println()

			if err := debugger.RunCLI(dbg); err != nil {
				fmt.Fprintf(os.Stderr, "Debugger error: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Direct execution mode
		if *verboseMode {
			fmt.Println("\nStarting execution...")
			fmt.Println("----------------------------------------")
		}

		// Run until halt
		machine.State = vm.StateRunning
		for machine.State == vm.StateRunning {
			if err := machine.Step(); err != nil {
				if machine.State == vm.StateHalted {
					// Normal exit
					break
				}
				fmt.Fprintf(os.Stderr, "\nRuntime error at PC=0x%08X: %v\n", machine.CPU.PC, err)
				os.Exit(1)
			}
		}

		if *verboseMode {
			fmt.Println("\n----------------------------------------")
			fmt.Println("Execution complete")
			fmt.Printf("Exit code: %d\n", machine.ExitCode)
			fmt.Printf("CPU cycles: %d\n", machine.CPU.Cycles)
			fmt.Printf("Instructions executed: %d\n", len(machine.InstructionLog))
		}

		os.Exit(int(machine.ExitCode))
	}
}

// loadProgramIntoVM loads a parsed program into the VM's memory
func loadProgramIntoVM(machine *vm.VM, program *parser.Program, entryPoint uint32) error {
	currentAddr := entryPoint

	// First, process directives that allocate data
	for _, directive := range program.Directives {
		switch directive.Name {
		case ".org":
			// Set current address
			if len(directive.Args) > 0 {
				var addr uint32
				if _, err := fmt.Sscanf(directive.Args[0], "0x%x", &addr); err != nil {
					if _, err := fmt.Sscanf(directive.Args[0], "%d", &addr); err != nil {
						return fmt.Errorf("invalid .org address: %s", directive.Args[0])
					}
				}
				currentAddr = addr
			}

		case ".word":
			// Write 32-bit words
			for _, arg := range directive.Args {
				var value uint32
				if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						return fmt.Errorf("invalid .word value: %s", arg)
					}
				}
				if err := machine.Memory.WriteWordUnsafe(currentAddr, value); err != nil {
					return err
				}
				currentAddr += 4
			}

		case ".byte":
			// Write bytes
			for _, arg := range directive.Args {
				var value uint32
				if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						return fmt.Errorf("invalid .byte value: %s", arg)
					}
				}
				if err := machine.Memory.WriteByteUnsafe(currentAddr, byte(value)); err != nil {
					return err
				}
				currentAddr++
			}

		case ".asciz", ".string":
			// Write null-terminated string
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Write string bytes
				for i := 0; i < len(str); i++ {
					if err := machine.Memory.WriteByteUnsafe(currentAddr, str[i]); err != nil {
						return err
					}
					currentAddr++
				}
				// Write null terminator
				if err := machine.Memory.WriteByteUnsafe(currentAddr, 0); err != nil {
					return err
				}
				currentAddr++
			}
		}
	}

	// Now encode and load instructions
	// Note: This is a simplified loader. A full implementation would need
	// to encode the parsed instructions into ARM machine code.
	// For now, we'll need an encoder that converts parser.Instruction -> uint32

	// TODO: Implement instruction encoding
	// This requires translating parsed mnemonics and operands into ARM opcodes

	fmt.Fprintln(os.Stderr, "Warning: Instruction encoding not yet fully implemented")
	fmt.Fprintln(os.Stderr, "The VM infrastructure is ready, but parsed instructions need to be encoded to ARM opcodes")

	return nil
}

func printHelp() {
	fmt.Printf(`ARM2 Emulator v%s

Usage: arm-emulator [options] <assembly-file>

Options:
  -help              Show this help message
  -version           Show version information
  -debug             Start in debugger mode (CLI)
  -tui               Start in TUI debugger mode
  -max-cycles N      Set maximum CPU cycles (default: 1000000)
  -stack-size N      Set stack size in bytes (default: %d)
  -entry ADDR        Set entry point address (default: 0x8000)
  -verbose           Enable verbose output

Examples:
  # Run a program directly
  arm-emulator examples/hello.s

  # Run with debugger
  arm-emulator -debug examples/fibonacci.s

  # Run with TUI debugger
  arm-emulator -tui examples/bubble_sort.s

  # Run with custom settings
  arm-emulator -max-cycles 5000000 -entry 0x10000 program.s

Debugger Commands (when in -debug mode):
  run, r             Start/restart program execution
  continue, c        Continue execution
  step, s            Execute single instruction
  next, n            Step over function calls
  break ADDR         Set breakpoint at address/label
  info registers     Show all registers
  print EXPR         Evaluate and print expression
  help               Show debugger help

For more information, see the README.md file.
`, version, vm.StackSegmentSize)
}
