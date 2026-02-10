# Phase 9 Testing Guide

This guide covers testing the file operations and dialogs implemented in Phase 9.

## Prerequisites

- .NET 10 SDK installed
- Go backend built (`make build` in project root)
- macOS 26.2 (primary development platform)

## Build & Run

```bash
cd avalonia-gui

# Build
dotnet build

# Run application
dotnet run --project ARMEmulator

# Run tests
dotnet test
```

## Test Checklist

### 1. File Menu Commands

- [ ] **Open File (Ctrl+O)**
  - Click File > Open or press Ctrl+O
  - Select an assembly file (.s)
  - Verify file content loads in editor
  - Verify file is added to Recent Files

- [ ] **Save File (Ctrl+S)**
  - Make changes to editor
  - Click File > Save or press Ctrl+S
  - Verify file is saved to current location
  - For new files, should show Save As dialog

- [ ] **Save As (Ctrl+Shift+S)**
  - Click File > Save As or press Ctrl+Shift+S
  - Enter new filename
  - Verify file is saved to new location
  - Verify file is added to Recent Files

- [ ] **Recent Files**
  - Open several files
  - Click File > Recent Files
  - Verify list shows recently opened files
  - Click a recent file to open it
  - Verify file loads correctly

- [ ] **Examples Browser (Ctrl+Shift+E)**
  - Click File > Examples or press Ctrl+Shift+E
  - Search for examples (try "fib", "hello", etc.)
  - Verify search filters list
  - Select an example
  - Verify preview shows on right
  - Click Load
  - Verify example loads in editor

- [ ] **Preferences (Ctrl+,)**
  - Click File > Preferences or press Ctrl+,
  - **General Tab:**
    - Modify Backend URL
    - Change Theme (Auto/Light/Dark)
    - Adjust Recent Files Limit
    - Toggle Auto-scroll memory writes
  - **Editor Tab:**
    - Adjust Font Size (10-24)
    - Verify live preview updates
  - Click OK
  - Verify settings are applied (font size should change)

- [ ] **About Dialog**
  - Click File > About
  - Verify app version displays
  - Verify backend version loads (if backend is running)
  - If backend not running, should show "Not available"
  - Verify commit hash and build date appear
  - Click Close

### 2. Keyboard Shortcuts

Verify all keyboard shortcuts work:
- Ctrl+O: Open
- Ctrl+S: Save
- Ctrl+Shift+S: Save As
- Ctrl+Shift+E: Examples
- Ctrl+,: Preferences

### 3. Dialog Behavior

- [ ] All dialogs center on parent window
- [ ] Dialogs are modal (parent disabled while open)
- [ ] Cancel closes without applying changes
- [ ] OK/Load buttons apply changes

### 4. File Service Integration

- [ ] Opening file updates editor content
- [ ] Opening file auto-loads program
- [ ] Saving file preserves content
- [ ] Recent files list updates correctly
- [ ] Recent files persists across sessions (TODO: needs persistence implementation)

### 5. Error Handling

- [ ] Try opening non-existent file from Recent Files
  - Should show error message
  - Should not crash
- [ ] Try opening invalid assembly file
  - Should load but show parse errors
- [ ] Try examples with backend not running
  - Should show error in examples browser

## Platform-Specific Testing

### macOS
- [ ] Native file dialogs appear
- [ ] Menu bar integrates with system
- [ ] Keyboard shortcuts follow macOS conventions
- [ ] App bundle structure correct

### Windows (if available)
- [ ] Native file dialogs appear
- [ ] Menu bar appears in window
- [ ] Keyboard shortcuts work
- [ ] .exe runs standalone

### Linux (if available)
- [ ] File dialogs work
- [ ] Menu bar appears
- [ ] Keyboard shortcuts work

## Known Limitations

1. **Settings Persistence**: Not yet implemented. Settings reset on restart.
2. **Recent Files Persistence**: Not yet implemented. List clears on restart.
3. **Theme Changes**: Require app restart to take effect.

## Test Results

All 223 unit tests passing:
- AppSettings: 6 tests
- FileService: 8 tests
- AboutWindowViewModel: 5 tests
- PreferencesWindowViewModel: 6 tests
- ExamplesBrowserViewModel: 6 tests
- MainWindowViewModel: 10 tests
- Other ViewModels and Services: 182 tests

Build status: ✅ Clean build with 0 warnings, 0 errors
