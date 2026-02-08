using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using ARMEmulator.Models;
using ARMEmulator.Services;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Tests for WebSocketClient - simplified to run quickly without blocking.
/// </summary>
public sealed class WebSocketClientTests
{
	[Fact]
	public void Constructor_CreatesClient()
	{
		using var client = new WebSocketClient("ws://localhost:8080/ws");
		client.IsConnected.Should().BeFalse();
	}

	[Fact]
	public void IsConnected_ReturnsFalse_WhenNotConnected()
	{
		using var client = new WebSocketClient("ws://localhost:8080/ws");
		client.IsConnected.Should().BeFalse();
	}

	// Note: Full integration tests with actual WebSocket connections will be in Phase 1 integration tests
	// These unit tests verify the basic structure and API contract
}
