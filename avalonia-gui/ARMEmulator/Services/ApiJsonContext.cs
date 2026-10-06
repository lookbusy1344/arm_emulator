using System.Text.Json.Serialization;
using ARMEmulator.Collections;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// JSON serializer context for AOT-friendly JSON serialization.
/// Source generator creates optimized serialization code at compile time.
/// </summary>
[JsonSerializable(typeof(SessionInfo))]
[JsonSerializable(typeof(SessionStatusResponse))]
[JsonSerializable(typeof(LoadProgramRequest))]
[JsonSerializable(typeof(LoadProgramWireResponse))]
[JsonSerializable(typeof(RegistersResponse))]
[JsonSerializable(typeof(VersionResponse))]
[JsonSerializable(typeof(ExampleInfo))]
[JsonSerializable(typeof(ExampleContentResponse))]
[JsonSerializable(typeof(WatchpointWire))]
[JsonSerializable(typeof(ApiErrorResponse))]
[JsonSerializable(typeof(MemoryResponse))]
[JsonSerializable(typeof(DisassemblyResponse))]
[JsonSerializable(typeof(SourceMapResponse))]
[JsonSerializable(typeof(BreakpointsResponse))]
[JsonSerializable(typeof(WatchpointsResponse))]
[JsonSerializable(typeof(EvaluationResponse))]
[JsonSerializable(typeof(ExamplesResponse))]
[JsonSerializable(typeof(BreakpointRequest))]
[JsonSerializable(typeof(AddWatchpointRequest))]
[JsonSerializable(typeof(EvaluateExpressionRequest))]
[JsonSerializable(typeof(StdinRequest))]
// Element types read by the EquatableArray / EquatableDictionary converters via options.GetTypeInfo
[JsonSerializable(typeof(string))]
[JsonSerializable(typeof(uint))]
[JsonSerializable(typeof(Dictionary<string, uint>))]
[JsonSerializable(typeof(DisassemblyInstructionWire))]
[JsonSerializable(typeof(SourceMapEntryWire))]
[JsonSourceGenerationOptions(
	PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
	PropertyNameCaseInsensitive = true)]
internal sealed partial class ApiJsonContext : JsonSerializerContext;

// Request bodies, matching the structs decoded in api/models.go and api/handlers.go

internal sealed record LoadProgramRequest(string Source);
internal sealed record BreakpointRequest(uint Address);
internal sealed record AddWatchpointRequest(uint Address, string Type);
internal sealed record EvaluateExpressionRequest(string Expression);
internal sealed record StdinRequest(string Data);

// Response bodies, mapped to domain models before leaving ApiClient

internal sealed record ApiErrorResponse(string? Error, string? Message);

internal sealed record SessionStatusResponse(
	string State,
	uint Pc,
	ulong Cycles,
	string? Error,
	bool HasWrite,
	uint? WriteAddr,
	uint? WriteSize)
{
	// The backend omits writeAddr and writeSize when they are zero
	public VMStatus ToVMStatus() => new(
		WireFormat.ParseState(State),
		Pc,
		Cycles,
		Error,
		HasWrite ? new MemoryWrite(WriteAddr ?? 0, WriteSize ?? 0) : null);
}

// Absent errors or symbols deserialize as empty collections
internal sealed record LoadProgramWireResponse(bool Success, EquatableArray<string> Errors, EquatableDictionary<string, uint> Symbols);

internal sealed record RegistersResponse(
	uint R0, uint R1, uint R2, uint R3, uint R4, uint R5, uint R6, uint R7,
	uint R8, uint R9, uint R10, uint R11, uint R12, uint Sp, uint Lr, uint Pc,
	CPSRFlags Cpsr)
{
	public RegisterState ToRegisterState() => RegisterState.Create(
		R0, R1, R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12, Sp, Lr, Pc, Cpsr);
}

internal sealed record DisassemblyInstructionWire(uint Address, uint MachineCode, string Disassembly, string? Symbol)
{
	// The backend returns the source line, including its indentation
	public DisassemblyInstruction ToModel() => new(Address, MachineCode, Disassembly.Trim(), Symbol);
}

internal sealed record SourceMapEntryWire(uint Address, int LineNumber, string Line)
{
	public SourceMapEntry ToModel() => new(Address, LineNumber, Line);
}

internal sealed record WatchpointWire(int Id, uint Address, string Type)
{
	public Watchpoint ToModel() => new(Id, Address, WireFormat.ParseWatchpointType(Type));
}

internal sealed record VersionResponse(string Version, string Commit, string Date)
{
	public BackendVersion ToModel() => new(Version, Commit, Date);
}

internal sealed record ExampleContentResponse(string Name, string Content);

// The backend encodes memory bytes as base64
internal sealed record MemoryResponse(string Data)
{
	public ImmutableArray<byte> ToBytes() => [.. Convert.FromBase64String(Data)];
}
internal sealed record DisassemblyResponse(EquatableArray<DisassemblyInstructionWire> Instructions);
internal sealed record SourceMapResponse(EquatableArray<SourceMapEntryWire> SourceMap);
internal sealed record BreakpointsResponse(EquatableArray<uint> Breakpoints);
internal sealed record WatchpointsResponse(EquatableArray<WatchpointWire> Watchpoints);
internal sealed record EvaluationResponse(uint Result);
internal sealed record ExamplesResponse(EquatableArray<ExampleInfo> Examples);

/// <summary>
/// String values the backend uses on the wire (service/types.go, api/broadcaster.go, api/handlers.go).
/// </summary>
internal static class WireFormat
{
	public static VMState ParseState(string value) => value switch {
		"running" => VMState.Running,
		"halted" => VMState.Halted,
		"breakpoint" => VMState.Breakpoint,
		"error" => VMState.Error,
		"waiting_for_input" => VMState.WaitingForInput,
		_ => throw new FormatException($"Unknown VM state '{value}'.")
	};

	public static ExecutionEventType ParseExecutionEvent(string value) => value switch {
		"breakpoint_hit" => ExecutionEventType.BreakpointHit,
		"halted" => ExecutionEventType.Halted,
		"error" => ExecutionEventType.Error,
		_ => throw new FormatException($"Unknown execution event '{value}'.")
	};

	public static WatchpointType ParseWatchpointType(string value) => value switch {
		"read" => WatchpointType.Read,
		"write" => WatchpointType.Write,
		"readwrite" => WatchpointType.ReadWrite,
		_ => throw new FormatException($"Unknown watchpoint type '{value}'.")
	};

	public static string ToWire(WatchpointType type) => type switch {
		WatchpointType.Read => "read",
		WatchpointType.Write => "write",
		WatchpointType.ReadWrite => "readwrite",
		_ => throw new ArgumentOutOfRangeException(nameof(type), type, "Unknown watchpoint type.")
	};
}
