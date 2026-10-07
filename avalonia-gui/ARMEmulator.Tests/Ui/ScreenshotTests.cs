using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia.Headless;
using Avalonia.Styling;
using Avalonia.Threading;
using NSubstitute;
using NSubstitute.ExceptionExtensions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Renders the main window with sample data to PNG files for visual review.
/// They write PNGs for review (<c>ARM_SCREENSHOT_DIR</c>), refresh the reviewed baselines (<c>ARM_UPDATE_BASELINES=1</c>) or compare against them; see <see cref="ScreenshotBaselines"/>.
/// </summary>
public sealed class ScreenshotTests
{
	private const double WindowWidth = 1200;
	private const double WindowHeight = 800;

	private const string Program = """
		; sum of 1..5
		_start:
		    MOV R0, #0
		    MOV R1, #5
		loop:
		    ADD R0, R0, R1
		    SUBS R1, R1, #1
		    BNE loop
		    MOV R7, #1
		    SWI #0
		""";

	public static bool WantsScreenshots => ScreenshotBaselines.IsActive;

	private static readonly ImmutableArray<(uint Address, int Line)> SourceMap = [(0x8000, 3), (0x8004, 4), (0x8008, 6), (0x800C, 7), (0x8010, 8), (0x8014, 9), (0x8018, 10)];

	[Theory(SkipUnless = nameof(WantsScreenshots), Skip = "Set ARM_SCREENSHOT_DIR or ARM_UPDATE_BASELINES, or add baselines for this platform")]
	[InlineData("light", false)]
	[InlineData("dark", true)]
	public Task MainWindow_RendersWithSampleData(string name, bool dark) =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			ui.Window.Width = WindowWidth;
			ui.Window.Height = WindowHeight;
			ui.Window.RequestedThemeVariant = dark ? ThemeVariant.Dark : ThemeVariant.Light;

			ui.ViewModel.SourceCode = Program;
			ui.ViewModel.AddressToLine = SourceMap.ToImmutableDictionary(entry => entry.Address, entry => entry.Line);
			ui.ViewModel.LineToAddress = SourceMap.ToImmutableDictionary(entry => entry.Line, entry => entry.Address);
			ui.ViewModel.Breakpoints = [0x8008];
			ui.ViewModel.UpdateRegisters(RegisterState.Create(r0: 0, r1: 5, sp: 0x50000, pc: 0x8004));
			ui.ViewModel.UpdateRegisters(RegisterState.Create(r0: 0, r1: 5, r2: 0xDEADBEEF, sp: 0x50000, lr: 0x8020, pc: 0x8008, cpsr: new CPSRFlags(N: false, Z: true, C: true, V: false)));
			ui.ViewModel.Status = dark ? VMState.WaitingForInput : VMState.Breakpoint;
			ui.Events.OnNext(new OutputEvent(MainWindowHarness.SessionId, OutputStreamType.Stdout, "Sum of 1..5\nResult: 15\n"));

