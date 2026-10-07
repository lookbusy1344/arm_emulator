namespace ARMEmulator.Models;

/// <summary>One line of the register table.</summary>
public sealed record RegisterRow(string Name, string Hex, string Base10, bool IsChanged);
