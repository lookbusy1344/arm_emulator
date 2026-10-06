using ARMEmulator.Collections;
using ARMEmulator.Models;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class LoadProgramResponseTests
{
	[Fact]
	public void Equals_SameSymbolsInDifferentInsertionOrder_AreEqual()
	{
		var first = new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(
			new Dictionary<string, uint> { ["_start"] = 0x8000, ["loop"] = 0x8008 }));
		var second = new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(
			new Dictionary<string, uint> { ["loop"] = 0x8008, ["_start"] = 0x8000 }));

		first.Should().Be(second);
		first.GetHashCode().Should().Be(second.GetHashCode());
	}

	[Fact]
	public void Equals_DifferentSymbolAddress_AreNotEqual()
	{
		var first = new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(
			new Dictionary<string, uint> { ["_start"] = 0x8000 }));
		var second = new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(
			new Dictionary<string, uint> { ["_start"] = 0x8004 }));

		first.Should().NotBe(second);
	}

	[Fact]
	public void Default_SymbolsBehaveAsEmpty()
	{
		new LoadProgramResponse(default).Symbols.Should().BeEmpty();
	}
}
