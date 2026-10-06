using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Windows.Input;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for the breakpoint and watchpoint list commands: validation, failures and success.
/// </summary>
public sealed class MainWindowViewModelWatchpointTests : IDisposable
{
	private const string SessionId = "s1";

	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	private readonly Subject<EmulatorEvent> events = new();

	public MainWindowViewModelWatchpointTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
	}

	public void Dispose() => events.Dispose();

	private MainWindowViewModel CreateViewModel() => new(api, ws, files) { SessionId = SessionId };

	private static void Execute(ICommand command, object? parameter = null) => command.Execute(parameter);

	// Add watchpoint

	[Theory]
	[InlineData("0x9000")]
	[InlineData("0X9000")]
	[InlineData("9000")]
	[InlineData("  0x9000  ")]
	public void AddWatchpoint_ParsesHexAddresses(string text)
	{
		var created = new Watchpoint(7, 0x9000, WatchpointType.Write);
		api.AddWatchpointAsync(SessionId, 0x9000, WatchpointType.Write, Arg.Any<CancellationToken>()).Returns(created);
		using var vm = CreateViewModel();
		vm.WatchpointAddressText = text;
		vm.SelectedWatchpointType = WatchpointType.Write;

		Execute(vm.AddWatchpointCommand);

		vm.Watchpoints.Should().Equal(created);
		vm.WatchpointAddressText.Should().BeEmpty();
		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public void AddWatchpoint_DefaultsToReadWrite()
	{
		using var vm = CreateViewModel();

		vm.SelectedWatchpointType.Should().Be(WatchpointType.ReadWrite);
	}

	[Theory]
	[InlineData("")]
	[InlineData("   ")]
	public void AddWatchpoint_WithoutAddress_ReportsItAndCallsNoApi(string text)
	{
		using var vm = CreateViewModel();
		vm.WatchpointAddressText = text;

		Execute(vm.AddWatchpointCommand);

		vm.ErrorMessage.Should().Be("Enter a watchpoint address");
		_ = api.DidNotReceiveWithAnyArgs().AddWatchpointAsync(default!, default, default, TestContext.Current.CancellationToken);
	}

	[Theory]
	[InlineData("xyz")]
	[InlineData("0x")]
	[InlineData("1FFFFFFFF")]
	[InlineData("-1")]
	public void AddWatchpoint_WithInvalidAddress_ReportsItKeepsTheTextAndCallsNoApi(string text)
	{
		using var vm = CreateViewModel();
		vm.WatchpointAddressText = text;

		Execute(vm.AddWatchpointCommand);

		vm.ErrorMessage.Should().Be($"Invalid watchpoint address '{text}': enter a 32-bit hex value such as 0x9000");
		vm.WatchpointAddressText.Should().Be(text);
		_ = api.DidNotReceiveWithAnyArgs().AddWatchpointAsync(default!, default, default, TestContext.Current.CancellationToken);
	}

	[Fact]
	public void AddWatchpoint_WithoutSession_ReportsIt()
	{
		using var vm = CreateViewModel();
		vm.SessionId = null;
		vm.WatchpointAddressText = "0x9000";

		Execute(vm.AddWatchpointCommand);

		vm.ErrorMessage.Should().Be("No active session");
		vm.WatchpointAddressText.Should().Be("0x9000");
	}

	[Fact]
	public void AddWatchpoint_WhenApiFails_ReportsErrorAndKeepsTheInput()
	{
		api.AddWatchpointAsync(SessionId, 0x9000, WatchpointType.ReadWrite, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ApiException("API error: address out of range"));
		using var vm = CreateViewModel();
		vm.WatchpointAddressText = "0x9000";

		Execute(vm.AddWatchpointCommand);

		vm.ErrorMessage.Should().Be("Add watchpoint failed: API error: address out of range");
		vm.WatchpointAddressText.Should().Be("0x9000");
		vm.Watchpoints.Should().BeEmpty();
	}

	// Remove watchpoint

	[Fact]
	public void RemoveWatchpoint_RemovesTheEntryWithThatId()
	{
		using var vm = CreateViewModel();
		vm.Watchpoints = [new Watchpoint(1, 0x9000, WatchpointType.Read), new Watchpoint(2, 0x9004, WatchpointType.Write)];

		Execute(vm.RemoveWatchpointCommand, 1);

		_ = api.Received(1).RemoveWatchpointAsync(SessionId, 1, Arg.Any<CancellationToken>());
		vm.Watchpoints.Select(w => w.Id).Should().Equal(2);
	}

	[Fact]
	public void RemoveWatchpoint_WhenApiFails_ReportsErrorAndKeepsTheList()
	{
		api.RemoveWatchpointAsync(SessionId, 1, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ApiException("API error: no such watchpoint"));
		using var vm = CreateViewModel();
		vm.Watchpoints = [new Watchpoint(1, 0x9000, WatchpointType.Read)];

		Execute(vm.RemoveWatchpointCommand, 1);

		vm.ErrorMessage.Should().Be("Remove watchpoint failed: API error: no such watchpoint");
		vm.Watchpoints.Should().HaveCount(1);
	}

	// Remove breakpoint

	[Fact]
	public void RemoveBreakpoint_RemovesTheAddress()
	{
		using var vm = CreateViewModel();
		vm.Breakpoints = [0x8000, 0x8004];

		Execute(vm.RemoveBreakpointCommand, 0x8000u);

		_ = api.Received(1).RemoveBreakpointAsync(SessionId, 0x8000, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEquivalentTo([0x8004u]);
	}

	[Fact]
	public void RemoveBreakpoint_WhenApiFails_ReportsErrorAndKeepsTheList()
	{
		api.RemoveBreakpointAsync(SessionId, 0x8000, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ApiException("API error: no breakpoint at address"));
		using var vm = CreateViewModel();
		vm.Breakpoints = [0x8000];

		Execute(vm.RemoveBreakpointCommand, 0x8000u);

		vm.ErrorMessage.Should().Be("Remove breakpoint failed: API error: no breakpoint at address");
		vm.Breakpoints.Should().BeEquivalentTo([0x8000u]);
	}
}
