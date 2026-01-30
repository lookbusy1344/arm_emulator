using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Client for the ARM Emulator REST API.
/// All methods throw <see cref="ApiException"/> or derived types on failure.
/// </summary>
public interface IApiClient
{
	// Session Management

	/// <exception cref="BackendUnavailableException">Backend not reachable</exception>
	/// <exception cref="ApiException">Request failed</exception>
	Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<VMStatus> GetStatusAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task DestroySessionAsync(string sessionId, CancellationToken ct = default);

	// Program Loading

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	/// <exception cref="ProgramLoadException">Assembly/parse errors</exception>
	Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default);

	// Execution Control

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RunAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task StopAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepOverAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> StepOutAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task ResetAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RestartAsync(string sessionId, CancellationToken ct = default);

	// State Inspection

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<RegisterState> GetRegistersAsync(string sessionId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<byte>> GetMemoryAsync(string sessionId, uint address, int length, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<DisassemblyInstruction>> GetDisassemblyAsync(string sessionId, uint address, int count, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<SourceMapEntry>> GetSourceMapAsync(string sessionId, CancellationToken ct = default);

	// Breakpoints

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task AddBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RemoveBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<uint>> GetBreakpointsAsync(string sessionId, CancellationToken ct = default);

	// Watchpoints

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<Watchpoint> AddWatchpointAsync(string sessionId, uint address, WatchpointType type, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task RemoveWatchpointAsync(string sessionId, int watchpointId, CancellationToken ct = default);

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task<ImmutableArray<Watchpoint>> GetWatchpointsAsync(string sessionId, CancellationToken ct = default);

	// Expression Evaluation

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	/// <exception cref="ExpressionEvaluationException">Invalid expression</exception>
	Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default);

	// Input

	/// <exception cref="SessionNotFoundException">Session not found</exception>
	Task SendStdinAsync(string sessionId, string data, CancellationToken ct = default);

	// Version

	/// <exception cref="BackendUnavailableException">Backend not reachable</exception>
	Task<BackendVersion> GetVersionAsync(CancellationToken ct = default);

	// Examples

	Task<ImmutableArray<ExampleInfo>> GetExamplesAsync(CancellationToken ct = default);

	/// <exception cref="ApiException">Example not found</exception>
	Task<string> GetExampleContentAsync(string name, CancellationToken ct = default);
}
