namespace ARMEmulator.Tests;

using System.Collections.Immutable;
using System.Reflection;
using Microsoft.AspNetCore.Razor.Language;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;

/// <summary>
///     Parses Razor with the compiler shipped by the selected .NET SDK, then exposes only syntax
///     nodes which map back to authored C# in the original document. Generated rendering machinery,
///     markup, HTML comments and Razor comments consequently cannot become architecture violations.
/// </summary>
internal sealed class RazorCSharpDocument
{
	private static readonly RazorProjectEngine ProjectEngine = RazorProjectEngine.Create(
		RazorConfiguration.Default,
		RazorProjectFileSystem.Empty);

	private static readonly MethodInfo GetCSharpDocumentMethod =
		typeof(RazorCodeDocument).GetMethod("GetCSharpDocument", BindingFlags.Instance | BindingFlags.NonPublic)
		?? throw new InvalidOperationException("The selected Razor compiler does not expose its generated C# document.");

	private readonly Microsoft.AspNetCore.Razor.Language.RazorCSharpDocument generated;

	private RazorCSharpDocument(Microsoft.AspNetCore.Razor.Language.RazorCSharpDocument generated)
	{
		this.generated = generated;
		Root = CSharpSyntaxTree.ParseText(
			generated.Text.ToString(),
			CSharpParseOptions.Default.WithLanguageVersion(LanguageVersion.Preview)).GetRoot();
	}

	public SyntaxNode Root { get; }

	public static RazorCSharpDocument Parse(string fileName, string source)
	{
		var sourceDocument = RazorSourceDocument.Create(source, fileName);
		var codeDocument = ProjectEngine.Process(
			sourceDocument,
			RazorFileKind.Legacy,
			ImmutableArray<RazorSourceDocument>.Empty,
			null);
		var generated = GetCSharpDocumentMethod.Invoke(codeDocument, null) as Microsoft.AspNetCore.Razor.Language.RazorCSharpDocument
						?? throw new InvalidOperationException("The Razor compiler did not produce a generated C# document.");

		return new(generated);
	}

	public int? OriginalLine(SyntaxNode node) => OriginalLine(node.SpanStart);

	public int? OriginalLine(SyntaxToken token) => OriginalLine(token.SpanStart);

	public IEnumerable<int> OriginalLines(SyntaxToken token)
	{
		foreach (var mapping in generated.SourceMappingsSortedByGenerated) {
			var generatedSpan = mapping.GeneratedSpan;
			var overlapStart = Math.Max(token.SpanStart, generatedSpan.AbsoluteIndex);
			var overlapEnd = Math.Min(token.Span.End, generatedSpan.AbsoluteIndex + generatedSpan.Length);
			if (overlapStart >= overlapEnd) {
				continue;
			}

			var originalStart = mapping.OriginalSpan.AbsoluteIndex + overlapStart - generatedSpan.AbsoluteIndex;
			var originalEnd = mapping.OriginalSpan.AbsoluteIndex + overlapEnd - generatedSpan.AbsoluteIndex - 1;
			var firstLine = generated.CodeDocument.Source.Text.Lines.GetLineFromPosition(originalStart).LineNumber + 1;
			var lastLine = generated.CodeDocument.Source.Text.Lines.GetLineFromPosition(originalEnd).LineNumber + 1;
			foreach (var line in Enumerable.Range(firstLine, lastLine - firstLine + 1)) {
				yield return line;
			}
		}
	}

	public IEnumerable<int[]> OriginalLineGroups(IEnumerable<SyntaxToken> tokens)
	{
		var materializedTokens = tokens.ToArray();
		foreach (var mapping in generated.SourceMappingsSortedByGenerated) {
			var lines = new HashSet<int>();
			foreach (var token in materializedTokens) {
				var generatedSpan = mapping.GeneratedSpan;
				var overlapStart = Math.Max(token.SpanStart, generatedSpan.AbsoluteIndex);
				var overlapEnd = Math.Min(token.Span.End, generatedSpan.AbsoluteIndex + generatedSpan.Length);
				if (overlapStart >= overlapEnd) {
					continue;
				}

				var originalStart = mapping.OriginalSpan.AbsoluteIndex + overlapStart - generatedSpan.AbsoluteIndex;
				var originalEnd = mapping.OriginalSpan.AbsoluteIndex + overlapEnd - generatedSpan.AbsoluteIndex - 1;
				var firstLine = generated.CodeDocument.Source.Text.Lines.GetLineFromPosition(originalStart).LineNumber + 1;
				var lastLine = generated.CodeDocument.Source.Text.Lines.GetLineFromPosition(originalEnd).LineNumber + 1;
				foreach (var line in Enumerable.Range(firstLine, lastLine - firstLine + 1)) {
					_ = lines.Add(line);
				}
			}

			if (lines.Count > 0) {
				yield return lines.Order().ToArray();
			}
		}
	}

	private int? OriginalLine(int generatedPosition)
	{
		foreach (var mapping in generated.SourceMappingsSortedByGenerated) {
			var generatedSpan = mapping.GeneratedSpan;
			if (generatedPosition < generatedSpan.AbsoluteIndex
				|| generatedPosition >= generatedSpan.AbsoluteIndex + generatedSpan.Length) {
				continue;
			}

			var originalPosition = mapping.OriginalSpan.AbsoluteIndex + generatedPosition - generatedSpan.AbsoluteIndex;
			return generated.CodeDocument.Source.Text.Lines.GetLineFromPosition(originalPosition).LineNumber + 1;
		}

		return null;
	}
}
