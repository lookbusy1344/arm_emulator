using ARMEmulator.Models;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

public sealed class JsonSettingsStoreTests : IDisposable
{
	private const string DamagedWarning = "Settings could not be read and were reset to defaults. The damaged file is kept as settings.json.bak.";

	private readonly DirectoryInfo directory = Directory.CreateTempSubdirectory("arm-settings-tests");

	public void Dispose() => directory.Delete(recursive: true);

	private string SettingsPath => Path.Combine(directory.FullName, "ARMEmulator", "settings.json");

	private string BackupPath => SettingsPath + ".bak";

	private JsonSettingsStore CreateStore() => new(SettingsPath);

	private void WriteSettingsFile(string content)
	{
		_ = Directory.CreateDirectory(Path.GetDirectoryName(SettingsPath)!);
		File.WriteAllText(SettingsPath, content);
	}

	private static readonly AppSettings Custom = new() {
		BackendUrl = "http://localhost:9090",
		EditorFontSize = 18,
		Theme = AppTheme.Dark,
		RecentFilesLimit = 5,
		AutoScrollToMemoryWrites = false
	};

	[Fact]
	public void Load_WithMissingFile_ReturnsDefaultsWithoutWarning()
	{
		var result = CreateStore().Load();

		result.Settings.Should().Be(AppSettings.Default);
		result.Warning.Should().BeNull();
	}

	[Fact]
	public void SaveThenLoad_RoundTripsEverySetting()
	{
		var store = CreateStore();

		store.Save(Custom);
		var result = store.Load();

		result.Settings.Should().Be(Custom);
		result.Warning.Should().BeNull();
	}

	[Fact]
	public void Save_WritesThemeAsName_AndCreatesDirectory()
	{
		CreateStore().Save(Custom);

		var json = File.ReadAllText(SettingsPath);
		json.Should().Contain("\"Theme\": \"Dark\"");
		Directory.GetFiles(Path.GetDirectoryName(SettingsPath)!).Should().Equal(SettingsPath);
	}

	[Fact]
	public void Load_WithMalformedJson_ReturnsDefaultsWarnsAndKeepsBackup()
	{
		const string damaged = "{ not json";
		WriteSettingsFile(damaged);

		var result = CreateStore().Load();

		result.Settings.Should().Be(AppSettings.Default);
		result.Warning.Should().Be(DamagedWarning);
		File.ReadAllText(BackupPath).Should().Be(damaged);
	}

	[Fact]
	public void Load_WithMissingRequiredProperty_TreatsFileAsInvalid()
	{
		WriteSettingsFile("{}");

		var result = CreateStore().Load();

		result.Settings.Should().Be(AppSettings.Default);
		result.Warning.Should().Be(DamagedWarning);
		File.Exists(BackupPath).Should().BeTrue();
	}

	[Fact]
	public void Load_WithNullDocument_TreatsFileAsInvalid()
	{
		WriteSettingsFile("null");

		var result = CreateStore().Load();

		result.Warning.Should().Be(DamagedWarning);
	}

	[Fact]
	public void Load_AfterDamagedFile_WarnsOnlyOnce()
	{
		WriteSettingsFile("{ not json");
		var store = CreateStore();

		_ = store.Load();
		var second = store.Load();

		second.Warning.Should().BeNull();
		second.Settings.Should().Be(AppSettings.Default);
	}

	[Fact]
	public void Load_ReplacesAnEarlierBackup()
	{
		WriteSettingsFile("first");
		_ = CreateStore().Load();
		WriteSettingsFile("second");

		_ = CreateStore().Load();

		File.ReadAllText(BackupPath).Should().Be("second");
	}

	[Fact]
	public void Load_ClampsOutOfRangeValues()
	{
		WriteSettingsFile("""
			{ "BackendUrl": "http://localhost:8080", "EditorFontSize": 99, "Theme": "Light",
			  "RecentFilesLimit": 0, "AutoScrollToMemoryWrites": true }
			""");

		var result = CreateStore().Load();

		result.Settings.EditorFontSize.Should().Be(24);
		result.Settings.RecentFilesLimit.Should().Be(1);
		result.Settings.Theme.Should().Be(AppTheme.Light);
		result.Warning.Should().BeNull();
	}

	[Fact]
	public void Load_WithInvalidBackendUrl_UsesDefaultUrl()
	{
		WriteSettingsFile("""
			{ "BackendUrl": "not a url", "EditorFontSize": 12, "Theme": "Auto",
			  "RecentFilesLimit": 3, "AutoScrollToMemoryWrites": true }
			""");

		var result = CreateStore().Load();

		result.Settings.BackendUrl.Should().Be(AppSettings.Default.BackendUrl);
		result.Settings.EditorFontSize.Should().Be(12);
	}

	[Fact]
	public void Load_WithUnknownThemeName_TreatsFileAsInvalid()
	{
		WriteSettingsFile("""
			{ "BackendUrl": "http://localhost:8080", "EditorFontSize": 12, "Theme": "Sepia",
			  "RecentFilesLimit": 3, "AutoScrollToMemoryWrites": true }
			""");

		var result = CreateStore().Load();

		result.Warning.Should().Be(DamagedWarning);
	}
}
