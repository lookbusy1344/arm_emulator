using System.Globalization;
using Avalonia.Data.Converters;
using ARMEmulator.Models;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts WatchpointType to an icon string.
/// </summary>
public class WatchpointTypeIconConverter : IValueConverter
{
	public static readonly WatchpointTypeIconConverter Instance = new();

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		if (value is not WatchpointType type)
		{
			return "👁";
		}

		return type switch
		{
			WatchpointType.Read => "👁",
			WatchpointType.Write => "✏️",
			WatchpointType.ReadWrite => "👁✏️",
			_ => "👁"
		};
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		throw new NotSupportedException();
	}
}
