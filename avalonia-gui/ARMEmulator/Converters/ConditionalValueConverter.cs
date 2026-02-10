using System.Globalization;
using Avalonia.Data.Converters;
using Avalonia.Media;

namespace ARMEmulator.Converters;

/// <summary>
/// Converter that returns one of two values based on a boolean condition.
/// Usage: {MultiBinding Converter={x:Static ConditionalValueConverter.Instance}}
///   - First binding: boolean condition
///   - Second binding: value if true
///   - Third binding: value if false
/// </summary>
public class ConditionalValueConverter : IMultiValueConverter
{
	public static readonly ConditionalValueConverter Instance = new();

	public object? Convert(IList<object?> values, Type targetType, object? parameter, CultureInfo culture)
	{
		if (values.Count != 3) {
			return null;
		}

		var condition = values[0] as bool? ?? false;
		var trueValue = values[1];
		var falseValue = values[2];

		return condition ? trueValue : falseValue;
	}
}
