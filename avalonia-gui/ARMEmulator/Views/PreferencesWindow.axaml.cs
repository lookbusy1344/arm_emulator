using ARMEmulator.Models;
using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator.Views;

/// <summary>
/// Preferences dialog for configuring application settings.
/// </summary>
public partial class PreferencesWindow : Window
{
	private readonly PreferencesWindowViewModel viewModel;

	public PreferencesWindow()
	{
		InitializeComponent();
		viewModel = new PreferencesWindowViewModel(AppSettings.Default);
		DataContext = viewModel;
	}

	/// <summary>
	/// Initializes the window with custom settings.
	/// </summary>
	public PreferencesWindow(AppSettings settings) : this()
	{
		viewModel = new PreferencesWindowViewModel(settings);
		DataContext = viewModel;
	}

	/// <summary>
	/// Gets the updated settings if the user clicked OK.
	/// </summary>
	public AppSettings? UpdatedSettings { get; private set; }

	private void OkButton_Click(object? sender, RoutedEventArgs e)
	{
		UpdatedSettings = viewModel.BuildSettings();
		Close();
	}

	private void CancelButton_Click(object? sender, RoutedEventArgs e)
	{
		UpdatedSettings = null;
		Close();
	}
}
