package debugger

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Command handler implementations

// cmdRun starts or restarts program execution
func (d *Debugger) cmdRun(args []string) error {
	// Reset VM state
	d.VM.Reset()
	d.VM.State = vm.StateRunning
	d.Running = true
	d.StepMode = StepNone

	d.Println("Starting program execution...")
	return nil
}

// cmdContinue continues execution from current point
func (d *Debugger) cmdContinue(args []string) error {
	if d.VM.State == vm.StateHalted {
		return fmt.Errorf("program is not running")
	}

	d.VM.State = vm.StateRunning
	d.Running = true
	d.StepMode = StepNone

	d.Println("Continuing...")
	return nil
}

// cmdStep executes a single instruction
func (d *Debugger) cmdStep(args []string) error {
	d.StepMode = StepSingle
	d.Running = true
	return nil
}

// cmdNext steps over function calls (step to next instruction at same level)
func (d *Debugger) cmdNext(args []string) error {
	// Store current PC for step over
	d.StepOverPC = d.VM.CPU.PC + 4
	d.StepMode = StepOver
	d.Running = true
	return nil
}

// cmdFinish steps out of current function
func (d *Debugger) cmdFinish(args []string) error {
	d.StepMode = StepOut
	d.Running = true
	return nil
}

// cmdBreak sets a breakpoint
func (d *Debugger) cmdBreak(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: break <address|label> [if <condition>]")
	}

	// Parse address/label
	address, err := d.ResolveAddress(args[0])
	if err != nil {
		return err
	}

	// Parse condition if present
	var condition string
	if len(args) > 1 && strings.ToLower(args[1]) == "if" {
		condition = strings.Join(args[2:], " ")
	}

	// Add breakpoint
	bp := d.Breakpoints.AddBreakpoint(address, false, condition)

	if condition != "" {
		d.Printf("Breakpoint %d at 0x%08X (condition: %s)\n", bp.ID, address, condition)
	} else {
		d.Printf("Breakpoint %d at 0x%08X\n", bp.ID, address)
	}

	return nil
}

// cmdTBreak sets a temporary breakpoint (auto-delete after hit)
func (d *Debugger) cmdTBreak(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: tbreak <address|label>")
	}

	address, err := d.ResolveAddress(args[0])
	if err != nil {
		return err
	}

	bp := d.Breakpoints.AddBreakpoint(address, true, "")
	d.Printf("Temporary breakpoint %d at 0x%08X\n", bp.ID, address)

	return nil
}

// cmdDelete deletes breakpoint(s)
func (d *Debugger) cmdDelete(args []string) error {
	if len(args) == 0 {
		// Delete all breakpoints
		d.Breakpoints.Clear()
		d.Println("All breakpoints deleted")
		return nil
	}

	// Delete specific breakpoint
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid breakpoint ID: %s", args[0])
	}

	if err := d.Breakpoints.DeleteBreakpoint(id); err != nil {
		return err
	}

	d.Printf("Breakpoint %d deleted\n", id)
	return nil
}

// cmdEnable enables breakpoint(s)
func (d *Debugger) cmdEnable(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: enable <breakpoint-id>")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid breakpoint ID: %s", args[0])
	}

	if err := d.Breakpoints.EnableBreakpoint(id); err != nil {
		return err
	}

	d.Printf("Breakpoint %d enabled\n", id)
	return nil
}

// cmdDisable disables breakpoint(s)
func (d *Debugger) cmdDisable(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: disable <breakpoint-id>")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid breakpoint ID: %s", args[0])
	}

	if err := d.Breakpoints.DisableBreakpoint(id); err != nil {
		return err
	}

	d.Printf("Breakpoint %d disabled\n", id)
	return nil
}

