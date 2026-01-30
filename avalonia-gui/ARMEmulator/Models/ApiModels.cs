namespace ARMEmulator.Models;

/// <summary>
/// Information about a created session.
/// </summary>
public sealed record SessionInfo(string SessionId);

/// <summary>
/// Response from loading a program.
/// </summary>
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
/// Application settings (persisted to user config).
/// </summary>
public sealed class AppSettings
{
    public string BackendUrl { get; set; } = "http://localhost:8080";
    public int FontSize { get; set; } = 14;
    public string ColorScheme { get; set; } = "Auto"; // Auto, Light, Dark
    public int RecentFilesLimit { get; set; } = 10;
    public List<string> RecentFiles { get; set; } = [];
}
