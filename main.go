package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lookbusy1344/arm-emulator/config"
	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/encoder"
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

		// Tracing and statistics flags
		enableTrace    = flag.Bool("trace", false, "Enable execution trace")
		traceFile      = flag.String("trace-file", "", "Trace output file (default: trace.log in log dir)")
		traceFilter    = flag.String("trace-filter", "", "Filter trace by registers (comma-separated, e.g., R0,R1,PC)")
		enableMemTrace = flag.Bool("mem-trace", false, "Enable memory access trace")
		memTraceFile   = flag.String("mem-trace-file", "", "Memory trace output file (default: memtrace.log)")
		enableStats    = flag.Bool("stats", false, "Enable performance statistics")
		statsFile      = flag.String("stats-file", "", "Statistics output file (default: stats.json)")
		statsFormat    = flag.String("stats-format", "json", "Statistics format (json, csv, html)")
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
	// If entry point is default and program has .org directive, use that
	if *entryPoint == "0x8000" && program.Origin != 0 {
		entryAddr = program.Origin
		if *verboseMode {
			fmt.Printf("Using .org directive address: 0x%08X\n", entryAddr)
		}
	} else {
		if _, err := fmt.Sscanf(*entryPoint, "0x%x", &entryAddr); err != nil {
			if _, err := fmt.Sscanf(*entryPoint, "%d", &entryAddr); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid entry point: %s\n", *entryPoint)
				os.Exit(1)
			}
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

	// Setup tracing and statistics (Phase 10)
	if *enableTrace {
		// Determine trace file path
		tracePath := *traceFile
		if tracePath == "" {
			tracePath = filepath.Join(config.GetLogPath(), "trace.log")
		}

		traceWriter, err := os.Create(tracePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating trace file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := traceWriter.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close trace file: %v\n", err)
			}
		}()

		machine.ExecutionTrace = vm.NewExecutionTrace(traceWriter)
		machine.ExecutionTrace.Start()

		// Apply filter if specified
		if *traceFilter != "" {
			regs := strings.Split(*traceFilter, ",")
			machine.ExecutionTrace.SetFilterRegisters(regs)
		}

		if *verboseMode {
			fmt.Printf("Execution trace enabled: %s\n", tracePath)
		}
	}

	if *enableMemTrace {
		// Determine memory trace file path
		memTracePath := *memTraceFile
		if memTracePath == "" {
			memTracePath = filepath.Join(config.GetLogPath(), "memtrace.log")
		}

		memTraceWriter, err := os.Create(memTracePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating memory trace file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := memTraceWriter.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close memory trace file: %v\n", err)
			}
		}()

		machine.MemoryTrace = vm.NewMemoryTrace(memTraceWriter)
		machine.MemoryTrace.Start()

		if *verboseMode {
			fmt.Printf("Memory trace enabled: %s\n", memTracePath)
		}
	}

	if *enableStats {
		machine.Statistics = vm.NewPerformanceStatistics()
		machine.Statistics.Start()

		if *verboseMode {
			fmt.Println("Performance statistics enabled")
		}
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

		// Flush traces and export statistics (Phase 10)
		if machine.ExecutionTrace != nil {
			if err := machine.ExecutionTrace.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "Error flushing execution trace: %v\n", err)
			}
			if *verboseMode {
				fmt.Printf("Execution trace written (%d entries)\n", len(machine.ExecutionTrace.GetEntries()))
			}
		}

		if machine.MemoryTrace != nil {
			if err := machine.MemoryTrace.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "Error flushing memory trace: %v\n", err)
			}
			if *verboseMode {
				fmt.Printf("Memory trace written (%d entries)\n", len(machine.MemoryTrace.GetEntries()))
			}
		}

		if machine.Statistics != nil {
			// Determine stats file path
			statPath := *statsFile
			if statPath == "" {
				ext := "json"
				if *statsFormat == "csv" {
					ext = "csv"
				} else if *statsFormat == "html" {
					ext = "html"
				}
				statPath = filepath.Join(config.GetLogPath(), "stats."+ext)
			}

			statsWriter, err := os.Create(statPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating statistics file: %v\n", err)
			} else {
				defer func() {
					if err := statsWriter.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to close statistics file: %v\n", err)
					}
				}()

				switch *statsFormat {
				case "json":
					err = machine.Statistics.ExportJSON(statsWriter)
				case "csv":
					err = machine.Statistics.ExportCSV(statsWriter)
				case "html":
					err = machine.Statistics.ExportHTML(statsWriter)
				default:
					err = machine.Statistics.ExportJSON(statsWriter)
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error exporting statistics: %v\n", err)
				} else if *verboseMode {
					fmt.Printf("Statistics exported: %s\n", statPath)
				}
			}

			// Also print summary if verbose
			if *verboseMode {
				fmt.Println()
				fmt.Println(machine.Statistics.String())
			}
		}

		os.Exit(int(machine.ExitCode))
	}
}