// cmdWatch sets a write watchpoint
func (d *Debugger) cmdWatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: watch <expression>")
	}

	expression := strings.Join(args, " ")

	// Parse expression to determine if register or memory
	isRegister, register, address, err := d.parseWatchExpression(expression)
	if err != nil {
		return err
	}

	wp := d.Watchpoints.AddWatchpoint(WatchWrite, expression, address, isRegister, register)

	// Initialize current value
	if err := d.Watchpoints.InitializeWatchpoint(wp.ID, d.VM); err != nil {
		d.Watchpoints.DeleteWatchpoint(wp.ID)
		return err
	}

	d.Printf("Watchpoint %d: %s\n", wp.ID, expression)
	return nil
}

// cmdRWatch sets a read watchpoint
func (d *Debugger) cmdRWatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: rwatch <expression>")
	}

	expression := strings.Join(args, " ")
	isRegister, register, address, err := d.parseWatchExpression(expression)
	if err != nil {
		return err
	}

	wp := d.Watchpoints.AddWatchpoint(WatchRead, expression, address, isRegister, register)

	if err := d.Watchpoints.InitializeWatchpoint(wp.ID, d.VM); err != nil {
		d.Watchpoints.DeleteWatchpoint(wp.ID)
		return err
	}

	d.Printf("Read watchpoint %d: %s\n", wp.ID, expression)
	return nil
}

// cmdAWatch sets a read/write watchpoint
func (d *Debugger) cmdAWatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: awatch <expression>")
	}

	expression := strings.Join(args, " ")
	isRegister, register, address, err := d.parseWatchExpression(expression)
	if err != nil {
		return err
	}

	wp := d.Watchpoints.AddWatchpoint(WatchReadWrite, expression, address, isRegister, register)

	if err := d.Watchpoints.InitializeWatchpoint(wp.ID, d.VM); err != nil {
		d.Watchpoints.DeleteWatchpoint(wp.ID)
		return err
	}

	d.Printf("Access watchpoint %d: %s\n", wp.ID, expression)
	return nil
}

// parseWatchExpression parses a watch expression (register or memory address)
func (d *Debugger) parseWatchExpression(expr string) (isRegister bool, register int, address uint32, err error) {
	expr = strings.ToLower(strings.TrimSpace(expr))

	// Check if it's a register (r0-r14, sp, lr, pc)
	if strings.HasPrefix(expr, "r") {
		regNum := -1
		if expr == "sp" || expr == "r13" {
			regNum = 13
		} else if expr == "lr" || expr == "r14" {
			regNum = 14
		} else if expr == "pc" || expr == "r15" {
			regNum = 15
		} else if len(expr) >= 2 {
			_, scanErr := fmt.Sscanf(expr, "r%d", &regNum)
			if scanErr == nil && regNum >= 0 && regNum <= 15 {
				return true, regNum, 0, nil
			}
		}

		if regNum >= 0 && regNum <= 15 {
			return true, regNum, 0, nil
		}
	}

	// Check if it's a memory address in brackets [0x1000] or [label]
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		addrStr := strings.TrimSuffix(strings.TrimPrefix(expr, "["), "]")
		addr, err := d.ResolveAddress(addrStr)
		if err != nil {
			return false, 0, 0, err
		}
		return false, 0, addr, nil
	}

	// Try to resolve as address or symbol
	addr, err := d.ResolveAddress(expr)
	if err != nil {
		return false, 0, 0, fmt.Errorf("invalid watch expression: %s", expr)
	}

	return false, 0, addr, nil
}

// cmdPrint evaluates and prints an expression
func (d *Debugger) cmdPrint(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: print <expression>")
	}

	expression := strings.Join(args, " ")
	result, err := d.Evaluator.EvaluateExpression(expression, d.VM, d.Symbols)
	if err != nil {
		return err
	}

	if result > uint32(math.MaxInt32) {
		d.Printf("$%d = 0x%08X (out of int32 range: %d)\n", d.Evaluator.GetValueNumber(), result, result)
	} else {
		d.Printf("$%d = 0x%08X (%d)\n", d.Evaluator.GetValueNumber(), result, int32(result))
	}
	return nil
}

