using System.Net;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Base exception for all API-related errors.
/// </summary>
public class ApiException : Exception
{
	/// <summary>HTTP status code if available.</summary>
	public HttpStatusCode? StatusCode { get; }

	public ApiException(string message, HttpStatusCode? statusCode = null, Exception? inner = null)
		: base(message, inner)
	{
		StatusCode = statusCode;
	}
}

/// <summary>
/// Thrown when a session is not found or has expired.
/// </summary>
public sealed class SessionNotFoundException(string sessionId)
	: ApiException($"Session '{sessionId}' not found or expired", HttpStatusCode.NotFound)
{
	/// <summary>The session ID that was not found.</summary>
	public string SessionId { get; } = sessionId;
}

/// <summary>
/// Thrown when program loading fails due to parse or assembly errors.
/// </summary>
public sealed class ProgramLoadException : ApiException
{
	/// <summary>List of parse errors from the assembler.</summary>
	public ImmutableArray<ParseError> Errors { get; }

	public ProgramLoadException(ImmutableArray<ParseError> errors)
		: base($"Program failed to load: {errors.Length} error(s)", HttpStatusCode.BadRequest)
	{
		Errors = errors;
	}
}

/// <summary>
/// Thrown when the backend is unreachable or not responding.
/// </summary>
public sealed class BackendUnavailableException(string message, Exception? inner = null)
	: ApiException(message, null, inner);

/// <summary>
/// Thrown when an expression evaluation fails.
/// </summary>
public sealed class ExpressionEvaluationException : ApiException
{
	/// <summary>The expression that failed to evaluate.</summary>
	public string Expression { get; }

	public ExpressionEvaluationException(string expression, string error)
		: base($"Failed to evaluate '{expression}': {error}", HttpStatusCode.BadRequest)
	{
		Expression = expression;
	}
}
