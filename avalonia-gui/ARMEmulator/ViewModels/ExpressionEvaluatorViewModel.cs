using System.Globalization;
using System.Reactive;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged, which triggers IL2026 warnings
// This is acceptable since we don't use AOT compilation for this project
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// ViewModel for the expression evaluator view.
/// Allows users to evaluate ARM expressions (registers, memory, arithmetic).
/// </summary>
public class ExpressionEvaluatorViewModel : ReactiveObject
{
	private readonly IApiClient api;

	public ExpressionEvaluatorViewModel(IApiClient api)
	{
		this.api = api;

		// Create commands
		EvaluateCommand = ReactiveCommand.CreateFromTask(EvaluateAsync);
		ClearHistoryCommand = ReactiveCommand.Create(ClearHistory);
	}

	// Reactive properties
	private string expression = "";

	public string Expression
	{
		get => expression;
		set => this.RaiseAndSetIfChanged(ref expression, value);
	}

	private uint? result;

	public uint? Result
	{
		get => result;
		set => this.RaiseAndSetIfChanged(ref result, value);
	}

	private string? errorMessage;

	public string? ErrorMessage
	{
		get => errorMessage;
		set => this.RaiseAndSetIfChanged(ref errorMessage, value);
	}

	private ImmutableArray<ExpressionResult> history = [];

	public ImmutableArray<ExpressionResult> History
	{
		get => history;
		set => this.RaiseAndSetIfChanged(ref history, value);
	}

	private string? sessionId;

	public string? SessionId
	{
		get => sessionId;
		set => this.RaiseAndSetIfChanged(ref sessionId, value);
	}

	// Commands
	public ReactiveCommand<Unit, Unit> EvaluateCommand { get; }
	public ReactiveCommand<Unit, Unit> ClearHistoryCommand { get; }

	// Evaluation logic
	private async Task EvaluateAsync(CancellationToken ct)
	{
		// Clear previous result and error
		Result = null;
		ErrorMessage = null;

		// Validate inputs
		if (string.IsNullOrWhiteSpace(Expression)) {
			return;
		}

		if (SessionId is null) {
			ErrorMessage = "No active session. Load a program first.";
			return;
		}

		try {
			var value = await api.EvaluateExpressionAsync(SessionId, Expression, ct);
			Result = value;

			// Add to history (most recent first)
			History = [new ExpressionResult(Expression, value, DateTime.Now), .. History];
		}
		catch (ExpressionEvaluationException ex) {
			ErrorMessage = ex.Message;
		}
		catch (SessionNotFoundException) {
			ErrorMessage = "Session expired. Please reload the program.";
		}
		catch (ApiException ex) {
			ErrorMessage = $"API error: {ex.Message}";
		}
		catch (Exception ex) {
			ErrorMessage = $"Unexpected error: {ex.Message}";
		}
	}

	private void ClearHistory()
	{
		History = [];
	}

	// Formatting helpers (static for easy testing)
	public static string FormatHex(uint value) => $"0x{value:X8}";

	public static string FormatDecimal(uint value) => value.ToString(CultureInfo.InvariantCulture);

	public static string FormatBinary(uint value) => Convert.ToString(value, 2).PadLeft(32, '0');
}