// cmdExamine examines memory at an address
func (d *Debugger) cmdExamine(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: x[/nfu] <address>\n  n: count, f: format (x/d/u/o/t/a/c/s), u: unit size (b/h/w)")
	}

	// Parse format specifier (e.g., "x/8xw")
	count := 1
	format := 'x'
	unit := 'w'
	addrArg := args[0]

	if strings.HasPrefix(args[0], "/") {
		// Parse format
		formatStr := args[0][1:]
		if len(args) < 2 {
			return fmt.Errorf("missing address")
		}
		addrArg = args[1]

		// Parse count
		i := 0
		for i < len(formatStr) && formatStr[i] >= '0' && formatStr[i] <= '9' {
			i++
		}
		if i > 0 {
			if n, err := strconv.Atoi(formatStr[:i]); err == nil {
				count = n
			}
			formatStr = formatStr[i:]
		}

		// Parse format character
		if len(formatStr) > 0 {
			format = rune(formatStr[0])
			formatStr = formatStr[1:]
		}

		// Parse unit size
		if len(formatStr) > 0 {
			unit = rune(formatStr[0])
		}
	}

	// Resolve address
	address, err := d.ResolveAddress(addrArg)
	if err != nil {
		return err
	}

	// Read and display memory
	d.Printf("0x%08X:", address)
	for i := 0; i < count; i++ {
		var value uint32
		var readErr error

		switch unit {
		case 'b': // byte
			val, e := d.VM.Memory.ReadByte(address)
			value = uint32(val)
			readErr = e
			address++
		case 'h': // halfword
			val, e := d.VM.Memory.ReadHalfword(address)
			value = uint32(val)
			readErr = e
			address += 2
		default: // 'w' - word
			value, readErr = d.VM.Memory.ReadWord(address)
			address += 4
		}

		if readErr != nil {
			return readErr
		}

		// Format output
		switch format {
		case 'x': // hex
			d.Printf(" 0x%08X", value)
		case 'd': // signed decimal
			d.Printf(" %d", int32(value))
		case 'u': // unsigned decimal
			d.Printf(" %d", value)
		case 'o': // octal
			d.Printf(" %o", value)
		case 't': // binary
			d.Printf(" %b", value)
		default:
			d.Printf(" 0x%08X", value)
		}
	}
	d.Println()

	return nil
}

// cmdInfo displays information about program state
func (d *Debugger) cmdInfo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: info <registers|breakpoints|watchpoints|stack>")
	}

	switch strings.ToLower(args[0]) {
	case "registers", "reg", "r":
		return d.showRegisters()
	case "breakpoints", "break", "b":
		return d.showBreakpoints()
	case "watchpoints", "watch", "w":
		return d.showWatchpoints()
	case "stack", "s":
		return d.showStack()
	default:
		return fmt.Errorf("unknown info command: %s", args[0])
	}
}

// showRegisters displays all register values
func (d *Debugger) showRegisters() error {
	d.Println("Registers:")
	for i := 0; i < 15; i++ {
		name := fmt.Sprintf("R%d", i)
		if i == 13 {
			name = "SP"
		} else if i == 14 {
			name = "LR"
		}
		d.Printf("  %-3s = 0x%08X (%d)\n", name, d.VM.CPU.R[i], int32(d.VM.CPU.R[i]))
	}
	d.Printf("  PC  = 0x%08X (%d)\n", d.VM.CPU.PC, int32(d.VM.CPU.PC))

	// Show CPSR flags
	flags := ""
	if d.VM.CPU.CPSR.N {
		flags += "N"
	} else {
		flags += "-"
	}
	if d.VM.CPU.CPSR.Z {
		flags += "Z"
	} else {
		flags += "-"
	}
	if d.VM.CPU.CPSR.C {
		flags += "C"
	} else {
		flags += "-"
	}
	if d.VM.CPU.CPSR.V {
		flags += "V"
	} else {
		flags += "-"
	}
	d.Printf("  CPSR = [%s]\n", flags)

	return nil
}

