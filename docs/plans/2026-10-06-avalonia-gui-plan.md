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
9. Work in a window that looks and behaves like a native desktop app on each platform, to the standard of the Swift GUI.

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

### 1.1 Show errors in the main window (done)

- An error bar below the toolbar, bound to `ErrorMessage`, with a dismiss button that clears it.
- Show multi-line assembler errors in full and keep them selectable for copying.
- **Tests:** view-model tests that a dismiss command clears `ErrorMessage`. A headless UI test (Phase 4) checks that the bar becomes visible.

### 1.2 Report command failures instead of crashing (done)

- Every command is created with an operation name and its `ThrownExceptions` routed to `ErrorMessage` as "<operation> failed: <message>". This covers API errors, lost connections (`HttpRequestException`) and wire-format errors. Cancellation is not reported.
- Gutter clicks run `ToggleBreakpointCommand`, so a failed breakpoint call is reported, not thrown from an event handler.
- **Tests:** for each command, the API throws `ApiException` or `SessionNotFoundException`, and the test asserts the exact `ErrorMessage` and that the state is unchanged.

### 1.3 Reset restarts the program (done)

- `ResetAsync` calls `RestartAsync`, clears the console, refreshes registers and status, sets `Idle` and clears register highlights. This matches Swift's `reset()`.
- **Tests:** reset calls `/restart` and not `/reset`. State after reset equals the state after load. A failure reports "Failed to restart: …".

### 1.4 Follow the PC in the editor (done)

- `ShowPcCommand` scrolls the editor to the line mapped from the PC.
- After each step, and after any state event that stops the VM, the editor scrolls to the PC line if it lies outside the visible range.
- Keep the scroll logic in the view; the view model exposes a request (an observable of line numbers), not editor calls.
- **Tests:** view-model tests that a step and the Show PC command emit the mapped line, and emit nothing when the PC has no source line.

### 1.5 Breakpoint toggling parity (done)

- `ToggleBreakpointCommand` (by source line) exists and reports lines without an instruction (done with 1.2).
- F9 toggles a breakpoint on the caret line.
- Clicking a row's marker column in the disassembly view toggles a breakpoint at that address.
- A line without an address (`ValidBreakpointLines` excludes it) shows a message rather than failing silently.
- **Tests:** a toggle on a valid line adds then removes the address; a toggle on an invalid line sets the message and calls no API.

### 1.6 Backend status (done)

- Show the backend state (`Starting`, `Running`, `Error`) in the status indicator tooltip and in the error bar on failure.
- Add a **Restart backend** command: stop, start, create a new session and reload the current source if one was loaded.
- **Tests:** restart sequence order with mocks; failure at each step reports its own message.

### 1.7 Application name (done)

- Set `Name="ARM Emulator"` on the `Application` element so the macOS menu bar shows the app name.

**Exit criteria:** every row in the gap table up to and including backend status is closed. A program with an assembler error, a backend that fails to start and a step after halt each show a message and leave the app usable.

## Phase 2: Daily use

### 2.1 Settings persistence (done)

- Add an `ISettingsStore` with a JSON implementation in the platform's application data directory (`Environment.SpecialFolder.ApplicationData/ARMEmulator/settings.json`). Serialise through `ApiJsonContext` or a dedicated source-generated context.
- Load settings in `App` before composing services. Pass the backend URL to `BackendManager`, `HttpClient` and the WebSocket URL.
- Apply font size to the editor and theme through `ThemeService` immediately on save. A backend URL change takes effect on backend restart (1.6).
- Treat an unreadable or invalid file as defaults, show a one-time message, and keep the damaged file as `settings.json.bak`.
- **Tests:** round trip, missing file, malformed JSON, out-of-range values clamped by `AppSettings.Validate`, and the `.bak` copy on failure.

### 2.2 Recent files persistence (done)

- Store recent files in the same settings file and honour `RecentFilesLimit`.
- Drop entries whose file no longer exists when the menu opens, and show a message if a selected entry is gone.
- **Tests:** ordering (most recent first), de-duplication, limit, and removal of missing files.

### 2.3 Command-line file argument (done)

- Read the first argument ending in `.s` (matching Swift), open it, and load it once the session exists.
- A missing or unreadable file shows a message; the app still starts.
- **Tests:** argument selection (ignores other flags, resolves relative paths), and load ordering after session creation.

