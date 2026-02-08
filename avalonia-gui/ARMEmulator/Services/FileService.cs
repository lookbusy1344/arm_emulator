namespace ARMEmulator.Services;

/// <summary>
/// File service implementation with recent file tracking.
/// File dialogs will be implemented when UI infrastructure is ready.
/// </summary>
public sealed class FileService : IFileService
{
	private const int MaxRecentFiles = 10;
	private readonly List<RecentFile> _recentFiles = [];

	public IReadOnlyList<RecentFile> RecentFiles => _recentFiles;

	public string? CurrentFilePath { get; set; }

	public Task<string?> OpenFileAsync()
	{
		// TODO: Implement with Avalonia file dialogs when UI is ready
		return Task.FromResult<string?>(null);
	}

	public async Task<string?> SaveFileAsync(string content, string? currentPath)
	{
		if (currentPath is not null)
		{
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
		_ = _recentFiles.RemoveAll(f => f.Path.Equals(path, StringComparison.OrdinalIgnoreCase));

		// Add to front
		_recentFiles.Insert(0, new RecentFile(path, DateTime.Now));

		// Trim to max
		if (_recentFiles.Count > MaxRecentFiles)
		{
			_recentFiles.RemoveRange(MaxRecentFiles, _recentFiles.Count - MaxRecentFiles);
		}
	}

	public void ClearRecentFiles()
	{
		_recentFiles.Clear();
	}
}
