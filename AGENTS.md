# ARM Emulator Project

ARM emulator in Go with HTTP REST API backend. Two GUI frontends (Swift + Avalonia) connect to the same API.

## Version control (read first)

Before the first VCS command, run `jj --ignore-working-copy root`. This may be a jj repo on one machine and plain Git on another.

**IMPORTANT:** if it succeeds, use `jj` for all VCS commands, including `log`, `show`, `status` and `diff`. Do not run `git` on jj repos. The `gitStatus` snapshot in the session context is not a reason to use git.

jj has no commit hooks. Run `scripts/pre-commit.sh` before `jj commit`, `jj describe` (when finalising a change) and `jj squash`. It reads the files changed in `@` and runs only the suites they select (`scripts/checks.sh`):

| Suite | Paths | Checks |
|-------|-------|--------|
| Go | `*.go`, `go.mod`, `go.sum`, `.golangci.yml`, `Makefile`, `examples/`, `include/`, `tests/` | `gofmt -l`, `make build`, `golangci-lint`, `go test -count=1` |
| Swift | `swift-gui/` | `make build`, `xcodegen generate`, `swiftformat --lint`, `swiftlint --strict`, `xcodebuild` build and test |
| .NET | `avalonia-gui/` | restore, build, `dotnet format --verify-no-changes`, `dotnet test` |

`*.md` and `docs/` select nothing. `scripts/checks.sh --all` runs every suite.

Before `jj git push` or moving a shared bookmark, run `scripts/pre-push.sh`. By default it formats the tip (newest non-empty mutable revision) with `jj fix` (`scripts/format-stdin.sh`: gofmt, swiftformat, `dotnet format whitespace`), then runs the suites selected by the files changed across the unpushed stack. `--full` formats every mutable revision in `::@` and checks each in its own checkout via `jj run`, with cold Xcode and .NET builds per revision. `scripts/agent-pre-push-hook.sh` runs the tip check on every push command and blocks the push on failure. Claude Code (`.claude/settings.json`) and Codex (`.codex/hooks.json`) call it as a `PreToolUse` hook. The hook checks the tip of `@`, not the bookmark being pushed, so push only the bookmark you are working on: the one on `@-`, with an empty `@` above it.

In a plain Git checkout both scripts work without jj. `pre-commit.sh` reads the files changed against `HEAD`. `pre-push.sh` checks the working tree for the files changed since the upstream, or every suite when there is no upstream, and does not format. Install them as Git hooks with `ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit` and `ln -sf ../../scripts/pre-push.sh .git/hooks/pre-push`. `git push` from an agent triggers the same `PreToolUse` hook.

Push only on explicit request. "Push this" means: if `@` is non-empty, `jj commit` it (after the pre-commit checks). Move the bookmark to `@-` with `jj bookmark set <name> -r @-`, then `jj git push --bookmark <name>`. Use the bookmark already on the stack; otherwise `main`. Never push any other bookmark. Do not use `jj git push -c`.

## Personal information

Exclude PII from every commit, commit message and bookmark name: real names, email addresses, usernames, machine paths such as `/Users/<name>/`, hostnames, tokens and credentials. Check the diff before `jj commit`, `jj describe` (finalising) and `git commit`.

## General

**⚠️ CRITICAL: API SYNCHRONIZATION**
- Go backend API is shared by **both Swift GUI and Avalonia GUI**
- **DO NOT make breaking API changes** - only additive changes allowed
- Any API modifications must work with both frontends
- Test both GUIs after backend changes

## Go Backend (Core Emulator + API)

### Build & Test

```bash
# Build with version info
make build

# Format, lint, test (MANDATORY before commit)
go fmt ./...
golangci-lint run ./...
go clean -testcache && go test ./...

# Run emulator
./arm-emulator program.s
```

**Test Organization:** Tests in `./tests/unit/` and `./tests/integration/` (exception: `debugger/tui_internal_test.go`)

## Swift GUI (macOS, Primary GUI)

**Prerequisites:** macOS 26.2, Swift 6.4, Xcode 27.0

```bash
cd swift-gui

# Build & test
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# Format & lint (MANDATORY before commit - 0 violations required)
swiftformat .
swiftlint
```

**See `swift-gui/AGENTS.md` for complete Swift development guide.**

## Avalonia GUI (Cross-Platform: Windows/macOS/Linux)

**Prerequisites:** .NET SDK 10.0+ from [https://dot.net](https://dot.net)

### Build & Test

```bash
cd avalonia-gui

# Build & run
dotnet build
dotnet run --project ARMEmulator

# Test
dotnet test

# Format (MANDATORY before commit - must build and pass tests)
# Note: dotnet format is run automatically after every file change
dotnet format
dotnet build
dotnet test
```

**Architecture:** MVVM with ReactiveUI. Connects via HTTP REST API + WebSocket to Go backend. Uses C# 13 features (primary constructors, collection expressions, records, pattern matching, immutable collections).

**Note:** `dotnet format` runs automatically after every change to the Avalonia project.

**Docs:** `docs/AVALONIA_IMPLEMENTATION_PLAN.md`

## Project Structure

```
├── api/           - HTTP REST API backend (port 8080)
├── service/       - Service layer for API/GUI integration
├── vm/            - Virtual machine implementation
├── parser/        - Assembly parser with macros
├── instructions/  - Instruction implementations
├── swift-gui/     - Swift native macOS GUI
├── avalonia-gui/  - Avalonia .NET cross-platform GUI
├── tests/         - Unit and integration tests
├── examples/      - 49 example ARM assembly programs
└── docs/          - Documentation
```

## Additional Documentation

- **Instructions & Syscalls:** [docs/INSTRUCTIONS.md](docs/INSTRUCTIONS.md)
- **Swift GUI Architecture:** `SWIFT_GUI_PLANNING.md`, `docs/SWIFT_CLI_AUTOMATION.md`, `docs/MCP_UI_DEBUGGING.md`
- **Avalonia GUI Plan:** `docs/AVALONIA_IMPLEMENTATION_PLAN.md`
- **Diagnostic Modes:** Code coverage, stack trace, flag trace, register trace (see `--help`)
