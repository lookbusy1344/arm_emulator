namespace ARMEmulator.Services;

/// <summary>
/// Asks the user what to do with unsaved edits before they would be replaced or lost.
/// </summary>
public interface IUnsavedChangesPrompt
{
	/// <summary>Shows the prompt for <paramref name="documentName"/>. Closing the prompt without a choice is <see cref="UnsavedChangesChoice.Cancel"/>.</summary>
	Task<UnsavedChangesChoice> AskAsync(string documentName);
}

/// <summary>The user's answer to <see cref="IUnsavedChangesPrompt"/>.</summary>
public enum UnsavedChangesChoice
{
	/// <summary>Keep the edits: stop the operation.</summary>
	Cancel,

	/// <summary>Save the edits, then continue.</summary>
	Save,

	/// <summary>Throw the edits away and continue.</summary>
	Discard
}
