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
}
