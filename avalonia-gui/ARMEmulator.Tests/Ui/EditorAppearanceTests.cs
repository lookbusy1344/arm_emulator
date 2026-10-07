using ARMEmulator.Controls;
using ARMEmulator.Models;
using Avalonia.Controls;
using Avalonia.Media;
using Avalonia.Styling;
using Avalonia.Threading;
using AvaloniaEdit;
using AvaloniaEdit.Highlighting;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Font, colours and line highlights of the source editor. Expected colours are the token values in <c>Themes/Tokens.axaml</c>.
/// </summary>
public sealed class EditorAppearanceTests
{
	private const string DataFontName = "JetBrains Mono";
	private const int LargeFontSize = 18;
	private static readonly Color LightComment = Color.Parse("#008000");
	private static readonly Color DarkComment = Color.Parse("#6A9955");
	private static readonly Color LightCurrentLine = Color.Parse("#0F000000");
	private static readonly Color DarkCurrentLine = Color.Parse("#14FFFFFF");
	private static readonly Color LightSecondaryText = Color.Parse("#6B6B6B");
	private static readonly Color DarkSecondaryText = Color.Parse("#A0A0A0");

	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	private static void UseTheme(MainWindowHarness ui, ThemeVariant variant)
	{
		ui.Window.RequestedThemeVariant = variant;
		Settle(ui);
	}

	private static TextEditor Editor(MainWindowHarness ui) => ui.Find<TextEditor>("TextEditor");

	private static Color ColourOf(IBrush? brush) => ((ISolidColorBrush)brush!).Color;

	private static Color CommentColour(TextEditor editor) =>
		((ISolidColorBrush)editor.SyntaxHighlighting.GetNamedColor("Comment").Foreground!.GetBrush(null)!).Color;

	[Fact]
	public Task Editor_UsesTheBundledMonospaceFont() =>
		UiTest.Run(ui => Editor(ui).FontFamily.Name.Should().Be(DataFontName));

	[Fact]
	public Task Editor_FontSizeFollowsTheSetting() =>
		UiTest.Run(ui => {
			ui.ViewModel.ApplySettings(AppSettings.Default with { EditorFontSize = LargeFontSize });
			Settle(ui);

			Editor(ui).FontSize.Should().Be(LargeFontSize);
		});

	[Fact]
	public Task SyntaxColours_FollowTheTheme() =>
		UiTest.Run(ui => {
			UseTheme(ui, ThemeVariant.Light);
			CommentColour(Editor(ui)).Should().Be(LightComment);

			UseTheme(ui, ThemeVariant.Dark);
			CommentColour(Editor(ui)).Should().Be(DarkComment);
		});

	[Fact]
	public Task CurrentLineBackground_FollowsTheTheme() =>
		UiTest.Run(ui => {
			UseTheme(ui, ThemeVariant.Light);
			ColourOf(Editor(ui).TextArea.TextView.CurrentLineBackground).Should().Be(LightCurrentLine);

			UseTheme(ui, ThemeVariant.Dark);
			ColourOf(Editor(ui).TextArea.TextView.CurrentLineBackground).Should().Be(DarkCurrentLine);
		});

	[Fact]
	public Task LineNumbers_UseTheSecondaryTextColour() =>
		UiTest.Run(ui => {
			UseTheme(ui, ThemeVariant.Light);
			ColourOf(Editor(ui).LineNumbersForeground).Should().Be(LightSecondaryText);

			UseTheme(ui, ThemeVariant.Dark);
			ColourOf(Editor(ui).LineNumbersForeground).Should().Be(DarkSecondaryText);
		});

	[Fact]
	public Task CurrentLineHighlight_IsOn() =>
		UiTest.Run(ui => Editor(ui).Options.HighlightCurrentLine.Should().BeTrue());

	[Fact]
	public Task PcLineRenderer_FollowsTheProgramCounterLine() =>
		UiTest.Run(ui => {
			const uint address = 0x8004;
			const int line = 2;
			ui.ViewModel.AddressToLine = ImmutableDictionary<uint, int>.Empty.Add(address, line);
			var renderer = Editor(ui).TextArea.TextView.BackgroundRenderers.OfType<PcLineBackgroundRenderer>().Single();
			renderer.Line.Should().BeNull();

			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: address));
			Settle(ui);

			renderer.Line.Should().Be(line);
		});

	[Fact]
	public Task PcLineRenderer_HasNoLineWhenThePcIsUnmapped() =>
		UiTest.Run(ui => {
			ui.ViewModel.AddressToLine = ImmutableDictionary<uint, int>.Empty.Add(0x8004, 2);
			var renderer = Editor(ui).TextArea.TextView.BackgroundRenderers.OfType<PcLineBackgroundRenderer>().Single();

			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: 0x9000));
			Settle(ui);

			renderer.Line.Should().BeNull();
		});
}
