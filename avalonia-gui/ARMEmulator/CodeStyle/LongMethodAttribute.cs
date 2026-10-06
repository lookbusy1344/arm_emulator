namespace ARMEmulator.CodeStyle;

/// <summary>
/// Marks a method-like declaration as a reviewed exception to the executable-line guideline
/// enforced by <c>ARMEmulator.Tests.CodeStyle_MethodLength</c>: keeping the operation together was
/// accepted on purpose, not missed. The required reason sits on the declaration rather than in a
/// separate allowlist, so the exception cannot drift from why it was made. Do not apply this attribute
/// without explicit user permission.
/// </summary>
[AttributeUsage(AttributeTargets.Method | AttributeTargets.Constructor)]
public sealed class LongMethodAttribute : Attribute
{
	/// <summary>Creates a reviewed exception carrying its required justification.</summary>
	/// <exception cref="ArgumentException"><paramref name="reason"/> is null, empty or white space.</exception>
	public LongMethodAttribute(string reason)
	{
		ArgumentException.ThrowIfNullOrWhiteSpace(reason);
		Reason = reason;
	}

	/// <summary>Why this declaration is exempt from the executable-line guideline.</summary>
	public string Reason { get; }
}
