using ARMEmulator.Models;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class RegisterTableTests
{
	private static ImmutableArray<string> ExpectedNames => ["R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8", "R9", "R10", "R11", "R12", "SP", "LR", "PC"];

	[Fact]
	public void Rows_ListSixteenRegistersInOrder()
	{
		var rows = RegisterTable.ToRows(RegisterState.Create(), new HashSet<string>());

		rows.Select(row => row.Name).Should().Equal(ExpectedNames);
	}

	[Fact]
	public void Rows_FormatHexAsEightDigitsAndDecimalUnsigned()
	{
		var rows = RegisterTable.ToRows(RegisterState.Create(r0: 5, r2: 0xDEADBEEF, pc: 0x8000), new HashSet<string>());

		rows[0].Should().Be(new RegisterRow("R0", "0x00000005", "5", false));
		rows[2].Should().Be(new RegisterRow("R2", "0xDEADBEEF", "3735928559", false));
		rows[15].Should().Be(new RegisterRow("PC", "0x00008000", "32768", false));
	}

	[Fact]
	public void Rows_MarkOnlyTheChangedRegisters()
	{
		var rows = RegisterTable.ToRows(RegisterState.Create(), new HashSet<string> { "R1", "LR" });

		rows.Where(row => row.IsChanged).Select(row => row.Name).Should().Equal("R1", "LR");
	}

	[Fact]
	public void Rows_IgnoreCpsrInTheChangedSet()
	{
		var rows = RegisterTable.ToRows(RegisterState.Create(), new HashSet<string> { "CPSR" });

		rows.Should().OnlyContain(row => !row.IsChanged);
	}

	[Fact]
	public void Flags_ListNZCVWithTheirState()
	{
		var flags = RegisterTable.ToFlags(RegisterState.Create(cpsr: new CPSRFlags(N: false, Z: true, C: true, V: false)));

		flags.Should().Equal(
			new FlagItem('N', "Negative", false),
			new FlagItem('Z', "Zero", true),
			new FlagItem('C', "Carry", true),
			new FlagItem('V', "Overflow", false));
	}
}
