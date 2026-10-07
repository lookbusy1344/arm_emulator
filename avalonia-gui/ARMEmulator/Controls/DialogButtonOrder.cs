using Avalonia.Controls;

namespace ARMEmulator.Controls;

/// <summary>
/// Orders the buttons of a dialog by platform convention. Views list them the macOS and Linux way, primary button last;
/// Windows puts the primary button first.
/// </summary>
public static class DialogButtonOrder
{
	/// <summary>Reverses the buttons in <paramref name="row"/> when <paramref name="isWindows"/> is true.</summary>
	public static void Apply(Panel row, bool isWindows)
	{
		if (!isWindows) {
			return;
		}

		var buttons = row.Children.Reverse().ToList();
		row.Children.Clear();
		row.Children.AddRange(buttons);
	}

	/// <summary>Orders <paramref name="row"/> for the platform the application runs on.</summary>
	public static void ForCurrentPlatform(Panel row) => Apply(row, OperatingSystem.IsWindows());
}
