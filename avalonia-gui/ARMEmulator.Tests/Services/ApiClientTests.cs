using System.Diagnostics.CodeAnalysis;
using System.Net;
using System.Text;
using System.Text.Json;
using ARMEmulator.Models;
using ARMEmulator.Services;
using FluentAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Services;

public sealed class ApiClientTests : IDisposable
{
	private readonly HttpClient httpClient;
	private readonly TestHttpMessageHandler handler;
	private readonly ApiClient apiClient;

	public ApiClientTests()
	{
		handler = new TestHttpMessageHandler();
		httpClient = new HttpClient(handler) { BaseAddress = new Uri("http://localhost:8080") };
		apiClient = new ApiClient(httpClient);
	}

	[Fact]
	public async Task CreateSessionAsync_WithSuccessResponse_ReturnsSessionInfo()
	{
		var sessionInfo = new SessionInfo("session-123");
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(sessionInfo, ApiJsonContext.Default.SessionInfo));

		var result = await apiClient.CreateSessionAsync();

		_ = result.SessionId.Should().Be("session-123");
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session");
		_ = handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task CreateSessionAsync_WhenBackendUnreachable_ThrowsBackendUnavailableException()
	{
		handler.SetException(new HttpRequestException("Connection refused"));

		var act = async () => await apiClient.CreateSessionAsync();

		_ = await act.Should().ThrowAsync<BackendUnavailableException>()
			.WithMessage("*Cannot connect to backend*");
	}

	[Fact]
	public async Task GetStatusAsync_WithValidSession_ReturnsStatus()
	{
		var status = new VMStatus(VMState.Idle, 0x8000, 0);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(status, ApiJsonContext.Default.VMStatus));

		var result = await apiClient.GetStatusAsync("session-123");

		_ = result.State.Should().Be(VMState.Idle);
		_ = result.PC.Should().Be(0x8000u);
	}

	[Fact]
	public async Task GetStatusAsync_WithInvalidSession_ThrowsSessionNotFoundException()
	{
		handler.SetResponse(HttpStatusCode.NotFound, "{\"error\":\"session not found\"}");

		var act = async () => await apiClient.GetStatusAsync("invalid-session");

		_ = await act.Should().ThrowAsync<SessionNotFoundException>()
			.Where(ex => ex.SessionId == "invalid-session");
	}

	[Fact]
	public async Task LoadProgramAsync_WithValidProgram_ReturnsLoadResponse()
	{
		var response = new LoadProgramResponse(true, [], 0x8000);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.LoadProgramResponse));

		var result = await apiClient.LoadProgramAsync("session-123", "MOV R0, #1");

		_ = result.Success.Should().BeTrue();
		_ = result.EntryPoint.Should().Be(0x8000u);
		_ = handler.LastRequest!.Content.Should().NotBeNull();
	}

	[Fact]
	public async Task LoadProgramAsync_WithParseErrors_ThrowsProgramLoadException()
	{
		var errorResponse = new ApiErrorResponse(
			"Parse error",
			[
				new ParseError(1, 5, "Invalid instruction"),
				new ParseError(2, 10, "Unknown register")
			]
		);
		handler.SetResponse(HttpStatusCode.BadRequest, JsonSerializer.Serialize(errorResponse, ApiJsonContext.Default.ApiErrorResponse));

		var act = async () => await apiClient.LoadProgramAsync("session-123", "INVALID");

		var exception = await act.Should().ThrowAsync<ProgramLoadException>();
		_ = exception.Which.Errors.Should().HaveCount(2);
		_ = exception.Which.Errors[0].Line.Should().Be(1);
		_ = exception.Which.Errors[0].Message.Should().Be("Invalid instruction");
	}

	[Fact]
	public async Task StepAsync_WithValidSession_ReturnsRegisters()
	{
		var registers = RegisterState.Create(r0: 42);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(registers, ApiJsonContext.Default.RegisterState));

		var result = await apiClient.StepAsync("session-123");

		_ = result.R0.Should().Be(42u);
		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/step");
		_ = handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task EvaluateExpressionAsync_WithValidExpression_ReturnsValue()
	{
		var response = new EvaluationResponse(42u);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.EvaluationResponse));

		var result = await apiClient.EvaluateExpressionAsync("session-123", "r0 + r1");

		_ = result.Should().Be(42u);
	}

	[Fact]
	public async Task EvaluateExpressionAsync_WithInvalidExpression_ThrowsExpressionEvaluationException()
	{
		var errorResponse = new ApiErrorResponse("Invalid syntax");
		handler.SetResponse(HttpStatusCode.BadRequest, JsonSerializer.Serialize(errorResponse, ApiJsonContext.Default.ApiErrorResponse));

		var act = async () => await apiClient.EvaluateExpressionAsync("session-123", "invalid");

		var exception = await act.Should().ThrowAsync<ExpressionEvaluationException>();
		_ = exception.Which.Expression.Should().Be("invalid");
	}

	[Fact]
	public async Task GetMemoryAsync_ReturnsMemoryData()
	{
		var memory = new byte[] { 0x01, 0x02, 0x03, 0x04 };
		var response = new MemoryResponse(memory);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.MemoryResponse));

		var result = await apiClient.GetMemoryAsync("session-123", 0x10000, 4);

		_ = result.Should().Equal(memory);
	}

	[Fact]
	public async Task AddBreakpointAsync_SendsCorrectRequest()
	{
		handler.SetResponse(HttpStatusCode.OK, "{}");

		await apiClient.AddBreakpointAsync("session-123", 0x8000);

		_ = handler.LastRequest!.RequestUri!.PathAndQuery.Should().Contain("/breakpoint");
		_ = handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task AddWatchpointAsync_ReturnsWatchpoint()
	{
		var watchpoint = new Watchpoint(1, 0x10000, WatchpointType.Write);
		handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(watchpoint, ApiJsonContext.Default.Watchpoint));

		var result = await apiClient.AddWatchpointAsync("session-123", 0x10000, WatchpointType.Write);

		_ = result.Id.Should().Be(1);
		_ = result.Address.Should().Be(0x10000u);
		_ = result.Type.Should().Be(WatchpointType.Write);
	}

	public void Dispose()
	{
		httpClient.Dispose();
		handler.Dispose();
	}
}

/// <summary>
/// Test HTTP message handler that allows setting responses and exceptions.
/// </summary>
internal sealed class TestHttpMessageHandler : HttpMessageHandler
{
	private HttpStatusCode statusCode = HttpStatusCode.OK;
	private string content = "{}";
	private Exception? exception;

	public HttpRequestMessage? LastRequest { get; private set; }

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

	protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
	{
		LastRequest = request;

		if (exception is not null) {
			throw exception;
		}

		var response = new HttpResponseMessage(statusCode) { Content = new StringContent(content, Encoding.UTF8, "application/json") };

		return Task.FromResult(response);
	}
}
