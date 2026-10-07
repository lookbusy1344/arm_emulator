using Avalonia.Input;

namespace ARMEmulator.Services;

/// <summary>A gesture and the command it triggers.</summary>
public sealed record ShortcutBinding(ShortcutId Id, KeyGesture Gesture);

/// <summary>
/// The shortcut table. Gestures use Command on macOS where other platforms use Control.
/// </summary>
public static class Shortcuts
{
	/// <summary>The bindings for a platform. A command with several gestures appears once per gesture, preferred first.</summary>
	public static ImmutableArray<ShortcutBinding> For(bool isMacOS)
	{
		var platform = isMacOS ? KeyModifiers.Meta : KeyModifiers.Control;
		return [
			Bind(ShortcutId.Open, Key.O, platform),
			Bind(ShortcutId.Save, Key.S, platform),
			Bind(ShortcutId.SaveAs, Key.S, platform | KeyModifiers.Shift),
			Bind(ShortcutId.Examples, Key.E, platform | KeyModifiers.Shift),
			Bind(ShortcutId.Preferences, Key.OemComma, platform),
			Bind(ShortcutId.Load, Key.L, platform),
			Bind(ShortcutId.Run, Key.F5, KeyModifiers.None),
			Bind(ShortcutId.Run, Key.R, platform),
			Bind(ShortcutId.Pause, Key.OemPeriod, platform),
			Bind(ShortcutId.Step, Key.F11, KeyModifiers.None),
			Bind(ShortcutId.Step, Key.T, platform),
			Bind(ShortcutId.StepOver, Key.F10, KeyModifiers.None),
			Bind(ShortcutId.StepOver, Key.T, platform | KeyModifiers.Shift),
			Bind(ShortcutId.StepOut, Key.T, platform | KeyModifiers.Alt),
			Bind(ShortcutId.Reset, Key.R, platform | KeyModifiers.Shift),
			Bind(ShortcutId.ShowPc, Key.J, platform)
		];
	}

	private static ShortcutBinding Bind(ShortcutId id, Key key, KeyModifiers modifiers) => new(id, new KeyGesture(key, modifiers));
}
