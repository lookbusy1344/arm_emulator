package debugger

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// RunCLI runs the command-line debugger interface
func RunCLI(dbg *Debugger) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Print prompt
		fmt.Print("(arm-dbg) ")

		// Read command
		if !scanner.Scan() {
			break
		}

		cmdLine := strings.TrimSpace(scanner.Text())

		// Exit commands
		if cmdLine == "quit" || cmdLine == "q" || cmdLine == "exit" {
			fmt.Println("Exiting debugger...")
			break
		}

		// Execute command
		if err := dbg.ExecuteCommand(cmdLine); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		// Print any output from the debugger
		output := dbg.GetOutput()
		if output != "" {
			fmt.Print(output)
		}

		// If running, execute until breakpoint or halt
		if dbg.Running {
			for dbg.Running {
				// Check for breakpoint before execution
				if shouldBreak, reason := dbg.ShouldBreak(); shouldBreak {
					dbg.Running = false
					fmt.Printf("Stopped: %s at PC=0x%08X\n", reason, dbg.VM.CPU.PC)
					break
				}

				// Execute one step
				if err := dbg.VM.Step(); err != nil {
					if dbg.VM.State == vm.StateHalted {
						dbg.Running = false
						fmt.Printf("Program exited with code %d\n", dbg.VM.ExitCode)
						break
					}
					fmt.Printf("Runtime error: %v\n", err)
					dbg.Running = false
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	return nil
}

// RunTUI runs the TUI (Text User Interface) debugger
func RunTUI(dbg *Debugger) error {
	tui := NewTUI(dbg)
	return tui.Run()
}
