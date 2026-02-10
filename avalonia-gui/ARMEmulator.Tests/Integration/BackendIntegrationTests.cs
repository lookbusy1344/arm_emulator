using System.Collections.Immutable;
using ARMEmulator.Models;
using ARMEmulator.Services;
using FluentAssertions;

namespace ARMEmulator.Tests.Integration;

/// <summary>
/// Integration tests that require a running backend.
/// Run the backend with: make build && ./arm-emulator
/// These tests verify end-to-end communication with the real backend.
/// </summary>
[Trait("Category", "Integration")]
public sealed class BackendIntegrationTests : IDisposable
{
	private readonly HttpClient _httpClient;
	private readonly ApiClient _apiClient;
	private readonly CancellationTokenSource _cts;

	public BackendIntegrationTests()
	{
		_httpClient = new HttpClient {
			BaseAddress = new Uri("http://localhost:8080"),
			Timeout = TimeSpan.FromSeconds(5)
		};
		_apiClient = new ApiClient(_httpClient);
		_cts = new CancellationTokenSource();
	}

#pragma warning disable xUnit1004 // Integration tests are skipped by default
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task HealthCheck_BackendAvailable_ReturnsVersion()
#pragma warning restore xUnit1004
	{
		// Act
		var version = await _apiClient.GetVersionAsync(_cts.Token);

		// Assert
		version.Should().NotBeNull();
		version.Version.Should().NotBeNullOrEmpty();
		version.Commit.Should().NotBeNullOrEmpty();
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task FullExecutionCycle_LoadStepRun_CompletesSuccessfully()
#pragma warning restore xUnit1004
	{
		// Arrange - Simple program that adds two numbers
		const string program = """
            .text
            .global _start

            _start:
                MOV R0, #5      ; R0 = 5
                MOV R1, #3      ; R1 = 3
                ADD R2, R0, R1  ; R2 = R0 + R1 = 8
                MOV R7, #1      ; syscall: exit
                SWI 0           ; exit
            """;

		// Act - Create session
		var session = await _apiClient.CreateSessionAsync(_cts.Token);
		session.SessionId.Should().NotBeNullOrEmpty();

		try {
			// Act - Load program
			var loadResponse = await _apiClient.LoadProgramAsync(session.SessionId, program, _cts.Token);
			loadResponse.Success.Should().BeTrue();

			// Act - Get initial status
			var status = await _apiClient.GetStatusAsync(session.SessionId, _cts.Token);
			status.State.Should().Be(VMState.Idle);

			// Act - Step through first instruction (MOV R0, #5)
			var registers = await _apiClient.StepAsync(session.SessionId, _cts.Token);
			registers.R0.Should().Be(5);

			// Act - Step through second instruction (MOV R1, #3)
			registers = await _apiClient.StepAsync(session.SessionId, _cts.Token);
			registers.R1.Should().Be(3);

			// Act - Step through third instruction (ADD R2, R0, R1)
			registers = await _apiClient.StepAsync(session.SessionId, _cts.Token);
			registers.R2.Should().Be(8);

			// Act - Run to completion
			await _apiClient.RunAsync(session.SessionId, _cts.Token);

			// Wait briefly for execution to complete
			await Task.Delay(100, _cts.Token);

			// Assert - Program should have halted
			status = await _apiClient.GetStatusAsync(session.SessionId, _cts.Token);
			status.State.Should().Be(VMState.Halted);
		}
		finally {
			// Cleanup
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task LoadProgram_WithSyntaxError_ThrowsProgramLoadException()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string invalidProgram = """
            .text
            _start:
                INVALID_INSTRUCTION R0, #5
            """;

		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			// Act
			var act = async () => await _apiClient.LoadProgramAsync(session.SessionId, invalidProgram, _cts.Token);

			// Assert
			await act.Should().ThrowAsync<ProgramLoadException>()
				.Where(ex => ex.Errors.Any());
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task Breakpoints_AddAndRemove_WorksCorrectly()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string program = """
            .text
            .global _start

            _start:
                MOV R0, #1      ; Address 0x00008000
                MOV R1, #2      ; Address 0x00008004
                MOV R2, #3      ; Address 0x00008008
                SWI 0
            """;

		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			await _apiClient.LoadProgramAsync(session.SessionId, program, _cts.Token);

			// Act - Add breakpoint at second instruction
			await _apiClient.AddBreakpointAsync(session.SessionId, 0x00008004, _cts.Token);

			// Act - Get breakpoints
			var breakpoints = await _apiClient.GetBreakpointsAsync(session.SessionId, _cts.Token);
			breakpoints.Should().Contain(0x00008004);

			// Act - Remove breakpoint
			await _apiClient.RemoveBreakpointAsync(session.SessionId, 0x00008004, _cts.Token);

			// Assert - Breakpoint should be removed
			breakpoints = await _apiClient.GetBreakpointsAsync(session.SessionId, _cts.Token);
			breakpoints.Should().NotContain(0x00008004);
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task Memory_ReadAndWrite_ReturnsCorrectData()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string program = """
            .text
            .global _start

            _start:
                MOV R0, #0x12345678
                SWI 0
            """;

		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			await _apiClient.LoadProgramAsync(session.SessionId, program, _cts.Token);

			// Act - Step to execute MOV instruction
			await _apiClient.StepAsync(session.SessionId, _cts.Token);

			// Act - Read memory at PC (program start)
			var memory = await _apiClient.GetMemoryAsync(session.SessionId, 0x00008000, 16, _cts.Token);

			// Assert - Should have read 16 bytes
			memory.Length.Should().Be(16);
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task Disassembly_GetInstructions_ReturnsFormattedCode()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string program = """
            .text
            .global _start

            _start:
                MOV R0, #5
                MOV R1, #3
                ADD R2, R0, R1
                SWI 0
            """;

		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			await _apiClient.LoadProgramAsync(session.SessionId, program, _cts.Token);

			// Act - Get disassembly starting at PC
			var disassembly = await _apiClient.GetDisassemblyAsync(session.SessionId, 0x00008000, 10, _cts.Token);

			// Assert
			disassembly.Should().NotBeEmpty();
			disassembly.Length.Should().BeGreaterThanOrEqualTo(4); // At least our 4 instructions
			disassembly.First().Address.Should().Be(0x00008000);
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task ExpressionEvaluation_ValidExpression_ReturnsValue()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string program = """
            .text
            _start:
                MOV R0, #42
                SWI 0
            """;

		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			await _apiClient.LoadProgramAsync(session.SessionId, program, _cts.Token);
			await _apiClient.StepAsync(session.SessionId, _cts.Token); // Execute MOV R0, #42

			// Act - Evaluate expression
			var result = await _apiClient.EvaluateExpressionAsync(session.SessionId, "r0", _cts.Token);

			// Assert
			result.Should().Be(42);
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task ExpressionEvaluation_InvalidExpression_ThrowsException()
#pragma warning restore xUnit1004
	{
		// Arrange
		var session = await _apiClient.CreateSessionAsync(_cts.Token);

		try {
			// Act
			var act = async () => await _apiClient.EvaluateExpressionAsync(
				session.SessionId,
				"invalid syntax!",
				_cts.Token);

			// Assert
			await act.Should().ThrowAsync<ExpressionEvaluationException>();
		}
		finally {
			await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
		}
	}

#pragma warning disable xUnit1004
	[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
	public async Task SessionNotFound_ThrowsSessionNotFoundException()
#pragma warning restore xUnit1004
	{
		// Arrange
		const string nonExistentSessionId = "does-not-exist";

		// Act
		var act = async () => await _apiClient.GetStatusAsync(nonExistentSessionId, _cts.Token);

		// Assert
		await act.Should().ThrowAsync<SessionNotFoundException>()
			.Where(ex => ex.Message.Contains(nonExistentSessionId));
	}

	public void Dispose()
	{
		_cts.Cancel();
		_cts.Dispose();
		_httpClient.Dispose();
	}
}
