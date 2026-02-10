using System.Reactive.Disposables;
using System.Reactive.Linq;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// ViewModel for the Examples Browser window.
/// Provides searchable list of example programs with preview.
/// </summary>
public sealed class ExamplesBrowserViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient api;
	private readonly CompositeDisposable disposables = [];
	private ImmutableArray<ExampleInfo> examples = [];
	private ImmutableArray<ExampleInfo> filteredExamples = [];
	private ExampleInfo? selectedExample;
	private string searchText = "";
	private string previewContent = "";
	private string? errorMessage;
	private bool isLoading;

	public ExamplesBrowserViewModel(IApiClient api)
	{
		this.api = api;

		// Update filtered examples when search text or examples change
		_ = this.WhenAnyValue(x => x.SearchText, x => x.Examples)
			.Throttle(TimeSpan.FromMilliseconds(300))
			.ObserveOn(RxApp.MainThreadScheduler)
			.Subscribe(_ => UpdateFilteredExamples())
			.DisposeWith(disposables);

		// Load preview content when selection changes
		// Fire-and-forget is acceptable - exceptions are handled in LoadPreviewContentAsync
#pragma warning disable VSTHRD101
		_ = this.WhenAnyValue(x => x.SelectedExample)
			.Throttle(TimeSpan.FromMilliseconds(100))
			.ObserveOn(RxApp.MainThreadScheduler)
			.Subscribe(async example => await LoadPreviewContentAsync(example))
			.DisposeWith(disposables);
#pragma warning restore VSTHRD101
	}

	/// <summary>All available examples.</summary>
	public ImmutableArray<ExampleInfo> Examples
	{
		get => examples;
		private set => this.RaiseAndSetIfChanged(ref examples, value);
	}

	/// <summary>Filtered examples based on search text.</summary>
	public ImmutableArray<ExampleInfo> FilteredExamples
	{
		get => filteredExamples;
		private set => this.RaiseAndSetIfChanged(ref filteredExamples, value);
	}

	/// <summary>Currently selected example.</summary>
	public ExampleInfo? SelectedExample
	{
		get => selectedExample;
		set => this.RaiseAndSetIfChanged(ref selectedExample, value);
	}

	/// <summary>Search text for filtering examples.</summary>
	public string SearchText
	{
		get => searchText;
		set => this.RaiseAndSetIfChanged(ref searchText, value);
	}

	/// <summary>Preview content of selected example.</summary>
	public string PreviewContent
	{
		get => previewContent;
		private set => this.RaiseAndSetIfChanged(ref previewContent, value);
	}

	/// <summary>Error message if loading fails.</summary>
	public string? ErrorMessage
	{
		get => errorMessage;
		private set => this.RaiseAndSetIfChanged(ref errorMessage, value);
	}

	/// <summary>Whether examples are currently loading.</summary>
	public bool IsLoading
	{
		get => isLoading;
		private set => this.RaiseAndSetIfChanged(ref isLoading, value);
	}

	/// <summary>
	/// Loads the list of examples from the backend.
	/// </summary>
	public async Task LoadExamplesAsync()
	{
		try {
			IsLoading = true;
			ErrorMessage = null;

			var exampleList = await api.GetExamplesAsync();
			Examples = exampleList;
			FilteredExamples = exampleList;
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to load examples: {ex.Message}";
			Examples = [];
			FilteredExamples = [];
		}
		finally {
			IsLoading = false;
		}
	}

	/// <summary>
	/// Updates filtered examples based on current search text.
	/// </summary>
	private void UpdateFilteredExamples()
	{
		if (string.IsNullOrWhiteSpace(SearchText)) {
			FilteredExamples = Examples;
			return;
		}

		var query = SearchText;
		FilteredExamples = Examples
			.Where(e => e.Name.Contains(query, StringComparison.OrdinalIgnoreCase) ||
						e.Description.Contains(query, StringComparison.OrdinalIgnoreCase))
			.ToImmutableArray();
	}

	/// <summary>
	/// Loads preview content for the selected example.
	/// </summary>
	private async Task LoadPreviewContentAsync(ExampleInfo? example)
	{
		if (example is null) {
			PreviewContent = "";
			return;
		}

		try {
			var content = await api.GetExampleContentAsync(example.Name);
			PreviewContent = content;
		}
		catch (Exception ex) {
			PreviewContent = $"Error loading preview: {ex.Message}";
		}
	}

	public void Dispose() => disposables.Dispose();
}