### 2.4 Unsaved changes and window title (done)

- Track a dirty flag: set on edit, cleared on open and save.
- Title: `ARM Emulator — file.s` with a `•` marker when dirty.
- Prompt before close, open, load example and open recent when dirty.
- **Tests:** dirty transitions; prompt decision logic in the view model, with the dialog behind an interface.

### 2.5 Breakpoint and watchpoint list errors (done)

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

## Phase 3: Visual design

Brings the look and layout up to the standard of the Swift GUI. The Swift app uses native split views, an SF Symbols toolbar, a compact view picker for the inspector panels, monospaced 10–11 pt data views and system colours. The Avalonia app today shows the problems below (screenshot of 2026-10-06):

| Problem | Where |
|---------|-------|
| Inspector tab headers use the large Fluent tab style and wrap onto three rows, taking a third of the panel | `RightPanelView.axaml` |
| Toolbar icons are emoji-like glyphs; separators are text dashes; button widths vary; the status light is a bare dot | `ToolbarView.axaml` |
| Splitters are 5 px dark bars; every panel draws its own border and corner radius, so frames nest | `MainWindow.axaml`, panel views |
| Register values render in a proportional font: `FontFamily="monospace"` does not resolve on macOS | `RegistersView.axaml` and other data views |
| Register cards leave large gaps; hex and decimal stack vertically | `RegistersView.axaml` |
| The editor has no current-line or PC-line highlight; the gutter has no theme colours | `EditorView`, `EditorGutterMargin` |
| The console has a heavy header and border; the input row looks like a form | `ConsoleView.axaml` |
| Shortcuts use `Ctrl` on macOS, where the platform uses `⌘` | `MainWindow.axaml` key bindings |
| No empty state: a fresh window is a blank editor and zeroed registers | Main window |

Every task checks light and dark themes, and 100 % and 200 % scaling.

### 3.1 Theme resources

- One resource dictionary (`Themes/`) holds the design tokens. Views use the tokens, never literal colours or sizes.
- **Colours:** semantic names (window background, panel background, divider, secondary text, accent, changed-register highlight, memory-write highlight, breakpoint, PC marker, error), with `ThemeVariant` light and dark entries.
- **Spacing and type:** a spacing scale (4 / 8 / 12 / 16), UI font sizes (11 / 12 / 13), and a data font size (11, from settings for the editor).
- **Monospace font:** a fallback list (`SF Mono, Menlo, Cascadia Mono, Consolas, DejaVu Sans Mono`), or a bundled font (JetBrains Mono, SIL Open Font License) for identical rendering on all platforms. **Decision needed.**
- Use `FluentTheme` with `DensityStyle="Compact"`.
- **Tests:** a test fails if any `.axaml` file under `Views/` sets a literal colour (`#…`, named colours) outside `Themes/`.

### 3.2 Toolbar

- Vector icons from one icon set (Fluent System Icons, MIT licence), embedded as `StreamGeometry` resources and tinted by theme. Icons follow the Swift set: Load (document), Run (play; continue at a breakpoint), Pause, Step, Step Over, Step Out, Reset (counter-clockwise arrow), Show PC.
- Uniform button size, icon over label or icon beside label, real separators, and tooltips showing the platform shortcut.
- The status indicator becomes a labelled pill (e.g. "Idle", "Running", "Breakpoint") in the theme's state colours.

### 3.3 Window layout

- Thin dividers: 1 px visible line with a wider hit area for dragging.
- Panels sit flush against the dividers; drop the per-panel border and corner radius.
- Minimum window size 800 × 600, as in Swift. Persist window size, position and splitter positions in settings (with 2.1).

### 3.4 Inspector navigation

- Replace the wrapping `TabControl` with a compact header: a "View:" selector with icon and label per panel (Registers, Memory, Stack, Disassembly, Evaluator, Watchpoints, Breakpoints), as in Swift.
- Persist the selected panel in settings.

### 3.5 Registers and status

- A compact monospace table: name, hex and decimal on one row per register.
- A status strip below it with VM state and N, Z, C, V flags as small pills, as in Swift's `StatusView`.
- Changed-register highlight uses the theme colour and fades out over the existing 1.5 s.

### 3.6 Memory, stack and disassembly

