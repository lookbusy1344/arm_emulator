using System.Reactive;
using System.Reactive.Subjects;
using System.Windows.Input;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using ARMEmulator.Views;
using Avalonia.Controls;
using AwesomeAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.Ui;

/// <summary>The macOS menu bar offers every command the in-window menu offers (MainWindow.axaml).</summary>
public sealed class NativeMenusTests : IDisposable
{
	private readonly Subject<EmulatorEvent> events = new();
	private readonly Subject<Unit> recentFilesChanged = new();
	private readonly IFileService files = Substitute.For<IFileService>();

	public void Dispose()
	{
		events.Dispose();
		recentFilesChanged.Dispose();
	}

	private MainWindowViewModel CreateViewModel(params RecentFile[] recent)
	{
		var ws = Substitute.For<IWebSocketClient>();
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(recentFilesChanged);
		files.RecentFiles.Returns(recent);
		return new MainWindowViewModel(Substitute.For<IApiClient>(), ws, files, Substitute.For<ISettingsStore>());
	}

	private static NativeMenuItem Submenu(NativeMenu menu, string header) =>
		menu.Items.OfType<NativeMenuItem>().Single(item => item.Header == header);

	private static string?[] Headers(NativeMenuItem submenu) =>
		[.. submenu.Menu!.Items.OfType<NativeMenuItem>().Where(item => item is not NativeMenuItemSeparator).Select(item => item.Header)];

	private static ICommand? CommandOf(NativeMenuItem submenu, string header) =>
		submenu.Menu!.Items.OfType<NativeMenuItem>().Single(item => item.Header == header).Command;

	[Fact]
	public Task WindowMenu_HasFileAndDebugMenus() =>
		UiTest.RunOnUiThread(() => {
			using var viewModel = CreateViewModel();

			var menu = NativeMenus.CreateWindowMenu(viewModel);

			menu.Items.OfType<NativeMenuItem>().Select(item => item.Header).Should().Equal("File", "Debug");
		});

	[Fact]
	public Task FileMenu_OffersEveryFileCommandOfTheInWindowMenu() =>
		UiTest.RunOnUiThread(() => {
			using var viewModel = CreateViewModel();

			var file = Submenu(NativeMenus.CreateWindowMenu(viewModel), "File");

			Headers(file).Should().Equal(
				"Open…", "Save", "Save As…", "Recent Files", "Examples…", "Preferences…", "Restart Backend", "About…");
		});

	[Fact]
	public Task FileMenu_ItemsRunTheirCommands() =>
		UiTest.RunOnUiThread(() => {
			using var viewModel = CreateViewModel();

			var file = Submenu(NativeMenus.CreateWindowMenu(viewModel), "File");

			CommandOf(file, "Open…").Should().BeSameAs(viewModel.OpenFileCommand);
			CommandOf(file, "Save").Should().BeSameAs(viewModel.SaveFileCommand);
			CommandOf(file, "Save As…").Should().BeSameAs(viewModel.SaveAsCommand);
			CommandOf(file, "Examples…").Should().BeSameAs(viewModel.OpenExampleCommand);
			CommandOf(file, "Preferences…").Should().BeSameAs(viewModel.ShowPreferencesCommand);
			CommandOf(file, "Restart Backend").Should().BeSameAs(viewModel.RestartBackendCommand);
			CommandOf(file, "About…").Should().BeSameAs(viewModel.ShowAboutCommand);
		});

	[Fact]
	public Task RecentFilesMenu_ListsEachRecentFileAndOpensItsPath() =>
		UiTest.RunOnUiThread(() => {
			using var viewModel = CreateViewModel(
				new RecentFile("/work/a.s", DateTimeOffset.UnixEpoch),
				new RecentFile("/work/b.s", DateTimeOffset.UnixEpoch));

			var recent = Submenu(Submenu(NativeMenus.CreateWindowMenu(viewModel), "File").Menu!, "Recent Files");

			var items = recent.Menu!.Items.OfType<NativeMenuItem>().ToArray();
			items.Select(item => item.Header).Should().Equal("a.s", "b.s");
			items.Select(item => item.CommandParameter).Should().Equal("/work/a.s", "/work/b.s");
			items.Should().OnlyContain(item => ReferenceEquals(item.Command, viewModel.OpenRecentFileCommand));
		});

	[Fact]
	public Task DebugMenu_OffersEveryDebugCommandOfTheInWindowMenu() =>
		UiTest.RunOnUiThread(() => {
			using var viewModel = CreateViewModel();

			var debug = Submenu(NativeMenus.CreateWindowMenu(viewModel), "Debug");

			Headers(debug).Should().Equal(
				"Assemble", "Run", "Pause", "Step", "Step Over", "Step Out", "Reset", "Show PC", "Toggle Breakpoint");
			CommandOf(debug, "Assemble").Should().BeSameAs(viewModel.AssembleCommand);
			CommandOf(debug, "Run").Should().BeSameAs(viewModel.RunCommand);
			CommandOf(debug, "Pause").Should().BeSameAs(viewModel.PauseCommand);
			CommandOf(debug, "Step").Should().BeSameAs(viewModel.StepCommand);
			CommandOf(debug, "Step Over").Should().BeSameAs(viewModel.StepOverCommand);
			CommandOf(debug, "Step Out").Should().BeSameAs(viewModel.StepOutCommand);
			CommandOf(debug, "Reset").Should().BeSameAs(viewModel.ResetCommand);
			CommandOf(debug, "Show PC").Should().BeSameAs(viewModel.ShowPcCommand);
			CommandOf(debug, "Toggle Breakpoint").Should().BeSameAs(viewModel.ToggleBreakpointAtCaretCommand);
		});
}
