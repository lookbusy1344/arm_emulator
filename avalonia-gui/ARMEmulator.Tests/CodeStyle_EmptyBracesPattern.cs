namespace ARMEmulator.Tests;

using AwesomeAssertions;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

/// <summary>
///     Architecture guard for CLAUDE.md's ban on the empty property pattern with binding
///     (<c>x is { } y</c> / <c>x is not { } y</c>). It hides a null test and a rebind behind empty braces;
///     a named type pattern (<c>x is SomeType y</c>) is real pattern matching and stays fine. This test scans
///     the tracked <c>.cs</c> and <c>.cshtml</c> sources under <c>src</c>, <c>tests</c>, and <c>samples</c> so
///     the pattern cannot creep back in.
/// </summary>
public sealed class CodeStyle_EmptyBracesPattern
{
	[Fact]
	public void Repository_sources_do_not_use_the_empty_braces_property_pattern()
	{
		var violations = RepositorySourceFiles.CSharpAndRazor()
											  .SelectMany(static file => EmptyBracesPatternGuard.FindViolations(file, File.ReadAllText(file)))
											  .ToArray();

		violations.Should().BeEmpty();
	}

	[Theory]
	[InlineData("if (x is { } y) { }")]
	[InlineData("if (x is not { } y) { }")]
	[InlineData("var z = x is { } y ? y : null;")]
	[InlineData("_ = value is not {  } bound;")]
	public void Empty_braces_binding_is_a_violation(string statement)
	{
		var source = $"class Example {{ void M(object? x, object? value) {{ {statement} }} }}";

		EmptyBracesPatternGuard.FindViolations("Example.cs", source).Should().NotBeEmpty();
	}

	[Theory]
	[InlineData("if (x is string y) { }")] // named type pattern — allowed
	[InlineData("if (x is not string y) { }")] // negated named type pattern — allowed
	[InlineData("if (x is not null) { }")] // plain null guard — no binding
	[InlineData("if (x is { Length: 0 } y) { }")] // non-empty property pattern — real matching
	[InlineData("if (x is (var a, var b)) { _ = a; _ = b; }")] // positional pattern — allowed
	public void Allowed_patterns_are_not_violations(string statement)
	{
		var source = $"class Example {{ void M(object? x, (int, int) t) {{ {statement} }} }}";

		EmptyBracesPatternGuard.FindViolations("Example.cs", source).Should().BeEmpty();
	}

	[Fact]
	public void Razor_empty_braces_binding_is_a_violation()
	{
		const string source = "@if (Model.Results is not { } results) { <p>@results</p> }";

		EmptyBracesPatternGuard.FindViolations("Example.cshtml", source).Should().NotBeEmpty();
	}

	[Fact]
	public void Razor_unicode_empty_braces_binding_is_a_violation()
	{
		const string source = "@if (Model.Results is not { } résultats) { <p>@résultats</p> }";

		EmptyBracesPatternGuard.FindViolations("Example.cshtml", source).Should().NotBeEmpty();
	}

	[Theory]
	[InlineData("@* Documentation: value is { } bound. *@")]
	[InlineData("<code>value is { } bound</code>")]
	public void Razor_non_code_empty_braces_text_is_not_a_violation(string source) =>
		EmptyBracesPatternGuard.FindViolations("Example.cshtml", source).Should().BeEmpty();
}

internal static class EmptyBracesPatternGuard
{
	public static IEnumerable<string> FindViolations(string fileName, string source) =>
		fileName.EndsWith(".cshtml", StringComparison.OrdinalIgnoreCase)
			? RazorViolations(fileName, source)
			: SyntaxViolations(fileName, source);

	private static IEnumerable<string> RazorViolations(string fileName, string source) =>
		RazorSyntaxViolations(fileName, RazorCSharpDocument.Parse(fileName, source));

	private static IEnumerable<string> RazorSyntaxViolations(string fileName, RazorCSharpDocument document) =>
		document.Root.DescendantNodes()
				.OfType<RecursivePatternSyntax>()
				.Where(IsEmptyBracesBinding)
				.Select(document.OriginalLine)
				.Where(static line => line.HasValue)
				.Select(line => Describe(fileName, line!.Value));

	private static IEnumerable<string> SyntaxViolations(string fileName, string source)
	{
		var root = CSharpSyntaxTree.ParseText(source).GetRoot();
		return root.DescendantNodes()
				   .OfType<RecursivePatternSyntax>()
				   .Where(IsEmptyBracesBinding)
				   .Select(pattern => Describe(fileName, pattern.GetLocation().GetLineSpan().StartLinePosition.Line + 1));
	}

	// `{ } y` and — under a `not` — `not { } y`: an empty property pattern that both binds a designation and
	// carries no type or positional clause. A named type pattern (`SomeType y`) is a DeclarationPattern, not a
	// RecursivePattern; a positional pattern populates PositionalPatternClause; a non-empty `{ Prop: ... }`
	// populates Subpatterns — none of those match here.
	private static bool IsEmptyBracesBinding(RecursivePatternSyntax pattern) =>
		pattern is { Type: null, PositionalPatternClause: null, Designation: not null, PropertyPatternClause.Subpatterns.Count: 0 };

	private static string Describe(string fileName, int line) =>
		$"{Path.GetFileName(fileName)}:{line}: forbidden empty-braces property pattern";
}
