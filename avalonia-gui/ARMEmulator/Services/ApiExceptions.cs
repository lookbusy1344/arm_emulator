using System.Diagnostics.CodeAnalysis;
using System.Net;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// Base exception for all API-related errors.
/// </summary>
[SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Uses custom constructor with domain-specific parameters")]
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
[SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Uses primary constructor with domain-specific parameters")]
public sealed class SessionNotFoundException(string sessionId)
	: ApiException($"Session '{sessionId}' not found or expired", HttpStatusCode.NotFound)
{
	/// <summary>The session ID that was not found.</summary>
	public string SessionId { get; } = sessionId;
}

/// <summary>
/// Thrown when program loading fails due to parse or assembly errors.
/// </summary>
[SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Uses primary constructor with domain-specific parameters")]
public sealed class ProgramLoadException(ImmutableArray<ParseError> errors)
	: ApiException($"Program failed to load: {errors.Length} error(s)", HttpStatusCode.BadRequest)
{
	/// <summary>List of parse errors from the assembler.</summary>
	public ImmutableArray<ParseError> Errors => errors;
}

/// <summary>
/// Thrown when the backend is unreachable or not responding.
/// </summary>
[SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Uses primary constructor with domain-specific parameters")]
public sealed class BackendUnavailableException(string message, Exception? inner = null)
	: ApiException(message, null, inner);

/// <summary>
/// Thrown when an expression evaluation fails.
/// </summary>
[SuppressMessage("Design", "RCS1194:Implement exception constructors", Justification = "Uses primary constructor with domain-specific parameters")]
public sealed class ExpressionEvaluationException(string expression, string error)
	: ApiException($"Failed to evaluate '{expression}': {error}", HttpStatusCode.BadRequest)
{
	/// <summary>The expression that failed to evaluate.</summary>
	public string Expression => expression;
}
