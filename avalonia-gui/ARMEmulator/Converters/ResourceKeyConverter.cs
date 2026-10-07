using System.Globalization;
using Avalonia;
using Avalonia.Data.Converters;

namespace ARMEmulator.Converters;

/// <summary>
/// Looks up an application resource by key, such as an icon geometry.
/// </summary>
public sealed class ResourceKeyConverter : IValueConverter
{
	public static ResourceKeyConverter Instance { get; } = new();

	public object? Convert(object? value, Type targetType, object? parameter, CultureInfo culture)
	{
		var app = Application.Current;
		return value is string key && app is not null && app.TryGetResource(key, app.ActualThemeVariant, out var resource)
			? resource
			: AvaloniaProperty.UnsetValue;
	}

	public object? ConvertBack(object? value, Type targetType, object? parameter, CultureInfo culture) =>
		throw new NotSupportedException();
}
