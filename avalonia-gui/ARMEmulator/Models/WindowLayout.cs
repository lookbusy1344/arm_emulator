using ARMEmulator.Collections;


namespace ARMEmulator.Models;

/// <summary>
/// The panels shown in the inspector area, in display order.
/// </summary>
public enum InspectorPanel
{
	Registers,
	Memory,
	Stack,
	Disassembly,
	Evaluator,
	Watchpoints,
	Breakpoints
}

/// <summary>
/// Selector entry for an <see cref="InspectorPanel"/>: its label and the resource key of its icon.
/// </summary>
public sealed record InspectorPanelItem(InspectorPanel Panel, string Label, string IconKey)
{
	/// <summary>Every panel in display order.</summary>
	public static EquatableArray<InspectorPanelItem> All { get; } = [
		new(InspectorPanel.Registers, "Registers", "IconRegisters"),
		new(InspectorPanel.Memory, "Memory", "IconMemory"),
		new(InspectorPanel.Stack, "Stack", "IconStack"),
		new(InspectorPanel.Disassembly, "Disassembly", "IconDisassembly"),
		new(InspectorPanel.Evaluator, "Evaluator", "IconEvaluate"),
		new(InspectorPanel.Watchpoints, "Watchpoints", "IconWatch"),
		new(InspectorPanel.Breakpoints, "Breakpoints", "IconBreakpoint")
	];
}

/// <summary>
/// Window layout kept across restarts.
/// </summary>
public sealed record WindowLayout
{
	/// <summary>The inspector panel on show.</summary>
	public InspectorPanel SelectedPanel { get; init; } = InspectorPanel.Registers;
}
