#pragma warning disable CS0618 // Bitmap.Save(string) is the only overload that needs no encoder options
using ARMEmulator.Models;
using Avalonia.Headless;
using Avalonia.Styling;
using Avalonia.Threading;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Renders the main window with sample data to PNG files for visual review.
/// Set <c>ARM_SCREENSHOT_DIR</c> to write them; the tests are skipped otherwise.
/// </summary>
public sealed class ScreenshotTests
{
	private const string DirectoryVariable = "ARM_SCREENSHOT_DIR";
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

	public static bool WantsScreenshots => Environment.GetEnvironmentVariable(DirectoryVariable) is { Length: > 0 };

	private static readonly ImmutableArray<(uint Address, int Line)> SourceMap = [(0x8000, 3), (0x8004, 4), (0x8008, 6), (0x800C, 7), (0x8010, 8), (0x8014, 9), (0x8018, 10)];

	[Theory(SkipUnless = nameof(WantsScreenshots), Skip = "Set ARM_SCREENSHOT_DIR to render screenshots")]
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
			ui.ViewModel.Status = VMState.Breakpoint;
			ui.Events.OnNext(new OutputEvent(MainWindowHarness.SessionId, OutputStreamType.Stdout, "Sum of 1..5\nResult: 15\n"));

			Dispatcher.UIThread.RunJobs();
			ui.Window.UpdateLayout();
			AvaloniaHeadlessPlatform.ForceRenderTimerTick();
			ui.Window.CaptureRenderedFrame()!.Save(Path.Combine(Environment.GetEnvironmentVariable(DirectoryVariable)!, $"main-{name}.png"));
		});
}
