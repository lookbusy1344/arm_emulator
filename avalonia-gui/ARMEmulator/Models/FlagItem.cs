namespace ARMEmulator.Models;

/// <summary>One CPSR condition flag.</summary>
public sealed record FlagItem(char Letter, string Name, bool IsSet);