			Dispatcher.UIThread.RunJobs();
			ui.Window.UpdateLayout();
			AvaloniaHeadlessPlatform.ForceRenderTimerTick();
			ScreenshotBaselines.Check($"main-{name}", ui.Window.CaptureRenderedFrame()!);
		});

	[Theory(SkipUnless = nameof(WantsScreenshots), Skip = "Set ARM_SCREENSHOT_DIR or ARM_UPDATE_BASELINES, or add baselines for this platform")]
	[InlineData(InspectorPanel.Memory, false)]
	[InlineData(InspectorPanel.Stack, false)]
	[InlineData(InspectorPanel.Disassembly, false)]
	[InlineData(InspectorPanel.Memory, true)]
	[InlineData(InspectorPanel.Stack, true)]
	[InlineData(InspectorPanel.Disassembly, true)]
	public Task Panel_RendersWithSampleData(InspectorPanel panel, bool dark) =>
		UiTest.RunAsync(async ui => {
			var ct = TestContext.Current.CancellationToken;
			await ui.ViewModel.StartAsync(ui.Backend, ct);
			ui.Window.Width = WindowWidth;
			ui.Window.Height = WindowHeight;
			ui.Window.RequestedThemeVariant = dark ? ThemeVariant.Dark : ThemeVariant.Light;
			ui.Api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(
				Enumerable.Range(0, 128).Select(i => (byte)(0x40 + (i % 0x30))).ToImmutableArray());
			ui.Api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(
				SourceMap.Select(entry => new DisassemblyInstruction(entry.Address, 0xE2800001, "ADD R0, R0, #1", entry.Address == 0x8008 ? "loop" : null)).ToImmutableArray());
			ui.ViewModel.Breakpoints = [0x8008];
			ui.ViewModel.UpdateRegisters(RegisterState.Create(r0: 0x8000, sp: 0x50000, pc: 0x8008));
			await ui.ViewModel.Memory.LoadMemoryAsync(0x8000);
			ui.ViewModel.Memory.LastWriteAddress = 0x8013;
			ui.ViewModel.SelectedInspectorPanel = panel;

			Dispatcher.UIThread.RunJobs();
			ui.Window.UpdateLayout();
			AvaloniaHeadlessPlatform.ForceRenderTimerTick();
			ScreenshotBaselines.Check($"{panel}-{(dark ? "dark" : "light")}", ui.Window.CaptureRenderedFrame()!);
		});

	[Theory(SkipUnless = nameof(WantsScreenshots), Skip = "Set ARM_SCREENSHOT_DIR or ARM_UPDATE_BASELINES, or add baselines for this platform")]
	[InlineData("preferences", false)]
	[InlineData("preferences", true)]
	[InlineData("about", false)]
	[InlineData("unsaved", true)]
	[InlineData("examples", false)]
	public Task Dialog_Renders(string name, bool dark) =>
		UiTest.RunAsync(async _ => {
			Avalonia.Controls.Window dialog = name switch {
				"preferences" => new Views.PreferencesWindow(AppSettings.Default),
				"about" => new Views.AboutWindow(),
				"unsaved" => new Views.UnsavedChangesWindow("prog.s"),
				_ => new Views.ExamplesBrowserWindow(await FailedExamplesViewModelAsync())
			};
			dialog.RequestedThemeVariant = dark ? ThemeVariant.Dark : ThemeVariant.Light;
			dialog.Show();
			Dispatcher.UIThread.RunJobs();
			dialog.UpdateLayout();
			AvaloniaHeadlessPlatform.ForceRenderTimerTick();
			ScreenshotBaselines.Check($"dialog-{name}-{(dark ? "dark" : "light")}", dialog.CaptureRenderedFrame()!);
			dialog.Close();
		});

	[Theory(SkipUnless = nameof(WantsScreenshots), Skip = "Set ARM_SCREENSHOT_DIR or ARM_UPDATE_BASELINES, or add baselines for this platform")]
	[InlineData("connecting")]
	[InlineData("failed")]
	[InlineData("empty")]
	public Task StartState_Renders(string name) =>
		UiTest.RunAsync(async ui => {
			ui.Window.Width = WindowWidth;
			ui.Window.Height = WindowHeight;
			if (name == "failed") {
				ui.Backend.StartAsync(default).ThrowsAsyncForAnyArgs(new BackendStartException("backend binary not found next to the application"));
			}

			if (name != "connecting") {
				await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			}

			Dispatcher.UIThread.RunJobs();
			ui.Window.UpdateLayout();
			AvaloniaHeadlessPlatform.ForceRenderTimerTick();
			ScreenshotBaselines.Check($"state-{name}", ui.Window.CaptureRenderedFrame()!);
		});

	/// <summary>A view model whose load failed: the error state is set at once, where the list and preview settle on timers.</summary>
	private static async Task<ExamplesBrowserViewModel> FailedExamplesViewModelAsync()
	{
		var api = Substitute.For<IApiClient>();
		api.GetExamplesAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new ApiException("connection refused"));
		var viewModel = new ExamplesBrowserViewModel(api);
		await viewModel.LoadExamplesAsync();
		return viewModel;
	}
}
