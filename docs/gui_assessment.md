# GUI Debugger Assessment

## Executive Summary

This document assesses the practicality of extending the current TUI (Text User Interface) debugger with a GUI (Graphical User Interface) debugger written in Go that works on Mac, Windows, and Linux.

**Recommendation: PRACTICAL and RECOMMENDED**

A GUI debugger can be implemented using the **Fyne** framework, providing a modern, cross-platform graphical interface that complements the existing TUI debugger.

## Current State

The ARM emulator currently provides two debugging interfaces:

1. **Command-line debugger** (`--debug` flag): Interactive text-based command prompt
2. **TUI debugger** (`--tui` flag): Full-screen text interface using tview/tcell

Both interfaces use the same underlying `Debugger` interface, which provides:
- Breakpoint management
- Watchpoint support
- Step/continue/run execution control
- Register and memory inspection
- Expression evaluation
- Symbol resolution

## Cross-Platform Go GUI Frameworks Analysis

### 1. Fyne (github.com/fyne-io/fyne/v2) ⭐ **RECOMMENDED**

**Pros:**
- **Pure Go**: No CGO dependencies for most platforms (minimal native code only for rendering)
- **Modern UI**: Material Design-inspired interface with native look and feel
- **Truly Cross-Platform**: Works on Mac, Windows, Linux, iOS, Android, and web
- **Active Development**: Regular updates, responsive maintainers, growing community
- **Good Widget Set**: Sufficient built-in widgets for debugger UI (labels, buttons, lists, tables, text areas, splits, tabs)
- **Straightforward API**: Clear documentation, many examples, reasonable learning curve
- **Reasonable Binary Size**: ~10-15MB compiled binaries (acceptable for a debugger)
- **Built-in Theming**: Dark/light mode support out of the box
- **Good Performance**: Hardware-accelerated rendering where available
- **Excellent Testing Support**: Built-in `fyne.io/fyne/v2/test` package for automated GUI testing (headless, CI/CD ready)

**Cons:**
- Somewhat limited advanced widgets (but adequate for debugger needs)
- Text rendering in tables could be more flexible
- Still maturing compared to decades-old frameworks

**Suitability: HIGH (9/10)**

### 2. Gio (gioui.org)

**Pros:**
- **Pure Go**: No CGO dependencies
- **Immediate Mode GUI**: Similar to Dear ImGui, good for dynamic interfaces
- **High Performance**: Efficient rendering, low overhead
- **Cross-Platform**: Mac, Windows, Linux, iOS, Android, web
- **Portable**: Single codebase for all platforms

**Cons:**
- **Steeper Learning Curve**: Immediate mode paradigm requires different thinking
- **Less Comprehensive Widget Library**: More manual implementation needed
- **More Layout Work**: Requires explicit layout calculations
- **Smaller Community**: Fewer examples and resources than Fyne

**Suitability: MEDIUM (6/10)** - Good framework but requires more development effort

### 3. Wails (github.com/wailsapp/wails/v2)

**Pros:**
- **Web Technologies**: Uses HTML/CSS/JavaScript for UI (familiar to web developers)
- **Modern Look**: Can achieve sophisticated, modern designs
- **Cross-Platform**: Mac, Windows, Linux
- **Rich Ecosystem**: Leverage existing web UI libraries (React, Vue, Svelte)
- **Good Documentation**: Well-documented with tutorials

**Cons:**
- **Not Pure Native**: Uses embedded browser (WebView2/WebKit)
- **Larger Binaries**: 15-30MB+ due to browser embedding
- **Platform Dependencies**: Requires WebView2 on Windows, WebKit on Mac/Linux
- **More Complex Build**: Requires Node.js toolchain for UI development
- **Separation of Concerns**: Go backend + JS frontend adds complexity

**Suitability: MEDIUM-LOW (5/10)** - Capable but adds significant complexity

### 4. Go-GTK (gotk3)

**Pros:**
- **Mature**: GTK is a well-established framework
- **Feature-Rich**: Comprehensive widget set
- **Native Look**: GTK styling on Linux

