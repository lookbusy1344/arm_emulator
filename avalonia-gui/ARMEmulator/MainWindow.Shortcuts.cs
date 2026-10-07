using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using ARMEmulator.Views;
using Avalonia.Controls;
using Avalonia.LogicalTree;

namespace ARMEmulator;

public partial class MainWindow
{
	private void ApplyPlatformShortcuts(MainWindowViewModel viewModel)
	{
		var isMacOS = OperatingSystem.IsMacOS();
		KeyBindings.AddRange(ShortcutBindings.Create(viewModel, isMacOS));

		foreach (var item in MenuBar.GetLogicalDescendants().OfType<MenuItem>()) {
			if (Enum.TryParse<ShortcutId>(item.Tag as string, out var id)) {
				item.InputGesture = ShortcutBindings.FirstGesture(id, isMacOS);
			}
		}

		// macOS shows a native menu bar; the in-window one would duplicate it.
		MenuBar.IsVisible = !isMacOS;
		NativeMenu.SetMenu(this, NativeMenus.CreateWindowMenu(viewModel));
	}
}
