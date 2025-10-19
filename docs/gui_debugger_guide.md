# GUI Debugger User Guide

## Overview

The ARM2 Emulator includes a modern graphical debugger built with the Fyne framework. It provides an intuitive interface for debugging ARM2 assembly programs with visual panels for source code, registers, memory, stack, and console output.

## Starting the GUI Debugger

Launch the GUI debugger with the `--gui` flag:

```bash
./arm-emulator --gui program.s
```

The GUI will open in a new window with multiple panels arranged for optimal debugging workflow.

## Interface Layout

The GUI debugger window is divided into several panels:

### Main Panels

**Left Side: Source Code View (55% width)**
- Displays the source code or disassembly
- Current execution line is marked with `→` indicator
- Shows line numbers for easy reference
- Updates automatically as program executes

**Right Side (45% width):**

**Top Section:**
- **Registers Panel** (60% height)
  - Shows all general-purpose registers (R0-R12)
  - Special registers: SP (R13), LR (R14), PC (R15)
  - CPSR flags: N, Z, C, V
  - Values displayed in both hex and decimal

- **Breakpoints Panel** (40% height)
  - Lists all active breakpoints
  - Shows address, symbol name (if available), and enabled/disabled status
  - Updates when breakpoints are added or removed

**Bottom Section: Tabbed Interface**
- **Memory Tab** - Hex dump and ASCII view of memory
- **Stack Tab** - Stack contents with current SP indicator
- **Console Tab** - Program output from SWI calls

### Control Elements

**Toolbar (Top)**
- ▶️ **Run/Play** - Start or restart program execution
- ⏭️ **Step** - Execute one instruction (step into)
- ⏩ **Continue** - Continue execution until breakpoint or exit
- ⏹️ **Stop** - Halt execution at current instruction
- ➕ **Add Breakpoint** - Set breakpoint at current PC
- ❌ **Clear Breakpoints** - Remove all breakpoints
- 🔄 **Refresh** - Manually refresh all views

**Status Bar (Bottom)**
- Shows current debugger state and messages
- Displays execution status, errors, and breakpoint information

## Using the GUI Debugger

### Basic Workflow

1. **Load a Program**: Start the debugger with your assembly file
   ```bash
   ./arm-emulator --gui examples/fibonacci.s
   ```

2. **Set Breakpoints**: 
   - Click **Add Breakpoint** (➕) to set a breakpoint at the current PC
   - Or let the program run and click Add Breakpoint when it stops

3. **Execute Code**:
   - Click **Run** (▶️) to start execution
   - Click **Step** (⏭️) to execute one instruction at a time
   - Click **Continue** (⏩) to run until next breakpoint

4. **Examine State**:
   - Watch registers update in the Registers panel
   - Switch to Memory tab to view memory contents
   - Check Stack tab to see stack operations
   - Read Console tab for program output

5. **Control Execution**:
   - Click **Stop** (⏹️) to halt execution
   - Use **Step** for fine-grained control
   - Add more breakpoints as needed

### Register Panel

The Registers panel shows all CPU registers in real-time:

```
General Purpose Registers:
──────────────────────────
R0 : 0x00000000  (0)
R1 : 0x00000001  (1)
R2 : 0x00000002  (2)
...
R12: 0x00000000  (0)

Special Registers:
──────────────────────────
SP:  0x00050000  (327680)
LR:  0x00000000  (0)
PC:  0x00008000  (32768)

Status Flags (CPSR):
──────────────────────────
Flags: NZCV or ----
```

**Note:** Future versions may highlight changed registers in color.

### Memory View

The Memory tab displays memory in hex dump format:

```
Memory at 0x00008000:
──────────────────────────────────────────────────
00008000: E3 A0 00 01 E3 A0 10 00 E0 80 10 01 E1 50 00 01  .............P..
00008010: 1A FF FF FB E1 A0 F0 0E 00 00 00 00 00 00 00 00  ................
```

- **Left column**: Memory address
- **Middle**: Hex bytes (16 bytes per row)
- **Right**: ASCII representation (. for non-printable characters)

The view is centered on the current PC or a specific address you're examining.

### Stack View

The Stack tab shows stack contents around the current SP:

```
Stack at SP=0x00050000:
──────────────────────────────
0004FFE0: 00000000  (0)
0004FFE4: 00000000  (0)
...
→ 00050000: 00000000  (0)  ← Current SP
  00050004: 00000000  (0)
```

- Stack grows downward (lower addresses)
- Current SP position is marked with `→`
- Each row shows address, hex value, and decimal value

### Breakpoint Management

