using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator.Views;

/// <summary>
/// Examples browser window for selecting and loading example programs.
/// </summary>
public partial class ExamplesBrowserWindow : Window
{
	private readonly ExamplesBrowserViewModel viewModel;

	public ExamplesBrowserWindow()
	{
		InitializeComponent();
		viewModel = null!; // Set via constructor parameter
	}

	/// <summary>
	/// Initializes the window with a ViewModel and loads examples.
	/// </summary>
	public ExamplesBrowserWindow(ExamplesBrowserViewModel viewModel) : this()
	{
		this.viewModel = viewModel;
		DataContext = viewModel;

		// Load examples asynchronously after window is shown
		// Fire-and-forget is acceptable - exceptions are handled in LoadExamplesAsync
#pragma warning disable VSTHRD101
		Opened += async (_, _) => await viewModel.LoadExamplesAsync();
#pragma warning restore VSTHRD101
	}

	/// <summary>
	/// Gets the content of the selected example if Load was clicked.
	/// </summary>
	public string? SelectedExampleContent { get; private set; }

	private void LoadButton_Click(object? sender, RoutedEventArgs e)
	{
		SelectedExampleContent = viewModel.PreviewContent;
		Close();
	}

	private void CancelButton_Click(object? sender, RoutedEventArgs e)
	{
		SelectedExampleContent = null;
		Close();
	}
}