- Monospace tables with column headers, consistent row height and subtle alternate-row shading.
- PC, SP, breakpoint and memory-write markers use theme colours and the same glyphs as the editor gutter.
- Address-entry and jump buttons in a compact header row matching 3.4.

### 3.7 Editor

- Monospace font and size from settings.
- Current-line highlight, and a full-width PC-line background in addition to the gutter arrow.
- Vector breakpoint and PC glyphs, a gutter background and separator from the theme, and line numbers in secondary text.
- Syntax colours defined for light and dark themes (the `.xshd` file references theme colours).

### 3.8 Console

- Monospace output on a terminal-style background.
- An inline input row with a prompt marker; the waiting-for-input state highlights the input row rather than the whole panel border.

### 3.9 Dialogs

- Preferences, About and Examples use the theme spacing, a standard button row (default and cancel), and the platform's button order.

### 3.10 Empty and connection states

- An empty-editor hint: "Open a file (⌘O) or choose an example (⇧⌘E)".
- A connection view shown while the backend starts or after it fails, with the error and a retry button (uses 1.6), in place of a blank window.
- The error bar (1.1) uses the theme's error colours and an icon.

### 3.11 Platform conventions

- Key gestures use the platform modifier: `⌘` on macOS, `Ctrl` elsewhere (Avalonia `PlatformHotkeyConfiguration` or per-platform bindings).
- macOS: a native application menu (`NativeMenu`) with About, Preferences (⌘,) and Quit, plus File, Debug and Window menus. Windows and Linux keep the in-window menu bar.
- Debug menu items mirror the toolbar, with shortcuts shown.

### 3.12 Visual verification

- Screenshot tests with Avalonia.Headless rendering the main window and each panel to PNG in light and dark themes, compared against reviewed baselines with a small pixel tolerance.
- A side-by-side review against the Swift GUI for each panel before closing the phase.

**Exit criteria:** every row in the problem table is closed; screenshot baselines exist for light and dark themes; a side-by-side review against the Swift GUI finds no layout or typography gap.

## Phase 4: Verification

### 4.1 Headless UI tests

Use `Avalonia.Headless.XUnit` (already referenced) with mocked services for:

- Startup with a backend failure shows the error bar.
- Load with assembler errors shows all messages.
- Gutter click and F9 toggle a breakpoint marker.
- Step updates registers, highlights changed registers and moves the PC marker.
- Console input is sent and cleared.
- Panels bind to their child view models (memory, stack and disassembly show data).

### 4.2 Integration suite

- Fix the three integration fixtures. **Decision needed:**
  - Add `.org 0x8000` to the breakpoint test.
  - Replace `MOV R0, #0x12345678` with an encodable immediate or `LDR R0, =0x12345678`.
  - Expect `Halted` from `GET status` straight after load (the backend reports `halted` before any run).
- Replace `[Fact(Skip = ...)]` with a trait-filtered run that starts a backend: CI builds the Go binary, starts it on a free port and runs `dotnet test --filter-trait Category=Integration` against it.
- Add a WebSocket integration test: subscribe, step, and receive a state event with registers.

### 4.3 Cross-platform CI

- Run the Avalonia unit tests on macOS and Windows as well as Ubuntu.
- Run the headless UI tests on all three.

**Exit criteria:** integration tests run unskipped in CI; headless UI tests cover the flows above on three platforms.

## Phase 5: Parity and polish

### 5.1 Parity check against the Swift GUI

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

### 5.2 Accessibility

- Keyboard navigation through every panel; automation names on controls (partly done).

**Exit criteria:** every parity row is checked and either matches or has a tracked follow-up.

## Phase 6: Release

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
3. Phase 3 after Phase 1, alongside Phase 2. 3.1 (theme resources) comes first; every later view task builds on it.
4. Phase 4 alongside Phases 2 and 3: each task adds its headless test once 4.1's harness exists. Screenshot baselines (3.12) wait for Phase 3 to settle.
5. Phase 5, then Phase 6.

## Open decisions

- Monospace font: platform fallback list or a bundled font (3.1).
- Integration fixture fixes (4.2).
- Whether backend items are fixed in Go or stay as GUI workarounds.
- Settings file location on macOS: `~/Library/Application Support` (via `ApplicationData`) or a shared location with the Swift GUI's `UserDefaults`. Sharing is not practical; separate files are assumed.
- Whether to keep the old implementation plan in `docs/` for reference or delete it.
