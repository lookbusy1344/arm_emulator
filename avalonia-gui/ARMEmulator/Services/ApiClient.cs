using System.Net;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// HTTP client for the ARM Emulator REST API.
/// Uses idiomatic .NET exception-based error handling.
/// </summary>
public sealed class ApiClient(HttpClient http) : IApiClient
{
	private readonly JsonSerializerOptions _jsonOptions = new()
	{
		PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
		PropertyNameCaseInsensitive = true
	};

	// Session Management

	public async Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default)
	{
		try
		{
			var response = await http.PostAsync("/api/v1/session", null, ct);
			return await ParseResponseOrThrowAsync<SessionInfo>(response, ct);
		}
		catch (HttpRequestException ex)
		{
			throw new BackendUnavailableException("Cannot connect to backend - is the emulator running?", ex);
		}
	}

	public async Task<VMStatus> GetStatusAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/status", ct);
		return await ParseResponseOrThrowAsync<VMStatus>(response, ct, sessionId);
	}

	public async Task DestroySessionAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.DeleteAsync($"/api/v1/session/{sessionId}", ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	// Program Loading

	public async Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default)
	{
		var content = new StringContent(source, Encoding.UTF8, "text/plain");
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/load", content, ct);

		// Special handling for parse errors (400 with error details)
		if (response.StatusCode == HttpStatusCode.BadRequest)
		{
			var errorResponse = await response.Content.ReadFromJsonAsync<ApiErrorResponse>(_jsonOptions, ct);
			if (errorResponse?.ParseErrors is { Length: > 0 } errors)
			{
				throw new ProgramLoadException([.. errors]);
			}
		}

		return await ParseResponseOrThrowAsync<LoadProgramResponse>(response, ct, sessionId);
	}

	// Execution Control

	public async Task RunAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/run", null, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task StopAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/stop", null, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task<RegisterState> StepAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/step", null, ct);
		return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
	}

	public async Task<RegisterState> StepOverAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/step-over", null, ct);
		return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
	}

	public async Task<RegisterState> StepOutAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/step-out", null, ct);
		return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
	}

	public async Task ResetAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/reset", null, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task RestartAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/restart", null, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	// State Inspection

	public async Task<RegisterState> GetRegistersAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/registers", ct);
		return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
	}

	public async Task<ImmutableArray<byte>> GetMemoryAsync(string sessionId, uint address, int length, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/memory?address={address}&length={length}", ct);
		var wrapper = await ParseResponseOrThrowAsync<MemoryResponse>(response, ct, sessionId);
		return [.. wrapper.Data];
	}

	public async Task<ImmutableArray<DisassemblyInstruction>> GetDisassemblyAsync(string sessionId, uint address, int count, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/disassembly?address={address}&count={count}", ct);
		var wrapper = await ParseResponseOrThrowAsync<DisassemblyResponse>(response, ct, sessionId);
		return [.. wrapper.Instructions];
	}

	public async Task<ImmutableArray<SourceMapEntry>> GetSourceMapAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/source-map", ct);
		var wrapper = await ParseResponseOrThrowAsync<SourceMapResponse>(response, ct, sessionId);
		return [.. wrapper.Entries];
	}

	// Breakpoints

	public async Task AddBreakpointAsync(string sessionId, uint address, CancellationToken ct = default)
	{
		var request = new { address };
		var response = await http.PostAsJsonAsync($"/api/v1/session/{sessionId}/breakpoint", request, _jsonOptions, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task RemoveBreakpointAsync(string sessionId, uint address, CancellationToken ct = default)
	{
		var response = await http.DeleteAsync($"/api/v1/session/{sessionId}/breakpoint/{address}", ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task<ImmutableArray<uint>> GetBreakpointsAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/breakpoints", ct);
		var wrapper = await ParseResponseOrThrowAsync<BreakpointsResponse>(response, ct, sessionId);
		return [.. wrapper.Breakpoints];
	}

	// Watchpoints

	public async Task<Watchpoint> AddWatchpointAsync(string sessionId, uint address, WatchpointType type, CancellationToken ct = default)
	{
		var request = new { address, type = type.ToString().ToLowerInvariant() };
		var response = await http.PostAsJsonAsync($"/api/v1/session/{sessionId}/watchpoint", request, _jsonOptions, ct);
		return await ParseResponseOrThrowAsync<Watchpoint>(response, ct, sessionId);
	}

	public async Task RemoveWatchpointAsync(string sessionId, int watchpointId, CancellationToken ct = default)
	{
		var response = await http.DeleteAsync($"/api/v1/session/{sessionId}/watchpoint/{watchpointId}", ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	public async Task<ImmutableArray<Watchpoint>> GetWatchpointsAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/session/{sessionId}/watchpoints", ct);
		var wrapper = await ParseResponseOrThrowAsync<WatchpointsResponse>(response, ct, sessionId);
		return [.. wrapper.Watchpoints];
	}

	// Expression Evaluation

	public async Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default)
	{
		var request = new { expression };
		var response = await http.PostAsJsonAsync($"/api/v1/session/{sessionId}/evaluate", request, _jsonOptions, ct);

		if (response.StatusCode == HttpStatusCode.BadRequest)
		{
			var errorResponse = await response.Content.ReadFromJsonAsync<ApiErrorResponse>(_jsonOptions, ct);
			throw new ExpressionEvaluationException(expression, errorResponse?.Error ?? "Unknown error");
		}

		var result = await ParseResponseOrThrowAsync<EvaluationResponse>(response, ct, sessionId);
		return result.Value;
	}

	// Input

	public async Task SendStdinAsync(string sessionId, string data, CancellationToken ct = default)
	{
		var content = new StringContent(data, Encoding.UTF8, "text/plain");
		var response = await http.PostAsync($"/api/v1/session/{sessionId}/stdin", content, ct);
		await ParseResponseOrThrowAsync<object>(response, ct, sessionId);
	}

	// Version

	public async Task<BackendVersion> GetVersionAsync(CancellationToken ct = default)
	{
		try
		{
			var response = await http.GetAsync("/api/v1/version", ct);
			return await ParseResponseOrThrowAsync<BackendVersion>(response, ct);
		}
		catch (HttpRequestException ex)
		{
			throw new BackendUnavailableException("Cannot reach backend", ex);
		}
	}

	// Examples

	public async Task<ImmutableArray<ExampleInfo>> GetExamplesAsync(CancellationToken ct = default)
	{
		var response = await http.GetAsync("/api/v1/examples", ct);
		var wrapper = await ParseResponseOrThrowAsync<ExamplesResponse>(response, ct);
		return [.. wrapper.Examples];
	}

	public async Task<string> GetExampleContentAsync(string name, CancellationToken ct = default)
	{
		var response = await http.GetAsync($"/api/v1/examples/{name}", ct);
		if (!response.IsSuccessStatusCode)
		{
			throw new ApiException($"Example '{name}' not found", response.StatusCode);
		}

		return await response.Content.ReadAsStringAsync(ct);
	}

	// Helper Methods

	private async Task<T> ParseResponseOrThrowAsync<T>(
		HttpResponseMessage response,
		CancellationToken ct,
		string? sessionId = null)
	{
		if (response.StatusCode == HttpStatusCode.NotFound && sessionId is not null)
		{
			throw new SessionNotFoundException(sessionId);
		}

		if (!response.IsSuccessStatusCode)
		{
			var error = await response.Content.ReadAsStringAsync(ct);
			throw new ApiException($"API error: {error}", response.StatusCode);
		}

		// For void methods, return default
		if (typeof(T) == typeof(object))
		{
			return default!;
		}

		var content = await response.Content.ReadFromJsonAsync<T>(_jsonOptions, ct);
		return content ?? throw new ApiException("Response deserialized to null");
	}
}

// Internal response wrapper types
file sealed record ApiErrorResponse(string Error, ParseError[]? ParseErrors = null);
file sealed record MemoryResponse(byte[] Data);
file sealed record DisassemblyResponse(DisassemblyInstruction[] Instructions);
file sealed record SourceMapResponse(SourceMapEntry[] Entries);
file sealed record BreakpointsResponse(uint[] Breakpoints);
file sealed record WatchpointsResponse(Watchpoint[] Watchpoints);
file sealed record EvaluationResponse(uint Value);
file sealed record ExamplesResponse(ExampleInfo[] Examples);