**Cons:**
- **CGO Required**: Must have CGO enabled
- **Platform Dependencies**: Requires GTK+ installation (3.x or 4.x)
- **Windows Support**: Problematic, requires MSYS2/MinGW
- **Mac Support**: Requires MacPorts/Homebrew installation
- **Not Pure Go**: Heavy reliance on C bindings
- **Distribution Challenges**: End users need GTK installed

**Suitability: LOW (3/10)** - Too many dependencies and platform issues

### 5. Walk (github.com/lxn/walk)

**Pros:**
- **Native Windows**: Uses Win32 API directly
- **Good for Windows**: Excellent Windows integration

**Cons:**
- **Windows Only**: Does not meet cross-platform requirement
- **CGO Required**: Uses C bindings to Win32

**Suitability: LOW (2/10)** - Not cross-platform

### 6. Qt (therecipe/qt or go-qt)

**Pros:**
- **Mature**: Qt is industry-standard, feature-complete
- **Professional**: High-quality widgets and tools

**Cons:**
- **Licensing**: LGPL or commercial license required
- **Large Dependency**: Requires Qt framework installation (~500MB+)
- **CGO Required**: Heavy C++ bindings
- **Complex Build**: Qt toolchain required
- **Very Large Binaries**: 20-50MB+ executables

**Suitability: LOW (2/10)** - Too heavy and complex for this use case

## Recommendation: Fyne

For the ARM emulator GUI debugger, **Fyne** is the most practical choice:

### Key Advantages

1. **Pure Go Philosophy**: Aligns with the project's Go-based architecture. Minimal native dependencies.

2. **True Cross-Platform**: Single codebase works on Mac, Windows, and Linux without platform-specific code or conditional compilation.

3. **No Installation Required**: End users don't need to install frameworks or dependencies. Single binary distribution.

4. **Sufficient Widget Set**: Has everything needed for a debugger:
   - `widget.Label` - Register displays, status information
   - `widget.Entry` / `widget.TextGrid` - Source code viewer, console output
   - `widget.Button` - Control buttons (Run, Step, Continue, etc.)
   - `widget.List` / `widget.Table` - Breakpoints, memory view
   - `container.Split` - Resizable panels
   - `container.Tabs` - Multiple views (source, disassembly, memory)
   - `widget.Toolbar` - Quick access to common actions

5. **Integrates Well**: Can reuse existing `Debugger` interface (same as TUI), minimal code changes.

6. **Modern Look**: Material Design provides a clean, professional appearance.

7. **Active Community**: Regular releases, responsive maintainers, growing ecosystem.

8. **Performance**: Hardware-accelerated rendering provides smooth UI updates.

## Implementation Plan

### Phase 1: Basic GUI Structure (2-4 hours)

Create `debugger/gui.go` with:
- Main window initialization
- Layout with panels for: source, registers, memory, stack, console, breakpoints
- Control toolbar (Run, Step, Step Over, Continue, Reset)
- Basic integration with existing Debugger interface

### Phase 2: Core Functionality (3-5 hours)

- Connect GUI controls to Debugger commands
- Implement register and memory display updates
- Add breakpoint management (add/remove/list)
- Console output redirection
- PC tracking and source highlighting

### Phase 3: Enhanced Features (2-4 hours)

- Memory search and editing
- Watch expressions
- Disassembly view
- Keyboard shortcuts (F5=Run, F9=Breakpoint, F10=Step Over, F11=Step Into)
- Source code syntax highlighting (basic)

### Phase 4: Polish (1-3 hours)

- Dark/light theme support
- Window state persistence (size, position)
- Preferences dialog
- Error handling and user feedback
- Documentation

**Total Estimated Effort: 8-16 hours** (manageable for a focused implementation)

### Integration with Existing Code

The GUI debugger will integrate seamlessly:

```go
// In main.go, add new flag:
guiMode := flag.Bool("gui", false, "Use GUI debugger")

// In debugger launch section:
if *guiMode {
    if err := debugger.RunGUI(dbg); err != nil {
        fmt.Fprintf(os.Stderr, "GUI error: %v\n", err)
        os.Exit(1)
    }
}
```

