using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia.Controls;

namespace ARMEmulator.Views;

/// <summary>
/// The macOS menu bar. Shortcuts are not set on native items: the window's key bindings handle them, and a second
/// registration would run a command twice.
/// </summary>
public static class NativeMenus
{
	/// <summary>The File and Debug menus of the main window.</summary>
	public static NativeMenu CreateWindowMenu(MainWindowViewModel viewModel) => [
		Submenu("File", [
			Item("Open…", viewModel, ShortcutId.Open),
			Item("Save", viewModel, ShortcutId.Save),
			Item("Save As…", viewModel, ShortcutId.SaveAs),
			new NativeMenuItemSeparator(),
			Item("Examples…", viewModel, ShortcutId.Examples),
			new NativeMenuItemSeparator(),
			new NativeMenuItem("Restart Backend") { Command = viewModel.RestartBackendCommand }
		]),
		Submenu("Debug", [
			Item("Load Program", viewModel, ShortcutId.Load),
			Item("Run", viewModel, ShortcutId.Run),
			Item("Pause", viewModel, ShortcutId.Pause),
			Item("Step", viewModel, ShortcutId.Step),
			Item("Step Over", viewModel, ShortcutId.StepOver),
			Item("Step Out", viewModel, ShortcutId.StepOut),
			Item("Reset", viewModel, ShortcutId.Reset),
			Item("Show PC", viewModel, ShortcutId.ShowPc)
		])
	];

	/// <summary>The items macOS shows in the application menu, above Quit.</summary>
	public static NativeMenu CreateApplicationMenu(MainWindowViewModel viewModel) => [
		new NativeMenuItem("About ARM Emulator") { Command = viewModel.ShowAboutCommand },
		Item("Preferences…", viewModel, ShortcutId.Preferences)
	];

	private static NativeMenuItem Item(string header, MainWindowViewModel viewModel, ShortcutId id) =>
		new(header) { Command = ShortcutBindings.CommandFor(viewModel, id) };

	private static NativeMenuItem Submenu(string header, IEnumerable<NativeMenuItemBase> items)
	{
		var menu = new NativeMenu();
		foreach (var item in items) {
			menu.Add(item);
		}

		return new NativeMenuItem(header) { Menu = menu };
	}
}
