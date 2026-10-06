namespace ARMEmulator.Tests;

using AwesomeAssertions;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;

/// <summary>
///     Hard line ceilings for authored C# and Razor files. A C# file is measured in code lines — a line
///     carrying at least one token, so blank and comment-only lines are free and documentation costs
///     nothing. Razor is measured in physical lines, since its markup and comments interleave. Production
///     and sample C# carry the base ceiling; Razor a tighter one because its declarative markup is denser;
///     test specifications a looser one, since one big linear fixture is often clearer whole than split.
///     There is no exception mechanism — an overlong file is divided along a cohesive type, capability,
///     or scenario boundary.
/// </summary>
public sealed class CodeStyle_FileLength
{
	[Fact]
	public void Repository_CSharp_and_Razor_files_do_not_exceed_the_line_count_ceiling()
	{
		var solutionRoot = Harness.RepoRoot;
		var violations = RepositorySourceFiles.CSharpAndRazor()
											  .Select(file => FileLengthGuard.FindViolation(solutionRoot, file, FileLengthGuard.LineCount(file, File.ReadAllText(file))))
											  .Where(static violation => violation is not null)
											  .ToArray();

		violations.Should().BeEmpty(
			"overlong source files should be divided along cohesive type, capability, or scenario boundaries:{0}{1}",
			Environment.NewLine,
			string.Join(Environment.NewLine, violations));
	}

	[Fact]
	public void Blank_and_comment_lines_do_not_count_towards_a_CSharp_file()
	{
		var source = """
					 namespace Example;

					 /// <summary>
					 ///     Documentation occupies no code line.
					 /// </summary>
					 internal sealed class Example
					 {
					 	// Nor does an explanation.
					 	public int Value => 1; /* nor a trailing block comment */
					 }
					 """;

		FileLengthGuard.LineCount("Example.cs", source).Should().Be(5);
	}

	[Fact]
	public void A_multi_line_string_counts_every_line_it_occupies()
	{
		var source = "internal sealed class Example\n{\n\tpublic const string Text = @\"one\ntwo\nthree\";\n}\n";

		FileLengthGuard.LineCount("Example.cs", source).Should().Be(6);
	}

	[Fact]
	public void Razor_file_length_counts_every_physical_line()
	{
		var source = "@* A Razor comment *@\n\n<p>Markup</p>\n";

		FileLengthGuard.LineCount("Example.cshtml", source).Should().Be(3);
	}

	[Theory]
	[InlineData("ARMEmulator/Example.cs", FileLengthGuard.MaxProductionCSharpLineCount)]
	[InlineData("ARMEmulator/Views/Example.cs", FileLengthGuard.MaxProductionCSharpLineCount)]
	[InlineData("ARMEmulator/Pages/Example.cshtml", FileLengthGuard.MaxRazorLineCount)]
	[InlineData("ARMEmulator.Tests/Example.cs", FileLengthGuard.MaxTestCSharpLineCount)]
	public void File_at_its_ceiling_is_not_a_violation(string relativePath, int lineCount) =>
		FileLengthGuard.FindViolation("/repo", Path.Combine("/repo", relativePath), lineCount).Should().BeNull();

	[Theory]
	[InlineData("ARMEmulator/Example.cs", FileLengthGuard.MaxProductionCSharpLineCount)]
	[InlineData("ARMEmulator/Views/Example.cs", FileLengthGuard.MaxProductionCSharpLineCount)]
	[InlineData("ARMEmulator/Pages/Example.cshtml", FileLengthGuard.MaxRazorLineCount)]
	[InlineData("ARMEmulator.Tests/Example.cs", FileLengthGuard.MaxTestCSharpLineCount)]
	public void File_over_its_ceiling_is_a_violation(string relativePath, int maximum)
	{
		var violation = FileLengthGuard.FindViolation("/repo", Path.Combine("/repo", relativePath), maximum + 1);

		violation.Should().Contain(relativePath.Replace(Path.DirectorySeparatorChar, '/'))
				 .And.Contain($"{maximum + 1} lines")
				 .And.Contain($"maximum is {maximum}");
	}

	[Fact]
	public void Test_CSharp_file_carries_the_looser_test_ceiling() =>
		FileLengthGuard.FindViolation("/repo", "/repo/ARMEmulator.Tests/Example.cs", FileLengthGuard.MaxTestCSharpLineCount)
					   .Should().BeNull();
}

internal static class FileLengthGuard
{
	public const int MaxProductionCSharpLineCount = 1_000;
	public const int MaxRazorLineCount = 500;
	public const int MaxTestCSharpLineCount = 2_000;

	/// <summary>
	///     Counts what the ceiling measures: code lines for C#, physical lines for Razor.
	/// </summary>
	public static int LineCount(string fileName, string source) => IsRazor(fileName)
		? PhysicalLineCount(source)
		: CodeLineCount(source);

	public static string? FindViolation(string solutionRoot, string fileName, int lineCount)
	{
		var relativePath = Path.GetRelativePath(solutionRoot, fileName)
							   .Replace(Path.DirectorySeparatorChar, '/')
							   .Replace(Path.AltDirectorySeparatorChar, '/');
		var maximum = MaximumFor(relativePath);

		if (!maximum.HasValue || lineCount <= maximum.Value) {
			return null;
		}

		return $"{relativePath}: {lineCount} lines; maximum is {maximum.Value}";
	}

	private static int? MaximumFor(string relativePath)
	{
		var topLevelDirectory = relativePath.Split('/', 2)[0];
		if (topLevelDirectory == "ARMEmulator.Tests") {
			return MaxTestCSharpLineCount;
		}

		if (topLevelDirectory != "ARMEmulator") {
			throw new ArgumentException("Source file must be under ARMEmulator or ARMEmulator.Tests.", nameof(relativePath));
		}

		return IsRazor(relativePath)
			? MaxRazorLineCount
			: MaxProductionCSharpLineCount;
	}

	private static bool IsRazor(string fileName) => fileName.EndsWith(".cshtml", StringComparison.OrdinalIgnoreCase);

	private static int PhysicalLineCount(string source)
	{
		using var reader = new StringReader(source);
		var lineCount = 0;
		while (reader.ReadLine() is not null) {
			++lineCount;
		}

		return lineCount;
	}

	private static int CodeLineCount(string source) => CSharpSyntaxTree.ParseText(source)
																	   .GetRoot()
																	   .DescendantTokens()
																	   .Where(static token => !token.IsKind(SyntaxKind.EndOfFileToken))
																	   .SelectMany(static token => LinesOccupiedBy(token))
																	   .Distinct()
																	   .Count();

	private static IEnumerable<int> LinesOccupiedBy(SyntaxToken token)
	{
		var span = token.GetLocation().GetLineSpan();
		var firstLine = span.StartLinePosition.Line;
		return Enumerable.Range(firstLine, span.EndLinePosition.Line - firstLine + 1);
	}
}
