using System.Globalization;
using Avalonia.Data.Converters;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts uint values to binary string representation (32-bit format).
/// </summary>
public class BinaryValueConverter : IValueConverter
{
	public static readonly BinaryValueConverter Instance = new();

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not uint uintValue)
		{
			return "00000000000000000000000000000000";
		}

		return System.Convert.ToString(uintValue, 2).PadLeft(32, '0');
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not string str)
		{
			return 0u;
		}

		try
		{
			return System.Convert.ToUInt32(str, 2);
		}
		catch
		{
			return 0u;
		}
	}
}
