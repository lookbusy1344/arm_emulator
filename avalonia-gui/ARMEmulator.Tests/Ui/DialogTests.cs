using ARMEmulator.Models;
using ARMEmulator.Views;
using Avalonia.Controls;
using Avalonia.Threading;
using Avalonia.VisualTree;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Structure shared by the dialogs: a button row with a default and a cancel button, themed fonts and text.
/// </summary>
public sealed class DialogTests
{
	private const string DataFontName = "JetBrains Mono";
	private const int EditorTabIndex = 1;
	private const string ThemeHint = "Auto follows the system theme.";

	private static async Task WithDialog<T>(Func<T> create, Action<T> check) where T : Window
	{
		await UiTest.RunOnUiThread(() => {
			var window = create();
			window.Show();
			Dispatcher.UIThread.RunJobs();
			try {
				check(window);
			}
			finally {
				window.Close();
			}
		});
	}

	private static T Named<T>(Window window, string name) where T : Control =>
		window.GetVisualDescendants().OfType<T>().Single(control => control.Name == name);

	[Fact]
	public Task Preferences_HasADefaultOkAndACancelButton() =>
		WithDialog(() => new PreferencesWindow(AppSettings.Default), window => {
			Named<Button>(window, "OkButton").IsDefault.Should().BeTrue();
			Named<Button>(window, "CancelButton").IsCancel.Should().BeTrue();
		});

	[Fact]
	public Task Preferences_ThemeHintDoesNotClaimARestartIsNeeded() =>
		WithDialog(() => new PreferencesWindow(AppSettings.Default), window =>
			Named<TextBlock>(window, "ThemeHint").Text.Should().Be(ThemeHint));

	[Fact]
	public Task Preferences_PreviewUsesTheBundledMonospaceFont() =>
		WithDialog(() => new PreferencesWindow(AppSettings.Default), window => {
			window.GetVisualDescendants().OfType<TabControl>().Single().SelectedIndex = EditorTabIndex;
			Dispatcher.UIThread.RunJobs();

			Named<TextBlock>(window, "FontPreview").FontFamily.Name.Should().Be(DataFontName);
		});

	[Fact]
	public Task About_CloseIsTheDefaultAndCancelButton() =>
		WithDialog(() => new AboutWindow(), window => {
			var close = Named<Button>(window, "CloseButton");
			close.IsDefault.Should().BeTrue();
			close.IsCancel.Should().BeTrue();
		});

	[Fact]
	public Task Examples_LoadIsDefaultAndCancelIsCancel() =>
		WithDialog(() => new ExamplesBrowserWindow(), window => {
			Named<Button>(window, "LoadButton").IsDefault.Should().BeTrue();
			Named<Button>(window, "CancelButton").IsCancel.Should().BeTrue();
		});

	[Fact]
	public Task Examples_PreviewUsesTheBundledMonospaceFont() =>
		WithDialog(() => new ExamplesBrowserWindow(), window =>
			Named<TextBox>(window, "PreviewBox").FontFamily.Name.Should().Be(DataFontName));

	[Fact]
	public Task Examples_ErrorHintDoesNotHardCodeTheBackendUrl() =>
		WithDialog(() => new ExamplesBrowserWindow(), window =>
			Named<TextBlock>(window, "ErrorHint").Text.Should().Be("Check that the backend is running."));

	[Fact]
	public Task UnsavedChanges_SaveIsDefaultAndCancelIsCancel() =>
		WithDialog(() => new UnsavedChangesWindow("prog.s"), window => {
			Named<Button>(window, "SaveButton").IsDefault.Should().BeTrue();
			Named<Button>(window, "CancelButton").IsCancel.Should().BeTrue();
		});
}
