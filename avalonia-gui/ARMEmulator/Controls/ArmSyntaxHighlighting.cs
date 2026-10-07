using System.Reflection;
using System.Xml;
using Avalonia.Media;
using AvaloniaEdit.Highlighting;
using AvaloniaEdit.Highlighting.Xshd;

namespace ARMEmulator.Controls;

/// <summary>
/// The ARM assembly highlighting definition. Rules come from the embedded <c>.xshd</c> file; colours come from the theme.
/// </summary>
public static class ArmSyntaxHighlighting
{
	private const string ResourceName = "ARMEmulator.Resources.ARMAssembly.xshd";

	/// <summary>The colour names in the definition. Each maps to the theme brush <c>Syntax{Name}Brush</c>.</summary>
	public static ImmutableArray<string> ColourNames { get; } = ["Comment", "Instruction", "Register", "Number", "String", "Label", "Directive"];

	/// <summary>Loads the definition from the embedded resource.</summary>
	/// <exception cref="InvalidOperationException">The resource is missing from the assembly</exception>
	public static IHighlightingDefinition Load()
	{
		using var stream = Assembly.GetExecutingAssembly().GetManifestResourceStream(ResourceName)
			?? throw new InvalidOperationException($"Embedded resource {ResourceName} is missing.");
		using var reader = new XmlTextReader(stream);
		return HighlightingLoader.Load(reader, HighlightingManager.Instance);
	}

	/// <summary>The theme resource key that colours <paramref name="colourName"/>.</summary>
	public static string BrushKey(string colourName) => $"Syntax{colourName}Brush";

	/// <summary>Sets each named colour's foreground from <paramref name="findBrush"/>. A colour without a solid brush keeps its current foreground.</summary>
	public static void ApplyColours(IHighlightingDefinition definition, Func<string, IBrush?> findBrush)
	{
		foreach (var name in ColourNames) {
			if (findBrush(BrushKey(name)) is ISolidColorBrush brush) {
				definition.GetNamedColor(name).Foreground = new SimpleHighlightingBrush(brush.Color);
			}
		}
	}
}
