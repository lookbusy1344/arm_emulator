using ARMEmulator.Controls;
using Avalonia.Media;
using AvaloniaEdit.Highlighting;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Controls;

public sealed class ArmSyntaxHighlightingTests
{
	private static readonly Color Applied = Color.Parse("#123456");

	private static Color? ForegroundOf(IHighlightingDefinition definition, string name) =>
		(definition.GetNamedColor(name).Foreground?.GetBrush(null) as ISolidColorBrush)?.Color;

	[Fact]
	public void Load_DefinesEveryNamedColour()
	{
		var definition = ArmSyntaxHighlighting.Load();

		ArmSyntaxHighlighting.ColourNames.Should().OnlyContain(name => definition.GetNamedColor(name) != null);
	}

	[Fact]
	public void ApplyColours_SetsEachColourFromItsThemeKey()
	{
		var definition = ArmSyntaxHighlighting.Load();
		var requested = new List<string>();

		ArmSyntaxHighlighting.ApplyColours(definition, key => {
			requested.Add(key);
			return new SolidColorBrush(Applied);
		});

		requested.Should().Equal("SyntaxCommentBrush", "SyntaxInstructionBrush", "SyntaxRegisterBrush", "SyntaxNumberBrush", "SyntaxStringBrush", "SyntaxLabelBrush", "SyntaxDirectiveBrush");
		ArmSyntaxHighlighting.ColourNames.Should().OnlyContain(name => ForegroundOf(definition, name) == Applied);
	}

	[Fact]
	public void ApplyColours_KeepsTheForegroundWhenThereIsNoBrush()
	{
		var definition = ArmSyntaxHighlighting.Load();
		var before = ForegroundOf(definition, "Comment");

		ArmSyntaxHighlighting.ApplyColours(definition, _ => null);

		ForegroundOf(definition, "Comment").Should().Be(before);
	}
}
