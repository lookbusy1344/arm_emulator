using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator.Views;

/// <summary>
/// About dialog displaying application and backend version information.
/// </summary>
public partial class AboutWindow : Window
{
	public AboutWindow()
	{
		InitializeComponent();
	}

	/// <summary>
	/// Initializes the window with a ViewModel and loads backend version.
	/// </summary>
	public AboutWindow(AboutWindowViewModel viewModel) : this()
	{
		DataContext = viewModel;

		// Load backend version asynchronously after window is shown
		// Fire-and-forget is acceptable here - exceptions are handled in LoadBackendVersionAsync
#pragma warning disable VSTHRD101
		Opened += async (_, _) => await viewModel.LoadBackendVersionAsync();
#pragma warning restore VSTHRD101
	}

	private void CloseButton_Click(object? sender, RoutedEventArgs e)
	{
		Close();
	}
}
