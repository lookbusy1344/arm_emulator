using ARMEmulator.Models;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Threading;
using AwesomeAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Ui;

public sealed class WindowGeometryUiTests
{
	private static readonly WindowGeometry Stored = new() {
		X = 30,
		Y = 40,
		Width = 1000,
		Height = 700,
		InspectorWidth = 380,
		ConsoleHeight = 210,
		IsMaximized = false
	};

	private static readonly AppSettings WithGeometry = AppSettings.Default with { Layout = new WindowLayout { Geometry = Stored } };

	[Fact]
	public Task Window_StartsAtTheStoredSizeAndSplitterPositions() =>
		UiTest.RunAsync(WithGeometry, ui => {
			ui.Window.Width.Should().Be(1000);
			ui.Window.Height.Should().Be(700);
			ui.Window.Position.Should().Be(new PixelPoint(30, 40));
			ui.Find<Grid>("TopSection").ColumnDefinitions[2].Width.Value.Should().Be(380);
			ui.Find<Grid>("ContentGrid").RowDefinitions[2].Height.Value.Should().Be(210);
			return Task.CompletedTask;
		});

	[Fact]
	public Task Closing_SavesTheCurrentGeometry() =>
		UiTest.RunAsync(WithGeometry, ui => {
			ui.Store.ClearReceivedCalls();
			ui.Window.Width = 1100;
			ui.Window.Height = 750;
			Dispatcher.UIThread.RunJobs();
			ui.Window.UpdateLayout();

			ui.Window.Close();

			ui.Store.Received().Save(Arg.Is<AppSettings>(saved =>
				saved.Layout.Geometry!.Width == 1100 && saved.Layout.Geometry.Height == 750 && !saved.Layout.Geometry.IsMaximized));
			return Task.CompletedTask;
		});

	[Fact]
	public Task Window_WithNoStoredGeometry_KeepsTheDefaultSize() =>
		UiTest.Run(ui => {
			ui.Window.Width.Should().Be(1200);
			ui.Window.Height.Should().Be(800);
		});
}
