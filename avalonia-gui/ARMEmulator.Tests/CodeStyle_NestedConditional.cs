namespace ARMEmulator.Tests;

using AwesomeAssertions;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

/// <summary>
///     Architecture guard for CLAUDE.md's ban on nested <c>?:</c> — a ternary inside either branch of another
///     ternary, in any position. The chained <c>a ? x : b ? y : z</c> form is the common offender; it hides a
///     multi-way decision inside an expression that reads as two-way. Rewrite as a <c>switch</c> expression, or
///     extract the inner conditional into a named local.
/// </summary>
/// <remarks>
///     Razor files are lowered through the SDK's Razor compiler and inspected as generated Roslyn syntax,
///     restricted through source mappings to C# authored in the original <c>.cshtml</c> document.
/// </remarks>
public sealed class CodeStyle_NestedConditional
{
	[Fact]
	public void Repository_sources_do_not_nest_conditional_expressions()
	{
		var violations = RepositorySourceFiles.CSharpAndRazor()
											  .SelectMany(static file => NestedConditionalGuard.FindViolations(file, File.ReadAllText(file)))
											  .ToArray();

		violations.Should().BeEmpty("nested conditionals found:{0}{1}", Environment.NewLine, string.Join(Environment.NewLine, violations));
	}

	[Theory]
	[InlineData("var y = a ? 1 : b ? 2 : 3;")] // chained — nested in the false branch
	[InlineData("var y = a ? (b ? 1 : 2) : 3;")] // nested in the true branch
	[InlineData("var y = a ? 1 : (b ? 2 : 3);")] // parenthesised chain
	[InlineData("var y = (a ? b : c) ? 1 : 2;")] // nested in the condition
	public void Nested_conditionals_are_violations(string statement)
	{
		var source = $"class Example {{ void M(bool a, bool b, bool c) {{ {statement} }} }}";

		NestedConditionalGuard.FindViolations("Example.cs", source).Should().NotBeEmpty();
	}

	[Theory]
	[InlineData("var y = a ? 1 : 2;")] // single conditional
	[InlineData("var y = (a ? 1 : 2) + (b ? 3 : 4);")] // siblings, not nested
	[InlineData("M(a ? 1 : 2, b ? 3 : 4);")] // siblings as separate arguments
	[InlineData("var y = a ? 1 : 2; var z = b ? 3 : 4;")] // separate statements
	public void Unnested_conditionals_are_allowed(string statement)
	{
		var source = $"class Example {{ void M(bool a, bool b) {{ {statement} }} static void M(int p, int q) {{ }} }}";

		NestedConditionalGuard.FindViolations("Example.cs", source).Should().BeEmpty();
	}

	[Theory]
	[InlineData("<p>@(a ? \"x\" : b ? \"y\" : \"z\")</p>")]
	[InlineData("<p class=\"@(a ? (b ? \"x\" : \"y\") : null)\">text</p>")]
	[InlineData("@{ var label = a ? \"x\" : b ? \"y\" : \"z\"; }")]
	[InlineData("@if (a ? b : c ? d : e) { <p>value</p> }")]
	[InlineData("@for (var i = 0; a ? b : c ? d : e; ++i) { <p>@i</p> }")]
	[InlineData("@functions { private string Label() => a ? \"x\" : b ? \"y\" : \"z\"; }")]
	public void Razor_nested_conditionals_are_violations(string markup) => NestedConditionalGuard.FindViolations("Example.cshtml", markup).Should().NotBeEmpty();

	[Theory]
	[InlineData("<p>@(a ? \"x\" : \"y\")</p>")]
	[InlineData("<p>@(a ? \"x\" : \"y\")@(b ? \"z\" : \"w\")</p>")] // siblings, not nested
	[InlineData("@{ var label = a ? \"x\" : \"y\"; var other = b ? \"z\" : \"w\"; }")]
	[InlineData("<a href=\"https://example.test/a?b=1\">link</a>")] // a query string is not a conditional
	[InlineData("@* Example only: @(a ? \"x\" : b ? \"y\" : \"z\") *@")]
	[InlineData("<code>a ? \"x\" : b ? \"y\" : \"z\"</code>")]
	public void Razor_unnested_conditionals_are_allowed(string markup) => NestedConditionalGuard.FindViolations("Example.cshtml", markup).Should().BeEmpty();
}

internal static class NestedConditionalGuard
{
	public static IEnumerable<string> FindViolations(string fileName, string source) =>
		fileName.EndsWith(".cshtml", StringComparison.OrdinalIgnoreCase)
			? NestedInRazor(fileName, RazorCSharpDocument.Parse(fileName, source))
			: NestedIn(fileName, source, 0);

	private static IEnumerable<string> NestedInRazor(string fileName, RazorCSharpDocument document) =>
		document.Root.DescendantNodes()
				.OfType<ConditionalExpressionSyntax>()
				.Where(static conditional => conditional.Ancestors().OfType<ConditionalExpressionSyntax>().Any())
				.Select(document.OriginalLine)
				.Where(static line => line.HasValue)
				.Select(line => Describe(fileName, line!.Value));

	private static IEnumerable<string> NestedIn(string fileName, string code, int lineOffset)
	{
		var root = CSharpSyntaxTree.ParseText(code).GetRoot();
		return root.DescendantNodes()
				   .OfType<ConditionalExpressionSyntax>()
				   .Where(static conditional => conditional.Ancestors().OfType<ConditionalExpressionSyntax>().Any())
				   .Select(conditional => Describe(fileName, conditional, lineOffset));
	}

	private static string Describe(string fileName, ConditionalExpressionSyntax conditional, int lineOffset) =>
		Describe(fileName, conditional.GetLocation().GetLineSpan().StartLinePosition.Line + 1 + lineOffset);

	private static string Describe(string fileName, int line) => $"{Path.GetFileName(fileName)}:{line}: forbidden nested conditional expression";
}
