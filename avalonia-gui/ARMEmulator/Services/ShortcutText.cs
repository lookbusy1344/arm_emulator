using Avalonia.Input;

namespace ARMEmulator.Services;

/// <summary>
/// Formats keyboard shortcuts the way the platform writes them.
/// </summary>
public static class ShortcutText
{
	/// <summary>The hint for the current platform.</summary>
	public static string CurrentEmptyEditorHint { get; } = EmptyEditorHint(OperatingSystem.IsMacOS());

	/// <summary>A shortcut made of the platform modifier, optionally Shift, and <paramref name="key"/>: "⇧⌘E" on macOS, "Ctrl+Shift+E" elsewhere.</summary>
	public static string Format(string key, bool shift, bool isMacOS)
	{
		if (isMacOS) {
			return shift ? $"⇧⌘{key}" : $"⌘{key}";
		}

		return shift ? $"Ctrl+Shift+{key}" : $"Ctrl+{key}";
	}

	/// <summary>The text shown over an empty editor.</summary>
	public static string EmptyEditorHint(bool isMacOS) =>
		$"Open a file ({Format("O", shift: false, isMacOS)}) or choose an example ({Format("E", shift: true, isMacOS)})";

	/// <summary>A gesture written the way the platform does: "⇧⌘E" on macOS, "Ctrl+Shift+E" elsewhere.</summary>
	public static string Display(KeyGesture gesture, bool isMacOS)
	{
		var key = KeyText(gesture.Key);
		var modifiers = gesture.KeyModifiers;
		return isMacOS
			? $"{Glyph(modifiers, KeyModifiers.Control, "⌃")}{Glyph(modifiers, KeyModifiers.Alt, "⌥")}{Glyph(modifiers, KeyModifiers.Shift, "⇧")}{Glyph(modifiers, KeyModifiers.Meta, "⌘")}{key}"
			: $"{Glyph(modifiers, KeyModifiers.Control, "Ctrl+")}{Glyph(modifiers, KeyModifiers.Alt, "Alt+")}{Glyph(modifiers, KeyModifiers.Shift, "Shift+")}{Glyph(modifiers, KeyModifiers.Meta, "Win+")}{key}";
	}

	/// <summary>A label with the gestures of a command: "Run (F5, ⌘R)".</summary>
	public static string Hint(string label, ShortcutId id, bool isMacOS)
	{
		var gestures = Shortcuts.For(isMacOS).Where(binding => binding.Id == id).Select(binding => Display(binding.Gesture, isMacOS));
		return $"{label} ({string.Join(", ", gestures)})";
	}

	private static string Glyph(KeyModifiers modifiers, KeyModifiers flag, string text) => modifiers.HasFlag(flag) ? text : "";

	private static string KeyText(Key key) => key switch {
		Key.OemComma => ",",
		Key.OemPeriod => ".",
		_ => key.ToString()
	};
}
