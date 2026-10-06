using System.Reactive;
using Avalonia.Controls;

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
	/// <param name="parent">Parent window for the dialog</param>
	/// <returns>Selected file path and content, or null if cancelled</returns>
	Task<(string path, string content)?> OpenFileAsync(Window parent);

	/// <summary>
	/// Opens a save dialog for the current file or a new file.
	/// </summary>
	/// <param name="parent">Parent window for the dialog</param>
	/// <param name="content">File content to save</param>
	/// <param name="currentPath">Current file path (null for new file)</param>
	/// <returns>Saved file path, or null if cancelled</returns>
	Task<string?> SaveFileAsync(Window parent, string content, string? currentPath);

	/// <summary>
	/// Snapshot of the recently opened files (most recent first). Each change produces a new list.
	/// </summary>
	IReadOnlyList<RecentFile> RecentFiles { get; }

	/// <summary>
	/// Maximum number of recent files kept. Lowering it drops the oldest entries.
	/// </summary>
	int RecentFilesLimit { get; set; }

	/// <summary>
	/// Emits after the recent files list changes. Does not emit for <see cref="LoadRecentFiles"/>.
	/// </summary>
	IObservable<Unit> RecentFilesChanged { get; }

	/// <summary>
	/// Replaces the list with stored paths (most recent first) without notifying.
	/// </summary>
	void LoadRecentFiles(IEnumerable<string> paths);

	/// <summary>
	/// Removes one entry from the recent files list.
	/// </summary>
	void RemoveRecentFile(string path);

	/// <summary>
	/// Removes entries whose file no longer exists.
	/// </summary>
	void RemoveMissingRecentFiles();

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
