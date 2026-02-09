using System.Globalization;
using Avalonia.Data.Converters;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts uint values to hexadecimal string representation (0x00000000 format).
/// </summary>
public class HexValueConverter : IValueConverter
{
	public static readonly HexValueConverter Instance = new();

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not uint uintValue) {
			return "0x00000000";
		}

		return $"0x{uintValue:X8}";
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not string str) {
			return 0u;
		}

		// Remove 0x prefix if present
		var hexStr = str.StartsWith("0x", StringComparison.OrdinalIgnoreCase)
			? str[2..]
			: str;

		return uint.TryParse(hexStr, NumberStyles.HexNumber, culture, out var result)
			? result
			: 0u;
	}
}
