using System.Reactive.Disposables;
using System.Reactive.Linq;
using System.Reflection;
using System.Xml;
using ARMEmulator.Controls;
using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.ReactiveUI;
using AvaloniaEdit.Highlighting;
using AvaloniaEdit.Highlighting.Xshd;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and WhenActivated, which triggers IL2026 warnings
// This is acceptable since we don't use AOT compilation for this project
#pragma warning disable IL2026

namespace ARMEmulator.Views;

public partial class EditorView : ReactiveUserControl<MainWindowViewModel>
{
	private EditorGutterMargin? _gutterMargin;

	public EditorView()
	{
		InitializeComponent();

		// Load ARM assembly syntax highlighting
		LoadSyntaxHighlighting();

		// Add custom gutter margin for breakpoints and PC indicator
		_gutterMargin = new EditorGutterMargin();
		TextEditor.TextArea.LeftMargins.Insert(0, _gutterMargin);

		_ = this.WhenActivated(disposables => {
			// Bind ViewModel.SourceCode to TextEditor.Text
			_ = this.WhenAnyValue(x => x.ViewModel!.SourceCode)
				.Where(text => text != TextEditor.Text)  // Avoid feedback loop
				.Subscribe(text => TextEditor.Text = text ?? "")
				.DisposeWith(disposables);

			// Bind TextEditor.Text changes back to ViewModel
			_ = Observable.FromEventPattern(
					handler => TextEditor.TextChanged += handler,
					handler => TextEditor.TextChanged -= handler)
				.Select(_ => TextEditor.Text)
				.Where(text => text != ViewModel?.SourceCode)  // Avoid feedback loop
				.Subscribe(text => {
					if (ViewModel is not null) {
						ViewModel.SourceCode = text ?? "";
					}
				})
				.DisposeWith(disposables);

			// Bind breakpoints (address-based) to gutter (line-based)
			_ = this.WhenAnyValue(
					x => x.ViewModel!.Breakpoints,
					x => x.ViewModel!.AddressToLine,
					(breakpoints, addressToLine) => ConvertBreakpointsToLines(breakpoints, addressToLine))
				.Subscribe(lines => _gutterMargin.BreakpointLines = lines)
				.DisposeWith(disposables);

			// Bind PC (address-based) to gutter (line-based)
			_ = this.WhenAnyValue(
					x => x.ViewModel!.Registers.PC,
					x => x.ViewModel!.AddressToLine,
					(pc, addressToLine) => addressToLine.TryGetValue(pc, out var line) ? line : (int?)null)
				.Subscribe(line => _gutterMargin.CurrentPCLine = line)
				.DisposeWith(disposables);

			// Handle gutter clicks to toggle breakpoints
			_gutterMargin.LineClicked += OnGutterLineClicked;
			_ = Disposable.Create(() => _gutterMargin.LineClicked -= OnGutterLineClicked).DisposeWith(disposables);
		});
	}

	private static System.Collections.Immutable.ImmutableHashSet<int> ConvertBreakpointsToLines(
		System.Collections.Immutable.ImmutableHashSet<uint> breakpoints,
		System.Collections.Immutable.ImmutableDictionary<uint, int> addressToLine)
	{
		return breakpoints
			.Where(addressToLine.ContainsKey)
			.Select(addr => addressToLine[addr])
			.ToImmutableHashSet();
	}

	// Justification: Event handler requires async void signature
#pragma warning disable VSTHRD100
	private async void OnGutterLineClicked(object? sender, int lineNumber)
#pragma warning restore VSTHRD100
	{
		if (ViewModel is null) {
			return;
		}

		// Convert line number to address
		if (!ViewModel.LineToAddress.TryGetValue(lineNumber, out var address)) {
			return;
		}

		// Toggle breakpoint
		if (ViewModel.Breakpoints.Contains(address)) {
			await ViewModel.RemoveBreakpointAsync(address);
		} else {
			await ViewModel.AddBreakpointAsync(address);
		}
	}

	/// <summary>
	/// Loads the ARM assembly syntax highlighting definition from embedded resources.
	/// </summary>
	private void LoadSyntaxHighlighting()
	{
		try {
			var assembly = Assembly.GetExecutingAssembly();
			var resourceName = "ARMEmulator.Resources.ARMAssembly.xshd";

			using var stream = assembly.GetManifestResourceStream(resourceName);
			if (stream is null) {
				return;  // Silently fail if resource not found
			}

			using var reader = new XmlTextReader(stream);
			var definition = HighlightingLoader.Load(reader, HighlightingManager.Instance);
			TextEditor.SyntaxHighlighting = definition;
		}
		catch {
			// Silently ignore syntax highlighting errors - editor still works without it
		}
	}
}
