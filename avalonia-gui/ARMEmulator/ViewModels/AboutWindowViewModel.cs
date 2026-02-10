using System.Reflection;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// ViewModel for the About dialog.
/// Displays app and backend version information.
/// </summary>
public sealed class AboutWindowViewModel : ReactiveObject
{
	private readonly IApiClient api;

	private string backendVersion = "Loading...";
	private string backendCommit = "";
	private string backendBuildDate = "";
	private bool isBackendAvailable = true;

	public AboutWindowViewModel(IApiClient api)
	{
		this.api = api;

		// Get app version from assembly
		var version = Assembly.GetExecutingAssembly().GetName().Version;
		AppVersion = version is not null
			? $"{version.Major}.{version.Minor}.{version.Build}"
			: "Unknown";
	}

	/// <summary>Application version.</summary>
	public string AppVersion { get; }

	/// <summary>Backend version string.</summary>
	public string BackendVersion
	{
		get => backendVersion;
		private set => this.RaiseAndSetIfChanged(ref backendVersion, value);
	}

	/// <summary>Backend commit hash.</summary>
	public string BackendCommit
	{
		get => backendCommit;
		private set => this.RaiseAndSetIfChanged(ref backendCommit, value);
	}

	/// <summary>Backend build date.</summary>
	public string BackendBuildDate
	{
		get => backendBuildDate;
		private set => this.RaiseAndSetIfChanged(ref backendBuildDate, value);
	}

	/// <summary>Whether backend is available.</summary>
	public bool IsBackendAvailable
	{
		get => isBackendAvailable;
		private set => this.RaiseAndSetIfChanged(ref isBackendAvailable, value);
	}

	/// <summary>
	/// Loads backend version information asynchronously.
	/// Call this after the dialog is shown to avoid blocking.
	/// </summary>
	public async Task LoadBackendVersionAsync()
	{
		try {
			var version = await api.GetVersionAsync();
			BackendVersion = version.Version;
			BackendCommit = version.Commit;
			BackendBuildDate = version.BuildDate;
			IsBackendAvailable = true;
		}
		catch (BackendUnavailableException) {
			BackendVersion = "Not available";
			BackendCommit = "";
			BackendBuildDate = "";
			IsBackendAvailable = false;
		}
		catch (Exception) {
			BackendVersion = "Error loading version";
			BackendCommit = "";
			BackendBuildDate = "";
			IsBackendAvailable = false;
		}
	}
}
