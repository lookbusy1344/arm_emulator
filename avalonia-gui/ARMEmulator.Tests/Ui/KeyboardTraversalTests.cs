using ARMEmulator.Models;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Headless;
using Avalonia.Input;
using Avalonia.Threading;
using Avalonia.VisualTree;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Keyboard-only users can reach every region of the main window.
/// </summary>
public sealed class KeyboardTraversalTests
{
	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	private static string FocusedName(MainWindowHarness ui) =>
		ui.Window.FocusManager?.GetFocusedElement() is Control control ? control.Name ?? control.GetType().Name : "none";

	private static void Press(MainWindowHarness ui, PhysicalKey key, RawInputModifiers modifiers = RawInputModifiers.None)
	{
		ui.Window.KeyPressQwerty(key, modifiers);
		Settle(ui);
	}

	[Fact]
	public Task F6_CyclesThroughEditorInspectorConsoleAndToolbar() =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);
			ui.Find<AvaloniaEdit.TextEditor>("TextEditor").TextArea.Focus();
			Settle(ui);

			var visited = Enumerable.Range(0, 4).Select(_ => {
				Press(ui, PhysicalKey.F6);
				return FocusedName(ui);
			}).ToList();

			visited.Should().Equal("InspectorSelector", "InputBox", "LoadButton", "TextArea");
		});

	[Fact]
	public Task ShiftF6_CyclesBackwards() =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);
			ui.Find<AvaloniaEdit.TextEditor>("TextEditor").TextArea.Focus();
			Settle(ui);

			Press(ui, PhysicalKey.F6, RawInputModifiers.Shift);

			FocusedName(ui).Should().Be("LoadButton");
		});

	[Fact]
	public Task Tab_InTheEditor_StillInsertsATab() =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			var editor = ui.Find<AvaloniaEdit.TextEditor>("TextEditor");
			editor.Text = "";
			editor.TextArea.Focus();
			Settle(ui);

			Press(ui, PhysicalKey.Tab);

			editor.Text.Should().Be("\t");
		});

	[Theory]
	[InlineData(InspectorPanel.Memory, typeof(Views.MemoryView))]
	[InlineData(InspectorPanel.Stack, typeof(Views.StackView))]
	[InlineData(InspectorPanel.Disassembly, typeof(Views.DisassemblyView))]
	[InlineData(InspectorPanel.Evaluator, typeof(Views.ExpressionEvaluatorView))]
	[InlineData(InspectorPanel.Watchpoints, typeof(Views.WatchpointsView))]
	[InlineData(InspectorPanel.Breakpoints, typeof(Views.BreakpointsListView))]
	public Task Tab_FromTheInspectorSelector_EntersThePanel(InspectorPanel panel, Type view) =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			ui.ViewModel.Breakpoints = [0x8000];
			ui.ViewModel.SelectedInspectorPanel = panel;
			Settle(ui);
			ui.Find<ComboBox>("InspectorSelector").Focus();

			Press(ui, PhysicalKey.Tab);

			var focused = ui.Window.FocusManager?.GetFocusedElement() as Visual;
			focused.Should().NotBeNull();
			focused!.GetSelfAndVisualAncestors().Should().Contain(ancestor => ancestor.GetType() == view);
		});
}
