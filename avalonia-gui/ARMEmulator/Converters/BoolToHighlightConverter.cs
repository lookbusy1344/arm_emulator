using System.Globalization;
using Avalonia.Data.Converters;
using Avalonia.Media;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a boolean value to a background brush for highlighting memory writes.
/// </summary>
public class BoolToHighlightConverter : IValueConverter
{
	private static readonly Color HighlightColor = Color.FromRgb(255, 255, 200); // Light yellow
	private static readonly IBrush NormalBrush = Brushes.Transparent;

	public object Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is bool isHighlighted && isHighlighted) {
			// A brush belongs to the thread that creates it, so each call makes its own
			return new SolidColorBrush(HighlightColor);
		}

		return NormalBrush;
	}

	public object ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException();
	}
}
