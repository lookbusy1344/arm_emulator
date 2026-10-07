namespace ARMEmulator.Models;

/// <summary>
/// Position, size and splitter positions of the main window, in device-independent pixels except for the position.
/// </summary>
public sealed record WindowGeometry
{
	public const double MinWidth = 800;
	public const double MinHeight = 600;
	public const double MinInspectorWidth = 250;
	public const double MinConsoleHeight = 100;

	/// <summary>The part of the window, on each axis, that must lie on a screen for the stored position to be reused.</summary>
	private const int VisibleMargin = 100;

	public required int X { get; init; }

	public required int Y { get; init; }

	public required double Width { get; init; }

	public required double Height { get; init; }

	/// <summary>Width of the inspector column.</summary>
	public required double InspectorWidth { get; init; }

	/// <summary>Height of the console row.</summary>
	public required double ConsoleHeight { get; init; }

	/// <summary>Size and position describe the restored window; the window opens maximised.</summary>
	public required bool IsMaximized { get; init; }

	/// <summary>Raises every size to its minimum.</summary>
	public WindowGeometry Clamp() => this with {
		Width = Math.Max(Width, MinWidth),
		Height = Math.Max(Height, MinHeight),
		InspectorWidth = Math.Max(InspectorWidth, MinInspectorWidth),
		ConsoleHeight = Math.Max(ConsoleHeight, MinConsoleHeight)
	};

	/// <summary>True when the window overlaps any screen by at least the visible margin on both axes.</summary>
	public bool IsVisibleOn(IEnumerable<ScreenArea> screens) => screens.Any(screen =>
		Overlap(X, Width, screen.X, screen.Width) >= VisibleMargin && Overlap(Y, Height, screen.Y, screen.Height) >= VisibleMargin);

	private static double Overlap(double start, double length, double screenStart, double screenLength) =>
		Math.Min(start + length, screenStart + screenLength) - Math.Max(start, screenStart);
}
