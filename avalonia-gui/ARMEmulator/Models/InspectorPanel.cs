namespace ARMEmulator.Models;

/// <summary>
/// The panels shown in the inspector area, in display order. Members are stored by name in the settings file:
/// renaming one discards stored settings.
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
