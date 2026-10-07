using ARMEmulator.Models;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class WindowGeometryTests
{
	private static readonly WindowGeometry Valid = new() {
		X = 40,
		Y = 60,
		Width = 1300,
		Height = 900,
		InspectorWidth = 420,
		ConsoleHeight = 220,
		IsMaximized = false
	};

	[Fact]
	public void Clamp_KeepsValuesAtOrAboveTheMinimums()
	{
		Valid.Clamp().Should().Be(Valid);
	}

	[Fact]
	public void Clamp_RaisesSizesBelowTheMinimums()
	{
		var clamped = (Valid with { Width = 799, Height = 599, InspectorWidth = 249, ConsoleHeight = 99 }).Clamp();

		clamped.Width.Should().Be(800);
		clamped.Height.Should().Be(600);
		clamped.InspectorWidth.Should().Be(250);
		clamped.ConsoleHeight.Should().Be(100);
	}

	[Fact]
	public void Clamp_RaisesNegativeSizesToTheMinimums()
	{
		var clamped = (Valid with { Width = -1, Height = -1, InspectorWidth = -1, ConsoleHeight = -1 }).Clamp();

		clamped.Should().Be(Valid with { Width = 800, Height = 600, InspectorWidth = 250, ConsoleHeight = 100 });
	}

	[Fact]
	public void AppSettingsValidate_ClampsTheStoredGeometry()
	{
		var settings = AppSettings.Default with {
			Layout = new WindowLayout { Geometry = Valid with { Width = 10 } }
		};

		settings.Validate().Layout.Geometry!.Width.Should().Be(800);
	}

	[Fact]
	public void IsVisibleOn_WhenTheWindowOverlapsAScreenByAtLeastTheVisibleMargin()
	{
		var screens = new[] { new ScreenArea(0, 0, 1920, 1080) };

		// Exactly the 100 px margin inside the right edge.
		(Valid with { X = 1820 }).IsVisibleOn(screens).Should().BeTrue();
	}

	[Fact]
	public void IsVisibleOn_IsFalseWhenTheWindowLiesOffEveryScreen()
	{
		var screens = new[] { new ScreenArea(0, 0, 1920, 1080) };

		(Valid with { X = 1900 }).IsVisibleOn(screens).Should().BeFalse();
		(Valid with { X = -1290 }).IsVisibleOn(screens).Should().BeFalse();
		(Valid with { Y = 1070 }).IsVisibleOn(screens).Should().BeFalse();
	}

	[Fact]
	public void IsVisibleOn_AcceptsAnyOfSeveralScreens()
	{
		var screens = new[] { new ScreenArea(0, 0, 1920, 1080), new ScreenArea(1920, 0, 1920, 1080) };

		(Valid with { X = 2500 }).IsVisibleOn(screens).Should().BeTrue();
	}

	[Fact]
	public void IsVisibleOn_IsFalseWithNoScreens()
	{
		Valid.IsVisibleOn([]).Should().BeFalse();
	}
}
