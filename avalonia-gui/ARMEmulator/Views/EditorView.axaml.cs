using System.Reactive.Disposables;
using System.Reactive.Disposables.Fluent;
using System.Reactive.Linq;
using System.Reflection;
using System.Windows.Input;
using System.Xml;
using ARMEmulator.Controls;
using ARMEmulator.ViewModels;
using Avalonia.Controls;
using AvaloniaEdit.Highlighting;
using AvaloniaEdit.Highlighting.Xshd;
using ReactiveUI.Avalonia.Reactive;
using ReactiveUI.Reactive;

// ReactiveUI uses reflection for WhenAnyValue and WhenActivated, which triggers IL2026 warnings
// This is acceptable since we don't use AOT compilation for this project
#pragma warning disable IL2026

namespace ARMEmulator.Views;

public partial class EditorView : ReactiveUserControl<MainWindowViewModel>
{
	private EditorGutterMargin? gutterMargin;

	public EditorView()
	{
		InitializeComponent();

		// Load ARM assembly syntax highlighting
		LoadSyntaxHighlighting();

		// Add custom gutter margin for breakpoints and PC indicator
		gutterMargin = new EditorGutterMargin();
		TextEditor.TextArea.LeftMargins.Insert(0, gutterMargin);

		_ = this.WhenActivated(disposables => {
			// Bind ViewModel.SourceCode to TextEditor.Text
			_ = this.WhenAnyValue(x => x.ViewModel!.SourceCode)
				.Where(text => text != TextEditor.Text) // Avoid feedback loop
				.Subscribe(text => TextEditor.Text = text ?? "")
				.DisposeWith(disposables);

			// Bind TextEditor.Text changes back to ViewModel
			_ = Observable.FromEventPattern(
					handler => TextEditor.TextChanged += handler,
					handler => TextEditor.TextChanged -= handler)
				.Select(_ => TextEditor.Text)
				.Where(text => text != ViewModel?.SourceCode) // Avoid feedback loop
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
				.Subscribe(lines => gutterMargin.BreakpointLines = lines)
				.DisposeWith(disposables);

			// Bind PC (address-based) to gutter (line-based). RegisterState is immutable and replaced on change.
			_ = this.WhenAnyValue(
					x => x.ViewModel!.Registers,
					x => x.ViewModel!.AddressToLine,
					(registers, addressToLine) => addressToLine.TryGetValue(registers.PC, out var line) ? line : (int?)null)
				.Subscribe(line => gutterMargin.CurrentPCLine = line)
				.DisposeWith(disposables);

			// Scroll to the PC line when the view model asks
			_ = (ViewModel?.ScrollToLineRequests
				.ObserveOn(RxSchedulers.MainThreadScheduler)
				.Subscribe(ScrollToLineIfHidden)
				.DisposeWith(disposables));

			// Handle gutter clicks to toggle breakpoints
			gutterMargin.LineClicked += OnGutterLineClicked;
			_ = Disposable.Create(() => gutterMargin.LineClicked -= OnGutterLineClicked).DisposeWith(disposables);
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

	private void ScrollToLineIfHidden(int line)
	{
		var textView = TextEditor.TextArea.TextView;
		var isVisible = textView.VisualLinesValid
			&& textView.VisualLines.Count > 0
			&& line >= textView.VisualLines[0].FirstDocumentLine.LineNumber
			&& line <= textView.VisualLines[^1].LastDocumentLine.LineNumber;
		if (!isVisible && line <= TextEditor.Document.LineCount) {
			TextEditor.ScrollTo(line, column: 0);
		}
	}

	private void OnGutterLineClicked(object? sender, int lineNumber) =>
		((ICommand?)ViewModel?.ToggleBreakpointCommand)?.Execute(lineNumber);

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
				return; // Silently fail if resource not found
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
