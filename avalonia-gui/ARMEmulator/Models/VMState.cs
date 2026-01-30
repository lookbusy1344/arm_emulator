namespace ARMEmulator.Models;

/// <summary>
/// Represents the current state of the virtual machine.
/// </summary>
public enum VMState
{
    /// <summary>VM is idle, ready to execute.</summary>
    Idle,

    /// <summary>VM is currently running.</summary>
    Running,

    /// <summary>VM has hit a breakpoint.</summary>
    Breakpoint,

    /// <summary>VM has halted (program completed).</summary>
    Halted,

    /// <summary>VM encountered an error.</summary>
    Error,

    /// <summary>VM is waiting for user input.</summary>
    WaitingForInput
}

/// <summary>
/// Extension methods for VMState that determine UI behavior.
/// </summary>
public static class VMStateExtensions
{
    /// <summary>
    /// Determines if the editor should be editable in this state.
    /// </summary>
    public static bool IsEditorEditable(this VMState state) =>
        state is VMState.Idle or VMState.Halted or VMState.Error;

    /// <summary>
    /// Determines if step commands can be executed in this state.
    /// </summary>
    public static bool CanStep(this VMState state) =>
        state is VMState.Idle or VMState.Breakpoint;

    /// <summary>
    /// Determines if the pause command can be executed in this state.
    /// </summary>
    public static bool CanPause(this VMState state) =>
        state is VMState.Running or VMState.WaitingForInput;
}
