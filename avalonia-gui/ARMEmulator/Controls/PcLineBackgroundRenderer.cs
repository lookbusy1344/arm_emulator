using Avalonia;
using Avalonia.Media;
using AvaloniaEdit.Rendering;

namespace ARMEmulator.Controls;

/// <summary>
/// Paints a full-width background behind the line the program counter is on.
/// </summary>
public sealed class PcLineBackgroundRenderer(TextView textView) : IBackgroundRenderer
{
	private int? line;
	private IBrush? brush;

	/// <summary>The line to paint, or null for none.</summary>
	public int? Line
	{
		get => line;
		set
		{
			if (line == value) {
				return;
			}

			line = value;
			textView.InvalidateLayer(Layer);
		}
	}

	/// <summary>The background colour. Null paints nothing.</summary>
	public IBrush? Brush
	{
		get => brush;
		set
		{
			brush = value;
			textView.InvalidateLayer(Layer);
		}
	}

	public KnownLayer Layer => KnownLayer.Background;

	public void Draw(TextView textView, DrawingContext drawingContext)
	{
		if (Line is not int pcLine || Brush is null || !textView.VisualLinesValid) {
			return;
		}

		var visualLine = textView.VisualLines.FirstOrDefault(candidate => candidate.FirstDocumentLine.LineNumber == pcLine);
		if (visualLine is null) {
			return;
		}

		var top = visualLine.VisualTop - textView.ScrollOffset.Y;
		drawingContext.FillRectangle(Brush, new Rect(0, top, textView.Bounds.Width, visualLine.Height));
	}
}
