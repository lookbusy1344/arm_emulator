using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.ViewModels;

public class ExpressionEvaluatorViewModelTests
{
	private readonly IApiClient mockApi;
	private readonly ExpressionEvaluatorViewModel viewModel;
	private const string TestSessionId = "test-session";

	public ExpressionEvaluatorViewModelTests()
	{
		mockApi = Substitute.For<IApiClient>();
		viewModel = new ExpressionEvaluatorViewModel(mockApi);
		viewModel.SessionId = TestSessionId;
	}

	[Fact]
	public void Constructor_InitializesProperties()
	{
		var vm = new ExpressionEvaluatorViewModel(mockApi);

		vm.Expression.Should().BeEmpty();
		vm.Result.Should().BeNull();
		vm.ErrorMessage.Should().BeNull();
		vm.History.Should().BeEmpty();
		vm.SessionId.Should().BeNull();
	}

	[Fact]
	public async Task EvaluateCommand_WithValidExpression_SetsResult()
	{
		// Arrange
		const uint expectedValue = 0x42;
		viewModel.Expression = "r0";
		mockApi.EvaluateExpressionAsync(TestSessionId, "r0", Arg.Any<CancellationToken>())
			.Returns(expectedValue);

		// Act
		await viewModel.EvaluateCommand.Execute();

		// Assert
		viewModel.Result.Should().Be(expectedValue);
		viewModel.ErrorMessage.Should().BeNull();
		viewModel.History.Should().HaveCount(1);
		viewModel.History[0].Expression.Should().Be("r0");
		viewModel.History[0].Result.Should().Be(expectedValue);
	}

	[Fact]
	public async Task EvaluateCommand_WithInvalidExpression_SetsErrorMessage()
	{
		// Arrange
		viewModel.Expression = "invalid";
		mockApi.EvaluateExpressionAsync(TestSessionId, "invalid", Arg.Any<CancellationToken>())
			.Returns<uint>(_ => throw new ExpressionEvaluationException("invalid", "Unknown register"));

		// Act
		await viewModel.EvaluateCommand.Execute();

		// Assert
		viewModel.Result.Should().BeNull();
		viewModel.ErrorMessage.Should().Contain("Unknown register");
		viewModel.History.Should().BeEmpty();
	}

	[Fact]
	public async Task EvaluateCommand_WithoutSessionId_SetsErrorMessage()
	{
		// Arrange
		var vm = new ExpressionEvaluatorViewModel(mockApi);
		vm.Expression = "r0";

		// Act
		await vm.EvaluateCommand.Execute();

		// Assert
		vm.Result.Should().BeNull();
		vm.ErrorMessage.Should().Contain("No active session");
	}

	[Fact]
	public async Task EvaluateCommand_WithEmptyExpression_DoesNothing()
	{
		// Arrange
		viewModel.Expression = "";

		// Act
		await viewModel.EvaluateCommand.Execute();

		// Assert
		await mockApi.DidNotReceiveWithAnyArgs().EvaluateExpressionAsync(default!, default!, default);
		viewModel.History.Should().BeEmpty();
	}

	[Fact]
	public async Task EvaluateCommand_AddsToHistory()
	{
		// Arrange
		mockApi.EvaluateExpressionAsync(Arg.Any<string>(), Arg.Any<string>(), Arg.Any<CancellationToken>())
			.Returns(callInfo => Task.FromResult((uint)(callInfo.ArgAt<string>(1).Length)));

		// Act
		viewModel.Expression = "r0";
		await viewModel.EvaluateCommand.Execute();

		viewModel.Expression = "r1";
		await viewModel.EvaluateCommand.Execute();

		viewModel.Expression = "r0+r1";
		await viewModel.EvaluateCommand.Execute();

		// Assert - History should show newest first
		viewModel.History.Should().HaveCount(3);
		viewModel.History[0].Expression.Should().Be("r0+r1"); // Most recent
		viewModel.History[1].Expression.Should().Be("r1");
		viewModel.History[2].Expression.Should().Be("r0"); // Oldest
	}

	[Fact]
	public void ClearHistoryCommand_ClearsHistory()
	{
		// Arrange
		var history = new List<ExpressionResult>
		{
			new("r0", 42, DateTime.Now),
			new("r1", 100, DateTime.Now)
		};
		viewModel.History = [.. history];

		// Act
		viewModel.ClearHistoryCommand.Execute().Subscribe();

		// Assert
		viewModel.History.Should().BeEmpty();
	}

	[Fact]
	public void FormatHex_FormatsCorrectly()
	{
		ExpressionEvaluatorViewModel.FormatHex(0x42).Should().Be("0x00000042");
		ExpressionEvaluatorViewModel.FormatHex(0xFFFFFFFF).Should().Be("0xFFFFFFFF");
		ExpressionEvaluatorViewModel.FormatHex(0).Should().Be("0x00000000");
	}

	[Fact]
	public void FormatDecimal_FormatsCorrectly()
	{
		ExpressionEvaluatorViewModel.FormatDecimal(42).Should().Be("42");
		ExpressionEvaluatorViewModel.FormatDecimal(0).Should().Be("0");
		ExpressionEvaluatorViewModel.FormatDecimal(4294967295).Should().Be("4294967295");
	}

	[Fact]
	public void FormatBinary_FormatsCorrectly()
	{
		ExpressionEvaluatorViewModel.FormatBinary(0b1010).Should().Be("00000000000000000000000000001010");
		ExpressionEvaluatorViewModel.FormatBinary(0xFFFFFFFF).Should().Be("11111111111111111111111111111111");
		ExpressionEvaluatorViewModel.FormatBinary(0).Should().Be("00000000000000000000000000000000");
	}
}
