using ARMEmulator.Models;
using Avalonia.Controls;
using Avalonia.Media;
using Avalonia.Styling;
using Avalonia.Threading;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Look of the console panel. Expected colours are the token values in <c>Themes/Tokens.axaml</c>.
/// </summary>
public sealed class ConsoleAppearanceTests
{
	private const string DataFontName = "JetBrains Mono";
	private const string WaitingClass = "waiting";
	private static readonly Color LightSurface = Color.Parse("#FFFFFF");
	private static readonly Color DarkSurface = Color.Parse("#1E1E1E");
	private static readonly Color LightText = Color.Parse("#1F1F1F");
	private static readonly Color DarkText = Color.Parse("#D4D4D4");

	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	private static Color ColourOf(IBrush? brush) => ((ISolidColorBrush)brush!).Color;

	[Fact]
	public Task Surface_IsTerminalColouredInEachTheme() =>
		UiTest.Run(ui => {
			ui.Window.RequestedThemeVariant = ThemeVariant.Light;
			Settle(ui);
			ColourOf(ui.Find<Border>("ConsoleSurface").Background).Should().Be(LightSurface);

			ui.Window.RequestedThemeVariant = ThemeVariant.Dark;
			Settle(ui);
			ColourOf(ui.Find<Border>("ConsoleSurface").Background).Should().Be(DarkSurface);
		});

	[Fact]
	public Task Output_UsesTheBundledMonospaceFontAndConsoleTextColour() =>
		UiTest.Run(ui => {
			var output = ui.Find<TextBox>("OutputBox");
			output.FontFamily.Name.Should().Be(DataFontName);

			ui.Window.RequestedThemeVariant = ThemeVariant.Light;
			Settle(ui);
			ColourOf(output.Foreground).Should().Be(LightText);

			ui.Window.RequestedThemeVariant = ThemeVariant.Dark;
			Settle(ui);
			ColourOf(output.Foreground).Should().Be(DarkText);
		});

	[Fact]
	public Task InputRow_HasAPromptMarker() =>
		UiTest.Run(ui => ui.Find<TextBlock>("PromptMarker").Text.Should().Be(">"));

	[Fact]
	public Task InputRow_IsHighlightedOnlyWhileTheProgramWaitsForInput() =>
		UiTest.Run(ui => {
			ui.Find<Border>("InputRow").Classes.Should().NotContain(WaitingClass);

			ui.ViewModel.Status = VMState.WaitingForInput;
			Settle(ui);
			ui.Find<Border>("InputRow").Classes.Should().Contain(WaitingClass);

			ui.ViewModel.Status = VMState.Idle;
			Settle(ui);
			ui.Find<Border>("InputRow").Classes.Should().NotContain(WaitingClass);
		});

	[Fact]
	public Task Surface_IsNotOutlinedWhileWaiting() =>
		UiTest.Run(ui => {
			ui.ViewModel.Status = VMState.WaitingForInput;
			Settle(ui);

			ui.Find<Border>("ConsoleSurface").BorderThickness.Should().Be(default(Avalonia.Thickness));
		});

	[Fact]
	public Task WaitingPill_ShowsOnlyWhileWaiting() =>
		UiTest.Run(ui => {
			ui.Find<Border>("WaitingPill").IsVisible.Should().BeFalse();

			ui.ViewModel.Status = VMState.WaitingForInput;
			Settle(ui);

			ui.Find<Border>("WaitingPill").IsVisible.Should().BeTrue();
		});
}
