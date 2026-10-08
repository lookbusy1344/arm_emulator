using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Subjects;
using ARMEmulator.Collections;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

public sealed class MainWindowViewModelSettingsTests : IDisposable
{
	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly ISettingsStore store = Substitute.For<ISettingsStore>();

	private readonly Subject<EmulatorEvent> events = new();

	private static readonly AppSettings Changed = AppSettings.Default with {
		EditorFontSize = 20,
		Theme = AppTheme.Dark,
		AutoScrollToMemoryWrites = false
	};

	private readonly Subject<Unit> recentFilesChanged = new();

	public MainWindowViewModelSettingsTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(recentFilesChanged);
		files.RecentFiles.Returns([]);
	}

	public void Dispose()
	{
		events.Dispose();
		recentFilesChanged.Dispose();
	}

	private MainWindowViewModel CreateViewModel() => new(api, ws, files, store);

	[Fact]
	public void Settings_DefaultToAppDefaults()
	{
		using var vm = CreateViewModel();

		vm.Settings.Should().Be(AppSettings.Default);
	}

	[Fact]
	public void SaveSettings_AppliesThenPersists()
	{
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		vm.Settings.Should().Be(Changed);
		vm.Memory.AutoScrollToWrites.Should().BeFalse();
		store.Received(1).Save(Changed);
		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public void SaveSettings_WhenStoreFails_ReportsErrorAndKeepsAppliedSettings()
	{
		store.When(s => s.Save(Arg.Any<AppSettings>())).Throw(new IOException("disk full"));
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		vm.ErrorMessage.Should().Be("Failed to save settings: disk full");
		vm.Settings.Should().Be(Changed);
	}

	[Fact]
	public void SelectedInspectorPanel_DefaultsToRegisters()
	{
		using var vm = CreateViewModel();

		vm.SelectedInspectorPanel.Should().Be(InspectorPanel.Registers);
	}

	[Fact]
	public void SelectingAnInspectorPanel_UpdatesTheLayoutWithoutTouchingTheStore()
	{
		using var vm = CreateViewModel();

		vm.SelectedInspectorPanel = InspectorPanel.Disassembly;

		vm.Settings.Layout.SelectedPanel.Should().Be(InspectorPanel.Disassembly);
		store.DidNotReceiveWithAnyArgs().Save(default!);
	}

	[Fact]
	public void SelectingAnInspectorPanel_RaisesPropertyChanged()
	{
		using var vm = CreateViewModel();
		var changed = new List<string?>();
		vm.PropertyChanged += (_, args) => changed.Add(args.PropertyName);

		vm.SelectedInspectorPanel = InspectorPanel.Stack;

		changed.Should().Contain(nameof(MainWindowViewModel.SelectedInspectorPanel));
	}

	[Fact]
	public void SelectingTheCurrentInspectorPanel_DoesNotPersist()
	{
		using var vm = CreateViewModel();

		vm.SelectedInspectorPanel = InspectorPanel.Registers;

		store.DidNotReceiveWithAnyArgs().Save(default!);
	}

	[Fact]
	public void ApplySettings_SelectsTheStoredInspectorPanelWithoutPersisting()
	{
		using var vm = CreateViewModel();

		vm.ApplySettings(Changed with { Layout = new WindowLayout { SelectedPanel = InspectorPanel.Memory } });

		vm.SelectedInspectorPanel.Should().Be(InspectorPanel.Memory);
		store.DidNotReceiveWithAnyArgs().Save(default!);
	}

	[Fact]
	public void SaveWindowGeometry_PersistsItAndKeepsTheSelectedPanel()
	{
		var geometry = new WindowGeometry { X = 1, Y = 2, Width = 900, Height = 700, InspectorWidth = 300, ConsoleHeight = 120, IsMaximized = false };
		using var vm = CreateViewModel();
		vm.SelectedInspectorPanel = InspectorPanel.Stack;
		store.ClearReceivedCalls();

		vm.SaveWindowGeometry(geometry);

		store.Received(1).Save(AppSettings.Default with { Layout = new WindowLayout { SelectedPanel = InspectorPanel.Stack, Geometry = geometry } });
	}

	[Fact]
	public void ApplySettings_DoesNotPersist()
	{
		using var vm = CreateViewModel();

		vm.ApplySettings(Changed);

		vm.Settings.Should().Be(Changed);
		store.DidNotReceiveWithAnyArgs().Save(default!);
	}

	[Fact]
	public void ApplySettings_SetsTheRecentFilesLimit()
	{
		using var vm = CreateViewModel();

		vm.ApplySettings(Changed with { RecentFilesLimit = 4 });

		files.RecentFilesLimit.Should().Be(4);
	}

	[Fact]
	public void SaveSettings_KeepsTheCurrentRecentFiles()
	{
		files.RecentFiles.Returns([new RecentFile("/p/b.s", DateTimeOffset.MinValue), new RecentFile("/p/a.s", DateTimeOffset.MinValue)]);
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		store.Received(1).Save(Changed with { RecentFiles = ["/p/b.s", "/p/a.s"] });
	}

	[Fact]
	public void RecentFilesChange_PersistsThePathsWithTheCurrentSettings()
	{
		using var vm = CreateViewModel();
		vm.ApplySettings(Changed);
		files.RecentFiles.Returns([new RecentFile("/p/a.s", DateTimeOffset.MinValue)]);

		recentFilesChanged.OnNext(Unit.Default);

		store.Received(1).Save(Changed with { RecentFiles = ["/p/a.s"] });
	}

	[Fact]
	public void RecentFilesChange_RaisesPropertyChangedForTheMenu()
	{
		using var vm = CreateViewModel();
		var raised = new List<string?>();
		vm.PropertyChanged += (_, e) => raised.Add(e.PropertyName);

		recentFilesChanged.OnNext(Unit.Default);

		raised.Should().Contain(nameof(MainWindowViewModel.RecentFiles));
	}

	[Fact]
	public void SaveSettings_WhenAccessIsDenied_ReportsError()
	{
		store.When(s => s.Save(Arg.Any<AppSettings>())).Throw(new UnauthorizedAccessException("access denied"));
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		vm.ErrorMessage.Should().Be("Failed to save settings: access denied");
	}

	[Fact]
	public void RecentFilesChange_WhenStoreFails_ReportsError()
	{
		store.When(s => s.Save(Arg.Any<AppSettings>())).Throw(new IOException("disk full"));
		using var vm = CreateViewModel();

		recentFilesChanged.OnNext(Unit.Default);

		vm.ErrorMessage.Should().Be("Failed to save recent files: disk full");
	}

	[Fact]
	public void RefreshRecentFiles_RemovesMissingEntries()
	{
		using var vm = CreateViewModel();

		vm.RefreshRecentFiles();

		files.Received(1).RemoveMissingRecentFiles();
	}

	[Fact]
	public async Task OpenRecentFile_WhenFileIsGone_ReportsItAndRemovesTheEntry()
	{
		const string missing = "/definitely/not/here.s";
		using var vm = CreateViewModel();

		await vm.OpenRecentFileCommand.Execute(missing);

		vm.ErrorMessage.Should().Be("File not found: /definitely/not/here.s");
		files.Received(1).RemoveRecentFile(missing);
	}

	[Fact]
	public async Task OpenRecentFile_WhenFileExists_MovesItToTheTopOfTheList()
	{
		var path = Path.GetTempFileName();
		try {
			using var vm = CreateViewModel();
			vm.SessionId = null;

			await vm.OpenRecentFileCommand.Execute(path);

			files.Received(1).AddRecentFile(path);
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task OpenSourceFile_WithSession_LoadsTheFileIntoTheEditorAndSession()
	{
		const string sessionId = "s1";
		const string source = "_start:\n  MOV R0, #1\n";
		var path = Path.GetTempFileName();
		try {
			await File.WriteAllTextAsync(path, source, TestContext.Current.CancellationToken);
			api.LoadProgramAsync(sessionId, source, Arg.Any<CancellationToken>())
				.Returns(new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(new Dictionary<string, uint>())));
			api.GetSourceMapAsync(sessionId, Arg.Any<CancellationToken>()).Returns([]);
			api.GetRegistersAsync(sessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create());
			using var vm = CreateViewModel();
			vm.SessionId = sessionId;

			await vm.OpenSourceFileAsync(path, TestContext.Current.CancellationToken);

			vm.SourceCode.Should().Be(source);
			files.CurrentFilePath.Should().Be(path);
			files.Received(1).AddRecentFile(path);
			await api.Received(1).LoadProgramAsync(sessionId, source, Arg.Any<CancellationToken>());
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task OpenSourceFile_WithoutSession_ShowsTheFileButDoesNotLoad()
	{
		var path = Path.GetTempFileName();
		try {
			await File.WriteAllTextAsync(path, "MOV R0, #1", TestContext.Current.CancellationToken);
			using var vm = CreateViewModel();

			await vm.OpenSourceFileAsync(path, TestContext.Current.CancellationToken);

			vm.SourceCode.Should().Be("MOV R0, #1");
			await api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, TestContext.Current.CancellationToken);
			vm.ErrorMessage.Should().BeNull();
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task OpenSourceFile_WhenMissing_ReportsItAndLeavesTheEditorAlone()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = "keep";

		await vm.OpenSourceFileAsync("/definitely/not/here.s", TestContext.Current.CancellationToken);

		vm.ErrorMessage.Should().Be("File not found: /definitely/not/here.s");
		vm.SourceCode.Should().Be("keep");
	}
}
