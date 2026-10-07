using System.Collections.Immutable;
using Avalonia;
using Avalonia.Input;
using Avalonia.Media;
using AvaloniaEdit.Editing;
using AvaloniaEdit.Rendering;

namespace ARMEmulator.Controls;

/// <summary>
/// Custom gutter margin that displays breakpoint markers and PC indicator.
/// Successfully uses AbstractMargin with Avalonia.AvaloniaEdit 11.x.
/// </summary>
public class EditorGutterMargin : AbstractMargin
{
	private const double GutterWidth = 30;
	private const double GlyphSize = 14;
	private const double IconGridSize = 20;
	private const double SeparatorWidth = 1;

	// The two glyph slots, left to right: breakpoint, then program counter. Both show on a breakpoint line at the PC.
	private const double BreakpointSlotX = 2;
	private const double PcSlotX = 14;

	/// <summary>
	/// Set of line numbers that have breakpoints.
	/// </summary>
	public static readonly StyledProperty<ImmutableHashSet<int>> BreakpointLinesProperty =
		AvaloniaProperty.Register<EditorGutterMargin, ImmutableHashSet<int>>(
			nameof(BreakpointLines),
			ImmutableHashSet<int>.Empty);

	/// <summary>
	/// Line number of the current program counter (null if not running).
	/// </summary>
	public static readonly StyledProperty<int?> CurrentPCLineProperty =
		AvaloniaProperty.Register<EditorGutterMargin, int?>(
			nameof(CurrentPCLine),
			null);

	public ImmutableHashSet<int> BreakpointLines
	{
		get => GetValue(BreakpointLinesProperty);
		set => SetValue(BreakpointLinesProperty, value);
	}

	public int? CurrentPCLine
	{
		get => GetValue(CurrentPCLineProperty);
		set => SetValue(CurrentPCLineProperty, value);
	}

	/// <summary>
	/// Event raised when a gutter line is clicked.
	/// </summary>
	public event EventHandler<int>? LineClicked;

	static EditorGutterMargin()
	{
		// Trigger re-render when properties change
		AffectsRender<EditorGutterMargin>(BreakpointLinesProperty, CurrentPCLineProperty);
	}

	public EditorGutterMargin()
	{
		Width = GutterWidth;
	}

	protected override Size MeasureOverride(Size availableSize)
	{
		return new Size(GutterWidth, 0);
	}

	public override void Render(DrawingContext context)
	{
		var host = TextView;
		if (host is null) {
			return;
		}

		var background = host.FindBrush("GutterBackgroundBrush");
		if (background is not null) {
			context.FillRectangle(background, new Rect(0, 0, Bounds.Width, Bounds.Height));
		}

		var divider = host.FindBrush("DividerBrush");
		if (divider is not null) {
			context.FillRectangle(divider, new Rect(Bounds.Width - SeparatorWidth, 0, SeparatorWidth, Bounds.Height));
		}

		if (!host.VisualLinesValid) {
			return;
		}

		foreach (var visualLine in host.VisualLines) {
			var lineNumber = visualLine.FirstDocumentLine.LineNumber;
			var top = visualLine.GetTextLineVisualYPosition(visualLine.TextLines[0], VisualYPosition.LineTop) - host.ScrollOffset.Y;

			if (BreakpointLines.Contains(lineNumber)) {
				DrawGlyph(context, host, "IconBreakpoint", "BreakpointBrush", BreakpointSlotX, top, visualLine.Height);
			}

			if (CurrentPCLine == lineNumber) {
				DrawGlyph(context, host, "IconPcArrow", "PcMarkerBrush", PcSlotX, top, visualLine.Height);
			}
		}
	}

	/// <summary>Draws an icon from the theme resources at <paramref name="x"/>, centred vertically within a line of the given height.</summary>
	private static void DrawGlyph(DrawingContext context, TextView host, string iconKey, string brushKey, double x, double top, double lineHeight)
	{
		var brush = host.FindBrush(brushKey);
		var geometry = host.FindGeometry(iconKey);
		if (brush is null || geometry is null) {
			return;
		}

		var scale = GlyphSize / IconGridSize;
		var offset = new Vector(x, top + ((lineHeight - GlyphSize) / 2));
		using (context.PushTransform(Matrix.CreateScale(scale, scale) * Matrix.CreateTranslation(offset.X, offset.Y))) {
			context.DrawGeometry(brush, null, geometry);
		}
	}

	protected override void OnPointerPressed(PointerPressedEventArgs e)
	{
		base.OnPointerPressed(e);

		var pos = e.GetPosition(this);
		var lineNumber = GetLineNumberFromY(pos.Y);

		if (lineNumber.HasValue) {
			LineClicked?.Invoke(this, lineNumber.Value);
			e.Handled = true;
		}
	}

	private int? GetLineNumberFromY(double y)
	{
		var textView = TextView;
		if (textView?.VisualLinesValid != true) {
			return null;
		}

		foreach (var visualLine in textView.VisualLines) {
			var lineY = visualLine.GetTextLineVisualYPosition(visualLine.TextLines[0], VisualYPosition.LineTop) - textView.ScrollOffset.Y;
			var lineHeight = visualLine.Height;

			if (y >= lineY && y <= lineY + lineHeight) {
				return visualLine.FirstDocumentLine.LineNumber;
			}
		}

		return null;
	}
}
