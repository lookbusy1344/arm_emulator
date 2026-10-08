using System.Net;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization.Metadata;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// HTTP client for the ARM Emulator REST API (api/server.go).
/// Wire records in <see cref="ApiJsonContext"/> mirror the backend's JSON; this class maps them to domain models.
/// </summary>
public sealed class ApiClient(HttpClient http) : IApiClient
{
	private const string EmptyJsonObject = "{}";
	private const string BackendUnreachableMessage = "Cannot connect to backend - is the emulator running?";

	// Session Management

	public async Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default)
	{
		var response = await SendAsync(HttpMethod.Post, "/api/v1/session", new StringContent(EmptyJsonObject, Encoding.UTF8, "application/json"), ct);
		return await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.SessionInfo, ct);
	}

	public async Task<VMStatus> GetStatusAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}", ct);
		var status = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.SessionStatusResponse, ct, sessionId);
		return status.ToVMStatus();
	}

	public async Task DestroySessionAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await DeleteAsync($"/api/v1/session/{sessionId}", ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	// Program Loading

	public async Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default)
	{
		var response = await PostJsonAsync(
			$"/api/v1/session/{sessionId}/load", new LoadProgramRequest(source), ApiJsonContext.Default.LoadProgramRequest, ct);

		// Assembler errors arrive as 400 with a LoadProgramResponse body; other 400s carry the generic error body
		if (response.StatusCode == HttpStatusCode.BadRequest) {
			var body = await response.Content.ReadAsStringAsync(ct);
			var failure = TryDeserialize(body, ApiJsonContext.Default.LoadProgramWireResponse);
			if (failure is not null && failure.Errors.Count > 0) {
				throw new ProgramLoadException([.. failure.Errors]);
			}

			throw CreateApiException(body, response.StatusCode);
		}

		var result = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.LoadProgramWireResponse, ct, sessionId);
		if (!result.Success) {
			throw new ProgramLoadException([.. result.Errors]);
		}

		return new LoadProgramResponse(result.Symbols);
	}

	// Execution Control

	public Task RunAsync(string sessionId, CancellationToken ct = default) =>
		PostCommandAsync(sessionId, "run", ct);

	public Task StopAsync(string sessionId, CancellationToken ct = default) =>
		PostCommandAsync(sessionId, "stop", ct);

	public Task<RegisterState> StepAsync(string sessionId, CancellationToken ct = default) =>
		PostForRegistersAsync(sessionId, "step", ct);

	public Task<RegisterState> StepOverAsync(string sessionId, CancellationToken ct = default) =>
		PostForRegistersAsync(sessionId, "step-over", ct);

	public Task<RegisterState> StepOutAsync(string sessionId, CancellationToken ct = default) =>
		PostForRegistersAsync(sessionId, "step-out", ct);

	public Task ResetAsync(string sessionId, CancellationToken ct = default) =>
		PostCommandAsync(sessionId, "reset", ct);

	public Task RestartAsync(string sessionId, CancellationToken ct = default) =>
		PostCommandAsync(sessionId, "restart", ct);

	// State Inspection

	public async Task<RegisterState> GetRegistersAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/registers", ct);
		var registers = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.RegistersResponse, ct, sessionId);
		return registers.ToRegisterState();
	}

	public async Task<ImmutableArray<byte>> GetMemoryAsync(string sessionId, uint address, int length, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/memory?address={address}&length={length}", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.MemoryResponse, ct, sessionId);
		return wrapper.ToBytes();
	}

	public async Task<ImmutableArray<DisassemblyInstruction>> GetDisassemblyAsync(string sessionId, uint address, int count, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/disassembly?address={address}&count={count}", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.DisassemblyResponse, ct, sessionId);
		return [.. wrapper.Instructions.Select(i => i.ToModel())];
	}

	public async Task<ImmutableArray<SourceMapEntry>> GetSourceMapAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/sourcemap", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.SourceMapResponse, ct, sessionId);
		return [.. wrapper.SourceMap.Select(e => e.ToModel())];
	}

	// Breakpoints

	public async Task AddBreakpointAsync(string sessionId, uint address, CancellationToken ct = default)
	{
		var response = await PostJsonAsync(
			$"/api/v1/session/{sessionId}/breakpoint", new BreakpointRequest(address), ApiJsonContext.Default.BreakpointRequest, ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	public async Task RemoveBreakpointAsync(string sessionId, uint address, CancellationToken ct = default)
	{
		var response = await SendAsync(
			HttpMethod.Delete, $"/api/v1/session/{sessionId}/breakpoint", JsonContent(new BreakpointRequest(address), ApiJsonContext.Default.BreakpointRequest), ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	public async Task<ImmutableArray<uint>> GetBreakpointsAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/breakpoints", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.BreakpointsResponse, ct, sessionId);
		return [.. wrapper.Breakpoints];
	}

	// Watchpoints

	public async Task<Watchpoint> AddWatchpointAsync(string sessionId, uint address, WatchpointType type, CancellationToken ct = default)
	{
		var response = await PostJsonAsync(
			$"/api/v1/session/{sessionId}/watchpoint",
			new AddWatchpointRequest(address, WireFormat.ToWire(type)),
			ApiJsonContext.Default.AddWatchpointRequest,
			ct);
		var watchpoint = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.WatchpointWire, ct, sessionId);
		return watchpoint.ToModel();
	}

	public async Task RemoveWatchpointAsync(string sessionId, int watchpointId, CancellationToken ct = default)
	{
		var response = await DeleteAsync($"/api/v1/session/{sessionId}/watchpoint/{watchpointId}", ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	public async Task<ImmutableArray<Watchpoint>> GetWatchpointsAsync(string sessionId, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/session/{sessionId}/watchpoints", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.WatchpointsResponse, ct, sessionId);
		return [.. wrapper.Watchpoints.Select(w => w.ToModel())];
	}

	// Expression Evaluation

	public async Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default)
	{
		var response = await PostJsonAsync(
			$"/api/v1/session/{sessionId}/evaluate",
			new EvaluateExpressionRequest(expression),
			ApiJsonContext.Default.EvaluateExpressionRequest,
			ct);

		if (response.StatusCode == HttpStatusCode.BadRequest) {
			var body = await response.Content.ReadAsStringAsync(ct);
			var error = TryDeserialize(body, ApiJsonContext.Default.ApiErrorResponse);
			throw new ExpressionEvaluationException(expression, error?.Message ?? error?.Error ?? body);
		}

		var result = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.EvaluationResponse, ct, sessionId);
		return result.Result;
	}

	// Input

	public async Task SendStdinAsync(string sessionId, string data, CancellationToken ct = default)
	{
		var response = await PostJsonAsync(
			$"/api/v1/session/{sessionId}/stdin", new StdinRequest(data), ApiJsonContext.Default.StdinRequest, ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	// Version

	public async Task<BackendVersion> GetVersionAsync(CancellationToken ct = default)
	{
		var response = await GetAsync("/api/v1/version", ct);
		var version = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.VersionResponse, ct);
		return version.ToModel();
	}

	// Examples

	public async Task<ImmutableArray<ExampleInfo>> GetExamplesAsync(CancellationToken ct = default)
	{
		var response = await GetAsync("/api/v1/examples", ct);
		var wrapper = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.ExamplesResponse, ct);
		return [.. wrapper.Examples];
	}

	public async Task<string> GetExampleContentAsync(string name, CancellationToken ct = default)
	{
		var response = await GetAsync($"/api/v1/examples/{name}", ct);
		if (!response.IsSuccessStatusCode) {
			throw new ApiException($"Example '{name}' not found", response.StatusCode);
		}

		var example = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.ExampleContentResponse, ct);
		return example.Content;
	}

	// Helper Methods

	private async Task PostCommandAsync(string sessionId, string action, CancellationToken ct)
	{
		var response = await SendAsync(HttpMethod.Post, $"/api/v1/session/{sessionId}/{action}", null, ct);
		await EnsureSuccessAsync(response, sessionId, ct);
	}

	private async Task<RegisterState> PostForRegistersAsync(string sessionId, string action, CancellationToken ct)
	{
		var response = await SendAsync(HttpMethod.Post, $"/api/v1/session/{sessionId}/{action}", null, ct);
		var registers = await ParseResponseOrThrowAsync(response, ApiJsonContext.Default.RegistersResponse, ct, sessionId);
		return registers.ToRegisterState();
	}

	private Task<HttpResponseMessage> PostJsonAsync<T>(string path, T body, JsonTypeInfo<T> typeInfo, CancellationToken ct) =>
		SendAsync(HttpMethod.Post, path, JsonContent(body, typeInfo), ct);

	private Task<HttpResponseMessage> GetAsync(string path, CancellationToken ct) => SendAsync(HttpMethod.Get, path, null, ct);

	private Task<HttpResponseMessage> DeleteAsync(string path, CancellationToken ct) => SendAsync(HttpMethod.Delete, path, null, ct);

	/// <summary>Sends one request, taking ownership of <paramref name="content"/>. A transport failure becomes <see cref="BackendUnavailableException"/>.</summary>
	private async Task<HttpResponseMessage> SendAsync(HttpMethod method, string path, HttpContent? content, CancellationToken ct)
	{
		using var request = new HttpRequestMessage(method, path) { Content = content };
		try {
			return await http.SendAsync(request, ct);
		}
		catch (HttpRequestException ex) {
			throw new BackendUnavailableException(BackendUnreachableMessage, ex);
		}
	}

	private static ByteArrayContent JsonContent<T>(T body, JsonTypeInfo<T> typeInfo)
	{
		var content = new ByteArrayContent(JsonSerializer.SerializeToUtf8Bytes(body, typeInfo));
		content.Headers.ContentType = new("application/json");
		return content;
	}

	private static async Task EnsureSuccessAsync(HttpResponseMessage response, string sessionId, CancellationToken ct)
	{
		ThrowIfSessionNotFound(response, sessionId);
		if (!response.IsSuccessStatusCode) {
			throw CreateApiException(await response.Content.ReadAsStringAsync(ct), response.StatusCode);
		}
	}

	private static async Task<T> ParseResponseOrThrowAsync<T>(
		HttpResponseMessage response,
		JsonTypeInfo<T> jsonTypeInfo,
		CancellationToken ct,
		string? sessionId = null)
	{
		if (sessionId is not null) {
			ThrowIfSessionNotFound(response, sessionId);
		}

		if (!response.IsSuccessStatusCode) {
			throw CreateApiException(await response.Content.ReadAsStringAsync(ct), response.StatusCode);
		}

		var stream = await response.Content.ReadAsStreamAsync(ct);
		var content = await JsonSerializer.DeserializeAsync(stream, jsonTypeInfo, ct);
		return content ?? throw new ApiException("Response deserialized to null");
	}

	private static void ThrowIfSessionNotFound(HttpResponseMessage response, string sessionId)
	{
		if (response.StatusCode == HttpStatusCode.NotFound) {
			throw new SessionNotFoundException(sessionId);
		}
	}

	private static ApiException CreateApiException(string body, HttpStatusCode statusCode)
	{
		var error = TryDeserialize(body, ApiJsonContext.Default.ApiErrorResponse);
		return new ApiException($"API error: {error?.Message ?? body}", statusCode);
	}

	// Error bodies are not guaranteed to be JSON (http.Error writes plain text)
	private static T? TryDeserialize<T>(string body, JsonTypeInfo<T> typeInfo) where T : class
	{
		try {
			return JsonSerializer.Deserialize(body, typeInfo);
		}
		catch (JsonException) {
			return null;
		}
	}
}
