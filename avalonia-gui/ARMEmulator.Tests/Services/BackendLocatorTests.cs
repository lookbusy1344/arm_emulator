using ARMEmulator.Services;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Backend discovery against the layouts the release workflow packages.
/// </summary>
public sealed class BackendLocatorTests
{
	private static string Join(params string[] parts) => Path.Combine(parts);

	private static Func<string, bool> Only(params string[] existing) => existing.Contains;

	[Fact]
	public void MacOS_InAppBundle_FindsBackendInResources()
	{
		var contents = Join("Apps", "ARMEmulator.app", "Contents");
		var appDir = Join(contents, "MacOS") + Path.DirectorySeparatorChar;
		var backend = Join(contents, "Resources", "arm-emulator");

		BackendLocator.Find(BackendPlatform.MacOS, appDir, Only(backend)).Should().Be(backend);
	}

	[Fact]
	public void MacOS_InAppBundle_PrefersResourcesOverTheExecutableDirectory()
	{
		var contents = Join("Apps", "ARMEmulator.app", "Contents");
		var appDir = Join(contents, "MacOS") + Path.DirectorySeparatorChar;
		var resources = Join(contents, "Resources", "arm-emulator");

		BackendLocator.Find(BackendPlatform.MacOS, appDir, Only(Join(appDir, "arm-emulator"), resources)).Should().Be(resources);
	}

	[Fact]
	public void MacOS_OutsideABundle_FindsBackendNextToTheExecutable()
	{
		var appDir = Join("opt", "arm") + Path.DirectorySeparatorChar;
		var backend = Join(appDir, "arm-emulator");

		BackendLocator.Find(BackendPlatform.MacOS, appDir, Only(backend)).Should().Be(backend);
	}

	[Fact]
	public void Windows_FindsBackendExeNextToTheExecutable()
	{
		var appDir = Join("Program Files", "ARM") + Path.DirectorySeparatorChar;
		var backend = Join(appDir, "arm-emulator.exe");

		BackendLocator.Find(BackendPlatform.Windows, appDir, Only(backend)).Should().Be(backend);
	}

	[Fact]
	public void Windows_IgnoresTheExtensionlessName() =>
		BackendLocator.Find(BackendPlatform.Windows, Join("ARM") + Path.DirectorySeparatorChar, Only(Join("ARM", "arm-emulator")))
			.Should().BeNull();

	[Fact]
	public void Linux_FindsBackendNextToTheExecutable()
	{
		var appDir = Join("opt", "arm") + Path.DirectorySeparatorChar;
		var backend = Join(appDir, "arm-emulator");

		BackendLocator.Find(BackendPlatform.Linux, appDir, Only(backend, "/usr/local/bin/arm-emulator")).Should().Be(backend);
	}

	[Fact]
	public void Linux_FallsBackToSystemLocationsInOrder()
	{
		var appDir = Join("opt", "arm") + Path.DirectorySeparatorChar;

		BackendLocator.Find(BackendPlatform.Linux, appDir, Only("/usr/share/arm-emulator/arm-emulator", "/usr/local/bin/arm-emulator"))
			.Should().Be("/usr/local/bin/arm-emulator");
		BackendLocator.Find(BackendPlatform.Linux, appDir, Only("/usr/share/arm-emulator/arm-emulator"))
			.Should().Be("/usr/share/arm-emulator/arm-emulator");
	}

	[Theory]
	[InlineData(BackendPlatform.Windows)]
	[InlineData(BackendPlatform.MacOS)]
	[InlineData(BackendPlatform.Linux)]
	public void WithNoBackend_ReturnsNull(BackendPlatform platform) =>
		BackendLocator.Find(platform, Join("opt", "arm") + Path.DirectorySeparatorChar, _ => false).Should().BeNull();
}
