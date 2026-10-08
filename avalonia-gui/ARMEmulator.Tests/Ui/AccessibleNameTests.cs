using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using ARMEmulator.Views;
using Avalonia;
using Avalonia.Automation;
using Avalonia.Controls;
using Avalonia.Controls.Primitives;
using Avalonia.Threading;
using Avalonia.VisualTree;
using AwesomeAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Every control a user can operate has a name a screen reader can announce.
/// </summary>
public sealed class AccessibleNameTests
{
	private static void Settle(Window window)
	{
		Dispatcher.UIThread.RunJobs();
		window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	private static bool IsOperable(Control control) =>
		control is Button or TextBox or ComboBox or NumericUpDown or CheckBox or ListBox
		&& control.TemplatedParent is null
		&& control.IsEffectivelyVisible;

	/// <summary>A name made of letters or digits: the explicit automation name, or text content.</summary>
	private static bool HasName(Control control)
	{
		var name = AutomationProperties.GetName(control) ?? (control as ContentControl)?.Content as string;
		return name?.Any(char.IsLetterOrDigit) == true;
	}

	private static IEnumerable<string> Unnamed(Visual root) =>
		root.GetVisualDescendants().OfType<Control>()
			.Where(control => IsOperable(control) && !HasName(control))
			.Select(control => $"{control.GetType().Name} {control.Name ?? "(unnamed)"} in {control.FindAncestorOfType<UserControl>()?.GetType().Name ?? "window"}");

	[Fact]
	public Task MainWindowPanels_NameTheirControls() =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			ui.ViewModel.Breakpoints = [0x8000];
			ui.ViewModel.Watchpoints = [new Watchpoint(1, 0x9000, WatchpointType.Write)];
			var unnamed = new List<string>();

			foreach (var panel in Enum.GetValues<InspectorPanel>()) {
				ui.ViewModel.SelectedInspectorPanel = panel;
				Settle(ui.Window);
				unnamed.AddRange(Unnamed(ui.Window).Select(description => $"{panel}: {description}"));
			}

			string.Join("; ", unnamed.Distinct()).Should().BeEmpty();
		});

	[Fact]
	public Task Dialogs_NameTheirControls() =>
		UiTest.RunAsync(async _ => {
			var api = Substitute.For<IApiClient>();
			api.GetExamplesAsync(Arg.Any<CancellationToken>()).Returns([new ExampleInfo("hello.s", "Prints a greeting", 120)]);
			var examples = new ExamplesBrowserViewModel(api);
			await examples.LoadExamplesAsync();
			Window[] dialogs = [new PreferencesWindow(AppSettings.Default), new AboutWindow(), new UnsavedChangesWindow("prog.s"), new ExamplesBrowserWindow(examples)];

			var unnamed = new List<string>();
			foreach (var dialog in dialogs) {
				dialog.Show();
				Settle(dialog);
				unnamed.AddRange(Unnamed(dialog).Select(description => $"{dialog.GetType().Name}: {description}"));
				dialog.Close();
			}

			string.Join("; ", unnamed.Distinct()).Should().BeEmpty();
		});
}
