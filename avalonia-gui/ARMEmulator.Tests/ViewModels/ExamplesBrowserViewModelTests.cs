using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for ExamplesBrowserViewModel.
/// </summary>
public sealed class ExamplesBrowserViewModelTests
{
	[Fact]
	public async Task Constructor_LoadsExamplesAsync()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns([
				new ExampleInfo("hello", "Hello World", 100),
				new ExampleInfo("fibonacci", "Fibonacci", 250)
			]);

		// Act
		using var vm = new ExamplesBrowserViewModel(mockApi);
		await vm.LoadExamplesAsync();

		// Assert
		vm.Examples.Should().HaveCount(2);
		vm.Examples[0].Name.Should().Be("hello");
		vm.Examples[1].Name.Should().Be("fibonacci");
	}

	[Fact]
	public async Task SearchText_FiltersExamples()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns([
				new ExampleInfo("hello", "Hello World", 100),
				new ExampleInfo("fibonacci", "Fibonacci", 250),
				new ExampleInfo("factorial", "Factorial", 200)
			]);

		using var vm = new ExamplesBrowserViewModel(mockApi);
		await vm.LoadExamplesAsync();

		// Act
		vm.SearchText = "fib";
		await Task.Delay(350); // Wait for throttle

		// Assert
		vm.FilteredExamples.Should().HaveCount(1);
		vm.FilteredExamples[0].Name.Should().Be("fibonacci");
	}

	[Fact]
	public async Task SearchText_IsCaseInsensitive()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns([
				new ExampleInfo("hello", "Hello World", 100)
			]);

		using var vm = new ExamplesBrowserViewModel(mockApi);
		await vm.LoadExamplesAsync();

		// Act
		vm.SearchText = "HELLO";
		await Task.Delay(350); // Wait for throttle

		// Assert
		vm.FilteredExamples.Should().HaveCount(1);
	}

	[Fact]
	public async Task SearchText_MatchesNameAndDescription()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns([
				new ExampleInfo("test1", "Example with loops", 100),
				new ExampleInfo("loops", "Loop demonstration", 150)
			]);

		using var vm = new ExamplesBrowserViewModel(mockApi);
		await vm.LoadExamplesAsync();

		// Act
		vm.SearchText = "loop";
		await Task.Delay(350); // Wait for throttle

		// Assert
		vm.FilteredExamples.Should().HaveCount(2);
	}

	[Fact]
	public async Task SelectedExample_LoadsContent()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns([new ExampleInfo("hello", "Hello World", 100)]);
		mockApi.GetExampleContentAsync("hello", Arg.Any<CancellationToken>())
			.Returns(".global _start\n_start:\n    MOV R0, #42\n");

		using var vm = new ExamplesBrowserViewModel(mockApi);
		await vm.LoadExamplesAsync();

		// Act
		vm.SelectedExample = vm.Examples[0];
		await Task.Delay(150); // Wait for debounce

		// Assert
		vm.PreviewContent.Should().Contain("MOV R0, #42");
	}

	[Fact]
	public async Task LoadExamplesAsync_HandlesError()
	{
		// Arrange
		var mockApi = Substitute.For<IApiClient>();
		mockApi.GetExamplesAsync(Arg.Any<CancellationToken>())
			.Returns<ImmutableArray<ExampleInfo>>(_ => throw new ApiException("Network error"));

		using var vm = new ExamplesBrowserViewModel(mockApi);

		// Act
		await vm.LoadExamplesAsync();

		// Assert
		vm.Examples.Should().BeEmpty();
		vm.ErrorMessage.Should().Contain("Failed to load");
	}
}
