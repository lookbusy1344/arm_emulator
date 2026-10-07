namespace ARMEmulator.Models;

/// <summary>
/// Follows a window's bounds and state so that closing saves the restored bounds and whether it was maximised.
/// </summary>
public sealed class WindowGeometryTracker(WindowGeometry initial)
{
	private WindowGeometry normal = initial with { IsMaximized = false };
	private bool isMaximized = initial.IsMaximized;

	/// <summary>Records the window's current state. Bounds count only while the window is in its normal state.</summary>
	public void Observe(ShownState state, int x, int y, double width, double height)
	{
		switch (state) {
			case ShownState.Normal:
				normal = normal with { X = x, Y = y, Width = width, Height = height };
				isMaximized = false;
				break;
			case ShownState.Maximized:
				isMaximized = true;
				break;
			case ShownState.Minimized:
				break;
			default:
				throw new ArgumentOutOfRangeException(nameof(state), state, "Unknown window state.");
		}
	}

	/// <summary>The geometry to save, with the splitter sizes read at the moment of closing.</summary>
	public WindowGeometry Capture(double inspectorWidth, double consoleHeight) =>
		(normal with { InspectorWidth = inspectorWidth, ConsoleHeight = consoleHeight, IsMaximized = isMaximized }).Clamp();
}
