namespace ARMEmulator.Models;

/// <summary>
/// Immutable snapshot of ARM register state including general-purpose registers and CPSR flags.
/// </summary>
public sealed record RegisterState
{
	/// <summary>All 16 general-purpose registers (R0-R15).</summary>
	public required ImmutableArray<uint> Registers { get; init; }

	/// <summary>Current Program Status Register flags.</summary>
	public required CPSRFlags CPSR { get; init; }

	// Custom equality for ImmutableArray
	public bool Equals(RegisterState? other)
	{
		if (other is null) {
			return false;
		}

		if (ReferenceEquals(this, other)) {
			return true;
		}

		return Registers.SequenceEqual(other.Registers) && CPSR.Equals(other.CPSR);
	}

	public override int GetHashCode()
	{
		var hash = new HashCode();
		foreach (var reg in Registers) {
			hash.Add(reg);
		}

		hash.Add(CPSR);
		return hash.ToHashCode();
	}

	// Named accessors for general-purpose registers
	public uint R0 => Registers[0];
	public uint R1 => Registers[1];
	public uint R2 => Registers[2];
	public uint R3 => Registers[3];
	public uint R4 => Registers[4];
	public uint R5 => Registers[5];
	public uint R6 => Registers[6];
	public uint R7 => Registers[7];
	public uint R8 => Registers[8];
	public uint R9 => Registers[9];
	public uint R10 => Registers[10];
	public uint R11 => Registers[11];
	public uint R12 => Registers[12];

	// Named accessors for special registers
	public uint SP => Registers[13];
	public uint LR => Registers[14];
	public uint PC => Registers[15];

	/// <summary>
	/// Get register value by name (case-insensitive).
	/// Supports: R0-R15, SP (R13), LR (R14), PC (R15).
	/// </summary>
	public uint this[string name] => name.ToUpperInvariant() switch {
		"R0" => R0,
		"R1" => R1,
		"R2" => R2,
		"R3" => R3,
		"R4" => R4,
		"R5" => R5,
		"R6" => R6,
		"R7" => R7,
		"R8" => R8,
		"R9" => R9,
		"R10" => R10,
		"R11" => R11,
		"R12" => R12,
		"SP" or "R13" => SP,
		"LR" or "R14" => LR,
		"PC" or "R15" => PC,
		_ => throw new ArgumentException($"Unknown register: {name}", nameof(name))
	};

	/// <summary>
	/// Factory method to create a RegisterState with named parameters.
	/// All registers default to 0 if not specified.
	/// </summary>
	public static RegisterState Create(
		uint r0 = 0, uint r1 = 0, uint r2 = 0, uint r3 = 0,
		uint r4 = 0, uint r5 = 0, uint r6 = 0, uint r7 = 0,
		uint r8 = 0, uint r9 = 0, uint r10 = 0, uint r11 = 0,
		uint r12 = 0, uint sp = 0, uint lr = 0, uint pc = 0,
		CPSRFlags? cpsr = null
	) => new() {
		Registers = [r0, r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, sp, lr, pc],
		CPSR = cpsr ?? default
	};

	/// <summary>
	/// Compare this register state with another and return the set of changed register names.
	/// Includes "CPSR" in the set if flags have changed.
	/// </summary>
	public ImmutableHashSet<string> Diff(RegisterState other)
	{
		var registerNames = new[] { "R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7",
									"R8", "R9", "R10", "R11", "R12", "SP", "LR", "PC" };

		var changedRegisters = registerNames
			.Where((name, i) => Registers[i] != other.Registers[i]);

		var cpsrChanged = CPSR != other.CPSR
			? new[] { "CPSR" }
			: [];

		return changedRegisters.Concat(cpsrChanged).ToImmutableHashSet();
	}
}

/// <summary>
/// Current Program Status Register flags (immutable value type).
/// </summary>
public readonly record struct CPSRFlags(bool N, bool Z, bool C, bool V)
{
	/// <summary>
	/// Format flags as a display string (e.g., "NZC-" for N=true, Z=true, C=true, V=false).
	/// </summary>
	public string DisplayString =>
		$"{(N ? 'N' : '-')}{(Z ? 'Z' : '-')}{(C ? 'C' : '-')}{(V ? 'V' : '-')}";
}
