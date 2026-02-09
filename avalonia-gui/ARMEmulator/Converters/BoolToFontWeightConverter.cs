using System.Globalization;
using Avalonia.Data.Converters;
using Avalonia.Media;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a boolean value to a font weight.
/// True = Bold (flag set), False = Normal (flag clear).
/// </summary>
public class BoolToFontWeightConverter : IValueConverter
{
	public static readonly BoolToFontWeightConverter Instance = new();

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not bool boolValue) {
			return FontWeight.Normal;
		}

		return boolValue ? FontWeight.Bold : FontWeight.Normal;
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException("BoolToFontWeightConverter does not support ConvertBack");
	}
}
