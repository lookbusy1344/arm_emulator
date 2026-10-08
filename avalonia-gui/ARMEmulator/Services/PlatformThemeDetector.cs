using System.ComponentModel;
using System.Diagnostics;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using Avalonia;
using Avalonia.Platform;

namespace ARMEmulator.Services;

/// <summary>
/// Follows the operating system's light or dark preference through Avalonia's platform settings.
/// Reads the platform, not the application, so a theme the user picked does not feed back into Auto.
/// </summary>
public sealed class PlatformThemeDetector : IPlatformThemeDetector, IDisposable
{
	private readonly IPlatformSettings? settings;
	private readonly BehaviorSubject<PlatformTheme> themeSubject;

	/// <summary>Follows the platform settings of the running application.</summary>
	public PlatformThemeDetector()
		: this(Application.Current?.PlatformSettings)
	{
	}

	/// <summary>Follows <paramref name="settings"/>. Without settings, macOS is asked directly and other systems report light.</summary>
	private PlatformThemeDetector(IPlatformSettings? settings)
	{
		this.settings = settings;
		themeSubject = new BehaviorSubject<PlatformTheme>(settings is null ? DetectWithoutAvalonia() : ToTheme(settings.GetColorValues()));
		if (settings is not null) {
			settings.ColorValuesChanged += OnColorValuesChanged;
		}
	}

	public PlatformTheme GetSystemTheme() => themeSubject.Value;

	public IObservable<PlatformTheme> ThemeChanged => themeSubject.AsObservable();

	public void Dispose()
	{
		if (settings is not null) {
			settings.ColorValuesChanged -= OnColorValuesChanged;
		}

		themeSubject.Dispose();
	}

	private void OnColorValuesChanged(object? sender, PlatformColorValues colors) => themeSubject.OnNext(ToTheme(colors));

	internal static PlatformTheme ToTheme(PlatformColorValues colors) =>
		colors.ThemeVariant == PlatformThemeVariant.Dark ? PlatformTheme.Dark : PlatformTheme.Light;

	private static PlatformTheme DetectWithoutAvalonia() =>
		OperatingSystem.IsMacOS() ? ReadMacOSAppearance() : PlatformTheme.Light;

	/// <summary>
	/// <c>defaults read -g AppleInterfaceStyle</c> prints "Dark" in dark mode and fails when the key is absent, which means light.
	/// </summary>
	private static PlatformTheme ReadMacOSAppearance()
	{
		var startInfo = new ProcessStartInfo {
			FileName = "defaults",
			UseShellExecute = false,
			RedirectStandardOutput = true,
			RedirectStandardError = true,
			CreateNoWindow = true,
			ArgumentList = { "read", "-g", "AppleInterfaceStyle" }
		};

		try {
			using var process = Process.Start(startInfo);
			if (process is null) {
				return PlatformTheme.Light;
			}

			var output = process.StandardOutput.ReadToEnd().Trim();
			process.WaitForExit();
			return output.Equals("Dark", StringComparison.OrdinalIgnoreCase) ? PlatformTheme.Dark : PlatformTheme.Light;
		}
		catch (Win32Exception) {
			// defaults is missing or cannot run
			return PlatformTheme.Light;
		}
	}
}