// showBreakpoints displays all breakpoints
func (d *Debugger) showBreakpoints() error {
	breakpoints := d.Breakpoints.GetAllBreakpoints()
	if len(breakpoints) == 0 {
		d.Println("No breakpoints")
		return nil
	}

	d.Println("Breakpoints:")
	for _, bp := range breakpoints {
		status := "enabled"
		if !bp.Enabled {
			status = "disabled"
		}

		temp := ""
		if bp.Temporary {
			temp = " (temporary)"
		}

		condition := ""
		if bp.Condition != "" {
			condition = fmt.Sprintf(" if %s", bp.Condition)
		}

		d.Printf("  %d: 0x%08X %s%s%s (hit %d times)\n",
			bp.ID, bp.Address, status, temp, condition, bp.HitCount)
	}

	return nil
}

// showWatchpoints displays all watchpoints
func (d *Debugger) showWatchpoints() error {
	watchpoints := d.Watchpoints.GetAllWatchpoints()
	if len(watchpoints) == 0 {
		d.Println("No watchpoints")
		return nil
	}

	d.Println("Watchpoints:")
	for _, wp := range watchpoints {
		status := "enabled"
		if !wp.Enabled {
			status = "disabled"
		}

		wpType := "write"
		if wp.Type == WatchRead {
			wpType = "read"
		} else if wp.Type == WatchReadWrite {
			wpType = "access"
		}

		d.Printf("  %d: %s %s %s (hit %d times, last value: 0x%08X)\n",
			wp.ID, wp.Expression, wpType, status, wp.HitCount, wp.LastValue)
	}

	return nil
}

// showStack displays stack contents
func (d *Debugger) showStack() error {
	sp := d.VM.CPU.GetSP()
	d.Printf("Stack (SP = 0x%08X):\n", sp)

	// Show 8 words from stack
	for i := 0; i < 8; i++ {
		addr := sp + uint32(i*4)
		value, err := d.VM.Memory.ReadWord(addr)
		if err != nil {
			break
		}
		d.Printf("  0x%08X: 0x%08X (%d)\n", addr, value, int32(value))
	}

	return nil
}

// cmdBacktrace shows the call stack
func (d *Debugger) cmdBacktrace(args []string) error {
	d.Println("Call stack:")
	d.Printf("  #0  PC=0x%08X\n", d.VM.CPU.PC)

	// Simple backtrace - would need call stack tracking for full implementation
	if d.VM.CPU.GetLR() != 0xFFFFFFFF {
		d.Printf("  #1  LR=0x%08X\n", d.VM.CPU.GetLR())
	}

	return nil
}

// cmdList shows source code around current PC
func (d *Debugger) cmdList(args []string) error {
	pc := d.VM.CPU.PC

	// Show current instruction
	if source, exists := d.SourceMap[pc]; exists {
		d.Printf("=> 0x%08X: %s\n", pc, source)
	} else {
		d.Printf("=> 0x%08X: <no source>\n", pc)
	}

	// Show nearby instructions
	for offset := uint32(4); offset <= 16; offset += 4 {
		addr := pc + offset
		if source, exists := d.SourceMap[addr]; exists {
			d.Printf("   0x%08X: %s\n", addr, source)
		}
	}

	return nil
}

// cmdSet modifies register or memory values
func (d *Debugger) cmdSet(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: set <register|*address> = <value>")
	}

	if args[1] != "=" {
		return fmt.Errorf("usage: set <register|*address> = <value>")
	}

	target := strings.ToLower(args[0])
	valueStr := args[2]

	// Parse value
	value, err := d.Evaluator.EvaluateExpression(valueStr, d.VM, d.Symbols)
	if err != nil {
		return err
	}

	// Check if memory dereference
	if strings.HasPrefix(target, "*") {
		addrStr := target[1:]
		address, err := d.ResolveAddress(addrStr)
		if err != nil {
			return err
		}

		if err := d.VM.Memory.WriteWord(address, value); err != nil {
			return err
		}

		d.Printf("Memory 0x%08X set to 0x%08X\n", address, value)
		return nil
	}

	// Parse register
	register := -1
	if target == "pc" || target == "r15" {
		register = 15
	} else if target == "sp" || target == "r13" {
		register = 13
	} else if target == "lr" || target == "r14" {
		register = 14
	} else if strings.HasPrefix(target, "r") {
		_, err := fmt.Sscanf(target, "r%d", &register)
		if err != nil || register < 0 || register > 14 {
			return fmt.Errorf("invalid register: %s", target)
		}
	} else {
		return fmt.Errorf("invalid target: %s", target)
	}

	// Set register value
	d.VM.CPU.SetRegister(register, value)
	d.Printf("Register %s set to 0x%08X\n", target, value)

	return nil
}

