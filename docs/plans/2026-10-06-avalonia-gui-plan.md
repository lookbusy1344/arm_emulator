# Avalonia GUI Plan

Plan to take the Avalonia GUI (`avalonia-gui/`) from a working core to a release that matches the Swift GUI for daily use. It supersedes `docs/AVALONIA_IMPLEMENTATION_PLAN.md`.

## Goal

The Avalonia GUI is ready for use when a user on Windows, macOS or Linux can:

1. Start the app with no backend running, or with one already running.
2. Open, edit, save and reload an assembly file, including from the command line and the recent files list.
3. Load, run, pause, step, step over, step out and restart a program.
4. Set breakpoints from the editor gutter, with F9, and from the disassembly view, and stop at them.
5. Inspect registers, flags, memory, stack and disassembly, and evaluate expressions.
6. Interact with a program through the console, including programs that wait for input.
7. See every failure as a message in the window. No backend or API error crashes the app.
8. Keep preferences and recent files across restarts.

## Current state

Verified against the code and a live backend on 2026-10-06 (commits `d03413e8`, `9e120b84`).

### Working

- **Startup.** `App` builds the services and the main view model. `BackendManager` reuses a healthy backend on the configured port or starts the bundled binary with `-api-server -port N`. The view model creates a session and connects the WebSocket.
- **REST and WebSocket clients.** Both match the Go backend's wire format. The fixtures in `ApiClientTests` and `WebSocketClientTests` are JSON captured from the backend.
- **Program loading.** Loading assembles the source, builds the address↔line maps from `/sourcemap`, loads registers and sets the state to `Idle`. Assembler errors arrive as `ProgramLoadException` with one message per error.
- **Execution.** Run, pause, step, step over and step out call the API. Status, registers and output arrive over the WebSocket.
- **Inspector panels.** Registers (with change highlighting), memory, stack, disassembly, expression evaluator, watchpoints and breakpoints. Memory, stack and disassembly receive session, register, breakpoint and memory-write updates. Stack and disassembly reload when the registers change while the VM is stopped.
- **Editor.** AvaloniaEdit with ARM syntax highlighting, a gutter showing breakpoints and the current PC, and breakpoint toggling by gutter click. The editor is read-only while the program runs.
- **Dialogs.** File open and save, examples browser, preferences and about.
- **Tests.** 426 pass and 9 integration tests are skipped. The CodeStyle_* guards cover empty-braces patterns, nested conditionals, postfix increments, mutable constant tables, file length, method length and struct size.
- **CI and release.** CI builds and tests on Ubuntu. The release workflow builds for Linux, macOS arm64 and Windows.

### Not working or missing

| Gap | Where | Effect |
|-----|-------|--------|
| `MainWindowViewModel.ErrorMessage` is bound in no view | `MainWindow.axaml` | Startup, load and execution errors are invisible |
| Execution commands do not catch API errors; no `ThrownExceptions` handler | `MainWindowViewModel` | A backend error during run, step or reset can terminate the app |
| Reset calls `/reset`, which unloads the program | `MainWindowViewModel.ResetAsync` | Stepping after reset fails with "no program loaded". Swift calls `/restart` |
| Show PC is a stub; the editor does not follow the PC | `MainWindowViewModel.ShowPcAsync` | The user loses the current line while stepping |
| No F9 breakpoint toggle; no breakpoint toggle in the disassembly view | Key bindings, `DisassemblyView` | Parity gap with Swift |
| No backend status display or restart action | Main window | A failed backend start leaves an empty window |
| Preferences are not loaded or saved | `ShowPreferencesAsync` | Settings reset on every start; font size is fixed at 14 |
| Recent files are kept in memory only | `FileService` | The list is empty on every start |
| Command-line file argument is ignored | `Program`, `App` | `ARMEmulator prog.s` opens an empty editor |
| Breakpoint and watchpoint list views swallow errors (six `TODO`s) | `BreakpointsListView`, `WatchpointsView` code-behind | Failed add or remove gives no feedback |
| WebSocket receive reads one frame of up to 8 KB | `WebSocketClient.ReceiveLoopAsync` | Long output arrives split or corrupt across a UTF-8 boundary |
| macOS application menu reads "Avalonia Application" | `App.axaml` | Wrong app name in the menu bar |
| No window title with the file name; no unsaved-changes prompt | `MainWindow` | Edits can be lost on close or load |
| Three integration tests have wrong fixtures | `BackendIntegrationTests` | Suite cannot run unskipped |

## Constraints

- **Backend API.** Additive changes only; the Swift GUI shares it. A backend change needs tests in Go and a check of both GUIs.
- **TDD.** Write the failing test first. Expected values come from the backend's JSON or a hand calculation.
- **Design.** Follow the framework-design-guidelines skill. Never suppress JSV01; use `EquatableArray` and `EquatableDictionary` from `ARMEmulator/Collections`.
- **Style.** The CodeStyle_* guards must pass. A `LongMethod` or `LargeStruct` exception needs explicit approval.
- **Checks.** Run `scripts/pre-commit.sh` before every commit. Every `dotnet` call runs outside the sandbox.

