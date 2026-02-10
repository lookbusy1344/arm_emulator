using System.Diagnostics.CodeAnalysis;

namespace ARMEmulator.Models;

/// <summary>
/// Information about a created session.
/// </summary>
public sealed record SessionInfo(string SessionId);

/// <summary>
/// Response from loading a program.
/// </summary>
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "ImmutableArray acceptable for small API response collections")]
public sealed record LoadProgramResponse(
	bool Success,
	ImmutableArray<ParseError> Errors,
	uint EntryPoint
);

/// <summary>
/// Parse error from the assembler (defined in Services but used in Models).
/// </summary>
public sealed record ParseError(int Line, int Column, string Message);

/// <summary>
/// Backend version information.
/// </summary>
public sealed record BackendVersion(
	string Version,
	string Commit,
	string BuildDate
);

/// <summary>
/// Information about an example program.
/// </summary>
public sealed record ExampleInfo(
	string Name,
	string Description,
	int Size
);

/// <summary>
/// Disassembled instruction.
/// </summary>
public sealed record DisassemblyInstruction(
	uint Address,
	uint MachineCode,
	string Mnemonic,
	string? Symbol = null
);

/// <summary>
/// Source map entry linking address to source line.
/// </summary>
public sealed record SourceMapEntry(
	uint Address,
	int LineNumber,
	string SourceText
);

/// <summary>
/// Application settings with persistent configuration.
/// Immutable record for thread-safe settings updates.
/// </summary>
public sealed record AppSettings
{
	/// <summary>Backend API base URL.</summary>
	public required string BackendUrl { get; init; }

	/// <summary>Editor font size (10-24pt range).</summary>
	public required int EditorFontSize { get; init; }

	/// <summary>Application color theme.</summary>
	public required AppTheme Theme { get; init; }

	/// <summary>Maximum number of recent files to track.</summary>
	public required int RecentFilesLimit { get; init; }

	/// <summary>Auto-scroll memory view to writes.</summary>
	public required bool AutoScrollToMemoryWrites { get; init; }

	/// <summary>Default settings instance.</summary>
	public static AppSettings Default { get; } = new() {
		BackendUrl = "http://localhost:8080",
		EditorFontSize = 14,
		Theme = AppTheme.Auto,
		RecentFilesLimit = 10,
		AutoScrollToMemoryWrites = true
	};

	/// <summary>
	/// Creates a validated copy with clamped font size.
	/// </summary>
	public AppSettings Validate() => this with {
		EditorFontSize = Math.Clamp(EditorFontSize, 10, 24),
		RecentFilesLimit = Math.Max(RecentFilesLimit, 1)
	};
}

/// <summary>
/// Application theme options.
/// </summary>
public enum AppTheme
{
	/// <summary>Automatically detect system theme.</summary>
	Auto,

	/// <summary>Light theme.</summary>
	Light,

	/// <summary>Dark theme.</summary>
	Dark
}
