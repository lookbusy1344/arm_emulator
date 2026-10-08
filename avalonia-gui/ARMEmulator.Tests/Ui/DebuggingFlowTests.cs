using System.Windows.Input;
using ARMEmulator.Controls;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Headless;
using Avalonia.Input;
using Avalonia.Media;
using Avalonia.Threading;
using Avalonia.VisualTree;
using AvaloniaEdit;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// The main window driven through its controls, with the backend mocked.
/// </summary>
public sealed class DebuggingFlowTests
{
	private const string SessionId = MainWindowHarness.SessionId;
	private const uint FirstAddress = 0x8000;
	private const uint SecondAddress = 0x8004;
	private const int FirstLine = 1;
	private const int SecondLine = 2;

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	/// <summary>Starts a session with a two-instruction program assembled, then clears the API call log.</summary>
	private static async Task StartSessionAsync(MainWindowHarness ui)
	{
		await ui.ViewModel.StartAsync(ui.Backend, Ct);
		ui.Api.GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>())
			.Returns([new SourceMapEntry(FirstAddress, FirstLine, "MOV R0, #5"), new SourceMapEntry(SecondAddress, SecondLine, "MOV R1, #6")]);
		ui.ViewModel.SourceCode = "MOV R0, #5\nMOV R1, #6\n";
		Execute(ui.ViewModel.AssembleCommand);
		Settle(ui);
		ui.Api.ClearReceivedCalls();
	}

	/// <summary>Runs pending UI work and a layout pass so bindings and visual lines are current.</summary>
	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	/// <summary>Runs a command as a binding would. Command pipelines complete on the UI thread, so assertions stay on it.</summary>
	private static void Execute(ICommand command) => command.Execute(null);

	private static EditorGutterMargin GutterOf(MainWindowHarness ui) =>
		ui.Find<TextEditor>("TextEditor").TextArea.LeftMargins.OfType<EditorGutterMargin>().Single();

	private static Border RegisterRow(MainWindowHarness ui, string register) =>
		ui.Find<ItemsControl>("RegisterRows").ContainerFromIndex(ui.ViewModel.RegisterRows.ToList().FindIndex(row => row.Name == register))!
			.GetVisualDescendants().OfType<Border>().First(border => border.Classes.Contains("registerRow"));

	private static string RegisterHex(MainWindowHarness ui, string register) =>
		RegisterRow(ui, register).GetVisualDescendants().OfType<TextBlock>().Single(text => text.Classes.Contains("hex")).Text!;

	// Assembler errors

	[Fact]
	public Task LoadWithAssemblerErrors_ShowsEveryMessage() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			ui.Api.LoadProgramAsync(SessionId, Arg.Any<string>(), Arg.Any<CancellationToken>())
				.ThrowsAsync(new ProgramLoadException(["line 2: unknown instruction 'MOVV'", "line 5: undefined label 'done'"]));

			Execute(ui.ViewModel.AssembleCommand);
			Settle(ui);

			ui.Find<Border>("ErrorBar").IsVisible.Should().BeTrue();
			ui.Find<SelectableTextBlock>("ErrorText").Text.Should().Be(
				"Failed to assemble program:\nline 2: unknown instruction 'MOVV'\nline 5: undefined label 'done'");
		});

	// Breakpoints

	[Fact]
	public Task F9_TogglesABreakpointMarkerOnTheCaretLine() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			var editor = ui.Find<TextEditor>("TextEditor");
			editor.TextArea.Focus();
			editor.TextArea.Caret.Line = SecondLine;
			editor.TextArea.IsFocused.Should().BeTrue();

			ui.Window.KeyPressQwerty(PhysicalKey.F9, RawInputModifiers.None);
			Settle(ui);

			await ui.Api.Received(1).AddBreakpointAsync(SessionId, SecondAddress, Arg.Any<CancellationToken>());
			GutterOf(ui).BreakpointLines.Should().BeEquivalentTo([SecondLine]);

			ui.Window.KeyPressQwerty(PhysicalKey.F9, RawInputModifiers.None);
			Settle(ui);

			await ui.Api.Received(1).RemoveBreakpointAsync(SessionId, SecondAddress, Arg.Any<CancellationToken>());
			GutterOf(ui).BreakpointLines.Should().BeEmpty();
		});

	[Fact]
	public Task GutterClick_TogglesABreakpointMarker() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			var gutter = GutterOf(ui);
			var insideFirstLine = gutter.TranslatePoint(new Point(gutter.Bounds.Width / 2, 2), ui.Window)!.Value;

			ui.Window.MouseDown(insideFirstLine, MouseButton.Left);
			ui.Window.MouseUp(insideFirstLine, MouseButton.Left);
			Settle(ui);

			await ui.Api.Received(1).AddBreakpointAsync(SessionId, FirstAddress, Arg.Any<CancellationToken>());
			gutter.BreakpointLines.Should().BeEquivalentTo([FirstLine]);
		});

	// Stepping

	[Fact]
	public Task Step_UpdatesRegistersHighlightsChangesAndMovesThePcMarker() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: FirstAddress));
			ui.Api.StepAsync(SessionId, Arg.Any<CancellationToken>())
				.Returns(RegisterState.Create(r0: 5, pc: SecondAddress));
			Settle(ui);
			GutterOf(ui).CurrentPCLine.Should().Be(FirstLine);

			Execute(ui.ViewModel.StepCommand);
			Settle(ui);

			RegisterHex(ui, "R0").Should().Be("0x00000005");
			RegisterHex(ui, "PC").Should().Be("0x00008004");
			RegisterRow(ui, "R0").Classes.Should().Contain("changed");
			RegisterRow(ui, "R1").Classes.Should().NotContain("changed");
			GutterOf(ui).CurrentPCLine.Should().Be(SecondLine);
		});

	// Console

	[Fact]
	public Task ConsoleInput_IsSentAndTheFieldCleared() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			ui.Api.StepAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: SecondAddress));
			var input = ui.Find<TextBox>("InputBox");

			input.Text = "42";
			Settle(ui);
			input.Focus();
			ui.Window.KeyPressQwerty(PhysicalKey.Enter, RawInputModifiers.None);
			Settle(ui);

			await ui.Api.Received(1).SendStdinAsync(SessionId, "42\n", Arg.Any<CancellationToken>());
			input.Text.Should().BeEmpty();
			ui.ViewModel.InputText.Should().BeEmpty();
		});

	[Fact]
	public Task ConsoleOutput_ShowsProgramOutput() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);

			ui.Events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "Hello\n"));
			Settle(ui);

			ui.Find<TextBox>("OutputBox").Text.Should().Be("Hello\n");
		});

	// Panels

	[Fact]
	public Task MemoryPanel_ShowsRowsFromItsViewModel() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			ui.Api.GetMemoryAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
				.Returns(Enumerable.Range(0, 32).Select(i => (byte)i).ToImmutableArray());
			SelectTab(ui, InspectorPanel.Memory);

			await ui.ViewModel.Memory.LoadMemoryAsync(0x8000);
			Settle(ui);

			ui.Find<ItemsControl>("MemoryRows").ItemCount.Should().Be(2);
		});

	[Fact]
	public Task StackPanel_ShowsEntriesFromItsViewModel() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			const uint stackPointer = 0x4FFF0;
			ui.Api.GetMemoryAsync(SessionId, stackPointer, Arg.Any<int>(), Arg.Any<CancellationToken>())
				.Returns(ImmutableArray.Create<byte>(1, 0, 0, 0, 2, 0, 0, 0));
			SelectTab(ui, InspectorPanel.Stack);

			ui.ViewModel.UpdateRegisters(RegisterState.Create(sp: stackPointer));
			Settle(ui);

			ui.Find<ItemsControl>("StackRows").ItemCount.Should().Be(2);
		});

	[Fact]
	public Task DisassemblyPanel_ShowsInstructionsFromItsViewModel() =>
		UiTest.RunAsync(async ui => {
			await StartSessionAsync(ui);
			ui.Api.GetDisassemblyAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
				.Returns(ImmutableArray.Create(
					new DisassemblyInstruction(FirstAddress, 0xE3A00005, "MOV R0, #5", null),
					new DisassemblyInstruction(SecondAddress, 0xE3A01006, "MOV R1, #6", null)));
			SelectTab(ui, InspectorPanel.Disassembly);

			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: FirstAddress));
			Settle(ui);

			ui.Find<ItemsControl>("DisassemblyRows").ItemCount.Should().Be(2);
		});

	private static void SelectTab(MainWindowHarness ui, InspectorPanel panel)
	{
		ui.Find<ComboBox>("InspectorSelector").SelectedValue = panel;
		Settle(ui);
	}
}
