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
	private readonly HttpClient _httpClient;
	private readonly TestHttpMessageHandler _handler;
	private readonly ApiClient _apiClient;

	public ApiClientTests()
	{
		_handler = new TestHttpMessageHandler();
		_httpClient = new HttpClient(_handler) {
			BaseAddress = new Uri("http://localhost:8080")
		};
		_apiClient = new ApiClient(_httpClient);
	}

	[Fact]
	public async Task CreateSessionAsync_WithSuccessResponse_ReturnsSessionInfo()
	{
		var sessionInfo = new SessionInfo("session-123");
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(sessionInfo, ApiJsonContext.Default.SessionInfo));

		var result = await _apiClient.CreateSessionAsync();

		result.SessionId.Should().Be("session-123");
		_handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session");
		_handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task CreateSessionAsync_WhenBackendUnreachable_ThrowsBackendUnavailableException()
	{
		_handler.SetException(new HttpRequestException("Connection refused"));

		var act = async () => await _apiClient.CreateSessionAsync();

		await act.Should().ThrowAsync<BackendUnavailableException>()
			.WithMessage("*Cannot connect to backend*");
	}

	[Fact]
	public async Task GetStatusAsync_WithValidSession_ReturnsStatus()
	{
		var status = new VMStatus(VMState.Idle, 0x8000, 0);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(status, ApiJsonContext.Default.VMStatus));

		var result = await _apiClient.GetStatusAsync("session-123");

		result.State.Should().Be(VMState.Idle);
		result.PC.Should().Be(0x8000u);
	}

	[Fact]
	public async Task GetStatusAsync_WithInvalidSession_ThrowsSessionNotFoundException()
	{
		_handler.SetResponse(HttpStatusCode.NotFound, "{\"error\":\"session not found\"}");

		var act = async () => await _apiClient.GetStatusAsync("invalid-session");

		await act.Should().ThrowAsync<SessionNotFoundException>()
			.Where(ex => ex.SessionId == "invalid-session");
	}

	[Fact]
	public async Task LoadProgramAsync_WithValidProgram_ReturnsLoadResponse()
	{
		var response = new LoadProgramResponse(true, [], 0x8000);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.LoadProgramResponse));

		var result = await _apiClient.LoadProgramAsync("session-123", "MOV R0, #1");

		result.Success.Should().BeTrue();
		result.EntryPoint.Should().Be(0x8000u);
		_handler.LastRequest!.Content.Should().NotBeNull();
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
		_handler.SetResponse(HttpStatusCode.BadRequest, JsonSerializer.Serialize(errorResponse, ApiJsonContext.Default.ApiErrorResponse));

		var act = async () => await _apiClient.LoadProgramAsync("session-123", "INVALID");

		var exception = await act.Should().ThrowAsync<ProgramLoadException>();
		exception.Which.Errors.Should().HaveCount(2);
		exception.Which.Errors[0].Line.Should().Be(1);
		exception.Which.Errors[0].Message.Should().Be("Invalid instruction");
	}

	[Fact]
	public async Task StepAsync_WithValidSession_ReturnsRegisters()
	{
		var registers = RegisterState.Create(r0: 42);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(registers, ApiJsonContext.Default.RegisterState));

		var result = await _apiClient.StepAsync("session-123");

		result.R0.Should().Be(42u);
		_handler.LastRequest!.RequestUri!.PathAndQuery.Should().Be("/api/v1/session/session-123/step");
		_handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task EvaluateExpressionAsync_WithValidExpression_ReturnsValue()
	{
		var response = new EvaluationResponse(42u);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.EvaluationResponse));

		var result = await _apiClient.EvaluateExpressionAsync("session-123", "r0 + r1");

		result.Should().Be(42u);
	}

	[Fact]
	public async Task EvaluateExpressionAsync_WithInvalidExpression_ThrowsExpressionEvaluationException()
	{
		var errorResponse = new ApiErrorResponse("Invalid syntax");
		_handler.SetResponse(HttpStatusCode.BadRequest, JsonSerializer.Serialize(errorResponse, ApiJsonContext.Default.ApiErrorResponse));

		var act = async () => await _apiClient.EvaluateExpressionAsync("session-123", "invalid");

		var exception = await act.Should().ThrowAsync<ExpressionEvaluationException>();
		exception.Which.Expression.Should().Be("invalid");
	}

	[Fact]
	public async Task GetMemoryAsync_ReturnsMemoryData()
	{
		var memory = new byte[] { 0x01, 0x02, 0x03, 0x04 };
		var response = new MemoryResponse(memory);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(response, ApiJsonContext.Default.MemoryResponse));

		var result = await _apiClient.GetMemoryAsync("session-123", 0x10000, 4);

		result.Should().Equal(memory);
	}

	[Fact]
	public async Task AddBreakpointAsync_SendsCorrectRequest()
	{
		_handler.SetResponse(HttpStatusCode.OK, "{}");

		await _apiClient.AddBreakpointAsync("session-123", 0x8000);

		_handler.LastRequest!.RequestUri!.PathAndQuery.Should().Contain("/breakpoint");
		_handler.LastRequest.Method.Should().Be(HttpMethod.Post);
	}

	[Fact]
	public async Task AddWatchpointAsync_ReturnsWatchpoint()
	{
		var watchpoint = new Watchpoint(1, 0x10000, WatchpointType.Write);
		_handler.SetResponse(HttpStatusCode.OK, JsonSerializer.Serialize(watchpoint, ApiJsonContext.Default.Watchpoint));

		var result = await _apiClient.AddWatchpointAsync("session-123", 0x10000, WatchpointType.Write);

		result.Id.Should().Be(1);
		result.Address.Should().Be(0x10000u);
		result.Type.Should().Be(WatchpointType.Write);
	}

	public void Dispose()
	{
		_httpClient.Dispose();
		_handler.Dispose();
	}
}

/// <summary>
/// Test HTTP message handler that allows setting responses and exceptions.
/// </summary>
internal sealed class TestHttpMessageHandler : HttpMessageHandler
{
	private HttpStatusCode _statusCode = HttpStatusCode.OK;
	private string _content = "{}";
	private Exception? _exception;

	public HttpRequestMessage? LastRequest { get; private set; }

	public void SetResponse(HttpStatusCode statusCode, string content)
	{
		_statusCode = statusCode;
		_content = content;
		_exception = null;
	}

	public void SetException(Exception exception)
	{
		_exception = exception;
	}

	protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
	{
		LastRequest = request;

		if (_exception is not null) {
			throw _exception;
		}

		var response = new HttpResponseMessage(_statusCode) {
			Content = new StringContent(_content, Encoding.UTF8, "application/json")
		};

		return Task.FromResult(response);
	}
}