## Phase 1: Safe core

Makes the app safe to use for a single file. Each task is a small TDD change.

### 1.1 Show errors in the main window

- Add an error bar above the console, bound to `ErrorMessage`, with a dismiss button that clears it.
- Show multi-line assembler errors in full and keep them selectable for copying.
- **Tests:** view-model tests that a dismiss command clears `ErrorMessage`. A headless UI test (Phase 3) checks that the bar becomes visible.

### 1.2 Report command failures instead of crashing

- Catch `ApiException` in run, pause, step, step over, step out, breakpoint and watchpoint operations, and set `ErrorMessage` with the operation name.
- Subscribe to `ThrownExceptions` on every command as a final boundary, so an unexpected exception shows a message and is logged rather than ending the process.
- **Tests:** for each command, the API throws `ApiException` or `SessionNotFoundException`, and the test asserts the exact `ErrorMessage` and that the state is unchanged.

### 1.3 Reset restarts the program

- `ResetAsync` calls `RestartAsync`, clears the console, refreshes registers and status, sets `Idle` and clears register highlights. This matches Swift's `reset()`.
- **Tests:** reset calls `/restart` and not `/reset`. State after reset equals the state after load. A failure reports "Failed to restart: …".

### 1.4 Follow the PC in the editor

- `ShowPcCommand` scrolls the editor to the line mapped from the PC.
- After each step, and after any state event that stops the VM, the editor scrolls to the PC line if it lies outside the visible range.
- Keep the scroll logic in the view; the view model exposes a request (an observable of line numbers), not editor calls.
- **Tests:** view-model tests that a step and the Show PC command emit the mapped line, and emit nothing when the PC has no source line.

### 1.5 Breakpoint toggling parity

- F9 toggles a breakpoint on the caret line.
- Clicking a row's marker column in the disassembly view toggles a breakpoint at that address.
- A line without an address (`ValidBreakpointLines` excludes it) shows a message rather than failing silently.
- **Tests:** a toggle on a valid line adds then removes the address; a toggle on an invalid line sets the message and calls no API.

### 1.6 Backend status

- Show the backend state (`Starting`, `Running`, `Error`) in the status indicator tooltip and in the error bar on failure.
- Add a **Restart backend** command: stop, start, create a new session and reload the current source if one was loaded.
- **Tests:** restart sequence order with mocks; failure at each step reports its own message.

### 1.7 Application name

- Set `Name="ARM Emulator"` on the `Application` element so the macOS menu bar shows the app name.

**Exit criteria:** every row in the gap table up to and including backend status is closed. A program with an assembler error, a backend that fails to start and a step after halt each show a message and leave the app usable.

## Phase 2: Daily use

### 2.1 Settings persistence

- Add an `ISettingsStore` with a JSON implementation in the platform's application data directory (`Environment.SpecialFolder.ApplicationData/ARMEmulator/settings.json`). Serialise through `ApiJsonContext` or a dedicated source-generated context.
- Load settings in `App` before composing services. Pass the backend URL to `BackendManager`, `HttpClient` and the WebSocket URL.
- Apply font size to the editor and theme through `ThemeService` immediately on save. A backend URL change takes effect on backend restart (1.6).
- Treat an unreadable or invalid file as defaults, show a one-time message, and keep the damaged file as `settings.json.bak`.
- **Tests:** round trip, missing file, malformed JSON, out-of-range values clamped by `AppSettings.Validate`, and the `.bak` copy on failure.

### 2.2 Recent files persistence

- Store recent files in the same settings file and honour `RecentFilesLimit`.
- Drop entries whose file no longer exists when the menu opens, and show a message if a selected entry is gone.
- **Tests:** ordering (most recent first), de-duplication, limit, and removal of missing files.

### 2.3 Command-line file argument

- Read the first argument ending in `.s` (matching Swift), open it, and load it once the session exists.
- A missing or unreadable file shows a message; the app still starts.
- **Tests:** argument selection (ignores other flags, resolves relative paths), and load ordering after session creation.

### 2.4 Unsaved changes and window title

- Track a dirty flag: set on edit, cleared on open and save.
- Title: `ARM Emulator — file.s` with a `•` marker when dirty.
- Prompt before close, open, load example and open recent when dirty.
- **Tests:** dirty transitions; prompt decision logic in the view model, with the dialog behind an interface.

### 2.5 Breakpoint and watchpoint list errors

- Move add and remove logic from the list views' code-behind into view-model commands with validation (address parse errors, API errors).
- **Tests:** invalid address text, API failure and success paths, with exact messages.

### 2.6 WebSocket message framing

- Accumulate frames until `EndOfMessage`, then decode the whole message once.
- **Tests:** a message split across frames, a multi-byte UTF-8 character split at a frame boundary, and a message larger than the buffer.

