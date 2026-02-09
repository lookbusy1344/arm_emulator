using System.Globalization;
using Avalonia.Data.Converters;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a boolean value to a PC (Program Counter) indicator symbol.
/// </summary>
public class BoolToPCIndicatorConverter : IValueConverter
{
	public object Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is bool isCurrentPC && isCurrentPC) {
			return "→";  // Arrow for current PC
		}

		return string.Empty;
	}

	public object ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException();
	}
}
