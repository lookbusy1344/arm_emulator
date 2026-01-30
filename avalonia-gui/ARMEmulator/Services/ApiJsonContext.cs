using System.Diagnostics.CodeAnalysis;
using System.Text.Json.Serialization;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// JSON serializer context for AOT-friendly JSON serialization.
/// Source generator creates optimized serialization code at compile time.
/// </summary>
[JsonSerializable(typeof(SessionInfo))]
[JsonSerializable(typeof(VMStatus))]
[JsonSerializable(typeof(LoadProgramResponse))]
[JsonSerializable(typeof(RegisterState))]
[JsonSerializable(typeof(BackendVersion))]
[JsonSerializable(typeof(ExampleInfo))]
[JsonSerializable(typeof(Watchpoint))]
[JsonSerializable(typeof(DisassemblyInstruction))]
[JsonSerializable(typeof(SourceMapEntry))]
[JsonSerializable(typeof(ParseError))]
[JsonSerializable(typeof(ApiErrorResponse))]
[JsonSerializable(typeof(MemoryResponse))]
[JsonSerializable(typeof(DisassemblyResponse))]
[JsonSerializable(typeof(SourceMapResponse))]
[JsonSerializable(typeof(BreakpointsResponse))]
[JsonSerializable(typeof(WatchpointsResponse))]
[JsonSerializable(typeof(EvaluationResponse))]
[JsonSerializable(typeof(ExamplesResponse))]
[JsonSerializable(typeof(AddBreakpointRequest))]
[JsonSerializable(typeof(AddWatchpointRequest))]
[JsonSerializable(typeof(EvaluateExpressionRequest))]
[JsonSourceGenerationOptions(
	PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
	PropertyNameCaseInsensitive = true)]
internal sealed partial class ApiJsonContext : JsonSerializerContext;

// Internal request types
internal sealed record AddBreakpointRequest(uint Address);
internal sealed record AddWatchpointRequest(uint Address, string Type);
internal sealed record EvaluateExpressionRequest(string Expression);

// Internal response wrapper types (moved from ApiClient.cs for source generator visibility)
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record ApiErrorResponse(string Error, ParseError[]? ParseErrors = null);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record MemoryResponse(byte[] Data);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record DisassemblyResponse(DisassemblyInstruction[] Instructions);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record SourceMapResponse(SourceMapEntry[] Entries);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record BreakpointsResponse(uint[] Breakpoints);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record WatchpointsResponse(Watchpoint[] Watchpoints);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record EvaluationResponse(uint Value);
[SuppressMessage("Design", "JSV01:Member does not have value semantics", Justification = "Internal wrapper for JSON deserialization, immediately converted to ImmutableArray")]
internal sealed record ExamplesResponse(ExampleInfo[] Examples);
