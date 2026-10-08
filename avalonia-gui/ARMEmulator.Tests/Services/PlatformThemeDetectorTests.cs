using ARMEmulator.Services;
using ARMEmulator.Tests.Ui;
using Avalonia;
using Avalonia.Platform;
using Avalonia.Styling;
using AwesomeAssertions;

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
	public Task GetSystemTheme_ReportsThePlatformThemeNotTheApplicationTheme() =>
		UiTest.RunOnUiThread(() => {
			var app = Application.Current!;
			var requested = app.RequestedThemeVariant;
			try {
				// The headless platform reports a light theme; the application asks for dark on top of it
				app.RequestedThemeVariant = ThemeVariant.Dark;

				using var detector = new PlatformThemeDetector();

				detector.GetSystemTheme().Should().Be(PlatformTheme.Light);
			}
			finally {
				app.RequestedThemeVariant = requested;
			}
		});

	[Theory]
	[InlineData(PlatformThemeVariant.Light, PlatformTheme.Light)]
	[InlineData(PlatformThemeVariant.Dark, PlatformTheme.Dark)]
	public void ToTheme_MapsThePlatformVariant(PlatformThemeVariant platform, PlatformTheme expected)
	{
		PlatformThemeDetector.ToTheme(new PlatformColorValues { ThemeVariant = platform }).Should().Be(expected);
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
