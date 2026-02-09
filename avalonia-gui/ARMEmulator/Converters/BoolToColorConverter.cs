using System.Globalization;
using Avalonia.Data.Converters;
using Avalonia.Media;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a boolean value to a color brush.
/// True = green (flag set), False = gray (flag clear).
/// </summary>
public class BoolToColorConverter : IValueConverter
{
	public static readonly BoolToColorConverter Instance = new();

	private static readonly IBrush TrueBrush = new SolidColorBrush(Color.FromRgb(0, 200, 0));
	private static readonly IBrush FalseBrush = Brushes.Gray;

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not bool boolValue) {
			return FalseBrush;
		}

		return boolValue ? TrueBrush : FalseBrush;
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException("BoolToColorConverter does not support ConvertBack");
	}
}
