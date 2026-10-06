namespace ARMEmulator.CodeStyle;

/// <summary>
/// Marks a value type authored on purpose above <c>ARMEmulator.Tests.CodeStyle_StructSize</c>'s
/// 24-byte guideline ceiling. The reason sits on the declaration rather than in a separate allowlist,
/// so the exception cannot drift from why it was made. Do not apply this attribute without explicit
/// user permission.
/// </summary>
[AttributeUsage(AttributeTargets.Struct)]
public sealed class LargeStructAttribute(string reason) : Attribute
{
	/// <summary>Why this value type is exempt from the size guideline.</summary>
	public string Reason { get; } = reason;
}
