using ARMEmulator.Services;
using FluentAssertions;

namespace ARMEmulator.Tests.Services;

public class PlatformThemeDetectorTests
{
	[Fact]
	public void GetSystemTheme_ReturnsValidTheme()
	{
		// Arrange
		using var detector = new PlatformThemeDetector();

		// Act
		var theme = detector.GetSystemTheme();

		// Assert
		theme.Should().BeOneOf(PlatformTheme.Light, PlatformTheme.Dark);
	}

	[Fact]
	public void ThemeChanged_IsObservable()
	{
		// Arrange
		using var detector = new PlatformThemeDetector();

		// Act & Assert
		detector.ThemeChanged.Should().NotBeNull();
	}

	[Fact]
	public void ThemeChanged_EmitsInitialValue()
	{
		// Arrange
		using var detector = new PlatformThemeDetector();
		PlatformTheme? receivedTheme = null;

		// Act
		using var subscription = detector.ThemeChanged.Subscribe(theme => receivedTheme = theme);

		// Assert
		receivedTheme.Should().BeOneOf(PlatformTheme.Light, PlatformTheme.Dark);
	}

	[Fact]
	public void Dispose_DoesNotThrow()
	{
		// Arrange
		using var detector = new PlatformThemeDetector();

		// Act
		var act = () => detector.Dispose();

		// Assert
		act.Should().NotThrow();
	}

	[Fact]
	public void MultipleDispose_DoesNotThrow()
	{
		// Arrange
		using var detector = new PlatformThemeDetector();

		// Act
		var act = () => {
			detector.Dispose();
			detector.Dispose();
		};

		// Assert
		act.Should().NotThrow();
	}
}
