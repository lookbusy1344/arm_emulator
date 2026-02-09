using System.Globalization;
using ARMEmulator.Converters;
using Avalonia.Media;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Converters;

/// <summary>
/// Tests for BoolToColorConverter used for CPSR flag visualization.
/// </summary>
public class BoolToColorConverterTests
{
	private readonly BoolToColorConverter converter = BoolToColorConverter.Instance;
	private readonly CultureInfo culture = CultureInfo.InvariantCulture;

	[Fact]
	public void Convert_TrueValue_ReturnsGreenBrush()
	{
		var result = converter.Convert(true, typeof(IBrush), null, culture);

		result.Should().BeOfType<SolidColorBrush>();
		var brush = (SolidColorBrush)result!;
		brush.Color.G.Should().Be(200); // Green component
		brush.Color.R.Should().Be(0);
		brush.Color.B.Should().Be(0);
	}

	[Fact]
	public void Convert_FalseValue_ReturnsGrayBrush()
	{
		var result = converter.Convert(false, typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Gray);
	}

	[Fact]
	public void Convert_NonBoolValue_ReturnsGrayBrush()
	{
		var result = converter.Convert("not a bool", typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Gray);
	}

	[Fact]
	public void Convert_NullValue_ReturnsGrayBrush()
	{
		var result = converter.Convert(null, typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Gray);
	}

	[Fact]
	public void ConvertBack_ThrowsNotSupportedException()
	{
		var action = () => converter.ConvertBack(Brushes.Green, typeof(bool), null, culture);

		action.Should().Throw<NotSupportedException>();
	}
}