// processEscapeSequences processes escape sequences in a string
func processEscapeSequences(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			// Process escape sequence
			switch s[i+1] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '0':
				result = append(result, '\x00')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			case 'a':
				result = append(result, '\a')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'v':
				result = append(result, '\v')
			default:
				// Unknown escape sequence, keep as is
				result = append(result, s[i])
				result = append(result, s[i+1])
			}
			i += 2
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

// loadProgramIntoVM loads a parsed program into the VM's memory
func loadProgramIntoVM(machine *vm.VM, program *parser.Program, entryPoint uint32) error {
	currentAddr := entryPoint

	// Create encoder
	enc := encoder.NewEncoder(program.SymbolTable)

	// Build address map for instructions and directives
	// First pass: calculate addresses
	addressMap := make(map[*parser.Instruction]uint32)
	dataAddr := currentAddr

	for _, inst := range program.Instructions {
		addressMap[inst] = dataAddr

		// Update symbol table with actual load address for instruction labels
		if inst.Label != "" {
			if err := program.SymbolTable.UpdateAddress(inst.Label, dataAddr); err != nil {
				return fmt.Errorf("failed to update instruction label %s address: %w", inst.Label, err)
			}
		}

		dataAddr += 4 // Each instruction is 4 bytes
	}

	// Data directives go after instructions (dataAddr is now positioned after all instructions)

	// Process data directives
	for _, directive := range program.Directives {
		// Update label address in symbol table if this directive has a label
		if directive.Label != "" {
			if err := program.SymbolTable.UpdateAddress(directive.Label, dataAddr); err != nil {
				return fmt.Errorf("failed to update label %s address: %w", directive.Label, err)
			}
		}

		switch directive.Name {
		case ".org":
			// .org directive is handled at parse time, skip it here
			continue

		case ".align":
			// Align to power of 2 (e.g., .align 2 means align to 2^2 = 4 bytes)
			if len(directive.Args) > 0 {
				var alignPower uint32
				if _, err := fmt.Sscanf(directive.Args[0], "0x%x", &alignPower); err != nil {
					if _, err := fmt.Sscanf(directive.Args[0], "%d", &alignPower); err != nil {
						return fmt.Errorf("invalid .align value: %s", directive.Args[0])
					}
				}
				alignBytes := uint32(1 << alignPower) // 2^alignPower
				mask := alignBytes - 1
				dataAddr = (dataAddr + mask) & ^mask
			}

		case ".balign":
			// Align to specified boundary
			if len(directive.Args) > 0 {
				var align uint32
				if _, err := fmt.Sscanf(directive.Args[0], "0x%x", &align); err != nil {
					if _, err := fmt.Sscanf(directive.Args[0], "%d", &align); err != nil {
						return fmt.Errorf("invalid .balign value: %s", directive.Args[0])
					}
				}
				if dataAddr%align != 0 {
					dataAddr += align - (dataAddr % align)
				}
			}

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

		case ".byte":
			// Write bytes
			for _, arg := range directive.Args {
				var value uint32
				if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						return fmt.Errorf("invalid .byte value: %s", arg)
					}
				}
				if err := machine.Memory.WriteByteUnsafe(dataAddr, byte(value)); err != nil {
					return err
				}
				dataAddr++
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
				processedStr := processEscapeSequences(str)
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
		}
	}

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

	// Set PC to entry point
	machine.CPU.PC = entryPoint

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

Tracing & Performance Options:
  -trace             Enable execution trace
  -trace-file FILE   Trace output file (default: trace.log in log dir)
  -trace-filter REGS Filter trace by registers (e.g., R0,R1,PC)
  -mem-trace         Enable memory access trace
  -mem-trace-file F  Memory trace file (default: memtrace.log)
  -stats             Enable performance statistics
  -stats-file FILE   Statistics output file (default: stats.json)
  -stats-format FMT  Statistics format: json, csv, html (default: json)

Examples:
  # Run a program directly
  arm-emulator examples/hello.s

  # Run with debugger
  arm-emulator -debug examples/fibonacci.s

  # Run with TUI debugger
  arm-emulator -tui examples/bubble_sort.s

  # Run with custom settings
  arm-emulator -max-cycles 5000000 -entry 0x10000 program.s

  # Run with execution trace
  arm-emulator -trace -trace-filter "R0,R1,PC" examples/factorial.s

  # Run with performance statistics
  arm-emulator -stats -stats-format html program.s

  # Run with all monitoring enabled
  arm-emulator -trace -mem-trace -stats -verbose program.s

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
