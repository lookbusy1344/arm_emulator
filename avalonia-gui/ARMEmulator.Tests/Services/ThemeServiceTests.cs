using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using Avalonia.Styling;
using FluentAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Services;

public class ThemeServiceTests : IDisposable
{
	private readonly List<IDisposable> disposables = [];

	[Fact]
	public void Constructor_DoesNotThrow()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Light);

		// Act
		var act = () => new ThemeService(detector);

		// Assert
		act.Should().NotThrow();
	}

	[Fact]
	public void GetEffectiveTheme_WithLightTheme_ReturnsLight()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Light);
		using var service = new ThemeService(detector);

		// Act
		service.ApplyTheme(AppTheme.Light);
		var theme = service.GetEffectiveTheme();

		// Assert
		theme.Should().Be(ThemeVariant.Light);
	}

	[Fact]
	public void GetEffectiveTheme_WithDarkTheme_ReturnsDark()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Dark);
		using var service = new ThemeService(detector);

		// Act
		service.ApplyTheme(AppTheme.Dark);
		var theme = service.GetEffectiveTheme();

		// Assert
		theme.Should().Be(ThemeVariant.Dark);
	}

	[Fact]
	public void GetEffectiveTheme_WithAutoAndPlatformLight_ReturnsLight()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Light);
		using var service = new ThemeService(detector);

		// Act
		service.ApplyTheme(AppTheme.Auto);
		var theme = service.GetEffectiveTheme();

		// Assert
		theme.Should().Be(ThemeVariant.Light);
	}

	[Fact]
	public void GetEffectiveTheme_WithAutoAndPlatformDark_ReturnsDark()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Dark);
		using var service = new ThemeService(detector);

		// Act
		service.ApplyTheme(AppTheme.Auto);
		var theme = service.GetEffectiveTheme();

		// Assert
		theme.Should().Be(ThemeVariant.Dark);
	}

	[Fact]
	public void Dispose_DoesNotThrow()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Light);
		using var service = new ThemeService(detector);

		// Act
		var act = () => service.Dispose();

		// Assert
		act.Should().NotThrow();
	}

	[Fact]
	public void MultipleDispose_DoesNotThrow()
	{
		// Arrange
		var detector = CreateMockDetector(PlatformTheme.Light);
		using var service = new ThemeService(detector);

		// Act
		var act = () => {
			service.Dispose();
			service.Dispose();
		};

		// Assert
		act.Should().NotThrow();
	}

	private IPlatformThemeDetector CreateMockDetector(PlatformTheme theme)
	{
		var subject = new BehaviorSubject<PlatformTheme>(theme);
		disposables.Add(subject);

		var detector = Substitute.For<IPlatformThemeDetector>();
		detector.GetSystemTheme().Returns(theme);
		detector.ThemeChanged.Returns(subject);

		return detector;
	}

	public void Dispose()
	{
		foreach (var disposable in disposables) {
			disposable.Dispose();
		}
		disposables.Clear();
		GC.SuppressFinalize(this);
	}
}
