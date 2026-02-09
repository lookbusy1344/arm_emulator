using System.Globalization;
using Avalonia.Data.Converters;
using Avalonia.Media;

namespace ARMEmulator.Converters;

/// <summary>
/// Converts a register name and changed registers set to a highlight brush.
/// Returns a green brush if the register is in the changed set, otherwise transparent.
/// </summary>
public class RegisterHighlightConverter : IMultiValueConverter
{
	public static readonly RegisterHighlightConverter Instance = new();

	// Green highlight for changed registers
	private static readonly IBrush HighlightBrush = new SolidColorBrush(Color.FromArgb(128, 0, 255, 0));
	private static readonly IBrush NormalBrush = Brushes.Transparent;

	public object? Convert(IList<object?> values, Type targetType, object? parameter, CultureInfo culture)
	{
		if (values.Count != 2) {
			return NormalBrush;
		}

		// values[0] is the register name (string)
		// values[1] is the changed registers set (ImmutableHashSet<string>)
		if (values[0] is not string registerName) {
			return NormalBrush;
		}

		if (values[1] is not ImmutableHashSet<string> changedRegisters) {
			return NormalBrush;
		}

		return changedRegisters.Contains(registerName) ? HighlightBrush : NormalBrush;
	}
}
