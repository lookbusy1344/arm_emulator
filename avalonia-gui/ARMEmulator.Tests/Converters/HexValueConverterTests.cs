using System.Globalization;
using ARMEmulator.Converters;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Converters;

/// <summary>
/// Tests for HexValueConverter to ensure proper uint to hex string conversion.
/// </summary>
public class HexValueConverterTests
{
	private readonly HexValueConverter converter = HexValueConverter.Instance;
	private readonly CultureInfo culture = CultureInfo.InvariantCulture;

	[Fact]
	public void Convert_ZeroValue_ReturnsZeroHexString()
	{
		var result = converter.Convert(0u, typeof(string), null, culture);

		result.Should().Be("0x00000000");
	}

	[Fact]
	public void Convert_SmallValue_ReturnsZeroPaddedHexString()
	{
		var result = converter.Convert(42u, typeof(string), null, culture);

		result.Should().Be("0x0000002A");
	}

	[Fact]
	public void Convert_MaxValue_ReturnsFullHexString()
	{
		var result = converter.Convert(uint.MaxValue, typeof(string), null, culture);

		result.Should().Be("0xFFFFFFFF");
	}

	[Fact]
	public void Convert_TypicalAddress_ReturnsCorrectHexString()
	{
		var result = converter.Convert(0x00008000u, typeof(string), null, culture);

		result.Should().Be("0x00008000");
	}

	[Fact]
	public void Convert_NonUIntValue_ReturnsDefaultZeroString()
	{
		var result = converter.Convert("not a uint", typeof(string), null, culture);

		result.Should().Be("0x00000000");
	}

	[Fact]
	public void Convert_NullValue_ReturnsDefaultZeroString()
	{
		var result = converter.Convert(null, typeof(string), null, culture);

		result.Should().Be("0x00000000");
	}

	[Fact]
	public void ConvertBack_ValidHexStringWithPrefix_ReturnsUInt()
	{
		var result = converter.ConvertBack("0x0000002A", typeof(uint), null, culture);

		result.Should().Be(42u);
	}

	[Fact]
	public void ConvertBack_ValidHexStringWithoutPrefix_ReturnsUInt()
	{
		var result = converter.ConvertBack("2A", typeof(uint), null, culture);

		result.Should().Be(42u);
	}

	[Fact]
	public void ConvertBack_InvalidString_ReturnsZero()
	{
		var result = converter.ConvertBack("not hex", typeof(uint), null, culture);

		result.Should().Be(0u);
	}

	[Fact]
	public void ConvertBack_NullValue_ReturnsZero()
	{
		var result = converter.ConvertBack(null, typeof(uint), null, culture);

		result.Should().Be(0u);
	}

	[Fact]
	public void ConvertBack_MaxValue_ReturnsCorrectUInt()
	{
		var result = converter.ConvertBack("0xFFFFFFFF", typeof(uint), null, culture);

		result.Should().Be(uint.MaxValue);
	}
}
