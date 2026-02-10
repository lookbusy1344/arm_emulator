using ARMEmulator.Models;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// ViewModel for the Preferences dialog.
/// Provides editable application settings with validation.
/// </summary>
public sealed class PreferencesWindowViewModel : ReactiveObject
{
	private string backendUrl;
	private int editorFontSize;
	private AppTheme selectedTheme;
	private int recentFilesLimit;
	private bool autoScrollToMemoryWrites;

	public PreferencesWindowViewModel(AppSettings settings)
	{
		backendUrl = settings.BackendUrl;
		editorFontSize = settings.EditorFontSize;
		selectedTheme = settings.Theme;
		recentFilesLimit = settings.RecentFilesLimit;
		autoScrollToMemoryWrites = settings.AutoScrollToMemoryWrites;
	}

	/// <summary>Backend API base URL.</summary>
	public string BackendUrl
	{
		get => backendUrl;
		set => this.RaiseAndSetIfChanged(ref backendUrl, value);
	}

	/// <summary>Editor font size (10-24pt).</summary>
	public int EditorFontSize
	{
		get => editorFontSize;
		set => this.RaiseAndSetIfChanged(ref editorFontSize, value);
	}

	/// <summary>Selected application theme.</summary>
	public AppTheme SelectedTheme
	{
		get => selectedTheme;
		set => this.RaiseAndSetIfChanged(ref selectedTheme, value);
	}

	/// <summary>Maximum recent files to track.</summary>
	public int RecentFilesLimit
	{
		get => recentFilesLimit;
		set => this.RaiseAndSetIfChanged(ref recentFilesLimit, value);
	}

	/// <summary>Auto-scroll memory view to writes.</summary>
	public bool AutoScrollToMemoryWrites
	{
		get => autoScrollToMemoryWrites;
		set => this.RaiseAndSetIfChanged(ref autoScrollToMemoryWrites, value);
	}

	/// <summary>Available theme options for selection.</summary>
	public IReadOnlyList<AppTheme> ThemeOptions { get; } = [AppTheme.Auto, AppTheme.Light, AppTheme.Dark];

	/// <summary>
	/// Builds an AppSettings instance from current values with validation.
	/// </summary>
	public AppSettings BuildSettings() => new AppSettings {
		BackendUrl = BackendUrl,
		EditorFontSize = EditorFontSize,
		Theme = SelectedTheme,
		RecentFilesLimit = RecentFilesLimit,
		AutoScrollToMemoryWrites = AutoScrollToMemoryWrites
	}.Validate();
}
