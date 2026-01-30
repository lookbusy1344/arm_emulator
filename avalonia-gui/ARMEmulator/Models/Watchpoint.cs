namespace ARMEmulator.Models;

/// <summary>
/// Represents a memory watchpoint for debugging.
/// </summary>
public sealed record Watchpoint(int Id, uint Address, WatchpointType Type);

/// <summary>
/// Type of memory access to watch.
/// </summary>
public enum WatchpointType
{
	/// <summary>Break on memory reads.</summary>
	Read,

	/// <summary>Break on memory writes.</summary>
	Write,

	/// <summary>Break on both reads and writes.</summary>
	ReadWrite
}
