namespace ARMEmulator.Models;

/// <summary>
/// Immutable snapshot of virtual machine status.
/// </summary>
public sealed record VMStatus(
    VMState State,
    uint PC,
    ulong Cycles,
    string? Error = null,
    MemoryWrite? LastWrite = null
);

/// <summary>
/// Represents a memory write operation for tracking and highlighting.
/// </summary>
public sealed record MemoryWrite(uint Address, uint Size);
