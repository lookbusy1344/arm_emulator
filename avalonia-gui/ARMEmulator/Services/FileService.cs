namespace ARMEmulator.Services;

/// <summary>
/// File service implementation with recent file tracking.
/// File dialogs will be implemented when UI infrastructure is ready.
/// </summary>
public sealed class FileService : IFileService
{
	private const int MaxRecentFiles = 10;
	private readonly List<RecentFile> recentFiles = [];

	public IReadOnlyList<RecentFile> RecentFiles => recentFiles;

	public string? CurrentFilePath { get; set; }

	public Task<string?> OpenFileAsync()
	{
		// TODO: Implement with Avalonia file dialogs when UI is ready
		return Task.FromResult<string?>(null);
	}

	public async Task<string?> SaveFileAsync(string content, string? currentPath)
	{
		if (currentPath is not null) {
			// Save to existing file
			await File.WriteAllTextAsync(currentPath, content);
			return currentPath;
		}

		// TODO: Implement save dialog when UI is ready
		return null;
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
