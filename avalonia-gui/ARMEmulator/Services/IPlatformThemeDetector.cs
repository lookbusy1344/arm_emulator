namespace ARMEmulator.Services;

/// <summary>
/// Detects the system theme (light/dark mode) on the current platform.
/// </summary>
public interface IPlatformThemeDetector
{
	/// <summary>
	/// Gets the current system theme preference.
	/// </summary>
	PlatformTheme GetSystemTheme();

	/// <summary>
	/// Observable that emits when the system theme changes.
	/// </summary>
	IObservable<PlatformTheme> ThemeChanged { get; }
}

/// <summary>
/// Platform theme variants.
/// </summary>
public enum PlatformTheme
{
	/// <summary>Unknown or unsupported platform.</summary>
	Unknown,

	/// <summary>Light theme.</summary>
	Light,

	/// <summary>Dark theme.</summary>
	Dark
}
