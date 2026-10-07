using ARMEmulator.Models;
using Avalonia;
using Avalonia.Controls;

namespace ARMEmulator;

public partial class MainWindow
{
	private const int InspectorColumnIndex = 2;
	private const int ConsoleRowIndex = 2;
	private const double DefaultWidth = 1200;
	private const double DefaultHeight = 800;
	private const double DefaultInspectorWidth = 400;
	private const double DefaultConsoleHeight = 250;

	private static readonly WindowGeometry DefaultGeometry = new() {
		X = 0,
		Y = 0,
		Width = DefaultWidth,
		Height = DefaultHeight,
		InspectorWidth = DefaultInspectorWidth,
		ConsoleHeight = DefaultConsoleHeight,
		IsMaximized = false
	};

	private WindowGeometryTracker geometryTracker = new(DefaultGeometry);

	private WindowGeometry? storedGeometry;

	private void ApplyGeometry(WindowGeometry? geometry)
	{
		storedGeometry = geometry;
		if (geometry is null) {
			return;
		}

		geometryTracker = new WindowGeometryTracker(geometry);
		Width = geometry.Width;
		Height = geometry.Height;
		Position = new PixelPoint(geometry.X, geometry.Y);
		WindowStartupLocation = WindowStartupLocation.Manual;
		TopSection.ColumnDefinitions[InspectorColumnIndex].Width = new GridLength(geometry.InspectorWidth);
		ContentGrid.RowDefinitions[ConsoleRowIndex].Height = new GridLength(geometry.ConsoleHeight);
	}

	protected override void OnOpened(EventArgs e)
	{
		base.OnOpened(e);
		if (storedGeometry is not null) {
			RestoreOnScreen(storedGeometry);
		}

		ObserveGeometry();
	}

	protected override void OnResized(WindowResizedEventArgs e)
	{
		base.OnResized(e);
		ObserveGeometry();
	}

	protected override void OnPropertyChanged(AvaloniaPropertyChangedEventArgs change)
	{
		base.OnPropertyChanged(change);
		if (change.Property == WindowStateProperty) {
			ObserveGeometry();
		}
	}

	private void RestoreOnScreen(WindowGeometry geometry)
	{
		if (Screens.All.Count > 0 && !geometry.IsVisibleOn(WorkingAreas())) {
			WindowStartupLocation = WindowStartupLocation.CenterScreen;
			var primary = Screens.Primary;
			if (primary is not null) {
				Position = CenterOn(primary.WorkingArea);
			}
		}

		if (geometry.IsMaximized) {
			WindowState = WindowState.Maximized;
		}
	}

	private void ObserveGeometry()
	{
		var state = WindowState switch {
			WindowState.Maximized or WindowState.FullScreen => ShownState.Maximized,
			WindowState.Minimized => ShownState.Minimized,
			_ => ShownState.Normal
		};
		geometryTracker.Observe(state, Position.X, Position.Y, ClientSize.Width, ClientSize.Height);
	}

	private IEnumerable<ScreenArea> WorkingAreas() =>
		Screens.All.Select(screen => new ScreenArea(screen.WorkingArea.X, screen.WorkingArea.Y, screen.WorkingArea.Width, screen.WorkingArea.Height));

	private PixelPoint CenterOn(PixelRect area) =>
		new(area.X + ((area.Width - (int)Width) / 2), area.Y + ((area.Height - (int)Height) / 2));

	private WindowGeometry CaptureGeometry()
	{
		ObserveGeometry();
		return geometryTracker.Capture(
			TopSection.ColumnDefinitions[InspectorColumnIndex].ActualWidth,
			ContentGrid.RowDefinitions[ConsoleRowIndex].ActualHeight);
	}
}
