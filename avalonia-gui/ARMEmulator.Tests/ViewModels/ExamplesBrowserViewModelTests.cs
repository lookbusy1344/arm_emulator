using System.Reactive.Linq;
using System.Reactive.Threading.Tasks;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using ReactiveUI.Reactive;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for ExamplesBrowserViewModel.
/// </summary>
public sealed class ExamplesBrowserViewModelTests
{
	private static readonly TimeSpan FilterTimeout = TimeSpan.FromSeconds(10);

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
		var filtered = vm.WhenAnyValue(x => x.FilteredExamples).Skip(1).FirstAsync().ToTask(TestContext.Current.CancellationToken);
		vm.SearchText = "fib";
		_ = await filtered.WaitAsync(FilterTimeout, TestContext.Current.CancellationToken); // The search is throttled

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
		var filtered = vm.WhenAnyValue(x => x.FilteredExamples).Skip(1).FirstAsync().ToTask(TestContext.Current.CancellationToken);
		vm.SearchText = "HELLO";
		_ = await filtered.WaitAsync(FilterTimeout, TestContext.Current.CancellationToken); // The search is throttled

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
		var filtered = vm.WhenAnyValue(x => x.FilteredExamples).Skip(1).FirstAsync().ToTask(TestContext.Current.CancellationToken);
		vm.SearchText = "loop";
		_ = await filtered.WaitAsync(FilterTimeout, TestContext.Current.CancellationToken); // The search is throttled

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
		var previewLoaded = vm.WhenAnyValue(x => x.PreviewContent).Where(content => content.Contains("MOV R0, #42", StringComparison.Ordinal)).FirstAsync().ToTask(TestContext.Current.CancellationToken);
		vm.SelectedExample = vm.Examples[0];
		_ = await previewLoaded.WaitAsync(FilterTimeout, TestContext.Current.CancellationToken); // The preview load is debounced

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
