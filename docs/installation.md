# Installation Guide

This guide will help you install and set up the ARM2 Emulator on your system.

## Prerequisites

- **Go 1.27 or later** (only required if building from source)
- **Git** (optional) - For cloning the repository
- **Xcode 27 on macOS 26.2** (only required to build the Swift GUI from source)
- **.NET SDK 10.0 or later** (only required to build the Avalonia GUI from source)

## Supported Platforms

The ARM2 Emulator is cross-platform and supports:

- **macOS** (Intel and Apple Silicon)
- **Windows** (10/11)
- **Linux** (Ubuntu, Fedora, Arch, and other distributions)

## Installation Methods

### Method 1: Download Pre-built Binaries (Recommended)

Pre-built binaries are available for download from the [Releases](https://github.com/lookbusy1344/arm_emulator/releases) page.

**Command-line emulator and TUI debugger:**
- **Linux (64-bit):** `arm-emulator-linux-amd64`
- **macOS (Apple Silicon):** `arm-emulator-macos-arm64`
- **Windows:** `arm-emulator-win-amd64.exe` (AMD64/x64) and `arm-emulator-win-arm64.exe` (ARM64)

**GUI apps** (each archive contains the app and the `arm-emulator` backend it starts):
- **Swift, macOS (Apple Silicon):** `arm-emulator-swift-gui.app.tar.gz`
- **Avalonia, macOS (Apple Silicon):** `arm-emulator-avalonia-macos-arm64.app.tar.gz`
- **Avalonia, Linux (64-bit):** `arm-emulator-avalonia-linux-amd64.tar.gz`
- **Avalonia, Windows:** `arm-emulator-avalonia-win-amd64.zip` and `arm-emulator-avalonia-win-arm64.zip`

`SHA256SUMS` lists the checksum of every file.

**To install:**
1. Visit the [Releases](https://github.com/lookbusy1344/arm_emulator/releases) page
2. Download the file for your platform
3. On Linux/macOS, make the command-line binary executable: `chmod +x arm-emulator-*`
4. Optionally verify the download: `shasum -a 256 -c SHA256SUMS --ignore-missing`

The macOS apps are ad-hoc signed, not notarised. On first launch, right-click the app and choose **Open**, or clear the quarantine flag with `xattr -dr com.apple.quarantine ARMEmulator.app`.

### Method 2: Build from Source

#### 1. Clone the Repository

```bash
git clone https://github.com/lookbusy1344/arm_emulator.git
cd arm_emulator
```

Or download and extract the source code archive.

#### 2. Build the Emulator

```bash
go build -o arm-emulator
```

This will create an executable named `arm-emulator` (or `arm-emulator.exe` on Windows) in the current directory.

#### 3. Verify Installation

```bash
./arm-emulator --version
```

You should see the version information displayed.

#### 4. (Optional) Install Globally

**On macOS/Linux:**

```bash
sudo mv arm-emulator /usr/local/bin/
```

**On Windows:**

Add the directory containing `arm-emulator.exe` to your PATH environment variable.

### Method 3: Build a GUI App from Source

Both GUI apps drive the `arm-emulator` backend over its HTTP API. Build the backend first with `make build` at the repository root; each app finds it there when run from source.

- **Swift app (macOS):** see [SWIFT_APP.md](SWIFT_APP.md)
- **Avalonia app (Windows, macOS, Linux):** see [avalonia-gui/README.md](../avalonia-gui/README.md)

[GUI.md](GUI.md) compares the two.

## Testing the Installation

### 1. Run a Simple Program

Create a file called `test.s`:

```asm
        .org    0x8000

_start:
        LDR     R0, =msg
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        MOV     R0, #0
        SWI     #0x00           ; EXIT

msg:
        .asciz  "ARM2 Emulator is working!"
```

### 2. Run the Program

```bash
./arm-emulator test.s
```

You should see:
```
ARM2 Emulator is working!
```

## Configuration (Optional)

The emulator can be configured using a configuration file:

**macOS/Linux:** `~/.config/arm-emu/config.toml`
**Windows:** `%APPDATA%\arm-emu\config.toml`

Example configuration:

```toml
[emulator]
default_origin = 0x8000
stack_size = 1048576       # 1 MB

[debugger]
enable_tui = true
history_size = 1000

[output]
show_cycle_count = false
show_register_changes = false
```

## Troubleshooting

### "Command not found" Error

**Problem:** The system can't find the `arm-emulator` command.

**Solution:**
- Make sure the emulator is in your PATH, or
- Run it with the full path: `/path/to/arm-emulator program.s`, or
- Use `./arm-emulator` if you're in the same directory

### Permission Denied (macOS/Linux)

**Problem:** `permission denied` when trying to run the emulator.

**Solution:** Make the file executable:
```bash
chmod +x arm-emulator
```

### Go Version Error

**Problem:** Error about Go version being too old.

**Solution:** Install Go 1.27 or later:
- **macOS:** `brew install go`
- **Linux:** Follow instructions at https://go.dev/doc/install
- **Windows:** Download installer from https://go.dev/dl/

### Build Errors

**Problem:** Errors when running `go build`.

**Solution:**
1. Make sure you're in the correct directory (should contain `go.mod`)
2. Run `go mod tidy` to update dependencies
3. Check that you have Go 1.27 or later: `go version`

### Terminal UI Not Working

**Problem:** TUI mode doesn't display correctly.

**Solution:**
- Make sure your terminal supports ANSI colors and cursor control
- On Windows, use Windows Terminal or PowerShell (not cmd.exe)
- Try running in non-TUI mode first

## Updating

To update to the latest version:

```bash
# If installed from source
cd arm_emulator
git pull
go build -o arm-emulator

# If using pre-built binary
# Download the latest release and replace your existing binary
```

## Uninstalling

### If Installed Globally

**macOS/Linux:**
```bash
sudo rm /usr/local/bin/arm-emulator
rm -rf ~/.config/arm-emu
```

**Windows:**
- Remove `arm-emulator.exe` from your installation directory
- Delete `%APPDATA%\arm-emu` folder

### If Installed Locally

Simply delete the emulator executable and source directory.

## Next Steps

- Read the [Tutorial](TUTORIAL.md) to learn ARM2 assembly from scratch
- Browse the [Example Programs](../examples/README.md) - 49 programs with descriptions
- Study the [Instruction Set Reference](INSTRUCTIONS.md) - ARM2 CPU instructions and syscalls
- Learn [Assembler Directives](ASSEMBLER.md) - .text, .data, .word, .ltorg, etc.
- Check the [Programming Reference](REFERENCE.md) - Condition codes, addressing modes, shifts
- Read the [Assembly Reference](assembly_reference.md) for comprehensive syntax guide
- Learn about the [Debugger](debugger_reference.md) and try the [Debugging Tutorial](debugging_tutorial.md)

## Getting Help

- **Documentation:** See `docs/` directory
- **Examples:** See `examples/` directory
- **Issues:** Report bugs at https://github.com/lookbusy1344/arm_emulator/issues
- **FAQ:** Check [FAQ.md](FAQ.md) for common questions

## System Requirements

- **RAM:** 100MB minimum (more for large programs)
- **Disk Space:** 20MB for emulator binary
- **Terminal:** ANSI-capable terminal (for TUI mode)
- **OS:** Any system supported by Go 1.27+
