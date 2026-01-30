using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Client for the ARM Emulator REST API.
/// All methods throw <see cref="ApiException"/> or derived types on failure.
/// </summary>
public interface IApiClient
{
	// Session Management

	/// <summary>Creates a new emulator session.</summary>
	/// <exception cref="BackendUnavailableException">Backend not reachable</exception>
	/// <exception cref="ApiException">Request failed</exception>
	Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default);

	/// <summary>Gets the current VM status for a session.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<VMStatus> GetStatusAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Destroys a session and frees its resources.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task DestroySessionAsync(string sessionId, CancellationToken ct = default);

	// Program Loading

	/// <summary>Loads and assembles ARM assembly source code into the session.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	/// <exception cref="ProgramLoadException">Assembly/parse errors</exception>
	Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default);

	// Execution Control

	/// <summary>Starts or continues execution of the loaded program.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RunAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Pauses execution of the running program.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task StopAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Executes a single instruction and returns updated register state.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Steps over function calls, stopping at the next instruction in the current function.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepOverAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Steps out of the current function, stopping after the return.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepOutAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Resets the VM to its initial state without reloading the program.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task ResetAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Reloads the program and resets the VM to entry point.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RestartAsync(string sessionId, CancellationToken ct = default);

	// State Inspection

	/// <summary>Gets the current values of all CPU registers.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> GetRegistersAsync(string sessionId, CancellationToken ct = default);

	/// <summary>Reads a range of memory bytes starting at the specified address.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<byte>> GetMemoryAsync(string sessionId, uint address, int length, CancellationToken ct = default);

	/// <summary>Disassembles instructions starting at the specified address.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<DisassemblyInstruction>> GetDisassemblyAsync(string sessionId, uint address, int count, CancellationToken ct = default);

	/// <summary>Gets the source-to-address mapping for the loaded program.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<SourceMapEntry>> GetSourceMapAsync(string sessionId, CancellationToken ct = default);

	// Breakpoints

	/// <summary>Adds a breakpoint at the specified address.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task AddBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

	/// <summary>Removes the breakpoint at the specified address.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RemoveBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

	/// <summary>Gets all active breakpoint addresses.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<uint>> GetBreakpointsAsync(string sessionId, CancellationToken ct = default);

	// Watchpoints

	/// <summary>Adds a memory watchpoint that breaks on read, write, or both.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<Watchpoint> AddWatchpointAsync(string sessionId, uint address, WatchpointType type, CancellationToken ct = default);

	/// <summary>Removes a watchpoint by its ID.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RemoveWatchpointAsync(string sessionId, int watchpointId, CancellationToken ct = default);

	/// <summary>Gets all active watchpoints.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<Watchpoint>> GetWatchpointsAsync(string sessionId, CancellationToken ct = default);

	// Expression Evaluation

	/// <summary>Evaluates an expression (e.g., "r0 + r1", "[sp]") in the current VM context.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	/// <exception cref="ExpressionEvaluationException">Invalid expression</exception>
	Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default);

	// Input

	/// <summary>Sends input data to the program's stdin.</summary>
	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task SendStdinAsync(string sessionId, string data, CancellationToken ct = default);

	// Version

	/// <summary>Gets the backend emulator version information.</summary>
	/// <exception cref="BackendUnavailableException">Backend not reachable</exception>
	Task<BackendVersion> GetVersionAsync(CancellationToken ct = default);

	// Examples

	/// <summary>Gets the list of available example programs.</summary>
	Task<ImmutableArray<ExampleInfo>> GetExamplesAsync(CancellationToken ct = default);

	/// <summary>Gets the source code for a specific example program.</summary>
	/// <exception cref="ApiException">Example not found</exception>
	Task<string> GetExampleContentAsync(string name, CancellationToken ct = default);
}
