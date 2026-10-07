using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.VisualTree;
using NSubstitute;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// A main window over a real view model with mocked services, shown on the headless platform.
/// </summary>
public sealed class MainWindowHarness : IDisposable
{
	public const string SessionId = "ui-session";

	public MainWindowHarness(AppSettings? settings = null)
	{
		Api = Substitute.For<IApiClient>();
		Ws = Substitute.For<IWebSocketClient>();
		Files = Substitute.For<IFileService>();
		Backend = Substitute.For<IBackendManager>();
		Events = new Subject<EmulatorEvent>();
		BackendStatus = new Subject<BackendStatus>();

		Ws.Events.Returns(Events);
		Files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		Files.RecentFiles.Returns([]);
		Backend.StatusChanged.Returns(BackendStatus);
		Api.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(new SessionInfo(SessionId));
		Api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		Api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
		Api.GetRegistersAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns(RegisterState.Create());

		Store = Substitute.For<ISettingsStore>();
		ViewModel = new MainWindowViewModel(Api, Ws, Files, Store);
		ViewModel.ApplySettings(settings ?? AppSettings.Default);
		Window = new MainWindow(ViewModel);
		Window.Show();
	}

	public IApiClient Api { get; }

	public IWebSocketClient Ws { get; }

	public IFileService Files { get; }

	public IBackendManager Backend { get; }

	public ISettingsStore Store { get; }

	public Subject<EmulatorEvent> Events { get; }

	public Subject<BackendStatus> BackendStatus { get; }

	public MainWindowViewModel ViewModel { get; }

	public MainWindow Window { get; }

	/// <summary>Finds a named control anywhere in the window.</summary>
	public T Find<T>(string name) where T : Control =>
		Window.GetVisualDescendants().OfType<T>().Single(control => control.Name == name);

	public void Dispose()
	{
		Window.Close();
		ViewModel.Dispose();
		Events.Dispose();
		BackendStatus.Dispose();
	}
}
