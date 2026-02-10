using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

/// <summary>
/// Tests for AppSettings model serialization and defaults.
/// </summary>
public class AppSettingsTests
{
	[Fact]
	public void DefaultSettings_ShouldHaveCorrectValues()
	{
		// Act
		var settings = AppSettings.Default;

		// Assert
		settings.BackendUrl.Should().Be("http://localhost:8080");
		settings.EditorFontSize.Should().Be(14);
		settings.Theme.Should().Be(AppTheme.Auto);
		settings.RecentFilesLimit.Should().Be(10);
		settings.AutoScrollToMemoryWrites.Should().BeTrue();
	}

	[Fact]
	public void Theme_ShouldSupportAllValues()
	{
		// Arrange
		var settings = AppSettings.Default;

		// Act & Assert
		settings = settings with { Theme = AppTheme.Light };
		settings.Theme.Should().Be(AppTheme.Light);

		settings = settings with { Theme = AppTheme.Dark };
		settings.Theme.Should().Be(AppTheme.Dark);

		settings = settings with { Theme = AppTheme.Auto };
		settings.Theme.Should().Be(AppTheme.Auto);
	}

	[Fact]
	public void Validate_ShouldClampFontSizeToValidRange()
	{
		// Arrange
		var settings = AppSettings.Default;

		// Act
		var tooSmall = (settings with { EditorFontSize = 5 }).Validate();
		var tooLarge = (settings with { EditorFontSize = 50 }).Validate();
		var valid = (settings with { EditorFontSize = 16 }).Validate();

		// Assert
		tooSmall.EditorFontSize.Should().Be(10, "font size should be clamped to minimum");
		tooLarge.EditorFontSize.Should().Be(24, "font size should be clamped to maximum");
		valid.EditorFontSize.Should().Be(16, "valid font size should be unchanged");
	}

	[Fact]
	public void RecentFilesLimit_ShouldBePositive()
	{
		// Arrange
		var settings = AppSettings.Default;

		// Act
		var withLimit = settings with { RecentFilesLimit = 15 };

		// Assert
		withLimit.RecentFilesLimit.Should().Be(15);
		withLimit.RecentFilesLimit.Should().BeGreaterThan(0);
	}

	[Fact]
	public void BackendUrl_ShouldAllowCustomValues()
	{
		// Arrange
		var settings = AppSettings.Default;

		// Act
		var custom = settings with { BackendUrl = "http://192.168.1.100:8080" };

		// Assert
		custom.BackendUrl.Should().Be("http://192.168.1.100:8080");
	}

	[Fact]
	public void Record_ShouldSupportImmutableUpdates()
	{
		// Arrange
		var original = AppSettings.Default;

		// Act
		var updated = original with {
			EditorFontSize = 18,
			Theme = AppTheme.Dark,
			AutoScrollToMemoryWrites = false
		};

		// Assert
		original.EditorFontSize.Should().Be(14);
		original.Theme.Should().Be(AppTheme.Auto);
		original.AutoScrollToMemoryWrites.Should().BeTrue();

		updated.EditorFontSize.Should().Be(18);
		updated.Theme.Should().Be(AppTheme.Dark);
		updated.AutoScrollToMemoryWrites.Should().BeFalse();
	}
}
