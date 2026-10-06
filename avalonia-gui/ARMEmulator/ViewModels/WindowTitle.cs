namespace ARMEmulator.ViewModels;

/// <summary>
/// Formats the main window title from the open file and its unsaved state.
/// </summary>
public static class WindowTitle
{
	private const string ApplicationName = "ARM Emulator";
	private const string UntitledName = "Untitled";
	private const string DirtyMarker = "•";

	/// <summary>Display name of a document: the file name, or "Untitled" without a path.</summary>
	public static string DocumentName(string? path) =>
		path is null ? UntitledName : Path.GetFileName(path);

	/// <summary>"ARM Emulator", then " — name" for an open file or unsaved edits, then " •" when there are unsaved edits.</summary>
	public static string Format(string? path, bool isDirty)
	{
		if (path is null && !isDirty) {
			return ApplicationName;
		}

		var marker = isDirty ? $" {DirtyMarker}" : "";
		return $"{ApplicationName} — {DocumentName(path)}{marker}";
	}
}
