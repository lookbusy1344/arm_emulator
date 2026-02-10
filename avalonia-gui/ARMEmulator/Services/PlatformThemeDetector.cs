using System.Diagnostics;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Runtime.InteropServices;
using Avalonia;
using Avalonia.Platform;
using Avalonia.Styling;

namespace ARMEmulator.Services;

/// <summary>
/// Platform-specific theme detection using Avalonia's platform services
/// and OS-specific APIs.
/// </summary>
public sealed class PlatformThemeDetector : IPlatformThemeDetector, IDisposable
{
	private readonly BehaviorSubject<PlatformTheme> themeSubject;
	private readonly IDisposable? themeChangeSubscription;

	public PlatformThemeDetector()
	{
		var initialTheme = DetectSystemTheme();
		themeSubject = new BehaviorSubject<PlatformTheme>(initialTheme);

		// Subscribe to Avalonia's actual theme changed event if available
		if (Application.Current is not null)
		{
			themeChangeSubscription = Observable
				.FromEventPattern(
					h => Application.Current.ActualThemeVariantChanged += h,
					h => Application.Current.ActualThemeVariantChanged -= h)
				.Select(_ => DetectSystemTheme())
				.Subscribe(theme => themeSubject.OnNext(theme));
		}
	}

	public PlatformTheme GetSystemTheme() => themeSubject.Value;

	public IObservable<PlatformTheme> ThemeChanged => themeSubject.AsObservable();

	private static PlatformTheme DetectSystemTheme()
	{
		try
		{
			// Use Avalonia's theme detection first
			if (Application.Current?.ActualThemeVariant is { } avaloniaTheme)
			{
				return avaloniaTheme == ThemeVariant.Dark ? PlatformTheme.Dark : PlatformTheme.Light;
			}

			// Fall back to platform-specific detection
			if (RuntimeInformation.IsOSPlatform(OSPlatform.OSX))
			{
				return DetectMacOSTheme();
			}

			if (RuntimeInformation.IsOSPlatform(OSPlatform.Windows))
			{
				return DetectWindowsTheme();
			}

			if (RuntimeInformation.IsOSPlatform(OSPlatform.Linux))
			{
				return DetectLinuxTheme();
			}
		}
		catch
		{
			// Fall back to light theme on detection failure
		}

		return PlatformTheme.Light;
	}

	private static PlatformTheme DetectMacOSTheme()
	{
		try
		{
			using var process = new Process
			{
				StartInfo = new ProcessStartInfo
				{
					FileName = "defaults",
					Arguments = "read -g AppleInterfaceStyle",
					UseShellExecute = false,
					RedirectStandardOutput = true,
					RedirectStandardError = true,
					CreateNoWindow = true
				}
			};

			if (process.Start())
			{
				var output = process.StandardOutput.ReadToEnd().Trim();
				process.WaitForExit();

				// If "Dark" is returned, it's dark mode. If the key doesn't exist (exit code 1), it's light mode.
				return output.Equals("Dark", StringComparison.OrdinalIgnoreCase)
					? PlatformTheme.Dark
					: PlatformTheme.Light;
			}
		}
		catch
		{
			// Fall back to light theme
		}

		return PlatformTheme.Light;
	}

	private static PlatformTheme DetectWindowsTheme()
	{
		// Windows 10/11 theme detection via registry
		// This is a simplified version - Avalonia should handle this better
		try
		{
			if (OperatingSystem.IsWindows())
			{
				// Avalonia's platform detection should handle this
				// For now, default to light
			}
		}
		catch
		{
			// Fall back to light theme
		}

		return PlatformTheme.Light;
	}

	private static PlatformTheme DetectLinuxTheme()
	{
		// Linux theme detection is complex (depends on DE: GNOME, KDE, etc.)
		// Avalonia's platform services should handle this
		// For now, default to light
		return PlatformTheme.Light;
	}

	public void Dispose()
	{
		themeChangeSubscription?.Dispose();
		themeSubject.Dispose();
	}
}
