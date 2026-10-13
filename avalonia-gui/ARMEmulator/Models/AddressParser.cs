using System.Globalization;

namespace ARMEmulator.Models;

/// <summary>
/// Reads a 32-bit address typed by the user: <c>0x</c> and hex digits, or a bare decimal number.
/// </summary>
public static class AddressParser
{
	private const string HexPrefix = "0x";

	/// <summary>Parses <paramref name="text"/>, ignoring surrounding whitespace.</summary>
	/// <returns>True when the text is a valid address; <paramref name="address"/> is 0 otherwise.</returns>
	public static bool TryParse(string? text, out uint address)
	{
		var trimmed = (text ?? "").AsSpan().Trim();
		return trimmed.StartsWith(HexPrefix, StringComparison.OrdinalIgnoreCase)
			? uint.TryParse(trimmed[HexPrefix.Length..], NumberStyles.AllowHexSpecifier, CultureInfo.InvariantCulture, out address)
			: uint.TryParse(trimmed, NumberStyles.None, CultureInfo.InvariantCulture, out address);
	}
}
