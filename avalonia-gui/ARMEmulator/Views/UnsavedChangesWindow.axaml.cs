using ARMEmulator.Controls;
using ARMEmulator.Services;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator.Views;

/// <summary>
/// Asks whether to save, discard or keep unsaved edits.
/// </summary>
public partial class UnsavedChangesWindow : Window
{
	public UnsavedChangesWindow()
	{
		InitializeComponent();
		DialogButtonOrder.ForCurrentPlatform(ButtonRow);
	}

	public UnsavedChangesWindow(string documentName) : this()
	{
		MessageText.Text = $"Save changes to {documentName} before continuing?";
	}

	/// <summary>The button the user chose. Closing the window with the title bar is Cancel.</summary>
	public UnsavedChangesChoice Choice { get; private set; } = UnsavedChangesChoice.Cancel;

	private void Save_Click(object? sender, RoutedEventArgs e) => Choose(UnsavedChangesChoice.Save);

	private void Discard_Click(object? sender, RoutedEventArgs e) => Choose(UnsavedChangesChoice.Discard);

	private void Cancel_Click(object? sender, RoutedEventArgs e) => Choose(UnsavedChangesChoice.Cancel);

	private void Choose(UnsavedChangesChoice choice)
	{
		Choice = choice;
		Close();
	}
}
