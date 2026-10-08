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
	/// <summary>The File and Debug menus of the main window. They offer the same commands as the in-window menu.</summary>
	public static NativeMenu CreateWindowMenu(MainWindowViewModel viewModel) => [
		Submenu("File", [
			Item("Open…", viewModel, ShortcutId.Open),
			Item("Save", viewModel, ShortcutId.Save),
			Item("Save As…", viewModel, ShortcutId.SaveAs),
			new NativeMenuItemSeparator(),
			RecentFilesItem(viewModel),
			new NativeMenuItemSeparator(),
			Item("Examples…", viewModel, ShortcutId.Examples),
			new NativeMenuItemSeparator(),
			Item("Preferences…", viewModel, ShortcutId.Preferences),
			new NativeMenuItem("Restart Backend") { Command = viewModel.RestartBackendCommand },
			new NativeMenuItem("About…") { Command = viewModel.ShowAboutCommand }
		]),
		Submenu("Debug", [
			Item("Assemble", viewModel, ShortcutId.Assemble),
			Item("Run", viewModel, ShortcutId.Run),
			Item("Pause", viewModel, ShortcutId.Pause),
			Item("Step", viewModel, ShortcutId.Step),
			Item("Step Over", viewModel, ShortcutId.StepOver),
			Item("Step Out", viewModel, ShortcutId.StepOut),
			Item("Reset", viewModel, ShortcutId.Reset),
			Item("Show PC", viewModel, ShortcutId.ShowPc),
			Item("Toggle Breakpoint", viewModel, ShortcutId.ToggleBreakpoint)
		])
	];

	/// <summary>The items macOS shows in the application menu, above Quit.</summary>
	public static NativeMenu CreateApplicationMenu(MainWindowViewModel viewModel) => [
		new NativeMenuItem("About ARM Emulator") { Command = viewModel.ShowAboutCommand },
		Item("Preferences…", viewModel, ShortcutId.Preferences)
	];

	/// <summary>Lists the recent files. Reopens the list from the file service each time the menu opens.</summary>
	private static NativeMenuItem RecentFilesItem(MainWindowViewModel viewModel)
	{
		var menu = new NativeMenu();
		void Populate()
		{
			menu.Items.Clear();
			foreach (var file in viewModel.RecentFiles) {
				menu.Items.Add(new NativeMenuItem(file.FileName) {
					Command = viewModel.OpenRecentFileCommand,
					CommandParameter = file.Path
				});
			}
		}

		menu.Opening += (_, _) => {
			viewModel.RefreshRecentFiles();
			Populate();
		};
		Populate();
		return new NativeMenuItem("Recent Files") { Menu = menu };
	}

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
