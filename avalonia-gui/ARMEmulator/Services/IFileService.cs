namespace ARMEmulator.Services;

/// <summary>
/// Service for file operations and recent file tracking.
/// Handles platform-specific file dialogs and persists recent file history.
/// </summary>
public interface IFileService
{
	/// <summary>
	/// Opens a file picker dialog for assembly files (.s extension).
	/// </summary>
	/// <returns>Selected file path, or null if cancelled</returns>
	Task<string?> OpenFileAsync();

	/// <summary>
	/// Opens a save dialog for the current file or a new file.
	/// </summary>
	/// <param name="content">File content to save</param>
	/// <param name="currentPath">Current file path (null for new file)</param>
	/// <returns>Saved file path, or null if cancelled</returns>
	Task<string?> SaveFileAsync(string content, string? currentPath);

	/// <summary>
	/// List of recently opened files (most recent first).
	/// </summary>
	IReadOnlyList<RecentFile> RecentFiles { get; }

	/// <summary>
	/// Adds a file to the recent files list.
	/// </summary>
	void AddRecentFile(string path);

	/// <summary>
	/// Clears all recent files.
	/// </summary>
	void ClearRecentFiles();

	/// <summary>
	/// Current file path being edited (null if new/unsaved file).
	/// </summary>
	string? CurrentFilePath { get; set; }
}

/// <summary>
/// Information about a recently opened file.
/// </summary>
public sealed record RecentFile(string Path, DateTime LastOpened)
{
	/// <summary>Gets the file name without path.</summary>
	public string FileName => System.IO.Path.GetFileName(Path);
}
