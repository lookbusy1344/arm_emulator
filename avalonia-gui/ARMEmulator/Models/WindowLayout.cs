namespace ARMEmulator.Models;

/// <summary>
/// Window layout kept across restarts.
/// </summary>
public sealed record WindowLayout
{
	/// <summary>The inspector panel on show.</summary>
	public InspectorPanel SelectedPanel { get; init; } = InspectorPanel.Registers;

	/// <summary>Where the window and its splitters were left. Null until the window has been closed once.</summary>
	public WindowGeometry? Geometry { get; init; }
}
