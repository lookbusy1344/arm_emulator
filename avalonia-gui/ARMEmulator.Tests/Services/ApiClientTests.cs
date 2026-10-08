using System.Net;
using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json.Nodes;
using ARMEmulator.Models;
using ARMEmulator.Services;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// ApiClient tests. Response fixtures are JSON captured from the Go backend (api/handlers.go);
/// request assertions check the shapes the backend's handlers decode.
/// </summary>
public sealed class ApiClientTests : IDisposable
{
	private const string SessionId = "session-123";

	private readonly HttpClient httpClient;
	private readonly TestHttpMessageHandler handler;
	private readonly ApiClient apiClient;

	public ApiClientTests()
	{
		handler = new TestHttpMessageHandler();
		httpClient = new HttpClient(handler) { BaseAddress = new Uri("http://localhost:8080") };
		apiClient = new ApiClient(httpClient);
	}

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	// Request bodies are compared as normalised JSON: escaping of characters such as + and ' is not part of the contract.
	private static string Json(string json) => JsonNode.Parse(json)!.ToJsonString();

	private string? RequestJson => handler.LastRequestBody is null ? null : Json(handler.LastRequestBody);

	// Session

	[Fact]
	public async Task CreateSessionAsync_PostsEmptyJsonObjectAndParsesSessionId()
	{
		handler.SetResponse(HttpStatusCode.Created, """{"sessionId":"246937e3ce77045eeddff1b78984d826","createdAt":"2026-10-06T13:44:40.096176+01:00"}""");

		var result = await apiClient.CreateSessionAsync(Ct);

		_ = result.SessionId.Should().Be("246937e3ce77045eeddff1b78984d826");
		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Post);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session");
		_ = RequestJson.Should().Be(Json("{}"));
		_ = handler.LastRequestContentType.Should().Be("application/json");
	}

	[Fact]
	public async Task CreateSessionAsync_WhenBackendUnreachable_ThrowsBackendUnavailableException()
	{
		handler.SetException(new HttpRequestException("Connection refused"));

		var act = async () => await apiClient.CreateSessionAsync(Ct);

		_ = await act.Should().ThrowAsync<BackendUnavailableException>()
			.WithMessage("*Cannot connect to backend*");
	}

	public static TheoryData<string> EveryOperation => [.. Operations.Keys];

	private static readonly ImmutableDictionary<string, Func<ApiClient, CancellationToken, Task>> Operations =
		new Dictionary<string, Func<ApiClient, CancellationToken, Task>> {
			[nameof(ApiClient.CreateSessionAsync)] = (api, ct) => api.CreateSessionAsync(ct),
			[nameof(ApiClient.GetStatusAsync)] = (api, ct) => api.GetStatusAsync(SessionId, ct),
			[nameof(ApiClient.DestroySessionAsync)] = (api, ct) => api.DestroySessionAsync(SessionId, ct),
			[nameof(ApiClient.LoadProgramAsync)] = (api, ct) => api.LoadProgramAsync(SessionId, "MOV R0, #1", ct),
			[nameof(ApiClient.RunAsync)] = (api, ct) => api.RunAsync(SessionId, ct),
			[nameof(ApiClient.StopAsync)] = (api, ct) => api.StopAsync(SessionId, ct),
			[nameof(ApiClient.StepAsync)] = (api, ct) => api.StepAsync(SessionId, ct),
			[nameof(ApiClient.StepOverAsync)] = (api, ct) => api.StepOverAsync(SessionId, ct),
			[nameof(ApiClient.StepOutAsync)] = (api, ct) => api.StepOutAsync(SessionId, ct),
			[nameof(ApiClient.ResetAsync)] = (api, ct) => api.ResetAsync(SessionId, ct),
			[nameof(ApiClient.RestartAsync)] = (api, ct) => api.RestartAsync(SessionId, ct),
			[nameof(ApiClient.GetRegistersAsync)] = (api, ct) => api.GetRegistersAsync(SessionId, ct),
			[nameof(ApiClient.GetMemoryAsync)] = (api, ct) => api.GetMemoryAsync(SessionId, 0x8000, 4, ct),
			[nameof(ApiClient.GetDisassemblyAsync)] = (api, ct) => api.GetDisassemblyAsync(SessionId, 0x8000, 2, ct),
			[nameof(ApiClient.GetSourceMapAsync)] = (api, ct) => api.GetSourceMapAsync(SessionId, ct),
			[nameof(ApiClient.AddBreakpointAsync)] = (api, ct) => api.AddBreakpointAsync(SessionId, 0x8000, ct),
			[nameof(ApiClient.RemoveBreakpointAsync)] = (api, ct) => api.RemoveBreakpointAsync(SessionId, 0x8000, ct),
			[nameof(ApiClient.GetBreakpointsAsync)] = (api, ct) => api.GetBreakpointsAsync(SessionId, ct),
			[nameof(ApiClient.AddWatchpointAsync)] = (api, ct) => api.AddWatchpointAsync(SessionId, 0x9000, WatchpointType.Write, ct),
			[nameof(ApiClient.RemoveWatchpointAsync)] = (api, ct) => api.RemoveWatchpointAsync(SessionId, 1, ct),
			[nameof(ApiClient.GetWatchpointsAsync)] = (api, ct) => api.GetWatchpointsAsync(SessionId, ct),
			[nameof(ApiClient.EvaluateExpressionAsync)] = (api, ct) => api.EvaluateExpressionAsync(SessionId, "r0", ct),
			[nameof(ApiClient.SendStdinAsync)] = (api, ct) => api.SendStdinAsync(SessionId, "x", ct),
			[nameof(ApiClient.GetVersionAsync)] = (api, ct) => api.GetVersionAsync(ct),
			[nameof(ApiClient.GetExamplesAsync)] = (api, ct) => api.GetExamplesAsync(ct),
			[nameof(ApiClient.GetExampleContentAsync)] = (api, ct) => api.GetExampleContentAsync("hello.s", ct)
		}.ToImmutableDictionary();

	[Theory]
	[MemberData(nameof(EveryOperation))]
	public async Task EveryOperation_WhenBackendUnreachable_ThrowsBackendUnavailableWrappingTheCause(string operation)
	{
		var cause = new HttpRequestException("Connection refused");
		handler.SetException(cause);

		var act = () => Operations[operation](apiClient, Ct);

		var exception = await act.Should().ThrowExactlyAsync<BackendUnavailableException>();
		_ = exception.Which.Message.Should().Be("Cannot connect to backend - is the emulator running?");
		_ = exception.Which.InnerException.Should().BeSameAs(cause);
	}

	[Fact]
	public void EveryOperation_CoversEveryInterfaceMethod()
	{
		_ = Operations.Keys.Should().BeEquivalentTo(typeof(IApiClient).GetMethods().Select(m => m.Name));
	}

	[Fact]
	public async Task GetStatusAsync_GetsSessionRouteAndParsesHaltedWithoutWrite()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"sessionId":"s","state":"halted","pc":32768,"cycles":7,"hasWrite":false}""");

		var result = await apiClient.GetStatusAsync(SessionId, Ct);

		_ = result.Should().Be(new VMStatus(VMState.Halted, 0x8000, 7));
		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Get);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123");
	}

	[Fact]
	public async Task GetStatusAsync_WithWrite_ParsesLastWrite()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"sessionId":"s","state":"breakpoint","pc":32780,"cycles":3,"hasWrite":true,"writeAddr":36864,"writeSize":4}""");

		var result = await apiClient.GetStatusAsync(SessionId, Ct);

		_ = result.State.Should().Be(VMState.Breakpoint);
		_ = result.LastWrite.Should().Be(new MemoryWrite(0x9000, 4));
	}

	[Theory]
	[InlineData("running", VMState.Running)]
	[InlineData("halted", VMState.Halted)]
	[InlineData("breakpoint", VMState.Breakpoint)]
	[InlineData("error", VMState.Error)]
	[InlineData("waiting_for_input", VMState.WaitingForInput)]
	public async Task GetStatusAsync_MapsEveryBackendState(string wire, VMState expected)
	{
		handler.SetResponse(HttpStatusCode.OK, $$"""{"state":"{{wire}}","pc":0,"cycles":0,"hasWrite":false}""");

		var result = await apiClient.GetStatusAsync(SessionId, Ct);

		_ = result.State.Should().Be(expected);
	}

	[Fact]
	public async Task GetStatusAsync_WithUnknownState_ThrowsFormatException()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"state":"exploded","pc":0,"cycles":0,"hasWrite":false}""");

		var act = async () => await apiClient.GetStatusAsync(SessionId, Ct);

		_ = await act.Should().ThrowAsync<FormatException>().WithMessage("*exploded*");
	}

	[Fact]
	public async Task GetStatusAsync_WithInvalidSession_ThrowsSessionNotFoundException()
	{
		handler.SetResponse(HttpStatusCode.NotFound, """{"error":"Not Found","message":"Session not found","code":404}""");

		var act = async () => await apiClient.GetStatusAsync("invalid-session", Ct);

		_ = await act.Should().ThrowAsync<SessionNotFoundException>()
			.Where(ex => ex.SessionId == "invalid-session");
	}

	// Program loading

	[Fact]
	public async Task LoadProgramAsync_PostsSourceAsJsonAndParsesSymbols()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"success":true,"symbols":{"_start":32768,"loop":32776}}""");

		var result = await apiClient.LoadProgramAsync(SessionId, "MOV R0, #1\n", Ct);

		_ = result.Symbols.Should().BeEquivalentTo(new Dictionary<string, uint> { ["_start"] = 0x8000, ["loop"] = 0x8008 });
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/load");
		_ = RequestJson.Should().Be(Json("""{"source":"MOV R0, #1\n"}"""));
		_ = handler.LastRequestContentType.Should().Be("application/json");
	}

	[Fact]
	public async Task LoadProgramAsync_SourceWithSpecialCharacters_RoundTripsExactly()
	{
		const string source = "\tMOV R0, #'A'\t; \"quote\" a+b <x> & ü\r\n";
		handler.SetResponse(HttpStatusCode.OK, """{"success":true}""");

		_ = await apiClient.LoadProgramAsync(SessionId, source, Ct);

		_ = JsonNode.Parse(handler.LastRequestBody!)!["source"]!.GetValue<string>().Should().Be(source);
	}

	[Fact]
	public async Task LoadProgramAsync_WithoutSymbols_ReturnsEmptySymbols()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"success":true}""");

		var result = await apiClient.LoadProgramAsync(SessionId, "", Ct);

		_ = result.Symbols.Should().BeEmpty();
	}

	[Fact]
	public async Task LoadProgramAsync_WithAssemblerErrors_ThrowsProgramLoadExceptionWithEveryError()
	{
		handler.SetResponse(HttpStatusCode.BadRequest,
			"""{"success":false,"errors":["api:2:2: unknown instruction: FOO","api:3:1: undefined label: bar"]}""");

		var act = async () => await apiClient.LoadProgramAsync(SessionId, "FOO R0", Ct);

		var exception = await act.Should().ThrowAsync<ProgramLoadException>();
		_ = exception.Which.Errors.Should().Equal("api:2:2: unknown instruction: FOO", "api:3:1: undefined label: bar");
		_ = exception.Which.StatusCode.Should().Be(HttpStatusCode.BadRequest);
	}

	[Fact]
	public async Task LoadProgramAsync_WithMalformedRequestError_ThrowsPlainApiException()
	{
		handler.SetResponse(HttpStatusCode.BadRequest, """{"error":"Bad Request","message":"Invalid request body","code":400}""");

		var act = async () => await apiClient.LoadProgramAsync(SessionId, "x", Ct);

		var exception = await act.Should().ThrowExactlyAsync<ApiException>();
		_ = exception.Which.Message.Should().Contain("Invalid request body");
		_ = exception.Which.StatusCode.Should().Be(HttpStatusCode.BadRequest);
	}

	// Execution and registers

	private const string RegistersJson =
		"""{"r0":5,"r1":36864,"r2":2,"r3":3,"r4":4,"r5":5,"r6":6,"r7":7,"r8":8,"r9":9,"r10":10,"r11":11,"r12":12,"sp":327680,"lr":32800,"pc":32772,"cpsr":{"n":true,"z":false,"c":true,"v":false},"cycles":1}""";

	private static readonly RegisterState ExpectedRegisters = RegisterState.Create(
		r0: 5, r1: 0x9000, r2: 2, r3: 3, r4: 4, r5: 5, r6: 6, r7: 7, r8: 8, r9: 9, r10: 10, r11: 11, r12: 12,
		sp: 0x50000, lr: 0x8020, pc: 0x8004, cpsr: new CPSRFlags(N: true, Z: false, C: true, V: false));

	[Fact]
	public async Task StepAsync_PostsStepAndParsesFlatRegisters()
	{
		handler.SetResponse(HttpStatusCode.OK, RegistersJson);

		var result = await apiClient.StepAsync(SessionId, Ct);

		_ = result.Should().Be(ExpectedRegisters);
		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Post);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/step");
	}

	[Fact]
	public async Task StepOverAsync_PostsStepOverAndParsesRegisters()
	{
		handler.SetResponse(HttpStatusCode.OK, RegistersJson);

		var result = await apiClient.StepOverAsync(SessionId, Ct);

		_ = result.Should().Be(ExpectedRegisters);
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/step-over");
	}

	[Fact]
	public async Task StepOutAsync_PostsStepOutAndParsesRegisters()
	{
		handler.SetResponse(HttpStatusCode.OK, RegistersJson);

		var result = await apiClient.StepOutAsync(SessionId, Ct);

		_ = result.Should().Be(ExpectedRegisters);
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/step-out");
	}

	[Fact]
	public async Task GetRegistersAsync_GetsRegistersAndParsesFlatRegisters()
	{
		handler.SetResponse(HttpStatusCode.OK, RegistersJson);

		var result = await apiClient.GetRegistersAsync(SessionId, Ct);

		_ = result.Should().Be(ExpectedRegisters);
		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Get);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/registers");
	}

	[Fact]
	public async Task StepOverAsync_WithServerError_ThrowsApiExceptionWithBackendMessage()
	{
		handler.SetResponse(HttpStatusCode.InternalServerError,
			"""{"error":"Internal Server Error","message":"Step over failed: no program loaded","code":500}""");

		var act = async () => await apiClient.StepOverAsync(SessionId, Ct);

		var exception = await act.Should().ThrowExactlyAsync<ApiException>();
		_ = exception.Which.Message.Should().Contain("Step over failed: no program loaded");
		_ = exception.Which.StatusCode.Should().Be(HttpStatusCode.InternalServerError);
	}

	// Inspection

	[Fact]
	public async Task GetMemoryAsync_DecodesBase64Data()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"address":32768,"data":"AQIDBA==","length":4}""");

		var result = await apiClient.GetMemoryAsync(SessionId, 0x8000, 4, Ct);

		_ = result.Should().Equal(0x01, 0x02, 0x03, 0x04);
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/memory?address=32768&length=4");
	}

	[Fact]
	public async Task GetDisassemblyAsync_MapsDisassemblyTextToTrimmedMnemonic()
	{
		handler.SetResponse(HttpStatusCode.OK,
			"""{"instructions":[{"address":32768,"machineCode":3818913797,"disassembly":"  MOV R0, #5","symbol":"_start"},{"address":32772,"machineCode":3818921104,"disassembly":"  LDR R1, =0x9000"}]}""");

		var result = await apiClient.GetDisassemblyAsync(SessionId, 0x8000, 2, Ct);

		_ = result.Should().Equal(
			new DisassemblyInstruction(0x8000, 0xE3A00005, "MOV R0, #5", "_start"),
			new DisassemblyInstruction(0x8004, 0xE3A01C90, "LDR R1, =0x9000"));
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/disassembly?address=32768&count=2");
	}

	[Fact]
	public async Task GetSourceMapAsync_GetsSourcemapRouteAndParsesEntries()
	{
		handler.SetResponse(HttpStatusCode.OK,
			"""{"sourceMap":[{"address":32768,"lineNumber":3,"line":"  MOV R0, #5"},{"address":32772,"lineNumber":4,"line":"  LDR R1, =0x9000"}]}""");

		var result = await apiClient.GetSourceMapAsync(SessionId, Ct);

		_ = result.Should().Equal(
			new SourceMapEntry(0x8000, 3, "  MOV R0, #5"),
			new SourceMapEntry(0x8004, 4, "  LDR R1, =0x9000"));
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/sourcemap");
	}

	// Breakpoints and watchpoints

	[Fact]
	public async Task AddBreakpointAsync_PostsAddressAsJson()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"success":true,"message":"Breakpoint added"}""");

		await apiClient.AddBreakpointAsync(SessionId, 0x8004, Ct);

		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Post);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/breakpoint");
		_ = RequestJson.Should().Be(Json("""{"address":32772}"""));
	}

	[Fact]
	public async Task RemoveBreakpointAsync_DeletesWithAddressInBody()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"success":true,"message":"Breakpoint removed"}""");

		await apiClient.RemoveBreakpointAsync(SessionId, 0x8004, Ct);

		_ = handler.LastRequest!.Method.Should().Be(HttpMethod.Delete);
		_ = handler.LastRequest.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/breakpoint");
		_ = RequestJson.Should().Be(Json("""{"address":32772}"""));
	}

	[Fact]
	public async Task GetBreakpointsAsync_ParsesAddresses()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"breakpoints":[32772,32780]}""");

		var result = await apiClient.GetBreakpointsAsync(SessionId, Ct);

		_ = result.Should().Equal(0x8004u, 0x800Cu);
	}

	[Fact]
	public async Task AddWatchpointAsync_PostsLowercaseTypeAndParsesWatchpoint()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"id":1,"address":36864,"type":"write"}""");

		var result = await apiClient.AddWatchpointAsync(SessionId, 0x9000, WatchpointType.Write, Ct);

		_ = result.Should().Be(new Watchpoint(1, 0x9000, WatchpointType.Write));
		_ = RequestJson.Should().Be(Json("""{"address":36864,"type":"write"}"""));
	}

	[Fact]
	public async Task GetWatchpointsAsync_ParsesEveryType()
	{
		handler.SetResponse(HttpStatusCode.OK,
			"""{"watchpoints":[{"id":1,"address":36864,"type":"read","enabled":true},{"id":2,"address":36868,"type":"readwrite","enabled":true}]}""");

		var result = await apiClient.GetWatchpointsAsync(SessionId, Ct);

		_ = result.Should().Equal(
			new Watchpoint(1, 0x9000, WatchpointType.Read),
			new Watchpoint(2, 0x9004, WatchpointType.ReadWrite));
	}

	// Evaluation and input

	[Fact]
	public async Task EvaluateExpressionAsync_PostsExpressionAndParsesResult()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"result":42}""");

		var result = await apiClient.EvaluateExpressionAsync(SessionId, "r0+1", Ct);

		_ = result.Should().Be(42u);
		_ = RequestJson.Should().Be(Json("""{"expression":"r0+1"}"""));
	}

	[Fact]
	public async Task EvaluateExpressionAsync_WithInvalidExpression_ThrowsWithBackendMessage()
	{
		handler.SetResponse(HttpStatusCode.BadRequest,
			"""{"error":"Bad Request","message":"Evaluation failed: failed to evaluate expression: unexpected token:  (EOF)","code":400}""");

		var act = async () => await apiClient.EvaluateExpressionAsync(SessionId, "r0+", Ct);

		var exception = await act.Should().ThrowAsync<ExpressionEvaluationException>();
		_ = exception.Which.Expression.Should().Be("r0+");
		_ = exception.Which.Message.Should().Contain("unexpected token");
	}

	[Fact]
	public async Task SendStdinAsync_PostsDataAsJson()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"success":true,"message":"Stdin sent"}""");

		await apiClient.SendStdinAsync(SessionId, "hi\n", Ct);

		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/stdin");
		_ = RequestJson.Should().Be(Json("""{"data":"hi\n"}"""));
		_ = handler.LastRequestContentType.Should().Be("application/json");
	}

	// Version and examples

	[Fact]
	public async Task GetVersionAsync_MapsDateToBuildDate()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"version":"v2.1.1-110-g55a024e","commit":"55a024e","date":"2026-10-06T12:44:28Z"}""");

		var result = await apiClient.GetVersionAsync(Ct);

		_ = result.Should().Be(new BackendVersion("v2.1.1-110-g55a024e", "55a024e", "2026-10-06T12:44:28Z"));
	}

	[Fact]
	public async Task GetExamplesAsync_ParsesNameAndSizeWithoutDescription()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"examples":[{"name":"add_128bit.s","size":4384},{"name":"hello.s","size":512}]}""");

		var result = await apiClient.GetExamplesAsync(Ct);

		_ = result.Should().Equal(new ExampleInfo("add_128bit.s", null, 4384), new ExampleInfo("hello.s", null, 512));
	}

	[Fact]
	public async Task GetExampleContentAsync_ReturnsContentField()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"name":"hello.s","content":"; hello.s\n_start:\n"}""");

		var result = await apiClient.GetExampleContentAsync("hello.s", Ct);

		_ = result.Should().Be("; hello.s\n_start:\n");
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/examples/hello.s");
	}

	[Fact]
	public async Task GetExampleContentAsync_WhenMissing_ThrowsNotFoundWithStatus()
	{
		handler.SetResponse(HttpStatusCode.NotFound, "404 page not found");

		var act = async () => await apiClient.GetExampleContentAsync("missing.s", Ct);

		var exception = await act.Should().ThrowExactlyAsync<ApiException>();
		_ = exception.Which.Message.Should().Be("Example 'missing.s' not found.");
		_ = exception.Which.StatusCode.Should().Be(HttpStatusCode.NotFound);
	}

	[Fact]
	public async Task GetExampleContentAsync_WithServerError_ThrowsWithBackendMessage()
	{
		handler.SetResponse(HttpStatusCode.InternalServerError, """{"error":"Internal Server Error","message":"Failed to read example","code":500}""");

		var act = async () => await apiClient.GetExampleContentAsync("hello.s", Ct);

		var exception = await act.Should().ThrowExactlyAsync<ApiException>();
		_ = exception.Which.Message.Should().Be("API error: Failed to read example");
		_ = exception.Which.StatusCode.Should().Be(HttpStatusCode.InternalServerError);
	}

	[Fact]
	public async Task GetExampleContentAsync_EscapesTheNameAsOnePathSegment()
	{
		handler.SetResponse(HttpStatusCode.OK, """{"name":"x","content":""}""");

		_ = await apiClient.GetExampleContentAsync("my file#1/../x.s", Ct);

		_ = handler.LastRequest!.RequestUri!.AbsolutePath.Should().Be("/api/v1/examples/my%20file%231%2F..%2Fx.s");
	}

	[Fact]
	public async Task SessionRoutes_EscapeTheSessionIdAsOnePathSegment()
	{
		handler.SetResponse(HttpStatusCode.OK, RegistersJson);

		_ = await apiClient.GetRegistersAsync("a/b?c", Ct);

		_ = handler.LastRequest!.RequestUri!.AbsolutePath.Should().Be("/api/v1/session/a%2Fb%3Fc/registers");
	}

	[Theory]
	[MemberData(nameof(EveryOperation))]
	public async Task EveryOperation_DisposesTheResponse(string operation)
	{
		handler.SetResponse(HttpStatusCode.InternalServerError, """{"error":"Internal Server Error","message":"boom","code":500}""");

		var act = () => Operations[operation](apiClient, Ct);

		_ = await act.Should().ThrowAsync<ApiException>();
		_ = handler.LastResponseDisposed.Should().BeTrue();
	}

	public void Dispose()
	{
		httpClient.Dispose();
		handler.Dispose();
	}
}

/// <summary>
/// Test HTTP message handler that records the last request and returns a canned response.
/// </summary>
internal sealed class TestHttpMessageHandler : HttpMessageHandler
{
	private HttpStatusCode statusCode = HttpStatusCode.OK;
	private string content = "{}";
	private Exception? exception;

	public HttpRequestMessage? LastRequest { get; private set; }

	public string? LastRequestBody { get; private set; }

	public string? LastRequestContentType { get; private set; }

	public void SetResponse(HttpStatusCode statusCode, string content)
	{
		this.statusCode = statusCode;
		this.content = content;
		exception = null;
	}

	public void SetException(Exception exception)
	{
		this.exception = exception;
	}

	protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
	{
		LastRequest = request;
		LastRequestBody = request.Content is null ? null : await request.Content.ReadAsStringAsync(cancellationToken);
		LastRequestContentType = request.Content?.Headers.ContentType?.MediaType;

		if (exception is not null) {
			throw exception;
		}

		lastResponseDisposed = new StrongBox<bool>();
		return new TrackedResponse(statusCode, lastResponseDisposed) { Content = new StringContent(content, Encoding.UTF8, "application/json") };
	}

	private StrongBox<bool>? lastResponseDisposed;

	/// <summary>True when the caller disposed the last response it received.</summary>
	public bool LastResponseDisposed => lastResponseDisposed?.Value ?? false;

	private sealed class TrackedResponse(HttpStatusCode statusCode, StrongBox<bool> disposed) : HttpResponseMessage(statusCode)
	{
		protected override void Dispose(bool disposing)
		{
			disposed.Value = true;
			base.Dispose(disposing);
		}
	}
}
