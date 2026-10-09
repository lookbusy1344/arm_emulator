using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia.Controls;

namespace ARMEmulator.Views;

/// <summary>
/// The items macOS shows in the application menu, above Quit. Avalonia.Native exports this menu once, after
/// <see cref="App.Initialize"/> and before the view model exists, so <see cref="Bind"/> sets the commands later.
/// Preferences has its own section: Avalonia appends the separator below it, ahead of Services.
/// </summary>
public sealed class ApplicationMenu
{
	private readonly NativeMenuItem about = new("About ARM Emulator");
	private readonly NativeMenuItem preferences = new("Preferences…");

	public ApplicationMenu() => Menu = [about, new NativeMenuItemSeparator(), preferences];

	public NativeMenu Menu { get; }

	public void Bind(MainWindowViewModel viewModel)
	{
		about.Command = viewModel.ShowAboutCommand;
		preferences.Command = ShortcutBindings.CommandFor(viewModel, ShortcutId.Preferences);
	}
}