### 2.7 Console behaviour

- Clear the console on load and restart.
- Cap retained console text (named constant) to keep long-running programs responsive.
- **Tests:** clear on load and restart; the cap keeps the newest text.

**Exit criteria:** preferences and recent files survive a restart; `ARMEmulator prog.s` opens and loads the file; no edit is lost without a prompt; long output arrives intact.

## Phase 3: Verification

### 3.1 Headless UI tests

Use `Avalonia.Headless.XUnit` (already referenced) with mocked services for:

- Startup with a backend failure shows the error bar.
- Load with assembler errors shows all messages.
- Gutter click and F9 toggle a breakpoint marker.
- Step updates registers, highlights changed registers and moves the PC marker.
- Console input is sent and cleared.
- Panels bind to their child view models (memory, stack and disassembly show data).

### 3.2 Integration suite

- Fix the three integration fixtures. **Decision needed:**
  - Add `.org 0x8000` to the breakpoint test.
  - Replace `MOV R0, #0x12345678` with an encodable immediate or `LDR R0, =0x12345678`.
  - Expect `Halted` from `GET status` straight after load (the backend reports `halted` before any run).
- Replace `[Fact(Skip = ...)]` with a trait-filtered run that starts a backend: CI builds the Go binary, starts it on a free port and runs `dotnet test --filter-trait Category=Integration` against it.
- Add a WebSocket integration test: subscribe, step, and receive a state event with registers.

### 3.3 Cross-platform CI

- Run the Avalonia unit tests on macOS and Windows as well as Ubuntu.
- Run the headless UI tests on all three.

**Exit criteria:** integration tests run unskipped in CI; headless UI tests cover the flows above on three platforms.

## Phase 4: Parity and polish

### 4.1 Parity check against the Swift GUI

Check each item against the Swift GUI and record the result in this file:

| Area | Items |
|------|-------|
| Editor | Line numbers, current-line highlight, read-only while running, font size from settings |
| Execution | Run/continue from breakpoint, pause while waiting for input, step after input (`docs/stepping-after-user-input-issue.md`) |
| Registers | Hex and decimal, CPSR flags, highlight fade timing |
| Memory | Jump to PC, SP, R0–R3; write highlight and auto-scroll setting from preferences |
| Stack | SP marker, offsets, annotations, LR detection |
| Disassembly | Window around PC, symbol labels, breakpoint markers |
| Debugging | Watchpoint types, expression history |
| Files | Examples browser search and preview |
| Shortcuts | Every Swift shortcut has an Avalonia binding; `KEYBOARD_SHORTCUTS.md` matches the bindings |

### 4.2 Accessibility and layout

- Keyboard navigation through every panel; automation names on controls (partly done).
- Minimum window size and splitter positions persisted in settings.

**Exit criteria:** every parity row is checked and either matches or has a tracked follow-up.

## Phase 5: Release

- Verify the release workflow artefacts on each platform: the app starts the bundled backend, loads an example and runs it.
- macOS: bundle name, icon, signing status documented; backend in `Contents/Resources`.
- Windows and Linux: backend next to the executable; `chmod +x` on Linux.
- Update `avalonia-gui/README.md`, `docs/GUI.md` and `KEYBOARD_SHORTCUTS.md`.
- Retire stale documents: `docs/AVALONIA_IMPLEMENTATION_PLAN.md`, `docs/AVALONIA_PHASE_12_SUMMARY.md`, `avalonia-gui/PHASE10_*`, `avalonia-gui/PHASE11_COMPLETION_SUMMARY.md`, and the "not yet implemented" notes in `CONFIGURATION.md` and `TESTING_GUIDE.md`.

## Backend items

These are backend behaviours the GUI works around today. Each needs a Go change with tests and a check of both GUIs:

- A failed load replaces the session's source map (seen with `GET /sourcemap` after a 400 load).
- `/reset` leaves the session without a loaded program. Document it as "clear VM" or make it keep the program; the GUIs use `/restart`.
- The backend reports `halted` for a loaded program that has not run. An `idle` (or `loaded`) state would remove the GUI's special case.
- Execution events (`breakpoint_hit`, `halted`, `error`) are defined in the broadcaster but no handler sends them.

## Order of work

1. Phase 1 in task order. 1.1 and 1.2 first: they make every later failure visible.
2. Phase 2.1 and 2.2 together (one settings file), then 2.3–2.7 in any order.
3. Phase 3 alongside Phase 2: each Phase 2 task adds its headless test once 3.1's harness exists.
4. Phase 4, then Phase 5.

## Open decisions

- Integration fixture fixes (3.2).
- Whether backend items are fixed in Go or stay as GUI workarounds.
- Settings file location on macOS: `~/Library/Application Support` (via `ApplicationData`) or a shared location with the Swift GUI's `UserDefaults`. Sharing is not practical; separate files are assumed.
- Whether to keep the old implementation plan in `docs/` for reference or delete it.
