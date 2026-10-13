# Front-end Feature Comparison

The emulator has three front ends:

| Front end | Source | Platform | Talks to the VM through |
|-----------|--------|----------|-------------------------|
| Swift GUI | `swift-gui/` | macOS | HTTP API + WebSocket (`api/`) |
| Avalonia GUI | `avalonia-gui/` | Windows, macOS, Linux | HTTP API + WebSocket (`api/`) |
| Go TUI | `debugger/tui.go` | Any terminal | In-process `debugger` package |

Launch the TUI with `arm-emulator -tui program.s`. Both GUIs start the bundled backend (`--api-server`) themselves.

The GUIs are edit-assemble-debug environments. The TUI is a GDB-style debugger over a program loaded from the command line; it has no editor and no file handling.

Legend: ✅ supported · ◐ partial · — absent

Address fields in all three front ends read `0x` and hex digits as hex and a bare number as decimal: `0x8000` and `32768` name the same address.

## Layout

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Editor + inspector + console layout | ✅ split views | ✅ grid splitters | ✅ fixed tview layout, all panels visible |
| Inspector panel picker | ✅ dropdown, 7 panels | ✅ dropdown, 7 panels | — (all panels shown at once) |
| Window geometry and splitters persist | ◐ window frame only | ✅ | — |
| Selected inspector panel persists | ✅ | ✅ | — |
| Move focus between panes | — | ✅ F6 / Shift+F6 | ✅ Tab / Shift+Tab |
| Native macOS menu bar | ✅ | ✅ | — |

## Files and editing

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Edit source | ✅ `NSTextView` | ✅ AvaloniaEdit | — |
| Syntax highlighting | ✅ system colours | ✅ themed `.xshd` | ◐ labels coloured in source view |
| Line numbers | ✅ | ✅ | ✅ |
| Editor font size preference | ✅ | ✅ | — |
| Editor read-only while running | ✅ | ✅ | n/a |
| Open / Save / Save As | ✅ | ✅ | — (`load` is a placeholder) |
| Recent files | ✅ limit 5–20, Clear Menu | ✅ limit 1–50 | — |
| Examples browser with search and preview | ✅ reads local `examples/` directory | ✅ reads `/api/v1/examples` | — |
| Unsaved-changes prompt (open, example, close, quit) | ✅ Save / Don't Save / Cancel | ✅ Save / Discard / Cancel | n/a |
| Dirty marker in window title | ✅ `name •`, dot in close button | ✅ `name •` | n/a |
| Open `.s` file from command line | ✅ | ✅ | ✅ |

## Execution control

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Assemble | ✅ ⌘L | ✅ ⌘L / Ctrl+L | — (assembled at launch) |
| Run / Continue | ✅ F5, ⌘R | ✅ F5, ⌘R / Ctrl+R | ✅ F5, `run`, `continue` |
| Pause | ✅ ⌘. | ✅ ⌘. / Ctrl+. | ✅ Ctrl+C while running; quits when stopped |
| Step | ✅ F11, ⌘T | ✅ F11, F7, ⌘T / Ctrl+T | ✅ F11, `step` |
| Step Over | ✅ F10, F8, ⌘⇧T | ✅ F10, F8, ⌘⇧T / Ctrl+Shift+T | ✅ F10, `next` |
| Step Out | ✅ ⌘⌥T | ✅ ⌘⌥T / Ctrl+Alt+T | ✅ `finish` |
| Reset | ✅ ⌘⇧R | ✅ ⌘⇧R / Ctrl+Shift+R | ✅ `reset` |
| Scroll editor to PC | ✅ ⌘J | ✅ ⌘J / Ctrl+J | ✅ F6 |
| Status indicator | ✅ toolbar icon + text | ✅ toolbar pill | ✅ Status panel |
| Restart backend from menu | ✅ File → Restart Backend | ✅ File → Restart Backend | n/a |
| Error display | modal alert | dismissible inline bar | Status panel |
| Command history | n/a | n/a | ✅ Up/Down recall; empty Enter repeats the last command |

