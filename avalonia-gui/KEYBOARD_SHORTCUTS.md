# Keyboard Shortcuts Reference

Comprehensive keyboard shortcuts for the ARM Emulator Avalonia GUI.

## File Operations

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Open File** | `Ctrl+O` | `Cmd+O` | Open an assembly file (.s) |
| **Save** | `Ctrl+S` | `Cmd+S` | Save current file |
| **Save As** | `Ctrl+Shift+S` | `Cmd+Shift+S` | Save with a new filename |
| **Open Example** | `Ctrl+Shift+E` | `Cmd+Shift+E` | Open examples browser |
| **Preferences** | `Ctrl+,` | `Cmd+,` | Open preferences window |

## Execution Control

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Run/Continue** | `F5` or `Ctrl+R` | `F5` or `Cmd+R` | Start or continue execution |
| **Pause** | `Ctrl+.` | `Cmd+.` | Pause running program |
| **Step** | `F11` or `Ctrl+T` | `F11` or `Cmd+T` | Execute single instruction |
| **Step Over** | `F10` or `Ctrl+Shift+T` | `F10` or `Cmd+Shift+T` | Step over function calls |
| **Step Out** | `Ctrl+Alt+T` | `Cmd+Option+T` | Step out of current function |
| **Reset** | `Ctrl+Shift+R` | `Cmd+Shift+R` | Reset VM to initial state |
| **Load Program** | `Ctrl+L` | `Cmd+L` | Load current program into VM |

## Navigation

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Show PC** | `Ctrl+J` | `Cmd+J` | Scroll editor to current PC |
| **Toggle Breakpoint** | `F9` | `F9` | Toggle breakpoint on current line |

## Editor

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Find** | `Ctrl+F` | `Cmd+F` | Open find dialog (AvaloniaEdit) |
| **Find Next** | `F3` | `F3` | Find next occurrence |
| **Replace** | `Ctrl+H` | `Cmd+H` | Open find/replace dialog |
| **Go to Line** | `Ctrl+G` | `Cmd+G` | Go to line number |
| **Select All** | `Ctrl+A` | `Cmd+A` | Select all text |
| **Undo** | `Ctrl+Z` | `Cmd+Z` | Undo last edit |
| **Redo** | `Ctrl+Y` or `Ctrl+Shift+Z` | `Cmd+Shift+Z` | Redo last undone edit |
| **Cut** | `Ctrl+X` | `Cmd+X` | Cut selected text |
| **Copy** | `Ctrl+C` | `Cmd+C` | Copy selected text |
| **Paste** | `Ctrl+V` | `Cmd+V` | Paste from clipboard |

## Window Management

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Close Window** | `Alt+F4` | `Cmd+W` | Close current window |
| **Quit Application** | `Alt+F4` | `Cmd+Q` | Quit application |

## Console

| Action | Windows/Linux | macOS | Description |
|--------|---------------|-------|-------------|
| **Send Input** | `Enter` | `Enter` | Send input to running program |

## Accessibility

All keyboard shortcuts are designed to work with:
- Screen readers
- High contrast themes
- Keyboard-only navigation

## Customization

Keyboard shortcuts are defined in the application and cannot currently be customized. Future versions may support custom key bindings.

## Platform-Specific Notes

### Windows
- Uses `Ctrl` as the primary modifier key
- `Alt+F4` closes windows (standard Windows behavior)
- Right-click context menus available in editor

### macOS
- Uses `Cmd` (⌘) as the primary modifier key
- `Cmd+W` closes windows, `Cmd+Q` quits application
- Native macOS menu bar integration
- `Option` key equivalent to `Alt` on Windows

### Linux
- Uses `Ctrl` as the primary modifier key (like Windows)
- Window management shortcuts depend on desktop environment
- GTK-style keyboard navigation

## Tips

- **Editor Focus**: Many shortcuts require the editor to have keyboard focus
- **Modal Dialogs**: Shortcuts are disabled when modal dialogs (Preferences, About) are open
- **Execution State**: Some shortcuts are only enabled in specific VM states:
  - **Run/Continue**: Only when idle or at breakpoint
  - **Pause**: Only when running or waiting for input
  - **Step/Step Over/Step Out**: Only when idle or at breakpoint

## See Also

- [README.md](README.md) - General build and run instructions
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration options
- [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md) - Running integration tests
