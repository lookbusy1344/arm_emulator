using System.Globalization;
using Avalonia.Data.Converters;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a boolean value to a breakpoint indicator symbol.
/// </summary>
public class BoolToBreakpointConverter : IValueConverter
{
	public object Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is bool hasBreakpoint && hasBreakpoint) {
			return "●";  // Red circle for breakpoint
		}

		return string.Empty;
	}

	public object ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException();
	}
}
