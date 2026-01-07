package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
	"github.com/lookbusy1344/arm-emulator/config"
	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Version information - can be overridden at build time with:
// go build -ldflags "-X main.Version=v1.2.3"
var (
	Version = "dev"     // Version number (set by git tag at build time)
	Commit  = "unknown" // Git commit hash
	Date    = "unknown" // Build date
)

func main() {
	// Command-line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		debugMode   = flag.Bool("debug", false, "Start in debugger mode")
		tuiMode     = flag.Bool("tui", false, "Use TUI (Text User Interface) debugger")
		apiServer   = flag.Bool("api-server", false, "Start HTTP API server mode")
		apiPort     = flag.Int("port", 8080, "API server port (used with -api-server)")
		maxCycles   = flag.Uint64("max-cycles", 1000000, "Maximum CPU cycles before halt")
		stackSize   = flag.Uint("stack-size", vm.StackSegmentSize, "Stack size in bytes")
		entryPoint  = flag.String("entry", "0x8000", "Entry point address (hex or decimal)")
		verboseMode = flag.Bool("verbose", false, "Verbose output")
		fsRoot      = flag.String("fsroot", "", "Restrict file operations to this directory (default: current directory)")

		// Tracing and statistics flags
		enableTrace    = flag.Bool("trace", false, "Enable execution trace")
		traceFile      = flag.String("trace-file", "", "Trace output file (default: trace.log in log dir)")
		traceFilter    = flag.String("trace-filter", "", "Filter trace by registers (comma-separated, e.g., R0,R1,PC)")
		enableMemTrace = flag.Bool("mem-trace", false, "Enable memory access trace")
		memTraceFile   = flag.String("mem-trace-file", "", "Memory trace output file (default: memtrace.log)")
		enableStats    = flag.Bool("stats", false, "Enable performance statistics")
		statsFile      = flag.String("stats-file", "", "Statistics output file (default: stats.json)")
		statsFormat    = flag.String("stats-format", "json", "Statistics format (json, csv, html)")

		// Additional diagnostic modes (Phase 11)
		enableCoverage      = flag.Bool("coverage", false, "Enable code coverage tracking")
		coverageFile        = flag.String("coverage-file", "", "Coverage output file (default: coverage.txt)")
		coverageFormat      = flag.String("coverage-format", "text", "Coverage format (text, json)")
		enableStackTrace    = flag.Bool("stack-trace", false, "Enable stack operation tracing")
		stackTraceFile      = flag.String("stack-trace-file", "", "Stack trace output file (default: stack_trace.txt)")
		stackTraceFormat    = flag.String("stack-trace-format", "text", "Stack trace format (text, json)")
		stackGuard          = flag.Bool("stack-guard", false, "Halt execution if stack overflows into heap segment")
		enableFlagTrace     = flag.Bool("flag-trace", false, "Enable CPSR flag change tracing")
		flagTraceFile       = flag.String("flag-trace-file", "", "Flag trace output file (default: flag_trace.txt)")
		flagTraceFormat     = flag.String("flag-trace-format", "text", "Flag trace format (text, json)")
		enableRegisterTrace = flag.Bool("register-trace", false, "Enable register access pattern tracing")
		registerTraceFile   = flag.String("register-trace-file", "", "Register trace output file (default: register_trace.txt)")
		registerTraceFormat = flag.String("register-trace-format", "text", "Register trace format (text, json)")

		// Symbol dump options
		dumpSymbols = flag.Bool("dump-symbols", false, "Dump symbol table and exit")
		symbolsFile = flag.String("symbols-file", "", "Symbol dump output file (default: stdout)")
	)

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("ARM2 Emulator %s\n", Version)
		if Commit != "unknown" {
			fmt.Printf("Commit: %s\n", Commit)
		}
		if Date != "unknown" {
			fmt.Printf("Built: %s\n", Date)
		}
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Start API server mode if requested
	if *apiServer {
		server := api.NewServerWithVersion(*apiPort, Version, Commit, Date)

		// Setup graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Create shutdown function with sync.Once to ensure it runs only once
		// This prevents race conditions between signal handler and process monitor
		var shutdownOnce sync.Once
		performShutdown := func() {
			shutdownOnce.Do(func() {
				fmt.Println("\nShutting down API server...")

				// Graceful shutdown with timeout
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := server.Shutdown(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
					os.Exit(1)
				}

				fmt.Println("API server stopped")
				os.Exit(0)
			})
		}

		// Start process monitor to detect parent death (Swift app crash/force-quit)
		// This prevents orphaned backend processes when the GUI terminates unexpectedly
		monitor := api.NewProcessMonitor(performShutdown)
		monitor.Start()

		// Start server in goroutine
		go func() {
			if err := server.Start(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "API server error: %v\n", err)
				os.Exit(1)
			}
		}()

		// Wait for shutdown signal (Ctrl+C or SIGTERM)
		<-sigChan
		performShutdown()
	}

	// Require assembly file for emulator mode
	if flag.NArg() == 0 {
		printHelp()
		os.Exit(0)
	}

	// Get assembly file from arguments
	asmFile := flag.Arg(0)
	if _, err := os.Stat(asmFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", asmFile)
		os.Exit(1)
	}

	// Parse assembly file (with preprocessing for .include, .ifdef, etc.)
	if *verboseMode {
		fmt.Printf("Loading and parsing assembly file: %s\n", asmFile)
	}

	program, _, err := parser.ParseFileSimple(asmFile)
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

	// Configure filesystem root for sandboxing
	filesystemRoot := *fsRoot
	if filesystemRoot == "" {
		// Default to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		filesystemRoot = cwd
	}
	// Convert to absolute path
	absRoot, err := filepath.Abs(filesystemRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving filesystem root path: %v\n", err)
		os.Exit(1)
	}
	machine.FilesystemRoot = absRoot

	if *verboseMode {
		fmt.Printf("Filesystem root: %s\n", absRoot)
	}

	// Initialize stack
	// Validate stack size to prevent integer overflow
	const maxStackSize = 0x10000000 // 256MB reasonable maximum
	if *stackSize > maxStackSize {
		fmt.Fprintf(os.Stderr, "Error: stack size %d exceeds maximum allowed %d\n", *stackSize, maxStackSize)
		os.Exit(1)
	}
	stackTop := uint32(vm.StackSegmentStart + *stackSize) // #nosec G115 -- Safe: validated maxStackSize ensures no overflow
	if err := machine.InitializeStack(stackTop); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing stack: %v\n", err)
		os.Exit(1)
	}

	// Parse entry point
	var entryAddr uint32
	// First, try to use _start symbol if it exists
	if startSym, exists := program.SymbolTable.Lookup("_start"); exists && startSym.Defined {
		entryAddr = startSym.Value
		if *verboseMode {
			fmt.Printf("Using _start symbol address: 0x%08X\n", entryAddr)
		}
	} else if *entryPoint == "0x8000" && program.OriginSet {
		// If entry point is default and program has .org directive, use that
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

	err = loader.LoadProgramIntoVM(machine, program, entryAddr)
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
		// Map every instruction's address to its raw source line
		sourceMap[inst.Address] = inst.RawLine
	}

	// Add data directives to source map (prefixed with [DATA] for TUI differentiation)
	for _, dir := range program.Directives {
		// Include directives that generate data in memory
		if dir.Name == ".word" || dir.Name == ".byte" || dir.Name == ".ascii" || dir.Name == ".asciz" || dir.Name == ".space" {
			// Prefix with [DATA] so TUI can display these in a different color
			sourceMap[dir.Address] = "[DATA]" + dir.RawLine
		}
	}

	if *verboseMode {
		fmt.Printf("Entry point: 0x%08X\n", entryAddr)
		fmt.Printf("Stack: 0x%08X - 0x%08X (%d bytes)\n",
			vm.StackSegmentStart, stackTop, *stackSize)
		fmt.Printf("Symbols: %d labels defined\n", len(symbols))
	}

	// Handle symbol dump if requested
	if *dumpSymbols {
		if err := dumpSymbolTable(program.SymbolTable, *symbolsFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error dumping symbols: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Setup tracing and statistics (Phase 10)
	if *enableTrace {
		// Determine trace file path
		tracePath := *traceFile
		if tracePath == "" {
			tracePath = filepath.Join(config.GetLogPath(), "trace.log")
		}

		traceWriter, err := os.Create(tracePath) // #nosec G304 -- user-specified trace output path
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
		machine.ExecutionTrace.LoadSymbols(symbols)
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

		memTraceWriter, err := os.Create(memTracePath) // #nosec G304 -- user-specified memory trace output path
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
		machine.MemoryTrace.LoadSymbols(symbols)
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

	// Setup additional diagnostic modes (Phase 11)
	if *enableCoverage {
		// Determine coverage file path
		covPath := *coverageFile
		if covPath == "" {
			ext := "txt"
			if *coverageFormat == "json" {
				ext = "json"
			}
			covPath = filepath.Join(config.GetLogPath(), "coverage."+ext)
		}

		covWriter, err := os.Create(covPath) // #nosec G304 -- user-specified coverage output path
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating coverage file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := covWriter.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close coverage file: %v\n", err)
			}
		}()

		machine.CodeCoverage = vm.NewCodeCoverage(covWriter)
		// Set code range based on program size
		if len(program.Instructions) > 0 {
			codeStart := entryAddr
			// Safe conversion: instruction count is bounded by memory size and parser limits
			codeEnd := entryAddr + uint32(len(program.Instructions)*4) // #nosec G115 -- program size is bounded by memory
			machine.CodeCoverage.SetCodeRange(codeStart, codeEnd)
		}
		machine.CodeCoverage.LoadSymbols(symbols)
		machine.CodeCoverage.Start()

		if *verboseMode {
			fmt.Printf("Code coverage enabled: %s\n", covPath)
		}
	}

	// Stack guard requires stack trace (even without output file)
	if *enableStackTrace || *stackGuard {
		var stWriter *os.File
		var stPath string

		if *enableStackTrace {
			// Determine stack trace file path
			stPath = *stackTraceFile
			if stPath == "" {
				ext := "txt"
				if *stackTraceFormat == "json" {
					ext = "json"
				}
				stPath = filepath.Join(config.GetLogPath(), "stack_trace."+ext)
			}

			var err error
			stWriter, err = os.Create(stPath) // #nosec G304 -- user-specified stack trace output path
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating stack trace file: %v\n", err)
				os.Exit(1)
			}
			defer func() {
				if err := stWriter.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to close stack trace file: %v\n", err)
				}
			}()
		}

		machine.StackTrace = vm.NewStackTrace(stWriter, stackTop, vm.StackSegmentStart)
		machine.StackTrace.LoadSymbols(symbols)
		machine.StackTrace.Start(stackTop)

		// Enable halt on overflow if stack guard is enabled
		if *stackGuard {
			machine.StackTrace.HaltOnOverflow = true
			if *verboseMode {
				fmt.Println("Stack guard enabled: execution will halt if SP enters heap segment")
			}
		}

		if *verboseMode && *enableStackTrace {
			fmt.Printf("Stack trace enabled: %s\n", stPath)
		}
	}

	if *enableFlagTrace {
		// Determine flag trace file path
		ftPath := *flagTraceFile
		if ftPath == "" {
			ext := "txt"
			if *flagTraceFormat == "json" {
				ext = "json"
			}
			ftPath = filepath.Join(config.GetLogPath(), "flag_trace."+ext)
		}

		ftWriter, err := os.Create(ftPath) // #nosec G304 -- user-specified flag trace output path
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating flag trace file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := ftWriter.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close flag trace file: %v\n", err)
			}
		}()

		machine.FlagTrace = vm.NewFlagTrace(ftWriter)
		machine.FlagTrace.LoadSymbols(symbols)
		machine.FlagTrace.Start(machine.CPU.CPSR)

		if *verboseMode {
			fmt.Printf("Flag trace enabled: %s\n", ftPath)
		}
	}

	if *enableRegisterTrace {
		// Determine register trace file path
		rtPath := *registerTraceFile
		if rtPath == "" {
			ext := "txt"
			if *registerTraceFormat == "json" {
				ext = "json"
			}
			rtPath = filepath.Join(config.GetLogPath(), "register_trace."+ext)
		}

		rtWriter, err := os.Create(rtPath) // #nosec G304 -- user-specified register trace output path
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating register trace file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := rtWriter.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close register trace file: %v\n", err)
			}
		}()

		machine.RegisterTrace = vm.NewRegisterTrace(rtWriter)
		machine.RegisterTrace.LoadSymbols(symbols)
		machine.RegisterTrace.Start()

		if *verboseMode {
			fmt.Printf("Register trace enabled: %s\n", rtPath)
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

			statsWriter, err := os.Create(statPath) // #nosec G304 -- user-specified stats output path
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

		// Flush additional diagnostic modes (Phase 11)
		if machine.CodeCoverage != nil {
			switch *coverageFormat {
			case "json":
				err := machine.CodeCoverage.ExportJSON(machine.CodeCoverage.Writer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error exporting coverage: %v\n", err)
				}
			default:
				err := machine.CodeCoverage.Flush()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing coverage: %v\n", err)
				}
			}
			if *verboseMode {
				fmt.Println()
				fmt.Println(machine.CodeCoverage.String())
			}
		}

		if machine.StackTrace != nil {
			switch *stackTraceFormat {
			case "json":
				err := machine.StackTrace.ExportJSON(machine.StackTrace.Writer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error exporting stack trace: %v\n", err)
				}
			default:
				err := machine.StackTrace.Flush()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing stack trace: %v\n", err)
				}
			}
			if *verboseMode {
				fmt.Println()
				fmt.Println(machine.StackTrace.String())
			}
		}

		if machine.FlagTrace != nil {
			switch *flagTraceFormat {
			case "json":
				err := machine.FlagTrace.ExportJSON(machine.FlagTrace.Writer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error exporting flag trace: %v\n", err)
				}
			default:
				err := machine.FlagTrace.Flush()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing flag trace: %v\n", err)
				}
			}
			if *verboseMode {
				fmt.Println()
				fmt.Println(machine.FlagTrace.String())
			}
		}

		if machine.RegisterTrace != nil {
			switch *registerTraceFormat {
			case "json":
				err := machine.RegisterTrace.ExportJSON(machine.RegisterTrace.Writer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error exporting register trace: %v\n", err)
				}
			default:
				err := machine.RegisterTrace.Flush()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing register trace: %v\n", err)
				}
			}
			if *verboseMode {
				fmt.Println()
				fmt.Println(machine.RegisterTrace.String())
			}
		}

		os.Exit(int(machine.ExitCode))
	}
}

