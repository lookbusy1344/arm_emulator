using System.Text.Json;
using System.Text.Json.Serialization;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Stores settings as JSON in a single file. A damaged file is kept as <c>.bak</c> and replaced with defaults.
/// </summary>
public sealed class JsonSettingsStore(string path) : ISettingsStore
{
	private const string BackupSuffix = ".bak";
	private const string DamagedWarning = "Settings could not be read and were reset to defaults. The damaged file is kept as settings.json.bak.";

	/// <summary>The per-user settings file: <c>ApplicationData/ARMEmulator/settings.json</c>.</summary>
	public static string DefaultPath { get; } = Path.Combine(
		Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
		"ARMEmulator",
		"settings.json");

	public SettingsLoadResult Load()
	{
		if (!File.Exists(path)) {
			return new SettingsLoadResult(AppSettings.Default);
		}

		try {
			var settings = JsonSerializer.Deserialize(File.ReadAllText(path), SettingsJsonContext.Default.AppSettings)
				?? throw new JsonException("Settings document is null");
			return new SettingsLoadResult(settings.Validate());
		}
		catch (JsonException) {
			File.Copy(path, path + BackupSuffix, overwrite: true);
			Save(AppSettings.Default);
			return new SettingsLoadResult(AppSettings.Default, DamagedWarning);
		}
	}

	public void Save(AppSettings settings)
	{
		_ = Directory.CreateDirectory(Path.GetDirectoryName(path)!);

		var temporaryPath = path + ".tmp";
		File.WriteAllText(temporaryPath, JsonSerializer.Serialize(settings, SettingsJsonContext.Default.AppSettings));
		File.Move(temporaryPath, path, overwrite: true);
	}
}

[JsonSerializable(typeof(AppSettings))]
[JsonSourceGenerationOptions(
	WriteIndented = true,
	UseStringEnumConverter = true,
	PropertyNameCaseInsensitive = true)]
internal sealed partial class SettingsJsonContext : JsonSerializerContext;
