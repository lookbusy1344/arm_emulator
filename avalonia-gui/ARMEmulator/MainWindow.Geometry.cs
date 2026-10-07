using ARMEmulator.Models;
using ARMEmulator.ViewModels;
using Avalonia;
using Avalonia.Controls;

namespace ARMEmulator;

public partial class MainWindow
{
	private const int InspectorColumnIndex = 2;
	private const int ConsoleRowIndex = 2;

	/// <summary>The geometry applied at start; kept as the restored size while the window is maximised.</summary>
	private WindowGeometry? restoredGeometry;

	private void ApplyGeometry(WindowGeometry? geometry)
	{
		restoredGeometry = geometry;
		if (geometry is null) {
			return;
		}

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
		if (restoredGeometry is null) {
			return;
		}

		if (Screens.All.Count > 0 && !restoredGeometry.IsVisibleOn(WorkingAreas())) {
			WindowStartupLocation = WindowStartupLocation.CenterScreen;
			var primary = Screens.Primary;
			if (primary is not null) {
				Position = CenterOn(primary.WorkingArea);
			}
		}

		if (restoredGeometry.IsMaximized) {
			WindowState = WindowState.Maximized;
		}
	}

	private IEnumerable<ScreenArea> WorkingAreas() =>
		Screens.All.Select(screen => new ScreenArea(screen.WorkingArea.X, screen.WorkingArea.Y, screen.WorkingArea.Width, screen.WorkingArea.Height));

	private PixelPoint CenterOn(PixelRect area) =>
		new(area.X + ((area.Width - (int)Width) / 2), area.Y + ((area.Height - (int)Height) / 2));

	private WindowGeometry CaptureGeometry()
	{
		var inspectorWidth = TopSection.ColumnDefinitions[InspectorColumnIndex].ActualWidth;
		var consoleHeight = ContentGrid.RowDefinitions[ConsoleRowIndex].ActualHeight;

		return WindowState == WindowState.Normal || restoredGeometry is null
			? new WindowGeometry {
				X = Position.X,
				Y = Position.Y,
				Width = ClientSize.Width,
				Height = ClientSize.Height,
				InspectorWidth = inspectorWidth,
				ConsoleHeight = consoleHeight,
				IsMaximized = false
			}.Clamp()
			: restoredGeometry with { InspectorWidth = inspectorWidth, ConsoleHeight = consoleHeight, IsMaximized = true };
	}
}
