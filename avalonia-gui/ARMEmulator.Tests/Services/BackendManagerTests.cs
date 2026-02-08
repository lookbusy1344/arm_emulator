using ARMEmulator.Services;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Tests for BackendManager.
/// Note: These are basic structural tests. Full process lifecycle tests require
/// integration testing with the actual backend binary.
/// </summary>
public sealed class BackendManagerTests
{
	[Fact]
	public void Constructor_InitializesWithStoppedStatus()
	{
		using var manager = new BackendManager();
		manager.Status.Should().Be(BackendStatus.Stopped);
		manager.BaseUrl.Should().NotBeNullOrEmpty();
	}

	[Fact]
	public void BaseUrl_ReturnsLocalhostUrl()
	{
		using var manager = new BackendManager();
		manager.BaseUrl.Should().StartWith("http://localhost:");
	}

	// Integration tests will verify:
	// - StartAsync launches the backend process
	// - StopAsync terminates the process gracefully
	// - HealthCheckAsync validates the /health endpoint
	// - StatusChanged emits events on state transitions
	// - Platform-specific binary discovery works on Windows/macOS/Linux
}
