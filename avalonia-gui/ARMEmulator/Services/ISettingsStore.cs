using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Persists <see cref="AppSettings"/>.
/// </summary>
public interface ISettingsStore
{
	/// <summary>Loads the stored settings. An unreadable store yields defaults and a warning.</summary>
	SettingsLoadResult Load();

	/// <summary>Writes the settings.</summary>
	/// <exception cref="IOException">The settings could not be written</exception>
	void Save(AppSettings settings);
}

/// <summary>Outcome of <see cref="ISettingsStore.Load"/>. <paramref name="Warning"/> is set when stored settings were discarded.</summary>
public sealed record SettingsLoadResult(AppSettings Settings, string? Warning = null);
