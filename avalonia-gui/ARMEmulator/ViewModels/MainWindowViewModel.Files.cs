using System.Reactive.Disposables.Fluent;
using ARMEmulator.Services;
using ARMEmulator.Views;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

/// <summary>
/// File commands, the unsaved-changes flag and the window title.
/// </summary>
public partial class MainWindowViewModel
{
	private readonly IUnsavedChangesPrompt? unsavedChangesPrompt;
#pragma warning disable CA2213 // Disposed via DisposeWith(disposables)
	private ObservableAsPropertyHelper<string> windowTitleHelper = null!;
#pragma warning restore CA2213
	private string? filePath;
	private bool isDirty;

	/// <summary>Path of the open file, or null for a new or example document.</summary>
	public string? FilePath
	{
		get => filePath;
		private set => this.RaiseAndSetIfChanged(ref filePath, value);
	}

	/// <summary>True when the source has edits that are not saved to <see cref="FilePath"/>.</summary>
	public bool IsDirty
	{
		get => isDirty;
		private set => this.RaiseAndSetIfChanged(ref isDirty, value);
	}

	public string WindowTitle => windowTitleHelper.Value;

	private void InitializeWindowTitle() =>
		windowTitleHelper = this.WhenAnyValue(x => x.FilePath, x => x.IsDirty, ViewModels.WindowTitle.Format)
			.ToProperty(this, x => x.WindowTitle)
			.DisposeWith(disposables);

	/// <summary>
	/// Returns true when it is safe to replace or close the document: it has no unsaved edits,
	/// or the user chose to discard them or saved them.
	/// </summary>
	public async Task<bool> ConfirmDiscardAsync()
	{
		if (!IsDirty || unsavedChangesPrompt is null) {
			return true;
		}

		var choice = await unsavedChangesPrompt.AskAsync(ViewModels.WindowTitle.DocumentName(FilePath));
		return choice switch {
			UnsavedChangesChoice.Discard => true,
			UnsavedChangesChoice.Save => await SaveDocumentAsync(FilePath),
			_ => false
		};
	}

	private void ShowDocument(string content, string? path)
	{
		SourceCode = content;
		SetCurrentFile(path);
		IsDirty = false;
	}

	private void SetCurrentFile(string? path)
	{
		FilePath = path;
		fileService.CurrentFilePath = path;
	}

	/// <summary>Saves to <paramref name="path"/>, or asks for a file when it is null. Returns false when no file was written.</summary>
	private async Task<bool> SaveDocumentAsync(string? path)
	{
		if (parentWindow is null) {
			return false;
		}

		var savedPath = await fileService.SaveFileAsync(parentWindow, SourceCode, path);
		if (savedPath is null) {
			return false;
		}

		SetCurrentFile(savedPath);
		IsDirty = false;
		return true;
	}

	private async Task OpenFileAsync(CancellationToken ct)
	{
		if (parentWindow is null || !await ConfirmDiscardAsync()) {
			return;
		}

		var opened = await fileService.OpenFileAsync(parentWindow);
		if (opened is null) {
			return;
		}

		ShowDocument(opened.Content, opened.Path);
		_ = await AssembleAsync(ct);
	}

	private async Task SaveFileAsync(CancellationToken ct) => _ = await SaveDocumentAsync(FilePath);

	private async Task SaveAsAsync(CancellationToken ct) => _ = await SaveDocumentAsync(null);

	private async Task OpenExampleAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var vm = new ExamplesBrowserViewModel(api);
		try {
			var window = new ExamplesBrowserWindow(vm);
			await window.ShowDialog(parentWindow);

			if (window.SelectedExampleContent is not null && await ConfirmDiscardAsync()) {
				ShowDocument(window.SelectedExampleContent, null);
				_ = await AssembleAsync(ct);
			}
		}
		finally {
			vm.Dispose();
		}
	}

	private async Task OpenRecentFileAsync(string path, CancellationToken ct)
	{
		if (await ConfirmDiscardAsync()) {
			await OpenSourceFileAsync(path, ct);
		}
	}

	/// <summary>
	/// Opens a file in the editor, records it as recent and loads it when a session exists.
	/// Failures are reported through <see cref="ErrorMessage"/>.
	/// </summary>
	public async Task OpenSourceFileAsync(string path, CancellationToken ct = default)
	{
		try {
			ShowDocument(await File.ReadAllTextAsync(path, ct), path);
			fileService.AddRecentFile(path);
			if (SessionId is not null) {
				_ = await AssembleAsync(ct);
			}
		}
		catch (Exception ex) when (ex is FileNotFoundException or DirectoryNotFoundException) {
			ErrorMessage = $"File not found: {path}";
			fileService.RemoveRecentFile(path);
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to open {path}: {ex.Message}";
		}
	}
}
