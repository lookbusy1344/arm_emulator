namespace ARMEmulator.Tests;

using System.Collections.Frozen;
using AwesomeAssertions;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

/// <summary>
///     Architecture guard for the Framework Design Guidelines rule (§9.12, folded into this repo's
///     conventions): never declare a constant table as a mutable array or collection behind
///     <c>static readonly</c> — <c>readonly</c> freezes the field's own reference, not its elements, so
///     a shared allowlist stays writable through the reference every caller holds. This also catches a
///     mutable collection nested inside an otherwise-readonly wrapper (e.g.
///     <c>IReadOnlyDictionary&lt;string, HashSet&lt;string&gt;&gt;</c>) — the outer interface stops
///     callers reassigning the field, not mutating the inner set. Prefer <c>FrozenSet&lt;T&gt;</c>/
///     <c>FrozenDictionary&lt;TKey, TValue&gt;</c> for membership,
///     <c>
///         static ReadOnlySpan&lt;T&gt; X =&gt;
///         [...]
///     </c>
///     for an ordered constant, or <c>IReadOnlyList&lt;T&gt;</c> only where an EF expression
///     tree must capture it. Scans the tracked <c>.cs</c> sources under <c>src</c> and <c>tests</c> so
///     the pattern cannot creep back in.
/// </summary>
public sealed class CodeStyle_MutableConstantTable
{
	[Fact]
	public void Repository_sources_have_no_mutable_static_readonly_constant_table()
	{
		var violations = RepositorySourceFiles.CSharpAndRazor()
											  .Where(static file => file.EndsWith(".cs", StringComparison.OrdinalIgnoreCase))
											  .SelectMany(static file => MutableConstantTableGuard.FindViolations(file, File.ReadAllText(file)))
											  .ToArray();

		violations.Should().BeEmpty();
	}

	[Theory]
	[InlineData("private static readonly int[] Numbers = [1, 2, 3];")]
	[InlineData("internal static readonly string[] Names = new[] { \"a\", \"b\" };")]
	[InlineData("public static readonly int[,] Grid = new int[2, 2];")]
	[InlineData("private readonly static double[] Weights = [1.0];")]
	[InlineData("private static readonly HashSet<string> Names = [\"a\"];")]
	[InlineData("private static readonly Dictionary<string, int> Values = new() { [\"a\"] = 1 };")]
	[InlineData("private static readonly List<string> OrderedNames = [\"a\"];")]
	[InlineData("private static readonly IReadOnlyDictionary<string, HashSet<string>> Nested = new Dictionary<string, HashSet<string>>();")]
	[InlineData("private static readonly System.Collections.Generic.Dictionary<int, string[]> ByKey = [];")]
	public void Mutable_static_readonly_constant_table_is_a_violation(string declaration)
	{
		var source = $"class Example {{ {declaration} }}";

		MutableConstantTableGuard.FindViolations("Example.cs", source).Should().ContainSingle();
	}

	[Theory]
	[InlineData("private static readonly System.Collections.Frozen.FrozenSet<string> Names = default!;")]
	[InlineData("private static readonly System.Collections.Generic.IReadOnlyList<int> Numbers = [];")]
	[InlineData("private readonly int[] instanceArray = [];")]
	[InlineData("private static int[] mutableTable = [];")]
	[InlineData("private static readonly System.Collections.Concurrent.ConcurrentDictionary<string, int> Cache = new();")]
	public void Allowed_declarations_are_not_violations(string declaration)
	{
		var source = $"class Example {{ {declaration} }}";

		MutableConstantTableGuard.FindViolations("Example.cs", source).Should().BeEmpty();
	}
}

internal static class MutableConstantTableGuard
{
	private static readonly FrozenSet<string> MutableCollectionTypeNames = FrozenSet.ToFrozenSet(
		[
			"Collection",
			"Dictionary",
			"HashSet",
			"LinkedList",
			"List",
			"ObservableCollection",
			"Queue",
			"SortedDictionary",
			"SortedSet",
			"Stack",
		],
		StringComparer.Ordinal);

	public static IEnumerable<string> FindViolations(string fileName, string source)
	{
		var root = CSharpSyntaxTree.ParseText(source).GetRoot();
		return root.DescendantNodes()
				   .OfType<FieldDeclarationSyntax>()
				   .Where(IsStaticReadonly)
				   .Where(IsMutableTableType)
				   .Select(field => Describe(fileName, field));
	}

	private static bool IsStaticReadonly(FieldDeclarationSyntax field) =>
		field.Modifiers.Any(SyntaxKind.StaticKeyword) && field.Modifiers.Any(SyntaxKind.ReadOnlyKeyword);

	// A bare array is always a violation regardless of naming. Otherwise, the innermost generic type
	// name in the declared type decides: this reaches past an outer read-only wrapper (e.g.
	// IReadOnlyDictionary<string, HashSet<string>>) to the mutable collection nested inside it, since a
	// caller holding that outer reference can still reach in and mutate the inner one.
	private static bool IsMutableTableType(FieldDeclarationSyntax field)
	{
		if (field.Declaration.Type is ArrayTypeSyntax) {
			return true;
		}

		var innermostGenericType = field.Declaration.Type
										.DescendantNodesAndSelf()
										.OfType<GenericNameSyntax>()
										.LastOrDefault();
		return innermostGenericType is not null && MutableCollectionTypeNames.Contains(innermostGenericType.Identifier.ValueText);
	}

	private static string Describe(string fileName, FieldDeclarationSyntax field)
	{
		var line = field.GetLocation().GetLineSpan().StartLinePosition.Line + 1;
		var names = string.Join(", ", field.Declaration.Variables.Select(static v => v.Identifier.Text));
		return $"{Path.GetFileName(fileName)}:{line}: mutable static readonly constant table ({names})";
	}
}