func printHelp() {
	fmt.Printf(`ARM2 Emulator %s

Usage: arm-emulator [options] <assembly-file>
       arm-emulator -api-server [-port N]

Options:
  -help              Show this help message
  -version           Show version information
  -api-server        Start HTTP API server mode (no assembly file required)
  -port N            API server port (default: 8080, used with -api-server)
  -debug             Start in debugger mode (CLI)
  -tui               Start in TUI debugger mode
  -max-cycles N      Set maximum CPU cycles (default: 1000000)
  -stack-size N      Set stack size in bytes (default: %d)
  -entry ADDR        Set entry point address (default: 0x8000)
  -verbose           Enable verbose output
  -fsroot DIR        Restrict file operations to directory (default: current directory)

Symbol Options:
  -dump-symbols      Dump symbol table and exit
  -symbols-file FILE Symbol dump output file (default: stdout)

Tracing & Performance Options:
  -trace             Enable execution trace
  -trace-file FILE   Trace output file (default: trace.log in log dir)
  -trace-filter REGS Filter trace by registers (e.g., R0,R1,PC)
  -mem-trace         Enable memory access trace
  -mem-trace-file F  Memory trace file (default: memtrace.log)
  -stats             Enable performance statistics
  -stats-file FILE   Statistics output file (default: stats.json)
  -stats-format FMT  Statistics format: json, csv, html (default: json)

Diagnostic Modes:
  -coverage          Enable code coverage tracking
  -coverage-file F   Coverage output file (default: coverage.txt)
  -coverage-format   Coverage format: text, json (default: text)
  -stack-trace       Enable stack operation tracing
  -stack-trace-file  Stack trace file (default: stack_trace.txt)
  -stack-trace-format Stack trace format: text, json (default: text)
  -stack-guard       Halt execution if stack overflows into heap segment
  -flag-trace        Enable CPSR flag change tracing
  -flag-trace-file   Flag trace file (default: flag_trace.txt)
  -flag-trace-format Flag trace format: text, json (default: text)
  -register-trace    Enable register access pattern tracing
  -register-trace-file Register trace file (default: register_trace.txt)
  -register-trace-format Register trace format: text, json (default: text)

Examples:
  # Start API server for GUI frontends (Swift app, Wails app)
  arm-emulator -api-server
  arm-emulator -api-server -port 3000

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

  # Run with code coverage
  arm-emulator -coverage -verbose program.s

  # Run with stack trace to debug stack issues
  arm-emulator -stack-trace program.s

  # Run with flag trace to debug conditional logic
  arm-emulator -flag-trace program.s

  # Run with register trace to analyze register usage patterns
  arm-emulator -register-trace program.s

  # Combine multiple diagnostic modes
  arm-emulator -coverage -stack-trace -flag-trace -register-trace program.s

  # Dump symbol table
  arm-emulator -dump-symbols program.s
  arm-emulator -dump-symbols -symbols-file symbols.txt program.s

  # Restrict file operations to a specific directory
  arm-emulator -fsroot /tmp/sandbox program.s
  arm-emulator -fsroot ./test_data program.s

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
`, Version, vm.StackSegmentSize)
}

