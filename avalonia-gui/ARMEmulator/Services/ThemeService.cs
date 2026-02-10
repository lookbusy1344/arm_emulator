using System.Reactive.Linq;
using ARMEmulator.Models;
using Avalonia;
using Avalonia.Styling;

namespace ARMEmulator.Services;

/// <summary>
/// Manages application theme switching based on settings and platform detection.
/// Integrates AppTheme settings with Avalonia's theme system.
/// </summary>
public sealed class ThemeService : IDisposable
{
	private readonly IPlatformThemeDetector platformDetector;
	private readonly IDisposable? themeSubscription;
	private AppTheme currentMode = AppTheme.Auto;

	public ThemeService(IPlatformThemeDetector platformDetector)
	{
		this.platformDetector = platformDetector;

		// Subscribe to platform theme changes
		themeSubscription = platformDetector.ThemeChanged
			.Where(_ => currentMode == AppTheme.Auto)
			.Subscribe(_ => ApplyTheme(currentMode));
	}

	/// <summary>
	/// Applies the specified theme mode to the application.
	/// </summary>
	public void ApplyTheme(AppTheme theme)
	{
		currentMode = theme;

		if (Application.Current is null) {
			return;
		}

		var effectiveTheme = theme switch {
			AppTheme.Light => ThemeVariant.Light,
			AppTheme.Dark => ThemeVariant.Dark,
			AppTheme.Auto => DetectPlatformThemeVariant(),
			_ => ThemeVariant.Default
		};

		Application.Current.RequestedThemeVariant = effectiveTheme;
	}

	/// <summary>
	/// Gets the current effective theme variant based on settings and platform.
	/// </summary>
	public ThemeVariant GetEffectiveTheme() =>
		currentMode switch {
			AppTheme.Light => ThemeVariant.Light,
			AppTheme.Dark => ThemeVariant.Dark,
			AppTheme.Auto => DetectPlatformThemeVariant(),
			_ => ThemeVariant.Default
		};

	private ThemeVariant DetectPlatformThemeVariant()
	{
		var platformTheme = platformDetector.GetSystemTheme();
		return platformTheme switch {
			PlatformTheme.Dark => ThemeVariant.Dark,
			PlatformTheme.Light => ThemeVariant.Light,
			_ => ThemeVariant.Default
		};
	}

	public void Dispose()
	{
		themeSubscription?.Dispose();
	}
}
