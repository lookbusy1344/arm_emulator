namespace ARMEmulator.Services;

/// <summary>The platforms the backend is packaged for.</summary>
public enum BackendPlatform
{
	Windows,
	MacOS,
	Linux
}

/// <summary>
/// Finds the backend executable in the layouts the release workflow packages: next to the application on
/// Windows and Linux, and in <c>Contents/Resources</c> of the application bundle on macOS.
/// </summary>
public static class BackendLocator
{
	private const string BundleMarker = ".app/Contents/";
	private const string UnixBinaryName = "arm-emulator";
	private const string WindowsBinaryName = "arm-emulator.exe";

	private static readonly ImmutableArray<string> LinuxSystemLocations = ["/usr/local/bin/arm-emulator", "/usr/share/arm-emulator/arm-emulator"];

	/// <summary>The first existing candidate for the platform, or null.</summary>
	/// <param name="platform">The platform to search for.</param>
	/// <param name="appDir">The application's base directory.</param>
	/// <param name="exists">Whether a file exists.</param>
	public static string? Find(BackendPlatform platform, string appDir, Func<string, bool> exists) =>
		Candidates(platform, appDir).FirstOrDefault(exists);

	/// <summary>The platform of the running process, or null when unsupported.</summary>
	public static BackendPlatform? Current
	{
		get
		{
			if (OperatingSystem.IsWindows()) {
				return BackendPlatform.Windows;
			}

			if (OperatingSystem.IsMacOS()) {
				return BackendPlatform.MacOS;
			}

			return OperatingSystem.IsLinux() ? BackendPlatform.Linux : null;
		}
	}

	private static IEnumerable<string> Candidates(BackendPlatform platform, string appDir)
	{
		var parentDir = Directory.GetParent(appDir)?.FullName;
		return platform switch {
			BackendPlatform.Windows => NextToApp(appDir, parentDir, WindowsBinaryName),
			BackendPlatform.MacOS => [.. BundleResources(appDir), .. NextToApp(appDir, parentDir, UnixBinaryName)],
			_ => [Path.Combine(appDir, UnixBinaryName), .. LinuxSystemLocations]
		};
	}

	private static IEnumerable<string> BundleResources(string appDir)
	{
		var markerIndex = appDir.IndexOf(BundleMarker, StringComparison.Ordinal);
		return markerIndex < 0
			? []
			: [Path.Combine(appDir[..markerIndex] + ".app/Contents", "Resources", UnixBinaryName)];
	}

	private static IEnumerable<string> NextToApp(string appDir, string? parentDir, string name) =>
		parentDir is null ? [Path.Combine(appDir, name)] : [Path.Combine(appDir, name), Path.Combine(parentDir, name)];
}