// dumpSymbolTable outputs the symbol table in a readable format
func dumpSymbolTable(st *parser.SymbolTable, filename string) error {
	var writer *os.File
	var err error

	if filename == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(filename) // #nosec G304 -- user-specified symbol output path
		if err != nil {
			return fmt.Errorf("failed to create symbol file: %w", err)
		}
		defer func() {
			if cerr := writer.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close symbol file: %v\n", cerr)
			}
		}()
	}

	allSymbols := st.GetAllSymbols()
	if len(allSymbols) == 0 {
		_, _ = fmt.Fprintln(writer, "No symbols defined")
		return nil
	}

	// Print header
	_, _ = fmt.Fprintln(writer, "Symbol Table")
	_, _ = fmt.Fprintln(writer, "============")
	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintf(writer, "%-30s %-12s %-10s %s\n", "Name", "Type", "Address", "Status")
	_, _ = fmt.Fprintln(writer, "--------------------------------------------------------------------------------")

	// Sort symbols by address for easier reading
	type symbolEntry struct {
		name   string
		symbol *parser.Symbol
	}
	entries := make([]symbolEntry, 0, len(allSymbols))
	for name, sym := range allSymbols {
		entries = append(entries, symbolEntry{name, sym})
	}

	// Sort by address using O(n log n) algorithm
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].symbol.Value < entries[j].symbol.Value
	})

	// Print each symbol
	for _, entry := range entries {
		name := entry.name
		sym := entry.symbol

		var symType string
		switch sym.Type {
		case parser.SymbolLabel:
			symType = "Label"
		case parser.SymbolConstant:
			symType = "Constant"
		case parser.SymbolVariable:
			symType = "Variable"
		default:
			symType = "Unknown"
		}

		status := "Defined"
		if !sym.Defined {
			status = "Undefined"
		}

		_, _ = fmt.Fprintf(writer, "%-30s %-12s 0x%08X %s\n", name, symType, sym.Value, status)
	}

	_, _ = fmt.Fprintln(writer)
	_, _ = fmt.Fprintf(writer, "Total symbols: %d\n", len(allSymbols))

	return nil
}
