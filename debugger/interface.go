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
				// For single-step mode, execute instruction first before checking if we should break
				// For other modes, check breakpoints before execution
				if dbg.StepMode != StepSingle {
					if shouldBreak, reason := dbg.ShouldBreak(); shouldBreak {
						dbg.Running = false
						fmt.Printf("Stopped: %s at PC=0x%08X\n", reason, dbg.VM.CPU.PC)
						break
					}
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

				// For single-step mode, check if we should break after execution
				if dbg.StepMode == StepSingle {
					if shouldBreak, reason := dbg.ShouldBreak(); shouldBreak {
						dbg.Running = false
						fmt.Printf("Stopped: %s at PC=0x%08X\n", reason, dbg.VM.CPU.PC)
						break
					}
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
	// Check terminal width before starting TUI
	const minWidth = 148
	fmt.Println(strings.Repeat("=", minWidth))
	fmt.Printf("The line above is %d characters wide.\n", minWidth)
	fmt.Print("Does the line above display as a single line without wrapping? (y/n/i=ignore): ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		// If there's an error reading (e.g., EOF), just continue
		fmt.Println("\nCould not read response, continuing anyway...")
	} else {
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "i" {
			// Ignore - just continue without warning
			fmt.Println("\nIgnoring width check...")
		} else if response != "y" {
			fmt.Printf("\nWARNING: Your terminal appears to be less than %d columns wide.\n", minWidth)
			fmt.Println("The TUI may not display correctly. Consider resizing your terminal.")
			fmt.Print("\nDo you want to continue anyway? (y/n): ")

			_, err = fmt.Scanln(&response)
			if err != nil || strings.ToLower(strings.TrimSpace(response)) != "y" {
				fmt.Println("Exiting. Please resize your terminal and try again.")
				return nil
			}
		}
	}

	fmt.Println("\nStarting TUI debugger...")
	tui := NewTUI(dbg)
	return tui.Run()
}
