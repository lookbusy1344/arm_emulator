namespace ARMEmulator.Tests;

using System.Text.RegularExpressions;
using AwesomeAssertions;

/// <summary>
///     Architecture guard: colours are design tokens. A view sets a brush from a resource (<c>{DynamicResource …}</c>
///     or a binding), never from a literal such as <c>#CC0000</c> or <c>Red</c>. Literals live in <c>Themes/</c>.
/// </summary>
public sealed class CodeStyle_LiteralColours
{
	[Fact]
	public void Views_do_not_set_literal_colours()
	{
		var violations = AxamlFiles()
			.SelectMany(static file => LiteralColourGuard.FindViolations(file, File.ReadAllText(file)))
			.ToArray();

		violations.Should().BeEmpty(
			"colours come from Themes/Tokens.axaml:{0}{1}",
			Environment.NewLine,
			string.Join(Environment.NewLine, violations));
	}

	private static IEnumerable<string> AxamlFiles()
	{
		var root = Path.Combine(Harness.RepoRoot, "ARMEmulator");
		return Directory.EnumerateFiles(root, "*.axaml", SearchOption.AllDirectories)
			.Where(static file => {
				var segments = file.Split(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar);
				return !segments.Contains("bin") && !segments.Contains("obj") && !segments.Contains("Themes");
			});
	}

	[Theory]
	[InlineData("<Border Background=\"#FFDDDD\" />")]
	[InlineData("<Border Background=\"#33FF0000\" />")]
	[InlineData("<TextBlock Foreground=\"Red\" />")]
	[InlineData("<Border BorderBrush=\"DodgerBlue\" />")]
	[InlineData("<Ellipse Fill=\"Gray\" />")]
	[InlineData("<Setter Property=\"Foreground\" Value=\"#CC0000\" />")]
	[InlineData("<Setter Property=\"Background\" Value=\"Orange\" />")]
	public void Guard_flags_literal_colours(string xaml) =>
		LiteralColourGuard.FindViolations("View.axaml", xaml).Should().HaveCount(1);

	[Theory]
	[InlineData("<Border Background=\"{DynamicResource PanelBackgroundBrush}\" />")]
	[InlineData("<Border Background=\"{Binding StatusBrush}\" />")]
	[InlineData("<Border Background=\"Transparent\" />")]
	[InlineData("<TextBlock Text=\"#1 Red\" />")]
	[InlineData("<Setter Property=\"Foreground\" Value=\"{DynamicResource SecondaryTextBrush}\" />")]
	[InlineData("<Setter Property=\"Padding\" Value=\"4\" />")]
	public void Guard_accepts_resources_bindings_and_transparent(string xaml) =>
		LiteralColourGuard.FindViolations("View.axaml", xaml).Should().BeEmpty();
}

internal static partial class LiteralColourGuard
{
	// Properties that take a brush or colour: Foreground, Background, BorderBrush, Fill, Stroke, Color
	[GeneratedRegex(@"\b(?:Foreground|Background|BorderBrush|Fill|Stroke|Color)=""(?<value>[^""{][^""]*)""")]
	private static partial Regex BrushAttribute();

	[GeneratedRegex(@"Property=""(?:Foreground|Background|BorderBrush|Fill|Stroke|Color)""\s+Value=""(?<value>[^""{][^""]*)""")]
	private static partial Regex BrushSetter();

	public static IEnumerable<string> FindViolations(string file, string xaml) =>
		BrushAttribute().Matches(xaml).Concat(BrushSetter().Matches(xaml))
			.Select(static match => match.Groups["value"].Value)
			.Where(static value => value != "Transparent")
			.Select(value => $"{file}: literal colour '{value}'");
}
