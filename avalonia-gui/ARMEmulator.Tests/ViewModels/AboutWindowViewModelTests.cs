using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for AboutWindowViewModel.
/// </summary>
public sealed class AboutWindowViewModelTests
{
	[Fact]
	public void Constructor_InitializesAppVersion()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();

		// Act
		var vm = new AboutWindowViewModel(mockApi);

		// Assert
		vm.AppVersion.Should().NotBeNullOrWhiteSpace();
		vm.AppVersion.Should().MatchRegex(@"\d+\.\d+\.\d+"); // Semantic version
	}

	[Fact]
	public void Constructor_InitializesBackendVersionAsLoading()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();

		// Act
		var vm = new AboutWindowViewModel(mockApi);

		// Assert
		vm.BackendVersion.Should().Be("Loading...");
		vm.BackendCommit.Should().Be("");
		vm.BackendBuildDate.Should().Be("");
	}

	[Fact]
	public async Task LoadBackendVersionAsync_UpdatesPropertiesOnSuccess()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetVersionAsync(Arg.Any<CancellationToken>())
			.Returns(new BackendVersion("1.2.3", "abc123", "2025-02-10"));

		var vm = new AboutWindowViewModel(mockApi);

		// Act
		await vm.LoadBackendVersionAsync();

		// Assert
		vm.BackendVersion.Should().Be("1.2.3");
		vm.BackendCommit.Should().Be("abc123");
		vm.BackendBuildDate.Should().Be("2025-02-10");
		vm.IsBackendAvailable.Should().BeTrue();
	}

	[Fact]
	public async Task LoadBackendVersionAsync_HandlesBackendUnavailable()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetVersionAsync(Arg.Any<CancellationToken>())
			.Returns<BackendVersion>(_ => throw new BackendUnavailableException("Not running"));

		var vm = new AboutWindowViewModel(mockApi);

		// Act
		await vm.LoadBackendVersionAsync();

		// Assert
		vm.BackendVersion.Should().Be("Not available");
		vm.IsBackendAvailable.Should().BeFalse();
	}

	[Fact]
	public async Task LoadBackendVersionAsync_HandlesGenericError()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetVersionAsync(Arg.Any<CancellationToken>())
			.Returns<BackendVersion>(_ => throw new InvalidOperationException("Unexpected"));

		var vm = new AboutWindowViewModel(mockApi);

		// Act
		await vm.LoadBackendVersionAsync();

		// Assert
		vm.BackendVersion.Should().Be("Error loading version");
		vm.IsBackendAvailable.Should().BeFalse();
	}
}