// cmdLoad loads a program (placeholder)
func (d *Debugger) cmdLoad(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: load <filename>")
	}

	d.Printf("Load command not yet implemented for file: %s\n", args[0])
	return nil
}

// cmdReset resets the VM
func (d *Debugger) cmdReset(args []string) error {
	d.VM.Reset()
	d.Println("VM reset")
	return nil
}

// cmdHelp displays help information
func (d *Debugger) cmdHelp(args []string) error {
	if len(args) > 0 {
		// Show help for specific command
		return d.showCommandHelp(args[0])
	}

	// Show general help
	d.Println("ARM2 Debugger Commands:")
	d.Println()
	d.Println("Execution Control:")
	d.Println("  run (r)           - Start program execution")
	d.Println("  continue (c)      - Continue execution")
	d.Println("  step (s, si)      - Execute single instruction")
	d.Println("  next (n)          - Step over function calls")
	d.Println("  finish (fin)      - Step out of current function")
	d.Println()
	d.Println("Breakpoints:")
	d.Println("  break (b) <addr>  - Set breakpoint")
	d.Println("  tbreak (tb) <addr>- Set temporary breakpoint")
	d.Println("  delete (d) [id]   - Delete breakpoint(s)")
	d.Println("  enable <id>       - Enable breakpoint")
	d.Println("  disable <id>      - Disable breakpoint")
	d.Println()
	d.Println("Watchpoints:")
	d.Println("  watch (w) <expr>  - Watch for writes")
	d.Println("  rwatch <expr>     - Watch for reads")
	d.Println("  awatch <expr>     - Watch for access")
	d.Println()
	d.Println("Inspection:")
	d.Println("  print (p) <expr>  - Evaluate expression")
	d.Println("  x[/nfu] <addr>    - Examine memory")
	d.Println("  info (i) <what>   - Show information")
	d.Println("  backtrace (bt)    - Show call stack")
	d.Println("  list (l)          - List source code")
	d.Println()
	d.Println("Modification:")
	d.Println("  set <var> = <val> - Modify register/memory")
	d.Println()
	d.Println("Control:")
	d.Println("  reset             - Reset VM")
	d.Println("  help (h, ?)       - Show this help")
	d.Println()
	d.Println("Type 'help <command>' for detailed help on a specific command.")

	return nil
}

// showCommandHelp shows detailed help for a specific command
func (d *Debugger) showCommandHelp(cmd string) error {
	helpText := map[string]string{
		"break": "break <address|label> [if <condition>]\n  Set a breakpoint at the specified address or label.\n  Optional condition will be evaluated each time.",
		"step":  "step\n  Execute a single instruction.",
		"next":  "next\n  Step over function calls (execute until next instruction at same level).",
		"print": "print <expression>\n  Evaluate and print an expression.\n  Expressions can include registers, memory, symbols, and arithmetic.",
		"x":     "x[/nfu] <address>\n  Examine memory.\n  n: count, f: format (x/d/u/o/t), u: unit (b/h/w)",
		"info":  "info <registers|breakpoints|watchpoints|stack>\n  Display information about program state.",
	}

	if help, exists := helpText[cmd]; exists {
		d.Println(help)
		return nil
	}

	return fmt.Errorf("no help available for command: %s", cmd)
}
