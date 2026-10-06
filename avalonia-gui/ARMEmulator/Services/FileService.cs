using System.Reactive;
using System.Reactive.Subjects;
using Avalonia.Controls;
using Avalonia.Platform.Storage;

namespace ARMEmulator.Services;

/// <summary>
/// File service implementation with platform-specific file dialogs.
/// Uses Avalonia's StorageProvider for cross-platform file pickers.
/// </summary>
public sealed class FileService : IFileService, IDisposable
{
	private const int DefaultRecentFilesLimit = 10;
	private readonly Subject<Unit> recentFilesChanged = new();
	private ImmutableList<RecentFile> recentFiles = [];
	private int recentFilesLimit = DefaultRecentFilesLimit;

	public IReadOnlyList<RecentFile> RecentFiles => recentFiles;

	public IObservable<Unit> RecentFilesChanged => recentFilesChanged;

	public int RecentFilesLimit
	{
		get => recentFilesLimit;
		set
		{
			recentFilesLimit = Math.Max(value, 1);
			Update(recentFiles);
		}
	}

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

	public void AddRecentFile(string path) =>
		Update(recentFiles
			.RemoveAll(f => IsSamePath(f.Path, path))
			.Insert(0, new RecentFile(path, DateTime.Now)));

	public void ClearRecentFiles() => Update([]);

	public void RemoveRecentFile(string path) =>
		Update(recentFiles.RemoveAll(f => IsSamePath(f.Path, path)));

	public void RemoveMissingRecentFiles() =>
		Update(recentFiles.RemoveAll(f => !File.Exists(f.Path)));

	public void LoadRecentFiles(IEnumerable<string> paths) =>
		recentFiles = Trim(paths
			.DistinctBy(path => path, StringComparer.OrdinalIgnoreCase)
			.Select(path => new RecentFile(path, DateTime.MinValue))
			.ToImmutableList());

	public void Dispose() => recentFilesChanged.Dispose();

	private static bool IsSamePath(string left, string right) =>
		left.Equals(right, StringComparison.OrdinalIgnoreCase);

	private ImmutableList<RecentFile> Trim(ImmutableList<RecentFile> files) =>
		files.Count > recentFilesLimit ? files.GetRange(0, recentFilesLimit) : files;

	/// <summary>Stores the new list, trimmed to the limit, and notifies when it differs from the current one.</summary>
	private void Update(ImmutableList<RecentFile> updated)
	{
		var trimmed = Trim(updated);
		if (trimmed.SequenceEqual(recentFiles)) {
			return;
		}

		recentFiles = trimmed;
		recentFilesChanged.OnNext(Unit.Default);
	}
}
