using ARMEmulator.Services;
using Avalonia.Controls;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;

namespace ARMEmulator.Tests.Ui;

public sealed class ErrorBarTests
{
	[Fact]
	public Task ErrorBar_IsHiddenWhenThereIsNoError() =>
		UiTest.Run(ui => ui.Find<Border>("ErrorBar").IsVisible.Should().BeFalse());

	[Fact]
	public Task BackendStartFailure_ShowsTheErrorBar() =>
		UiTest.RunAsync(async ui => {
			ui.Backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("binary not found"));

			await ui.ViewModel.StartAsync(ui.Backend, TestContext.Current.CancellationToken);

			ui.Find<Border>("ErrorBar").IsVisible.Should().BeTrue();
			ui.Find<SelectableTextBlock>("ErrorText").Text.Should().Be("Failed to start backend: binary not found");
		});
}
