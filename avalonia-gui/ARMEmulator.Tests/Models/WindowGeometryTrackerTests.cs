using ARMEmulator.Models;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class WindowGeometryTrackerTests
{
	private static readonly WindowGeometry Initial = new() {
		X = 10,
		Y = 20,
		Width = 1000,
		Height = 700,
		InspectorWidth = 300,
		ConsoleHeight = 150,
		IsMaximized = false
	};

	private const double Inspector = 410;
	private const double Console = 220;

	[Fact]
	public void Capture_WithoutObservations_KeepsTheInitialBoundsAndTakesTheCurrentSplitters()
	{
		var tracker = new WindowGeometryTracker(Initial);

		tracker.Capture(Inspector, Console).Should().Be(Initial with { InspectorWidth = Inspector, ConsoleHeight = Console });
	}

	[Fact]
	public void Capture_AfterANormalResize_UsesTheNewBounds()
	{
		var tracker = new WindowGeometryTracker(Initial);

		tracker.Observe(ShownState.Normal, 50, 60, 1200, 900);

		tracker.Capture(Inspector, Console).Should().Be(new WindowGeometry {
			X = 50,
			Y = 60,
			Width = 1200,
			Height = 900,
			InspectorWidth = Inspector,
			ConsoleHeight = Console,
			IsMaximized = false
		});
	}

	[Fact]
	public void Capture_WhileMaximized_KeepsTheLastNormalBoundsAndSetsTheFlag()
	{
		var tracker = new WindowGeometryTracker(Initial);
		tracker.Observe(ShownState.Normal, 50, 60, 1200, 900);

		tracker.Observe(ShownState.Maximized, 0, 0, 2560, 1400);

		var captured = tracker.Capture(Inspector, Console);
		(captured.X, captured.Y, captured.Width, captured.Height, captured.IsMaximized).Should().Be((50, 60, 1200, 900, true));
	}

	[Fact]
	public void Capture_AfterMaximizingAFreshWindow_FlagsMaximizedWithTheInitialBounds()
	{
		var tracker = new WindowGeometryTracker(Initial);

		tracker.Observe(ShownState.Maximized, 0, 0, 2560, 1400);

		var captured = tracker.Capture(Inspector, Console);
		(captured.Width, captured.Height, captured.IsMaximized).Should().Be((1000, 700, true));
	}

	[Fact]
	public void Capture_WhileMinimizedFromMaximized_StaysMaximized()
	{
		var tracker = new WindowGeometryTracker(Initial);
		tracker.Observe(ShownState.Maximized, 0, 0, 2560, 1400);

		tracker.Observe(ShownState.Minimized, -32000, -32000, 160, 28);

		var captured = tracker.Capture(Inspector, Console);
		(captured.Width, captured.IsMaximized).Should().Be((1000, true));
	}

	[Fact]
	public void Capture_WhileMinimizedFromNormal_IgnoresTheMinimizedBounds()
	{
		var tracker = new WindowGeometryTracker(Initial);
		tracker.Observe(ShownState.Normal, 50, 60, 1200, 900);

		tracker.Observe(ShownState.Minimized, -32000, -32000, 160, 28);

		tracker.Capture(Inspector, Console).Should().Be(new WindowGeometry {
			X = 50,
			Y = 60,
			Width = 1200,
			Height = 900,
			InspectorWidth = Inspector,
			ConsoleHeight = Console,
			IsMaximized = false
		});
	}

	[Fact]
	public void Capture_AfterRestoringFromMaximized_ClearsTheFlag()
	{
		var tracker = new WindowGeometryTracker(Initial);
		tracker.Observe(ShownState.Maximized, 0, 0, 2560, 1400);

		tracker.Observe(ShownState.Normal, 70, 80, 1100, 800);

		var captured = tracker.Capture(Inspector, Console);
		(captured.X, captured.Width, captured.IsMaximized).Should().Be((70, 1100, false));
	}

	[Fact]
	public void Capture_ClampsUndersizedBounds()
	{
		var tracker = new WindowGeometryTracker(Initial);
		tracker.Observe(ShownState.Normal, 0, 0, 300, 200);

		var captured = tracker.Capture(10, 10);

		(captured.Width, captured.Height, captured.InspectorWidth, captured.ConsoleHeight).Should().Be((800, 600, 250, 100));
	}
}
