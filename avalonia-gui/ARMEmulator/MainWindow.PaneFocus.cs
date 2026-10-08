using Avalonia.Controls;
using Avalonia.Input;
using Avalonia.VisualTree;
using AvaloniaEdit;

namespace ARMEmulator;

public partial class MainWindow
{
	/// <summary>Controls that take focus in turn when F6 moves between regions.</summary>
	private static readonly ImmutableArray<string> PaneNames = ["LoadButton", "TextEditor", "InspectorSelector", "InputBox"];

	/// <summary>Moves focus between regions with F6, and back with Shift+F6. Tab cannot leave the editor, which inserts tabs.</summary>
	protected override void OnKeyDown(KeyEventArgs e)
	{
		base.OnKeyDown(e);
		if (e.Handled || e.Key != Key.F6) {
			return;
		}

		var panes = PaneNames.Select(FindPane).OfType<InputElement>().Where(pane => pane.IsEffectivelyVisible).ToList();
		if (panes.Count == 0) {
			return;
		}

		var forward = !e.KeyModifiers.HasFlag(KeyModifiers.Shift);
		_ = panes[NextPaneIndex(panes.FindIndex(pane => pane.IsKeyboardFocusWithin), panes.Count, forward)].Focus();
		e.Handled = true;
	}

	/// <summary>The pane after (or before) the focused one; the first (or last) when none has focus.</summary>
	internal static int NextPaneIndex(int current, int count, bool forward)
	{
		if (current < 0) {
			return forward ? 0 : count - 1;
		}

		return (current + (forward ? 1 : -1) + count) % count;
	}

	private InputElement? FindPane(string name)
	{
		var control = this.GetVisualDescendants().OfType<Control>().FirstOrDefault(candidate => candidate.Name == name);
		return control is TextEditor editor ? editor.TextArea : control;
	}
}
