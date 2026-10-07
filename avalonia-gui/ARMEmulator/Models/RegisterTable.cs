using ARMEmulator.Collections;

namespace ARMEmulator.Models;

/// <summary>One line of the register table.</summary>
public sealed record RegisterRow(string Name, string Hex, string Base10, bool IsChanged);

/// <summary>One CPSR condition flag.</summary>
public sealed record FlagItem(char Letter, string Name, bool IsSet);

/// <summary>
/// Projects a <see cref="RegisterState"/> to what the register panel shows.
/// </summary>
public static class RegisterTable
{
	private static readonly ImmutableArray<string> RowNames = ["R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "R9", "R10", "R11", "R12", "SP", "LR", "PC"];

	/// <summary>The sixteen registers in order. <paramref name="changed"/> holds the names to highlight.</summary>
	public static EquatableArray<RegisterRow> Rows(RegisterState state, IReadOnlySet<string> changed) =>
		[.. RowNames.Select((name, index) => new RegisterRow(name, $"0x{state.Registers[index]:X8}", state.Registers[index].ToString(System.Globalization.CultureInfo.InvariantCulture), changed.Contains(name)))];

	/// <summary>The N, Z, C and V flags in that order.</summary>
	public static EquatableArray<FlagItem> Flags(RegisterState state) => [
		new('N', "Negative", state.CPSR.N),
		new('Z', "Zero", state.CPSR.Z),
		new('C', "Carry", state.CPSR.C),
		new('V', "Overflow", state.CPSR.V)
	];
}
