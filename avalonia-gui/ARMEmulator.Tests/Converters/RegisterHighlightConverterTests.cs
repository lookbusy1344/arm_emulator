using System.Globalization;
using ARMEmulator.Converters;
using Avalonia.Media;
using FluentAssertions;
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
	public void Convert_RegisterInChangedSet_ReturnsHighlightBrush()
	{
		var registerName = "R0";
		var changedRegisters = ImmutableHashSet.Create("R0", "R1", "R2");

		var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

		result.Should().BeOfType<SolidColorBrush>();
		var brush = (SolidColorBrush)result!;
		brush.Color.A.Should().Be(128); // Semi-transparent green
		brush.Color.G.Should().Be(255);
	}

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
	public void Convert_CpsrInChangedSet_ReturnsHighlightBrush()
	{
		var registerName = "CPSR";
		var changedRegisters = ImmutableHashSet.Create("CPSR");

		var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

		result.Should().BeOfType<SolidColorBrush>();
	}

	[Fact]
	public void Convert_SpecialRegisterInChangedSet_ReturnsHighlightBrush()
	{
		var registerName = "PC";
		var changedRegisters = ImmutableHashSet.Create("PC");

		var result = converter.Convert([registerName, changedRegisters], typeof(IBrush), null, culture);

		result.Should().BeOfType<SolidColorBrush>();
	}

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
