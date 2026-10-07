using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator;

public partial class MainWindow : Window
{
	public MainWindow()
	{
		InitializeComponent();

		// Set DataContext for design-time preview
		// At runtime, this will be replaced by DI-injected ViewModel
		if (Design.IsDesignMode) {
			// For design-time, use a mock ViewModel
			DataContext = null;
		}
	}

	/// <summary>
	/// Constructor for DI with ViewModel injection.
	/// </summary>
	public MainWindow(MainWindowViewModel viewModel) : this()
	{
		DataContext = viewModel;
		viewModel.SetParentWindow(this);
		ApplyGeometry(viewModel.Settings.Layout.Geometry);
		ApplyPlatformShortcuts(viewModel);
	}

	private bool closeConfirmed;

	protected override void OnClosing(WindowClosingEventArgs e)
	{
		base.OnClosing(e);
		if (DataContext is not MainWindowViewModel viewModel) {
			return;
		}

		if (!closeConfirmed && viewModel.IsDirty) {
			e.Cancel = true;
			_ = CloseIfDiscardConfirmedAsync(viewModel);
			return;
		}

		// The platform window still exists here, so its position and size are readable.
		viewModel.SaveWindowGeometry(CaptureGeometry());
	}

	private async Task CloseIfDiscardConfirmedAsync(MainWindowViewModel viewModel)
	{
		if (await viewModel.ConfirmDiscardAsync()) {
			closeConfirmed = true;
			Close();
		}
	}

	private void RecentFilesMenu_SubmenuOpened(object? sender, RoutedEventArgs e) =>
		(DataContext as MainWindowViewModel)?.RefreshRecentFiles();
}
