using ARMEmulator.Collections;
using ARMEmulator.Models;

namespace ARMEmulator.Views;

/// <summary>
/// Selector entry for an <see cref="InspectorPanel"/>: its label and the resource key of its icon.
/// </summary>
public sealed record InspectorPanelChoice(InspectorPanel Panel, string Label, string IconKey)
{
	/// <summary>Every panel in display order.</summary>
	public static EquatableArray<InspectorPanelChoice> All { get; } = [
		new(InspectorPanel.Registers, "Registers", "IconRegisters"),
		new(InspectorPanel.Memory, "Memory", "IconMemory"),
		new(InspectorPanel.Stack, "Stack", "IconStack"),
		new(InspectorPanel.Disassembly, "Disassembly", "IconDisassembly"),
		new(InspectorPanel.Evaluator, "Evaluator", "IconEvaluate"),
		new(InspectorPanel.Watchpoints, "Watchpoints", "IconWatch"),
		new(InspectorPanel.Breakpoints, "Breakpoints", "IconBreakpoint")
	];
}
