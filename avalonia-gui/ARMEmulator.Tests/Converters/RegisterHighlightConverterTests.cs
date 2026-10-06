using System.Globalization;
using ARMEmulator.Converters;
using ARMEmulator.Tests.Ui;
using Avalonia.Media;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Converters;

/// <summary>
/// Tests for RegisterHighlightConverter to ensure proper highlight logic.
/// </summary>
public class RegisterHighlightConverterTests
{
	private readonly RegisterHighlightConverter converter = RegisterHighlightConverter.Instance;
	private readonly CultureInfo culture = CultureInfo.InvariantCulture;

	[Fact]
	public Task Convert_RegisterInChangedSet_ReturnsHighlightBrush() =>
		UiTest.RunOnUiThread(() => {
			var registerName = "R0";
			var changedRegisters = ImmutableHashSet.Create("R0", "R1", "R2");

			var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

			result.Should().BeOfType<SolidColorBrush>();
			var brush = (SolidColorBrush)result!;
			brush.Color.A.Should().Be(128); // Semi-transparent green
			brush.Color.G.Should().Be(255);
		});

	[Fact]
	public void Convert_RegisterNotInChangedSet_ReturnsTransparentBrush()
	{
		var registerName = "R5";
		var changedRegisters = ImmutableHashSet.Create("R0", "R1", "R2");

		var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Transparent);
	}

	[Fact]
	public void Convert_EmptyChangedSet_ReturnsTransparentBrush()
	{
		var registerName = "R0";
		var changedRegisters = ImmutableHashSet<string>.Empty;

		var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Transparent);
	}

	[Fact]
	public Task Convert_CpsrInChangedSet_ReturnsHighlightBrush() =>
		UiTest.RunOnUiThread(() => {
			var registerName = "CPSR";
			var changedRegisters = ImmutableHashSet.Create("CPSR");

			var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

			result.Should().BeOfType<SolidColorBrush>();
		});

	[Fact]
	public Task Convert_SpecialRegisterInChangedSet_ReturnsHighlightBrush() =>
		UiTest.RunOnUiThread(() => {
			var registerName = "PC";
			var changedRegisters = ImmutableHashSet.Create("PC");

			var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

			result.Should().BeOfType<SolidColorBrush>();
		});

	[Fact]
	public void Convert_InvalidValueCount_ReturnsTransparentBrush()
	{
		var result = converter.Convert(["R0"], typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Transparent);
	}

	[Fact]
	public void Convert_InvalidRegisterNameType_ReturnsTransparentBrush()
	{
		var changedRegisters = ImmutableHashSet.Create("R0");

		var result = converter.Convert([42, changedRegisters], typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Transparent);
	}

	[Fact]
	public void Convert_InvalidChangedRegistersType_ReturnsTransparentBrush()
	{
		var result = converter.Convert(["R0", "not a set"], typeof(IBrush), null, culture);

		result.Should().Be(Brushes.Transparent);
	}
}
