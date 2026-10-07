using ARMEmulator.Models;
using ARMEmulator.Services;
using Avalonia.Controls;
using Avalonia.Headless;
using Avalonia.Input;
using Avalonia.Threading;
using AwesomeAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// The shortcut table applied to the window, toolbar and menus.
/// </summary>
public sealed class ShortcutBindingTests
{
	private static readonly bool IsMacOS = OperatingSystem.IsMacOS();
	private static readonly RawInputModifiers CommandModifier = IsMacOS ? RawInputModifiers.Meta : RawInputModifiers.Control;

	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	[Fact]
	public Task Window_HasOneKeyBindingPerTableEntry() =>
		UiTest.Run(ui =>
			ui.Window.KeyBindings.Select(binding => binding.Gesture)
				.Should().BeEquivalentTo(Shortcuts.For(IsMacOS).Select(binding => binding.Gesture)));

	[Fact]
	public Task Window_BindsEachGestureToItsCommand() =>
		UiTest.Run(ui => {
			var vm = ui.ViewModel;
			var expected = new Dictionary<ShortcutId, object> {
				[ShortcutId.Open] = vm.OpenFileCommand,
				[ShortcutId.Save] = vm.SaveFileCommand,
				[ShortcutId.SaveAs] = vm.SaveAsCommand,
				[ShortcutId.Examples] = vm.OpenExampleCommand,
				[ShortcutId.Preferences] = vm.ShowPreferencesCommand,
				[ShortcutId.Load] = vm.LoadProgramCommand,
				[ShortcutId.Run] = vm.RunCommand,
				[ShortcutId.Pause] = vm.PauseCommand,
				[ShortcutId.Step] = vm.StepCommand,
				[ShortcutId.StepOver] = vm.StepOverCommand,
				[ShortcutId.StepOut] = vm.StepOutCommand,
				[ShortcutId.Reset] = vm.ResetCommand,
				[ShortcutId.ShowPc] = vm.ShowPcCommand
			};

			foreach (var table in Shortcuts.For(IsMacOS)) {
				var binding = ui.Window.KeyBindings.Single(candidate => candidate.Gesture == table.Gesture);
				binding.Command.Should().BeSameAs(expected[table.Id], $"{table.Id} is bound to {table.Gesture}");
			}
		});

	[Fact]
	public Task PlatformStepGesture_StepsTheProgram() =>
		UiTest.RunAsync(async ui => {
			const string session = MainWindowHarness.SessionId;
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			ui.Api.StepAsync(session, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: 0x8004));
			Settle(ui);

			ui.Window.KeyPressQwerty(PhysicalKey.T, CommandModifier);
			Settle(ui);

			await ui.Api.Received(1).StepAsync(session, Arg.Any<CancellationToken>());
		});

	[Theory]
	[InlineData("LoadButton", "Load Program", ShortcutId.Load)]
	[InlineData("RunButton", "Run", ShortcutId.Run)]
	[InlineData("PauseButton", "Pause", ShortcutId.Pause)]
	[InlineData("StepButton", "Step", ShortcutId.Step)]
	[InlineData("StepOverButton", "Step Over", ShortcutId.StepOver)]
	[InlineData("StepOutButton", "Step Out", ShortcutId.StepOut)]
	[InlineData("ResetButton", "Reset", ShortcutId.Reset)]
	[InlineData("ShowPcButton", "Show PC", ShortcutId.ShowPc)]
	public Task ToolbarTooltip_ShowsThePlatformShortcuts(string buttonName, string label, ShortcutId id) =>
		UiTest.Run(ui =>
			ToolTip.GetTip(ui.Find<Button>(buttonName)).Should().Be(ShortcutText.Hint(label, id, IsMacOS)));

	[Fact]
	public Task InWindowMenuBar_IsHiddenOnlyOnMacOS() =>
		UiTest.Run(ui => ui.Find<Menu>("MenuBar").IsVisible.Should().Be(!IsMacOS));

	[Fact]
	public Task DebugMenu_MirrorsTheToolbar() =>
		UiTest.Run(ui => {
			var debug = ui.Find<Menu>("MenuBar").Items.OfType<MenuItem>().Single(item => (string?)item.Header == "_Debug");

			debug.Items.OfType<MenuItem>().Select(item => (string?)item.Header).Should().Equal(
				"_Load Program", "_Run", "_Pause", "_Step", "Step _Over", "Step Ou_t", "R_eset", "Show _PC");
			debug.Items.OfType<MenuItem>().Select(item => item.Command).Should().Equal(
				ui.ViewModel.LoadProgramCommand, ui.ViewModel.RunCommand, ui.ViewModel.PauseCommand, ui.ViewModel.StepCommand,
				ui.ViewModel.StepOverCommand, ui.ViewModel.StepOutCommand, ui.ViewModel.ResetCommand, ui.ViewModel.ShowPcCommand);
		});

	[Fact]
	public Task MenuItems_ShowTheFirstShortcutOfTheirCommand() =>
		UiTest.Run(ui => {
			var debug = ui.Find<Menu>("MenuBar").Items.OfType<MenuItem>().Single(item => (string?)item.Header == "_Debug");
			var run = debug.Items.OfType<MenuItem>().Single(item => (string?)item.Header == "_Run");

			run.InputGesture.Should().Be(Shortcuts.For(IsMacOS).First(binding => binding.Id == ShortcutId.Run).Gesture);
		});

	[Fact]
	public Task NativeMenu_ListsFileAndDebugWithTheSameCommands() =>
		UiTest.Run(ui => {
			var menu = NativeMenu.GetMenu(ui.Window)!;

			menu.Items.OfType<NativeMenuItem>().Select(item => item.Header).Should().Equal("File", "Debug");
			var debug = menu.Items.OfType<NativeMenuItem>().Single(item => item.Header == "Debug").Menu!;
			debug.Items.OfType<NativeMenuItem>().Select(item => item.Command).Should().Equal(
				ui.ViewModel.LoadProgramCommand, ui.ViewModel.RunCommand, ui.ViewModel.PauseCommand, ui.ViewModel.StepCommand,
				ui.ViewModel.StepOverCommand, ui.ViewModel.StepOutCommand, ui.ViewModel.ResetCommand, ui.ViewModel.ShowPcCommand);
		});
}
