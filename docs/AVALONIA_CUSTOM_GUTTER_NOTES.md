# Custom Gutter Implementation Notes - Avalonia GUI

> **Status Update (2026-02-09):** The original version incompatibility issue has been resolved by upgrading to `Avalonia.AvaloniaEdit 11.*`. This document now serves as:
> 1. Historical reference for the problem encountered
> 2. Comprehensive guide for implementing custom gutter features (adorner-based approach remains valid)
> 3. Technical reference for future custom editor extensions

## Problem Statement

The ARM Emulator's Avalonia GUI requires a custom editor gutter that displays:
- **Line numbers** (handled by AvaloniaEdit's built-in feature)
- **Breakpoint markers** - Red circles indicating active breakpoints
- **Current PC indicator** - Blue arrow showing the current program counter position
- **Interactive clicks** - Toggle breakpoints by clicking the gutter

## Challenge: Version Incompatibility (RESOLVED)

### Root Cause (Historical)

**AvaloniaEdit 0.10.12** was built against **Avalonia 0.10.x**, while the project uses **Avalonia 11.3.x**. This created a type mismatch:

```
error CS7069: Reference to type 'IControl' claims it is defined in
'Avalonia.Controls', but it could not be found

error CS1503: Argument 2: cannot convert from 'EditorGutterMargin'
to 'Avalonia.Controls.IControl'
```

### Resolution (2026-02-09)

**Package renamed for Avalonia 11:**
- Old package: `AvaloniaEdit` (0.10.12, Avalonia 0.10.x)
- New package: `Avalonia.AvaloniaEdit` (11.4.1+, Avalonia 11.x)
- **Status:** ✅ Version mismatch resolved, project now uses `Avalonia.AvaloniaEdit 11.*`

### Technical Details (Historical)

- `AbstractMargin` from AvaloniaEdit 0.10 inherited from an older `Control` type
- `TextEditor.TextArea.LeftMargins` collection expected `IControl` from Avalonia 0.10
- Custom `EditorGutterMargin : AbstractMargin` couldn't be added due to interface mismatch
- **Update:** The package was renamed, not abandoned - see `Avalonia.AvaloniaEdit` on NuGet

### Attempted Solution

Created `EditorGutterMargin` extending `AbstractMargin`:
```csharp
public class EditorGutterMargin : AbstractMargin
{
    public static readonly StyledProperty<ImmutableHashSet<int>> BreakpointLinesProperty = ...;
    public static readonly StyledProperty<int?> CurrentPCLineProperty = ...;

    public override void Render(DrawingContext context) {
        // Render line numbers, breakpoints, PC arrow
    }

    protected override void OnPointerPressed(...) {
        // Handle clicks for breakpoint toggle
    }
}
```

**Failed because:** `TextEditor.TextArea.LeftMargins.Insert(0, _gutterMargin)` throws type conversion error.

---

## Potential Solutions

### Option 1: Adorner Layer Overlay (Recommended)

**Approach:** Use Avalonia's adorner layer to draw breakpoints and PC indicators on top of the editor.

**Pros:**
- No version conflicts - uses pure Avalonia APIs
- Clean separation of concerns
- Full control over rendering and z-order
- Can animate breakpoint/PC indicators

**Cons:**
- More complex coordinate calculation
- Need to track scroll position manually
- Slightly more boilerplate

**Implementation Sketch:**
```csharp
public class EditorGutterAdorner : Control
{
    private readonly TextEditor _editor;

    public EditorGutterAdorner(TextEditor editor)
    {
        _editor = editor;
        ClipToBounds = false;

        // Subscribe to editor scroll/layout changes
        _editor.TextArea.TextView.ScrollOffsetChanged += (s, e) => InvalidateVisual();
        _editor.TextArea.TextView.VisualLinesChanged += (s, e) => InvalidateVisual();
    }

    public override void Render(DrawingContext context)
    {
        var textView = _editor.TextArea.TextView;
        if (!textView.VisualLinesValid) return;

        foreach (var visualLine in textView.VisualLines)
        {
            var lineNumber = visualLine.FirstDocumentLine.LineNumber;
            var y = CalculateYPosition(visualLine);

            if (BreakpointLines.Contains(lineNumber))
                RenderBreakpoint(context, y);

            if (CurrentPCLine == lineNumber)
                RenderPCArrow(context, y);
        }
    }

    private double CalculateYPosition(VisualLine line)
    {
        var textView = _editor.TextArea.TextView;
        var visualTop = line.GetTextLineVisualYPosition(line.TextLines[0], VisualYPosition.LineTop);
        return visualTop - textView.VerticalOffset + GutterOffset;
    }
}

// Usage in EditorView:
var adorner = new EditorGutterAdorner(TextEditor);
AdornerLayer.SetAdorner(TextEditor, adorner);
```

**Binding Strategy:**
```csharp
// In EditorView.axaml.cs WhenActivated block:
this.WhenAnyValue(
    x => x.ViewModel!.Breakpoints,
    x => x.ViewModel!.AddressToLine)
    .Select(tuple => ConvertToLineNumbers(tuple.Item1, tuple.Item2))
    .Subscribe(lines => adorner.BreakpointLines = lines);

this.WhenAnyValue(
    x => x.ViewModel!.Registers.PC,
    x => x.ViewModel!.AddressToLine)
    .Select(tuple => GetLineForAddress(tuple.Item1, tuple.Item2))
    .Subscribe(line => adorner.CurrentPCLine = line);
```

**Click Handling:**
```csharp
protected override void OnPointerPressed(PointerPressedEventArgs e)
{
    var pos = e.GetPosition(_editor.TextArea.TextView);
    var line = GetLineNumberFromPosition(pos);
    if (line.HasValue)
        LineClicked?.Invoke(line.Value);
}
```

---

### Option 2: Canvas Overlay

**Approach:** Place a `Canvas` control over the editor's left margin area.

**Pros:**
- Simple to implement
- Familiar WPF/Avalonia pattern
- Easy to position elements

**Cons:**
- May interfere with editor hit testing
- Z-order management required
- Less integrated feel

**Implementation Sketch:**
```csharp
<Grid>
    <avaloniaEdit:TextEditor Name="TextEditor" ... />
    <Canvas Name="GutterCanvas"
            Width="50"
            HorizontalAlignment="Left"
            Background="Transparent"
            PointerPressed="OnGutterCanvasClicked">
        <!-- Breakpoints and PC indicators added dynamically -->
    </Canvas>
</Grid>
```

**Dynamic Element Management:**
```csharp
private void UpdateGutterCanvas()
{
    GutterCanvas.Children.Clear();

    var textView = TextEditor.TextArea.TextView;
    foreach (var visualLine in textView.VisualLines)
    {
        var lineNumber = visualLine.FirstDocumentLine.LineNumber;
        var y = CalculateYPosition(visualLine);

        if (BreakpointLines.Contains(lineNumber))
            GutterCanvas.Children.Add(CreateBreakpointEllipse(y));

        if (CurrentPCLine == lineNumber)
            GutterCanvas.Children.Add(CreatePCArrowPolygon(y));
    }
}
```

---

### Option 3: Avalonia.AvaloniaEdit 11.x (NOW IN USE)

**Approach:** Use the renamed `Avalonia.AvaloniaEdit` package for Avalonia 11 compatibility.

**Status:** ✅ **Implemented** - Project now uses `Avalonia.AvaloniaEdit 11.*`

**Pros:**
- Official package for Avalonia 11.x
- Version compatibility resolved
- Maintained by AvaloniaUI team
- Drop-in replacement for AvaloniaEdit 0.10

**Notes:**
- Requires theme inclusion in App.axaml: `<StyleInclude Source="avares://AvaloniaEdit/Themes/Fluent/AvaloniaEdit.xaml" />`
- Still need to investigate if margin extensibility APIs have improved in 11.x
- The adorner-based approach (Option 1) may still be preferred for custom gutters depending on API capabilities

**Additional Package:** `AvaloniaEdit.TextMate 11.4.1+` is available separately for TextMate grammar support, but not currently needed for basic syntax highlighting

---

### Option 4: Custom TextEditor Clone

**Approach:** Fork/copy AvaloniaEdit and update to Avalonia 11.

**Pros:**
- Full control over implementation
- Can fix all version issues
- Can optimize for our use case

**Cons:**
- Massive maintenance burden
- Lose upstream updates
- Not recommended for this project scale

**Verdict:** ❌ Not worth it for this project.

---

## Next Steps: Re-evaluate with Avalonia.AvaloniaEdit 11.x

Now that we're using `Avalonia.AvaloniaEdit 11.*`, we should investigate whether the margin extensibility APIs have improved:

**Before committing to the adorner approach, verify:**
1. Does `AbstractMargin` from `Avalonia.AvaloniaEdit 11.x` work with Avalonia 11 types?
2. Can we extend `AbstractMargin` to create custom gutters without type mismatches?
3. Are there new/improved APIs for margin rendering and interaction?

**Test with minimal implementation:**
```csharp
public class TestMargin : AbstractMargin
{
    protected override void OnRender(DrawingContext context)
    {
        // Simple test rendering
    }
}

// Try adding: TextEditor.TextArea.LeftMargins.Insert(0, new TestMargin());
```

**If margin extension now works:**
- Consider reverting to margin-based approach (cleaner, more integrated)
- Update this document with findings

**If margin extension still has issues:**
- Proceed with adorner-based approach as documented below

---

## Recommended Implementation Path

### Phase 1: Adorner-Based Breakpoint Markers

1. **Create `EditorGutterAdorner` class:**
   ```csharp
   public class EditorGutterAdorner : Control
   {
       // Styled properties for breakpoints and PC
       // Render method
       // Click handling
   }
   ```

2. **Wire up in `EditorView.axaml.cs`:**
   ```csharp
   var adorner = new EditorGutterAdorner(TextEditor);
   AdornerLayer.SetAdorner(TextEditor, adorner);
   ```

3. **Bind to ViewModel properties:**
   - `Breakpoints` → `BreakpointLines` (with address-to-line conversion)
   - `Registers.PC` → `CurrentPCLine` (with address-to-line lookup)

4. **Implement click-to-toggle:**
   - Detect click position
   - Map to line number
   - Call `ViewModel.AddBreakpointAsync()` or `RemoveBreakpointAsync()`

### Phase 2: Animation and Polish

1. **Fade-in animation for new breakpoints**
2. **Smooth scrolling to PC on step**
3. **Hover effects on gutter**
4. **Tooltip showing address on hover**

### Phase 3: Advanced Features

1. **Conditional breakpoints** (future)
2. **Breakpoint enable/disable toggle**
3. **Multiple breakpoint types** (hardware vs. software)

---

## Code Structure

### File Organization

```
avalonia-gui/ARMEmulator/
├── Controls/
│   └── EditorGutterAdorner.cs        # Adorner for breakpoints/PC
├── Views/
│   ├── EditorView.axaml              # TextEditor + adorner setup
│   └── EditorView.axaml.cs           # Wire up adorner to ViewModel
└── ViewModels/
    └── MainWindowViewModel.cs        # Breakpoint management methods
```

### Key Classes

**`EditorGutterAdorner`**
- Extends `Control` (not `AbstractMargin`)
- Renders breakpoints and PC arrow
- Handles pointer events for click-to-toggle
- Subscribes to TextView scroll/layout changes

**`EditorView`**
- Hosts `TextEditor`
- Creates and attaches `EditorGutterAdorner`
- Binds adorner properties to ViewModel
- Handles adorner click events → ViewModel methods

**`MainWindowViewModel`**
- `Breakpoints` property (address-based)
- `AddressToLine` / `LineToAddress` mappings
- `AddBreakpointAsync()` / `RemoveBreakpointAsync()`
- `Registers.PC` for current line highlight

---

## Implementation Challenges & Solutions

### Challenge 1: Coordinate Mapping

**Problem:** Need to map editor pixel positions to line numbers.

**Solution:**
```csharp
private int? GetLineNumberFromY(double y)
{
    var textView = _editor.TextArea.TextView;
    var visualLine = textView.GetVisualLineFromVisualTop(y + textView.VerticalOffset);
    return visualLine?.FirstDocumentLine.LineNumber;
}
```

### Challenge 2: Scroll Synchronization

**Problem:** Adorner must update when editor scrolls.

**Solution:**
```csharp
textView.ScrollOffsetChanged += (s, e) => InvalidateVisual();
textView.VisualLinesChanged += (s, e) => InvalidateVisual();
```

### Challenge 3: Address-to-Line Mapping

**Problem:** Breakpoints are address-based, gutter is line-based.

**Solution:**
```csharp
private ImmutableHashSet<int> ConvertAddressesToLines(
    ImmutableHashSet<uint> addresses,
    ImmutableDictionary<uint, int> addressToLine)
{
    return addresses
        .Where(addressToLine.ContainsKey)
        .Select(addr => addressToLine[addr])
        .ToImmutableHashSet();
}
```

### Challenge 4: Click Hit Testing

**Problem:** Determine if click is on gutter (not editor content).

**Solution:**
```csharp
protected override void OnPointerPressed(PointerPressedEventArgs e)
{
    var pos = e.GetPosition(_editor);
    if (pos.X < 0 || pos.X > GutterWidth) return;  // Outside gutter

    var lineNumber = GetLineNumberFromY(pos.Y);
    if (lineNumber.HasValue)
        OnLineClicked(lineNumber.Value);
}
```

---

## Testing Strategy

### Unit Tests

```csharp
[Fact]
public void EditorGutterAdorner_BreakpointLines_TriggersRender()
{
    var adorner = new EditorGutterAdorner(mockEditor);
    var renderCalled = false;
    adorner.PropertyChanged += (s, e) => {
        if (e.PropertyName == nameof(adorner.BreakpointLines))
            renderCalled = true;
    };

    adorner.BreakpointLines = [1, 2, 3];

    renderCalled.Should().BeTrue();
}
```

### Integration Tests

1. **Load program with breakpoints** → Verify gutter shows markers
2. **Click gutter line** → Verify breakpoint toggles
3. **Step execution** → Verify PC arrow moves
4. **Scroll editor** → Verify markers stay in sync

### Manual Testing

1. Open example program
2. Click various gutter lines → breakpoints appear/disappear
3. Run program → PC arrow tracks execution
4. Resize window → gutter renders correctly
5. Switch themes → colors adapt

---

## Performance Considerations

### Optimize Rendering

- **Only render visible lines** - Query `textView.VisualLines`
- **Cache geometry objects** - Reuse `Geometry` for repeated shapes
- **Throttle updates** - Debounce rapid scroll events

### Memory Management

- **Dispose subscriptions** - Use `DisposeWith(disposables)` for all subscriptions
- **Weak event handlers** - Prevent memory leaks from editor events

---

## Future Enhancements

1. **Conditional Breakpoints**
   - Right-click menu → "Edit Condition"
   - Expression evaluator integration

2. **Breakpoint Persistence**
   - Save/load breakpoints with program
   - Recent breakpoints list

3. **Visual Feedback**
   - Pulse animation on breakpoint hit
   - Glow effect on current line
   - Disabled breakpoint grayed out

4. **Keyboard Shortcuts**
   - F9: Toggle breakpoint on current line
   - Ctrl+Shift+F9: Delete all breakpoints

---

## References

- **AvaloniaEdit GitHub:** https://github.com/AvaloniaUI/AvaloniaEdit
- **Avalonia Adorner Docs:** https://docs.avaloniaui.net/docs/controls/adorner
- **Swift GUI Reference:** `swift-gui/ARMEmulator/Views/EditorView.swift`
- **Implementation Plan:** `docs/AVALONIA_IMPLEMENTATION_PLAN.md`

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-09 | Use adorner-based approach | Version conflicts ruled out `AbstractMargin` extension (historical) |
| 2026-02-09 | Defer gutter to Phase 4.3+ | Focus on core editor functionality first |
| 2026-02-09 | ~~Keep AvaloniaEdit 0.10.12~~ | ~~Latest stable, syntax highlighting works~~ **SUPERSEDED** |
| 2026-02-09 | **Upgrade to Avalonia.AvaloniaEdit 11.*** | **Resolves version mismatch, uses renamed package for Avalonia 11.x** |
| 2026-02-09 | Re-evaluate margin extensibility | Check if `Avalonia.AvaloniaEdit 11.x` has improved APIs before committing to adorner approach |

---

## Implementation Checklist

When implementing the adorner-based gutter:

- [ ] Create `EditorGutterAdorner.cs` in `Controls/`
- [ ] Extend `Control`, not `AbstractMargin`
- [ ] Implement `Render()` method with breakpoint/PC drawing
- [ ] Subscribe to `ScrollOffsetChanged` and `VisualLinesChanged`
- [ ] Implement `OnPointerPressed()` for click handling
- [ ] Add styled properties: `BreakpointLines`, `CurrentPCLine`
- [ ] Wire up in `EditorView.axaml.cs`
- [ ] Bind properties to ViewModel in `WhenActivated`
- [ ] Handle adorner clicks → ViewModel breakpoint methods
- [ ] Test with real programs and breakpoints
- [ ] Add animations (fade-in, pulse on hit)
- [ ] Document usage in `AVALONIA_IMPLEMENTATION_PLAN.md`

---

**Status:** Ready for implementation
**Priority:** Medium (Phase 4.3 - after register/memory views)
**Estimated Effort:** 4-6 hours
