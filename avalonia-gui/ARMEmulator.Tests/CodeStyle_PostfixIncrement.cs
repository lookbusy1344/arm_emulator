namespace ARMEmulator.Tests;

using AwesomeAssertions;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

/// <summary>
///     Architecture guard for the house preference for prefix increment/decrement wherever the operator's
///     result is discarded: a bare <c>count++;</c> statement and a <c>for</c> loop's <c>i++</c> incrementor
///     both mean exactly what <c>++count</c> / <c>++i</c> mean, so the prefix form is used consistently and
///     the postfix form is reserved for the cases that genuinely read the pre-increment value
///     (<c>var b = a++;</c>, <c>buffer[i++] = x;</c>, <c>return count--;</c>), where it stays untouched.
///     Scans the tracked <c>.cs</c> and <c>.cshtml</c> sources under <c>src</c>, <c>tests</c>, and
///     <c>samples</c>.
/// </summary>
public sealed class CodeStyle_PostfixIncrement
{
	[Fact]
	public void Repository_sources_do_not_discard_a_postfix_increment_result()
	{
		var violations = RepositorySourceFiles.CSharpAndRazor()
											  .SelectMany(static file => PostfixIncrementGuard.FindViolations(file, File.ReadAllText(file)))
											  .ToArray();

		violations.Should().BeEmpty(
			"the prefix form should be used wherever the result is discarded:{0}{1}",
			Environment.NewLine,
			string.Join(Environment.NewLine, violations));
	}

	[Theory]
	[InlineData("count++;")]
	[InlineData("count--;")]
	[InlineData("this.count++;")]
	[InlineData("values[0]++;")]
	[InlineData("for (var i = 0; i < 10; i++) { }")]
	[InlineData("for (var i = 10; i > 0; i--) { }")]
	[InlineData("for (var i = 0, j = 0; i < 10; i++, j++) { }")]
	[InlineData("while (count < 10) { count++; }")]
	[InlineData("while (count < 10) count++;")] // no braces — still an expression statement
	[InlineData("do { count++; } while (count < 10);")]
	public void Discarded_postfix_result_is_a_violation(string statement)
	{
		var source = $"class Example {{ int count; void M(int[] values) {{ {statement} }} }}";

		PostfixIncrementGuard.FindViolations("Example.cs", source).Should().NotBeEmpty();
	}

	[Theory]
	[InlineData("++count;")] // already the prefix form
	[InlineData("for (var i = 0; i < 10; ++i) { }")] // prefix incrementor
	[InlineData("var b = count++;")] // result used — the pre-increment value is the point
	[InlineData("values[count++] = 1;")] // result used as an index
	[InlineData("M2(count--);")] // result used as an argument
	[InlineData("if (count++ > 0) { }")] // result used as a condition
	[InlineData("while (count++ < 10) { }")] // result used as a loop condition
	[InlineData("do { } while (count-- > 0);")] // result used as a loop condition
	public void Used_or_prefix_forms_are_not_violations(string statement)
	{
		var source = $"class Example {{ int count; void M(int[] values) {{ {statement} }} static void M2(int v) {{ }} }}";

		PostfixIncrementGuard.FindViolations("Example.cs", source).Should().BeEmpty();
	}

	[Theory]
	[InlineData("@for (var i = 0; i < Model.Rows.Count; i++) { <p>@i</p> }")]
	[InlineData("@{ counter++; }")]
	[InlineData("@while (counter < 10) { counter++; <p>@counter</p> }")]
	[InlineData("@do { counter++; } while (counter < 10)")]
	[InlineData("@{ values[index]++; }")]
	[InlineData("@for (var i = 0, j = 0; i < Model.Rows.Count; i++, j++) { <p>@i @j</p> }")]
	public void Razor_discarded_postfix_result_is_a_violation(string source) =>
		PostfixIncrementGuard.FindViolations("Example.cshtml", source).Should().NotBeEmpty();

	[Theory]
	[InlineData("<div class=\"jt-card jt-card--narrow\"></div>")]
	[InlineData("<option value=\"\">-- none --</option>")]
	[InlineData("@* a comment -- with dashes *@")]
	[InlineData("@for (var i = 0; i < Model.Rows.Count; ++i) { <p>@i</p> }")]
	[InlineData("@{ var id = Next(counter++); }")] // result used
	[InlineData("@while (counter++ < 10) { <p>@counter</p> }")] // result used as a loop condition
	[InlineData("<div style=\"color: var(--bs-body-color)\"></div>")]
	[InlineData("@* Example only: counter++; *@")]
	[InlineData("<code>counter++;</code>")]
	public void Razor_markup_and_prefix_forms_are_not_violations(string source) =>
		PostfixIncrementGuard.FindViolations("Example.cshtml", source).Should().BeEmpty();
}

internal static class PostfixIncrementGuard
{
	public static IEnumerable<string> FindViolations(string fileName, string source) =>
		fileName.EndsWith(".cshtml", StringComparison.OrdinalIgnoreCase)
			? RazorViolations(fileName, source)
			: SyntaxViolations(fileName, source);

	private static IEnumerable<string> RazorViolations(string fileName, string source)
	{
		var document = RazorCSharpDocument.Parse(fileName, source);
		return document.Root.DescendantNodes()
					   .OfType<PostfixUnaryExpressionSyntax>()
					   .Where(IsIncrementOrDecrement)
					   .Where(IsResultDiscarded)
					   .Select(document.OriginalLine)
					   .Where(static line => line.HasValue)
					   .Select(line => Describe(fileName, line!.Value));
	}

	private static IEnumerable<string> SyntaxViolations(string fileName, string source)
	{
		var root = CSharpSyntaxTree.ParseText(source).GetRoot();
		return root.DescendantNodes()
				   .OfType<PostfixUnaryExpressionSyntax>()
				   .Where(IsIncrementOrDecrement)
				   .Where(IsResultDiscarded)
				   .Select(expression => Describe(fileName, expression.GetLocation().GetLineSpan().StartLinePosition.Line + 1));
	}

	private static bool IsIncrementOrDecrement(PostfixUnaryExpressionSyntax expression) =>
		expression.IsKind(SyntaxKind.PostIncrementExpression) || expression.IsKind(SyntaxKind.PostDecrementExpression);

	// Two shapes throw the operator's value away, so `a++` and `++a` are interchangeable: a bare expression
	// statement, and a `for` header's incrementor list. Anywhere else the pre-increment value may be read, and
	// syntax alone cannot prove otherwise — leave those alone.
	private static bool IsResultDiscarded(PostfixUnaryExpressionSyntax expression) =>
		expression.Parent switch {
			ExpressionStatementSyntax => true,
			ForStatementSyntax parent => parent.Incrementors.Contains(expression),
			_ => false,
		};

	private static string Describe(string fileName, int line) =>
		$"{Path.GetFileName(fileName)}:{line}: discarded postfix increment/decrement — use the prefix form";
}
