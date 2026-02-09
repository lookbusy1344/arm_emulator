using System.Globalization;
using ARMEmulator.Converters;
using Avalonia.Media;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Converters;

/// <summary>
/// Tests for BoolToFontWeightConverter used for CPSR flag emphasis.
/// </summary>
public class BoolToFontWeightConverterTests
{
	private readonly BoolToFontWeightConverter converter = BoolToFontWeightConverter.Instance;
	private readonly CultureInfo culture = CultureInfo.InvariantCulture;

	[Fact]
	public void Convert_TrueValue_ReturnsBold()
	{
		var result = converter.Convert(true, typeof(FontWeight), null, culture);

		result.Should().Be(FontWeight.Bold);
	}

	[Fact]
	public void Convert_FalseValue_ReturnsNormal()
	{
		var result = converter.Convert(false, typeof(FontWeight), null, culture);

		result.Should().Be(FontWeight.Normal);
	}

	[Fact]
	public void Convert_NonBoolValue_ReturnsNormal()
	{
		var result = converter.Convert("not a bool", typeof(FontWeight), null, culture);

		result.Should().Be(FontWeight.Normal);
	}

	[Fact]
	public void Convert_NullValue_ReturnsNormal()
	{
		var result = converter.Convert(null, typeof(FontWeight), null, culture);

		result.Should().Be(FontWeight.Normal);
	}

	[Fact]
	public void ConvertBack_ThrowsNotSupportedException()
	{
		var action = () => converter.ConvertBack(FontWeight.Bold, typeof(bool), null, culture);

		action.Should().Throw<NotSupportedException>();
	}
}
