namespace ARMEmulator.Models;

/// <summary>
/// Base class for all emulator events received via WebSocket.
/// Discriminated union implemented via abstract record with sealed derived types.
/// </summary>
public abstract record EmulatorEvent(string SessionId);

/// <summary>
/// VM state update event containing register and status information.
/// </summary>
public sealed record StateEvent(
    string SessionId,
    VMStatus Status,
    RegisterState Registers
) : EmulatorEvent(SessionId);

/// <summary>
/// Program output event (stdout or stderr).
/// </summary>
public sealed record OutputEvent(
    string SessionId,
    OutputStreamType Stream,
    string Content
) : EmulatorEvent(SessionId);

/// <summary>
/// Execution event (breakpoint, halt, error).
/// </summary>
public sealed record ExecutionEvent(
    string SessionId,
    ExecutionEventType EventType,
    uint? Address = null,
    string? Symbol = null,
    string? Message = null
) : EmulatorEvent(SessionId);

/// <summary>
/// Type of output stream.
/// </summary>
public enum OutputStreamType
{
    /// <summary>Standard output stream.</summary>
    Stdout,

    /// <summary>Standard error stream.</summary>
    Stderr
}

/// <summary>
/// Type of execution event.
/// </summary>
public enum ExecutionEventType
{
    /// <summary>Breakpoint was hit.</summary>
    BreakpointHit,

    /// <summary>Program halted normally.</summary>
    Halted,

    /// <summary>Error occurred during execution.</summary>
    Error
}
