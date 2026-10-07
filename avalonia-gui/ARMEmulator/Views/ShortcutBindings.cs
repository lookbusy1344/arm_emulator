using System.Windows.Input;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia.Input;

namespace ARMEmulator.Views;

/// <summary>
/// Connects the shortcut table to the main window's commands.
/// </summary>
public static class ShortcutBindings
{
	/// <summary>The command a shortcut triggers.</summary>
	public static ICommand CommandFor(MainWindowViewModel viewModel, ShortcutId id) => id switch {
		ShortcutId.Open => viewModel.OpenFileCommand,
		ShortcutId.Save => viewModel.SaveFileCommand,
		ShortcutId.SaveAs => viewModel.SaveAsCommand,
		ShortcutId.Examples => viewModel.OpenExampleCommand,
		ShortcutId.Preferences => viewModel.ShowPreferencesCommand,
		ShortcutId.Load => viewModel.LoadProgramCommand,
		ShortcutId.Run => viewModel.RunCommand,
		ShortcutId.Pause => viewModel.PauseCommand,
		ShortcutId.Step => viewModel.StepCommand,
		ShortcutId.StepOver => viewModel.StepOverCommand,
		ShortcutId.StepOut => viewModel.StepOutCommand,
		ShortcutId.Reset => viewModel.ResetCommand,
		ShortcutId.ShowPc => viewModel.ShowPcCommand,
		_ => throw new ArgumentOutOfRangeException(nameof(id), id, "No command for this shortcut.")
	};

	/// <summary>Key bindings for every entry of the table on the given platform.</summary>
	public static IEnumerable<KeyBinding> Create(MainWindowViewModel viewModel, bool isMacOS) =>
		Shortcuts.For(isMacOS).Select(binding => new KeyBinding { Gesture = binding.Gesture, Command = CommandFor(viewModel, binding.Id) });

	/// <summary>The gesture to show beside a menu item: the first one of the command.</summary>
	public static KeyGesture? FirstGesture(ShortcutId id, bool isMacOS) =>
		Shortcuts.For(isMacOS).FirstOrDefault(binding => binding.Id == id)?.Gesture;
}