The GUI will use the same `Debugger` interface as TUI:
- `dbg.Step()` - Execute one instruction
- `dbg.Continue()` - Run until breakpoint
- `dbg.AddBreakpoint()` - Set breakpoint
- `dbg.VM.CPU.Registers` - Access register values
- `dbg.VM.Memory` - Access memory
- Symbol table integration (existing)

### Dependencies

Only one new dependency required:

```go
require (
    github.com/fyne-io/fyne/v2 v2.4.5
)
```

Current project already has:
- `github.com/rivo/tview` - TUI framework (24KB in go.mod)
- `github.com/gdamore/tcell/v2` - Terminal handling

Adding Fyne will increase binary size by ~10-15MB (acceptable for a debugger with GUI).

## Platform-Specific Considerations

### macOS
- **Works**: Fyne supports macOS 10.12+ (native Cocoa integration)
- **Build**: `go build` works directly
- **Distribution**: Single binary or .app bundle

### Windows
- **Works**: Fyne supports Windows 10+
- **Build**: `go build` works directly (with Go 1.16+ no CGO needed for most cases)
- **Distribution**: Single .exe binary
- **Note**: Windows Defender may flag unsigned executables (standard for all Go GUI apps)

### Linux
- **Works**: Fyne supports major distributions (Ubuntu, Fedora, Arch, etc.)
- **Requirements**: X11 or Wayland, OpenGL (standard on modern Linux desktops)
- **Build**: `go build` works directly
- **Distribution**: Single binary
- **Packaging**: Can create .deb, .rpm, AppImage, or Flatpak

## Risks and Mitigations

### Risk 1: Binary Size Increase
**Impact**: Medium
**Mitigation**: 
- Fyne adds ~10-15MB to binary (acceptable for modern systems)
- GUI is optional feature (users can still use TUI or CLI)
- Use build tags to optionally exclude GUI if needed

### Risk 2: Platform-Specific Issues
**Impact**: Low
**Mitigation**:
- Fyne has good cross-platform support and testing
- Large user base helps catch platform-specific bugs quickly
- Fallback to TUI/CLI if GUI fails to initialize

### Risk 3: Learning Curve
**Impact**: Low
**Mitigation**:
- Fyne has clear documentation and many examples
- Similar concepts to other GUI frameworks
- Can prototype basic UI in a few hours

### Risk 4: Maintenance
**Impact**: Low
**Mitigation**:
- Fyne is actively maintained with regular releases
- Breaking changes are rare and well-documented
- Strong backward compatibility commitment

## Alternatives Considered

### Alternative 1: Web-Based UI (HTTP server)
**Approach**: Embed HTTP server, serve web UI
**Pros**: Universal browser access, rich UI possibilities
**Cons**: More complex (server + client), security concerns (open port), requires browser
**Verdict**: Overkill for desktop debugger

### Alternative 2: Keep TUI Only
**Approach**: Don't add GUI, enhance TUI instead
**Pros**: No new dependencies, consistent with current design
**Cons**: TUI inherently limited (no mouse, limited colors, text-only)
**Verdict**: TUI is good but GUI would be valuable addition

### Alternative 3: Hybrid (TUI + Web)
**Approach**: Enhance TUI to optionally serve web interface
**Pros**: No GUI framework needed
**Cons**: Complex architecture, security concerns, not truly native
**Verdict**: Unnecessarily complex

## Testing Strategy

### Automated Testing with Fyne

**Fyne provides excellent built-in testing support** through the `fyne.io/fyne/v2/test` package:

- **Headless Testing**: Tests run without requiring a display server
- **CI/CD Ready**: Works in GitHub Actions, GitLab CI, and other automated environments
- **Widget Interaction**: Simulate clicks, typing, and other user interactions
- **Visual Testing**: Capture and compare widget renderings
- **Full Coverage**: Can test all GUI functionality automatically

See `docs/gui_testing.md` for comprehensive testing documentation.

### Development Testing
1. Write automated tests for GUI components (using `fyne.io/fyne/v2/test`)
2. Test on Linux first (typically easiest)
3. Add unit tests for GUI-independent logic
4. Manual testing of GUI interactions
5. Automated screenshot comparison tests (optional)

### Example Automated Test