## Breakpoints and watchpoints

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Toggle breakpoint in editor gutter | ✅ | ✅ | — |
| F9 toggles breakpoint | ✅ at editor caret | ✅ at editor caret | ◐ at current PC |
| Toggle breakpoint in disassembly | ✅ click marker column | ✅ click marker column | — |
| Breakpoint by address or label | — | — | ✅ `break <addr\|label>` |
| Conditional breakpoints | — | — | ✅ `break <addr> if <expr>` |
| Temporary breakpoints | — | — | ✅ `tbreak` |
| Enable / disable without deleting | — | — | ✅ `enable`, `disable` |
| Breakpoints survive re-assembly on their source line | ✅ | ✅ | n/a |
| Breakpoints list with remove | ✅ | ✅ with counts | ✅ panel; `delete [id]` |
| Watchpoints: read, write, read/write | ✅ | ✅ | ✅ `rwatch`, `watch`, `awatch` |
| Watchpoint target | address | address | expression |

## Inspection

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Registers R0–R12, SP, LR, PC (hex + decimal) | ✅ 1–3 columns by width | ✅ table | ✅ |
| Changed registers highlighted | ✅ green, fades over 1.5 s | ✅ row background | ✅ green until next step |
| CPSR flags | ✅ N/Z/C/V pills in status strip | ✅ N/Z/C/V pills | ✅ |
| Memory hex + ASCII dump, 16 bytes per row | ✅ 256 bytes | ✅ | ✅ |
| Memory: go to address | ✅ | ✅ | ◐ `x/nfu` prints to Status; panel only follows writes |
| Memory: jump to PC, SP, R0–R3 | ✅ | ✅ | — |
| Memory: follow writes (toggle) | ✅ | ✅ also a preference | ✅ always on |
| Memory: write highlight | ✅ per byte, fades | ✅ per row, fades | ✅ per byte |
| Stack view with SP offset, value, ASCII | ✅ | ✅ | ✅ |
| Stack annotations | ✅ return address, stack pointer, register match, LR | ✅ LR, return address, stack address | ◐ symbol names only |
| Backtrace | — | — | ✅ `backtrace` |
| Disassembly ±32 instructions around PC | ✅ | ✅ | ✅ |
| Disassembly symbols, PC and breakpoint markers | ✅ | ✅ | ✅ |
| Expression evaluator with history | ✅ hex/dec/bin per entry | ✅ result panel + history, Clear | ✅ `print` with `$n` value history |
| Modify registers or memory | — | — | ✅ `set r0 = 5`, `set *addr = v` |
| Source listing command | n/a | n/a | ✅ `list` |

## Console

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Program output with auto-scroll | ✅ | ✅ | ✅ |
| Program input (stdin) | ✅ | ✅ | ✅ separate input field |
| Waiting-for-input cue | ✅ orange outline | ✅ pill + styled input row | — |

## Preferences

| | Swift | Avalonia | TUI |
|---|---|---|---|
| Theme: auto / light / dark | ✅ | ✅ | — |
| Backend URL | ✅ applied at next launch | ✅ applied on backend restart | n/a |
| Editor font size | ✅ live preview | ✅ live preview | — |
| Recent files limit | ✅ | ✅ | — |

## Gaps

Swift GUI:

- Splitter positions do not persist.
- No key to move focus between panes.
- A new Backend URL applies at the next launch, not on backend restart.

Avalonia GUI:

- Matches or exceeds the Swift GUI in most rows. It trails on memory write highlighting (per row, not per byte).

Both GUIs:

- No conditional or temporary breakpoints, no enable/disable, no register or memory editing, no backtrace. The TUI has all of these.
- The API exposes `trace` and `stats` endpoints that neither GUI uses.

Go TUI:

- The memory panel cannot be moved to an arbitrary address.
