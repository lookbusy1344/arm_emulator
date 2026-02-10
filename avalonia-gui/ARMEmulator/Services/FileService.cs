using Avalonia.Controls;
using Avalonia.Platform.Storage;

namespace ARMEmulator.Services;

/// <summary>
/// File service implementation with platform-specific file dialogs.
/// Uses Avalonia's StorageProvider for cross-platform file pickers.
/// </summary>
public sealed class FileService : IFileService
{
	private const int MaxRecentFiles = 10;
	private readonly List<RecentFile> recentFiles = [];

	public IReadOnlyList<RecentFile> RecentFiles => recentFiles;

	public string? CurrentFilePath { get; set; }

	public async Task<(string path, string content)?> OpenFileAsync(Window parent)
	{
		var storage = parent.StorageProvider;
		if (!storage.CanOpen) {
			return null;
		}

		var files = await storage.OpenFilePickerAsync(new FilePickerOpenOptions {
			Title = "Open Assembly File",
			AllowMultiple = false,
			FileTypeFilter = [
				new FilePickerFileType("Assembly Files") { Patterns = ["*.s", "*.asm"] },
				new FilePickerFileType("All Files") { Patterns = ["*.*"] }
			]
		});

		if (files.Count == 0) {
			return null;
		}

		var file = files[0];
		var path = file.Path.LocalPath;

		// Read file content
		await using var stream = await file.OpenReadAsync();
		using var reader = new StreamReader(stream);
		var content = await reader.ReadToEndAsync();

		AddRecentFile(path);
		CurrentFilePath = path;

		return (path, content);
	}

	public async Task<string?> SaveFileAsync(Window parent, string content, string? currentPath)
	{
		if (currentPath is not null) {
			// Save to existing file
			await File.WriteAllTextAsync(currentPath, content);
			return currentPath;
		}

		// Show save dialog for new file
		var storage = parent.StorageProvider;
		if (!storage.CanSave) {
			return null;
		}

		var file = await storage.SaveFilePickerAsync(new FilePickerSaveOptions {
			Title = "Save Assembly File",
			SuggestedFileName = "program.s",
			DefaultExtension = "s",
			FileTypeChoices = [
				new FilePickerFileType("Assembly Files") { Patterns = ["*.s"] },
				new FilePickerFileType("All Files") { Patterns = ["*.*"] }
			]
		});

		if (file is null) {
			return null;
		}

		var path = file.Path.LocalPath;

		// Write file content
		await using var stream = await file.OpenWriteAsync();
		await using var writer = new StreamWriter(stream);
		await writer.WriteAsync(content);

		AddRecentFile(path);
		CurrentFilePath = path;

		return path;
	}

	public void AddRecentFile(string path)
	{
		// Remove if already exists (to move to top)
		_ = recentFiles.RemoveAll(f => f.Path.Equals(path, StringComparison.OrdinalIgnoreCase));

		// Add to front
		recentFiles.Insert(0, new RecentFile(path, DateTime.Now));

		// Trim to max
		if (recentFiles.Count > MaxRecentFiles) {
			recentFiles.RemoveRange(MaxRecentFiles, recentFiles.Count - MaxRecentFiles);
		}
	}

	public void ClearRecentFiles()
	{
		recentFiles.Clear();
	}
}
