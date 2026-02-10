namespace ARMEmulator.Models;

/// <summary>
/// Represents the result of an expression evaluation with timestamp.
/// </summary>
public sealed record ExpressionResult(string Expression, uint Result, DateTime Timestamp);
