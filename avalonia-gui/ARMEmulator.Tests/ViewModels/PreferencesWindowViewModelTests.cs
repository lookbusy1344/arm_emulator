using ARMEmulator.Models;
using ARMEmulator.ViewModels;
using FluentAssertions;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for PreferencesWindowViewModel.
/// </summary>
public sealed class PreferencesWindowViewModelTests
{
	[Fact]
	public void Constructor_InitializesWithDefaultSettings()
	{
		// Arrange
		var settings = AppSettings.Default;

		// Act
		var vm = new PreferencesWindowViewModel(settings);

		// Assert
		vm.BackendUrl.Should().Be("http://localhost:8080");
		vm.EditorFontSize.Should().Be(14);
		vm.SelectedTheme.Should().Be(AppTheme.Auto);
		vm.RecentFilesLimit.Should().Be(10);
		vm.AutoScrollToMemoryWrites.Should().BeTrue();
	}

	[Fact]
	public void Constructor_InitializesWithCustomSettings()
	{
		// Arrange
		var settings = AppSettings.Default with {
			BackendUrl = "http://192.168.1.100:8080",
			EditorFontSize = 18,
			Theme = AppTheme.Dark,
			RecentFilesLimit = 15,
			AutoScrollToMemoryWrites = false
		};

		// Act
		var vm = new PreferencesWindowViewModel(settings);

		// Assert
		vm.BackendUrl.Should().Be("http://192.168.1.100:8080");
		vm.EditorFontSize.Should().Be(18);
		vm.SelectedTheme.Should().Be(AppTheme.Dark);
		vm.RecentFilesLimit.Should().Be(15);
		vm.AutoScrollToMemoryWrites.Should().BeFalse();
	}

	[Fact]
	public void BuildSettings_CreatesValidatedAppSettings()
	{
		// Arrange
		var vm = new PreferencesWindowViewModel(AppSettings.Default);
		vm.BackendUrl = "http://custom:9090";
		vm.EditorFontSize = 5; // Too small, should be clamped
		vm.SelectedTheme = AppTheme.Light;

		// Act
		var settings = vm.BuildSettings();

		// Assert
		settings.BackendUrl.Should().Be("http://custom:9090");
		settings.EditorFontSize.Should().Be(10, "font size should be clamped to minimum");
		settings.Theme.Should().Be(AppTheme.Light);
	}

	[Fact]
	public void EditorFontSize_SupportsValidRange()
	{
		// Arrange
		var vm = new PreferencesWindowViewModel(AppSettings.Default);

		// Act
		vm.EditorFontSize = 10;
		var min = vm.EditorFontSize;

		vm.EditorFontSize = 24;
		var max = vm.EditorFontSize;

		vm.EditorFontSize = 16;
		var mid = vm.EditorFontSize;

		// Assert
		min.Should().Be(10);
		max.Should().Be(24);
		mid.Should().Be(16);
	}

	[Fact]
	public void ThemeOptions_ContainsAllValues()
	{
		// Arrange
		var vm = new PreferencesWindowViewModel(AppSettings.Default);

		// Assert
		vm.ThemeOptions.Should().BeEquivalentTo([AppTheme.Auto, AppTheme.Light, AppTheme.Dark]);
	}

	[Fact]
	public void RecentFilesLimit_SupportsCustomValues()
	{
		// Arrange
		var vm = new PreferencesWindowViewModel(AppSettings.Default);

		// Act
		vm.RecentFilesLimit = 20;

		// Assert
		vm.RecentFilesLimit.Should().Be(20);
	}
}
