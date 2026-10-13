using ARMEmulator.Models;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class AddressParserTests
{
	[Theory]
	[InlineData("0x9000", 0x9000u)]
	[InlineData("0X9000", 0x9000u)]
	[InlineData("0xabcd", 0xABCDu)]
	[InlineData("  0x9000  ", 0x9000u)]
	[InlineData("0xFFFFFFFF", uint.MaxValue)]
	[InlineData("0x0", 0u)]
	public void TryParse_ReadsHexWithPrefix(string text, uint expected)
	{
		_ = AddressParser.TryParse(text, out var address).Should().BeTrue();
		_ = address.Should().Be(expected);
	}

	[Theory]
	[InlineData("9000", 9000u)]
	[InlineData("36864", 0x9000u)]
	[InlineData("0", 0u)]
	[InlineData("  4096 ", 4096u)]
	[InlineData("4294967295", uint.MaxValue)]
	public void TryParse_ReadsBareNumbersAsDecimal(string text, uint expected)
	{
		_ = AddressParser.TryParse(text, out var address).Should().BeTrue();
		_ = address.Should().Be(expected);
	}

	[Theory]
	[InlineData(null)]
	[InlineData("")]
	[InlineData("   ")]
	[InlineData("xyz")]
	[InlineData("0x")]
	[InlineData("0xG1")]
	[InlineData("0x 10")]
	[InlineData("0x1FFFFFFFF")]
	[InlineData("4294967296")]
	[InlineData("1FFFFFFFF")]
	[InlineData("9000h")]
	[InlineData("-1")]
	[InlineData("+5")]
	[InlineData("1 2")]
	[InlineData("1,000")]
	public void TryParse_RejectsInvalidText(string? text)
	{
		_ = AddressParser.TryParse(text, out var address).Should().BeFalse();
		_ = address.Should().Be(0u);
	}
}