**Adding Breakpoints:**
1. Execute code until you reach the desired location (using Step or Continue)
2. Click the **Add Breakpoint** (➕) button
3. A breakpoint is added at the current PC
4. The Breakpoints panel updates to show the new breakpoint

**Viewing Breakpoints:**
- The Breakpoints panel lists all breakpoints
- Format: `0x00008010 [function_name] (enabled)`
- Symbol names are shown when available

**Clearing Breakpoints:**
- Click **Clear Breakpoints** (❌) to remove all breakpoints at once

**Note:** Future versions will support:
- Clicking in the source view to toggle breakpoints
- Right-click context menu for breakpoint management
- Conditional breakpoints
- Disabled vs. enabled state toggling

### Console Output

The Console tab captures all program output:
- SWI #0x01 (write_char)
- SWI #0x02 (write_string)
- SWI #0x03 (write_int)
- Other console-related syscalls

Output appears in real-time as the program executes.

## Execution Control

### Run/Play (▶️)

**Behavior:**
- Starts program execution from current PC
- Runs continuously until:
  - A breakpoint is hit
  - Program exits (SWI #0x00)
  - An error occurs

**Use Case:**
- Initial program startup
- Resuming after examining state at a breakpoint
- Running to the next interesting point

**Note:** The GUI remains responsive during execution. You can click Stop to interrupt.

### Step (⏭️)

**Behavior:**
- Executes exactly one ARM instruction
- Updates all views (registers, memory, stack)
- Stops immediately after instruction

**Use Case:**
- Fine-grained debugging
- Examining each instruction's effect
- Understanding algorithm flow step-by-step

**Example Workflow:**
```
1. Click Step - execute MOV R0, #5
2. See R0 change to 5 in Registers panel
3. Click Step - execute ADD R1, R0, #1
4. See R1 change to 6 in Registers panel
```

### Continue (⏩)

**Behavior:**
- Same as Run/Play
- Semantic difference: "continue" implies resuming from a breakpoint

**Use Case:**
- After examining state at a breakpoint
- Letting program run to next breakpoint
- Skipping over uninteresting code

### Stop (⏹️)

**Behavior:**
- Immediately halts program execution
- Useful for interrupting long-running code
- Can resume with Continue or Step

**Use Case:**
- Investigating unexpected behavior
- Stopping infinite loops
- Pausing to examine current state

## Tips and Best Practices

### Debugging Strategy

1. **Start with Breakpoints**: Set breakpoints at key locations (function entry, loops)
2. **Use Step Wisely**: Step through critical sections, use Continue for routine code
3. **Watch Registers**: Keep an eye on registers that matter for your algorithm
4. **Check Stack**: Verify stack operations (PUSH/POP) are balanced
5. **Monitor Memory**: Watch for unexpected memory writes in Memory view

### Performance Considerations

**GUI Refresh Rate:**
- Views update after each step or when execution stops
- For very fast execution, updates occur at breakpoints only
- Click Refresh (🔄) if views seem out of sync

**Large Programs:**
- Memory view shows 256 bytes (16 lines) at a time
- Use Step to navigate through code
- Breakpoints help skip to interesting locations

### Common Workflows

**Workflow 1: First Run Through**
```
1. Load program: ./arm-emulator --gui examples/fibonacci.s
2. Click Run (▶️) - see if it completes successfully
3. Check Console tab for expected output
4. If issues, restart and use Step
```

**Workflow 2: Debugging a Problem**
```
1. Set breakpoint at suspected problem area
2. Click Run to reach breakpoint
3. Use Step to execute line by line
4. Watch Registers and Memory for unexpected changes
5. Check Stack for corruption or imbalance
```

**Workflow 3: Understanding an Algorithm**
```
1. Set breakpoint at algorithm start
2. Click Continue to reach it
3. Use Step repeatedly through algorithm
4. Watch how registers change with each instruction
5. Verify against expected algorithm behavior
```

## System Requirements

### Platform Support

**Linux:**
- X11 or Wayland display server required
- OpenGL support recommended (available on most modern systems)
- Tested on Ubuntu 22.04+, Fedora 38+

**macOS:**
- macOS 10.12+ (Sierra or later)
- Native Cocoa support
- Hardware acceleration via Metal

**Windows:**
- Windows 10+ (64-bit)
- No additional dependencies needed
- Hardware acceleration via DirectX

### Dependencies

The GUI debugger uses the **Fyne** framework (v2.7.0+), which has minimal dependencies:
- Pure Go implementation (no CGO required for most features)
- System graphics libraries (provided by OS)
- No separate installation of GUI frameworks needed

**Building from Source:**
```bash
# Clone repository
git clone <repository-url>
cd arm_emulator

# Install Go 1.25.2 or later
# Build (all dependencies downloaded automatically)
go build -o arm-emulator

# Run GUI debugger
./arm-emulator --gui examples/hello.s
```

### Known Limitations

1. **Headless Environments**: The GUI requires a display server. For headless systems (CI, servers), use `--tui` or `--debug` instead.

2. **Remote X11**: GUI may work over X11 forwarding but performance will be slower. Consider TUI for remote debugging.

3. **Source View**: Currently shows basic source/disassembly. Syntax highlighting planned for future versions.

4. **Breakpoint UI**: Setting breakpoints currently requires using Add Breakpoint button at current PC. Click-to-set in source view coming in future version.

## Keyboard Shortcuts

**Note:** Full keyboard shortcut support is planned for a future version. Current version uses mouse/toolbar interaction.

Planned shortcuts:
- `F5` - Continue
- `F9` - Toggle breakpoint at current line
- `F10` - Step over (when implemented)
- `F11` - Step into (current Step behavior)
- `Ctrl+R` - Refresh views
- `Ctrl+B` - Add breakpoint
- `Ctrl+Q` - Quit

## Troubleshooting

### GUI Won't Start

**Symptom:** Error message when running `--gui` flag

**Solutions:**
1. Verify display is available (not headless environment)
2. On Linux, check `$DISPLAY` environment variable is set
3. Try TUI mode instead: `--tui`

### Display Issues

**Symptom:** Garbled text, overlapping panels, or display corruption

**Solutions:**
1. Resize window (panels should adjust)
2. Click Refresh (🔄) button
3. Restart debugger
4. Update graphics drivers

### Slow Performance

**Symptom:** GUI feels sluggish or unresponsive

**Solutions:**
1. Close other applications to free resources
2. Reduce window size (fewer pixels to render)
3. Consider using TUI mode for better performance
4. Check for hardware acceleration support

### Program Output Not Visible

**Symptom:** Console tab is empty despite program output

**Solutions:**
1. Switch to Console tab (program output only shown there)
2. Verify program actually produces output (check with `--tui` mode)
3. Click Refresh to update view

## Comparison: GUI vs TUI vs CLI

| Feature | GUI | TUI | CLI |
|---------|-----|-----|-----|
| **Visual Appeal** | ★★★★★ | ★★★★☆ | ★★☆☆☆ |
| **Ease of Use** | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| **Performance** | ★★★☆☆ | ★★★★☆ | ★★★★★ |
| **Remote Access** | ★★☆☆☆ | ★★★★★ | ★★★★★ |
| **Scriptability** | ★☆☆☆☆ | ★☆☆☆☆ | ★★★★★ |
| **Learning Curve** | ★★★★★ | ★★★☆☆ | ★★☆☆☆ |
| **Resource Usage** | ★★☆☆☆ | ★★★★☆ | ★★★★★ |

**Choose GUI when:**
- You have a graphical environment
- You want the easiest debugging experience
- Visual feedback is important
- You're new to the emulator

**Choose TUI when:**
- You're working over SSH
- You want good performance with visual feedback
- You're comfortable with keyboard navigation
- You need multiple panels but not a full GUI

**Choose CLI when:**
- You want to script debugging sessions
- Maximum performance is critical
- You're working in a minimal environment
- You prefer command-line workflows

## Advanced Features (Planned)

Future versions of the GUI debugger will include:

**Enhanced Breakpoints:**
- Click in source view to toggle breakpoints
- Conditional breakpoints with expression support
- Hardware/software breakpoint distinction
- Breakpoint hit counts

**Memory Editing:**
- Click to edit memory bytes
- Search memory for patterns
- Memory regions visualization
- Goto address feature

**Watch Expressions:**
- Add custom expressions to watch
- Automatic evaluation on each step
- Complex expressions (registers, memory, symbols)

**Themes:**
- Dark mode / Light mode toggle
- Customizable colors
- Font size adjustment
- Layout presets

**Session Management:**
- Save/load debugger state
- Breakpoint persistence
- Window layout saving
- Recent files list

**Disassembly Enhancements:**
- Syntax highlighting for assembly
- Instruction tooltips (what each instruction does)
- Mixed source/assembly view
- Symbol resolution inline

## Feedback and Contributions

The GUI debugger is actively being improved. Feedback and contributions are welcome:

- Report issues on GitHub
- Suggest features
- Contribute code improvements
- Share debugging workflows

See the project repository for more information on contributing.

## See Also

- [Debugger Reference](./debugger_reference.md) - Complete debugger command reference
- [GUI Assessment](./gui_assessment.md) - Technical details and framework evaluation
- [Architecture](./architecture.md) - System architecture and design
- [Tutorial](./TUTORIAL.md) - Learn ARM2 assembly programming
