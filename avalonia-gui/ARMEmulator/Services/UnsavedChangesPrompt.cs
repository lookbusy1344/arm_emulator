using ARMEmulator.Views;
using Avalonia.Controls;

namespace ARMEmulator.Services;

/// <summary>
/// Shows <see cref="UnsavedChangesWindow"/> as a dialog over the main window.
/// </summary>
public sealed class UnsavedChangesPrompt(Func<Window?> ownerProvider) : IUnsavedChangesPrompt
{
	public async Task<UnsavedChangesChoice> AskAsync(string documentName)
	{
		var owner = ownerProvider();
		if (owner is null) {
			return UnsavedChangesChoice.Cancel;
		}

		var window = new UnsavedChangesWindow(documentName);
		await window.ShowDialog(owner);
		return window.Choice;
	}
}
