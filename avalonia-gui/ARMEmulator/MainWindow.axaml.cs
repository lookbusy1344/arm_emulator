using ARMEmulator.ViewModels;
using Avalonia.Controls;

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
	}
}
