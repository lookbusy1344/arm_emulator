using System.Globalization;
using ARMEmulator.Converters;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Converters;

public sealed class BinaryValueConverterTests
{
	private readonly BinaryValueConverter converter = BinaryValueConverter.Instance;
	private readonly CultureInfo culture = CultureInfo.InvariantCulture;

	[Fact]
	public void Convert_PadsToThirtyTwoDigits()
	{
		converter.Convert(5u, typeof(string), null, culture).Should().Be("00000000000000000000000000000101");
	}

	[Theory]
	[InlineData("101", 5u)]
	[InlineData("11111111111111111111111111111111", uint.MaxValue)]
	public void ConvertBack_ValidBinary_ReturnsValue(string text, uint expected)
	{
		converter.ConvertBack(text, typeof(uint), null, culture).Should().Be(expected);
	}

	[Theory]
	[InlineData("102")]
	[InlineData("")]
	[InlineData("111111111111111111111111111111111")]
	public void ConvertBack_InvalidBinary_ReturnsZero(string text)
	{
		converter.ConvertBack(text, typeof(uint), null, culture).Should().Be(0u);
	}
}
