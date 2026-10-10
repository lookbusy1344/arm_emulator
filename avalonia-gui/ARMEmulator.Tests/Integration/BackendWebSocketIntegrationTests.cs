using ARMEmulator.Models;
using ARMEmulator.Services;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Integration;

/// <summary>
/// Event streaming against a running backend.
/// </summary>
[Trait("Category", "Integration")]
public sealed class BackendWebSocketIntegrationTests
{
	private static readonly TimeSpan EventTimeout = TimeSpan.FromSeconds(10);

	private const string Program = """
		.org 0x8000
		.text
		.global _start
		_start:
		    MOV R0, #5
		    MOV R7, #1
		    SWI 0
		""";

	[Fact(SkipUnless = nameof(BackendAvailability.IsRunning), SkipType = typeof(BackendAvailability), Skip = BackendAvailability.SkipReason)]
	public async Task Step_BroadcastsAStateEventWithRegisters()
	{
		var ct = TestContext.Current.CancellationToken;
		using var http = BackendAvailability.CreateHttpClient();
		var api = new ApiClient(http);
		using var client = BackendAvailability.CreateWebSocketClient();
		var stepped = new TaskCompletionSource<StateEvent>(TaskCreationOptions.RunContinuationsAsynchronously);
		using var subscription = client.Events
			.OfType<StateEvent>()
			.Where(evt => evt.Registers?.R0 == 5)
			.Subscribe(stepped.SetResult);

		var session = await api.CreateSessionAsync(ct);
		try {
			await client.ConnectAsync(session.SessionId, ct);
			_ = await api.LoadProgramAsync(session.SessionId, Program, ct);

			_ = await api.StepAsync(session.SessionId, ct);

			var evt = await stepped.Task.WaitAsync(EventTimeout, ct);
			evt.SessionId.Should().Be(session.SessionId);
			evt.Registers!.R0.Should().Be(5u);
			evt.Registers.PC.Should().Be(0x8004u);
		}
		finally {
			await client.DisconnectAsync();
			await api.DestroySessionAsync(session.SessionId, ct);
		}
	}
}
