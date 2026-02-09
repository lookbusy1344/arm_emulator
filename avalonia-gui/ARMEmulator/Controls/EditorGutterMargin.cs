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
	private const double MarkerSize = 12;
	private const double MarkerMargin = 9; // (30 - 12) / 2

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
		// Draw gutter background
		context.FillRectangle(
			new SolidColorBrush(Color.FromRgb(245, 245, 245)),
			new Rect(0, 0, Bounds.Width, Bounds.Height));

		var textView = TextView;
		if (textView?.VisualLinesValid != true) {
			return;
		}

		// Render markers for each visible line
		foreach (var visualLine in textView.VisualLines) {
			var lineNumber = visualLine.FirstDocumentLine.LineNumber;
			var y = visualLine.GetTextLineVisualYPosition(visualLine.TextLines[0], VisualYPosition.LineTop) - textView.ScrollOffset.Y;

			// Draw breakpoint marker (red circle)
			if (BreakpointLines.Contains(lineNumber)) {
				DrawBreakpointMarker(context, y);
			}

			// Draw PC indicator (blue arrow)
			if (CurrentPCLine == lineNumber) {
				DrawPCIndicator(context, y);
			}
		}
	}

	private static void DrawBreakpointMarker(DrawingContext context, double y)
	{
		var center = new Point(GutterWidth / 2, y + MarkerSize / 2 + 2);
		var brush = new SolidColorBrush(Color.FromRgb(220, 50, 50)); // Red

		context.DrawEllipse(
			brush,
			new Pen(new SolidColorBrush(Color.FromRgb(180, 40, 40)), 1),
			center,
			MarkerSize / 2,
			MarkerSize / 2);
	}

	private static void DrawPCIndicator(DrawingContext context, double y)
	{
		var arrowY = y + 8;
		var arrowPoints = new[]
		{
			new Point(4, arrowY),
			new Point(16, arrowY + 4),
			new Point(4, arrowY + 8)
		};

		var geometry = new PolylineGeometry(arrowPoints, true);
		var brush = new SolidColorBrush(Color.FromRgb(50, 120, 220)); // Blue

		context.DrawGeometry(
			brush,
			new Pen(new SolidColorBrush(Color.FromRgb(30, 90, 180)), 1),
			geometry);
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
