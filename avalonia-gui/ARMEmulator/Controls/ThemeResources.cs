using Avalonia.Controls;
using Avalonia.Media;

namespace ARMEmulator.Controls;

/// <summary>
/// Looks up theme resources for code that draws or styles outside XAML.
/// </summary>
public static class ThemeResources
{
	/// <summary>The brush stored under <paramref name="key"/> for the control's current theme variant, or null when there is none.</summary>
	public static IBrush? FindBrush(this Control host, string key) =>
		host.TryFindResource(key, host.ActualThemeVariant, out var resource) ? resource as IBrush : null;

	/// <summary>The geometry stored under <paramref name="key"/> for the control's current theme variant, or null when there is none.</summary>
	public static Geometry? FindGeometry(this Control host, string key) =>
		host.TryFindResource(key, host.ActualThemeVariant, out var resource) ? resource as Geometry : null;
}
