using ARMEmulator.Services;
using Avalonia.Controls;
using Avalonia.Threading;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// The empty-editor hint, the connection view and the error bar.
/// </summary>
public sealed class EmptyStateTests
{
	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	[Fact]
	public Task EmptyEditorHint_ShowsOnlyWhileThereIsNoSource() =>
		UiTest.Run(ui => {
			var hint = ui.Find<TextBlock>("EmptyEditorHint");
			hint.IsVisible.Should().BeTrue();
			hint.Text.Should().Be(ShortcutText.EmptyEditorHint(OperatingSystem.IsMacOS()));

			ui.ViewModel.SourceCode = "MOV R0, #1\n";
			Settle(ui);
			hint.IsVisible.Should().BeFalse();

			ui.ViewModel.SourceCode = "";
			Settle(ui);
			hint.IsVisible.Should().BeTrue();
		});

	[Fact]
	public Task ConnectionView_ShowsProgressBeforeTheBackendIsUp() =>
		UiTest.Run(ui => {
			ui.Find<Border>("ConnectionView").IsVisible.Should().BeTrue();
			ui.Find<ProgressBar>("ConnectionProgress").IsVisible.Should().BeTrue();
			ui.Find<Button>("RetryButton").IsVisible.Should().BeFalse();
		});

	[Fact]
	public Task ConnectionView_IsHiddenOnceConnected() =>
		UiTest.RunAsync(async ui => {
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);

			ui.Find<Border>("ConnectionView").IsVisible.Should().BeFalse();
		});

	[Fact]
	public Task ConnectionView_AfterAFailure_ShowsTheReasonAndARetryButton() =>
		UiTest.RunAsync(async ui => {
			ui.Backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("binary not found"));

			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);

			ui.Find<Border>("ConnectionView").IsVisible.Should().BeTrue();
			ui.Find<ProgressBar>("ConnectionProgress").IsVisible.Should().BeFalse();
			ui.Find<TextBlock>("ConnectionFailureText").Text.Should().Be("Failed to start backend: binary not found");
			ui.Find<Button>("RetryButton").IsVisible.Should().BeTrue();
		});

	[Fact]
	public Task RetryButton_RunsTheRestartBackendCommand() =>
		UiTest.Run(ui => ui.Find<Button>("RetryButton").Command.Should().BeSameAs(ui.ViewModel.RestartBackendCommand));

	[Fact]
	public Task ConnectionView_HidesAfterARetrySucceeds() =>
		UiTest.RunAsync(async ui => {
			ui.Backend.StartAsync(Arg.Any<CancellationToken>()).Returns(
				_ => Task.FromException(new BackendStartException("binary not found")),
				_ => Task.CompletedTask);
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);

			await ui.ViewModel.RestartBackendCommand.Execute();
			Settle(ui);

			ui.Find<Border>("ConnectionView").IsVisible.Should().BeFalse();
		});

	[Fact]
	public Task ErrorBar_ShowsAnIconAndADismissIconButton() =>
		UiTest.RunAsync(async ui => {
			ui.Backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("binary not found"));
			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);
			Settle(ui);

			ui.Find<PathIcon>("ErrorIcon").IsVisible.Should().BeTrue();
			ui.Find<Button>("DismissErrorButton").Command.Should().BeSameAs(ui.ViewModel.DismissErrorCommand);
		});
}