```go
func TestGUIBreakpoints(t *testing.T) {
    // Create test app (no display needed)
    app := test.NewApp()
    defer app.Quit()
    
    // Create and test GUI components
    gui := newGUI(debugger)
    gui.addBreakpoint()
    
    if len(gui.breakpoints) != 1 {
        t.Error("Breakpoint not added")
    }
}
```

### Cross-Platform Testing
1. Use CI/CD to build for all platforms
2. Automated tests run in CI (headless)
3. Manual testing on Mac, Windows, Linux
4. Community testing through releases

### Integration Testing
1. Verify GUI works with all debugger commands
2. Test with various ARM programs
3. Ensure feature parity with TUI where applicable
4. Test edge cases (no source file, memory errors, etc.)
5. Automated workflow tests (step, breakpoint, continue, etc.)

## Documentation Requirements

### User Documentation
1. Update `README.md` with `--gui` flag
2. Create `docs/gui_debugger_guide.md` with screenshots
3. Add GUI section to debugger reference
4. Update installation guide (Fyne requirements if any)

### Developer Documentation
1. Document GUI architecture in `docs/architecture.md`
2. Add GUI extension points
3. Document testing approach
4. Add troubleshooting guide

## Conclusion

**Implementing a GUI debugger using Fyne is practical and recommended.**

### Key Points

✅ **Technically Feasible**: Fyne provides all necessary capabilities
✅ **Cross-Platform**: True Mac, Windows, Linux support in single codebase  
✅ **Reasonable Effort**: 8-16 hours for full implementation
✅ **Good Integration**: Works with existing Debugger interface
✅ **No CGO**: Pure Go (mostly) simplifies builds and distribution
✅ **Modern**: Professional appearance and good UX
✅ **Maintainable**: Active project with good community support
✅ **Testable**: Built-in headless testing framework for automated GUI tests

### Next Steps

1. **Prototype**: Create basic Fyne window with debugger panels (2-4 hours)
2. **Integrate**: Connect to existing Debugger interface (2-3 hours)
3. **Enhance**: Add features to match/exceed TUI capabilities (3-6 hours)
4. **Polish**: Themes, shortcuts, preferences (1-3 hours)
5. **Test**: Cross-platform testing and bug fixes (2-4 hours)
6. **Document**: User guide with screenshots (1-2 hours)

**Total: 11-22 hours** for complete, polished implementation

The GUI debugger would significantly enhance the project by:
- Providing a more user-friendly debugging experience
- Attracting users who prefer graphical tools
- Enabling advanced features (drag-and-drop, visual memory editing, etc.)
- Complementing (not replacing) the excellent TUI debugger

## Appendix: Code Structure

### Proposed File Structure

```
debugger/
├── interface.go       # Existing - RunCLI(), RunTUI()
├── gui.go            # NEW - RunGUI() and GUI implementation
├── gui_panels.go     # NEW - GUI panel implementations
├── tui.go            # Existing - TUI implementation
├── debugger.go       # Existing - Core Debugger interface
├── commands.go       # Existing - Command execution
├── breakpoints.go    # Existing - Breakpoint management
└── expressions.go    # Existing - Expression evaluation
```

### Example GUI Initialization

```go
// gui.go
package debugger

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type GUI struct {
    Debugger *Debugger
    App      fyne.App
    Window   fyne.Window
    
    // Panels
    SourceView      *widget.TextGrid
    RegisterView    *widget.TextGrid
    MemoryView      *widget.TextGrid
    StackView       *widget.TextGrid
    BreakpointsList *widget.List
    Console         *widget.TextGrid
    
    // Controls
    Toolbar         *widget.Toolbar
}

func RunGUI(dbg *Debugger) error {
    gui := newGUI(dbg)
    gui.Window.ShowAndRun()
    return nil
}
```

This structure mirrors the TUI implementation, making it easy to maintain both interfaces.

## References

- [Fyne Documentation](https://developer.fyne.io/)
- [Fyne GitHub](https://github.com/fyne-io/fyne)
- [Fyne Examples](https://github.com/fyne-io/examples)
- [ARM2 Specification](../SPECIFICATION.md)
- [Current Debugger Reference](./debugger_reference.md)
