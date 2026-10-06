using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// The status pill: a key that selects the pill's theme colour, and a short label.
/// </summary>
public sealed class MainWindowViewModelStatusTests : IDisposable
{
	private readonly Subject<EmulatorEvent> events = new();
	private readonly IApiClient api = Substitute.For<IApiClient>();
	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();
	private readonly IFileService files = Substitute.For<IFileService>();

	public MainWindowViewModelStatusTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
	}

	public void Dispose() => events.Dispose();

	[Fact]
	public void WhenNotConnected_KeyAndLabelAreDisconnected()
	{
		using var vm = new MainWindowViewModel(api, ws, files);
		vm.Status = VMState.Running;

		vm.StatusKey.Should().Be("Disconnected");
		vm.StatusLabel.Should().Be("Disconnected");
	}

	[Theory]
	[InlineData(VMState.Idle, "Idle", "Idle")]
	[InlineData(VMState.Running, "Running", "Running")]
	[InlineData(VMState.Breakpoint, "Breakpoint", "Breakpoint")]
	[InlineData(VMState.Halted, "Halted", "Halted")]
	[InlineData(VMState.Error, "Error", "Error")]
	[InlineData(VMState.WaitingForInput, "WaitingForInput", "Waiting for input")]
	public void WhenConnected_KeyNamesTheStateAndLabelIsReadable(VMState state, string key, string label)
	{
		using var vm = new MainWindowViewModel(api, ws, files) { IsConnected = true, Status = state };

		vm.StatusKey.Should().Be(key);
		vm.StatusLabel.Should().Be(label);
	}
}
