using System.Reactive.Disposables;
using System.Reactive.Linq;
using System.Reflection;
using System.Xml;
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
	public EditorView()
	{
		InitializeComponent();

		// Load ARM assembly syntax highlighting
		LoadSyntaxHighlighting();

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

			// TODO: Bind breakpoints and PC indicator when custom gutter is implemented
		});
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
