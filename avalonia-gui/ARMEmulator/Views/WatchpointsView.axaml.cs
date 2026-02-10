using System.Globalization;
using Avalonia.Controls;
using Avalonia.Interactivity;
using ARMEmulator.Models;
using ARMEmulator.ViewModels;

namespace ARMEmulator.Views;

public partial class WatchpointsView : UserControl
{
	public WatchpointsView()
	{
		InitializeComponent();
	}

	// Justification: Event handlers must be async void per Avalonia framework requirements
#pragma warning disable VSTHRD100
	private async void AddWatchpointButton_Click(object? sender, RoutedEventArgs e)
#pragma warning restore VSTHRD100
	{
		if (DataContext is not MainWindowViewModel vm)
		{
			return;
		}

		if (vm.SessionId is null)
		{
			// TODO: Show error message
			return;
		}

		// Get address input
		var addressInput = this.FindControl<TextBox>("WatchpointAddressInput");
		var typeCombo = this.FindControl<ComboBox>("WatchpointTypeCombo");

		if (addressInput?.Text is not { } addressStr || string.IsNullOrWhiteSpace(addressStr))
		{
			return;
		}

		// Parse address
		var hexStr = addressStr.StartsWith("0x", StringComparison.OrdinalIgnoreCase)
			? addressStr[2..]
			: addressStr;

		if (!uint.TryParse(hexStr, NumberStyles.HexNumber, CultureInfo.InvariantCulture, out var address))
		{
			// TODO: Show error message
			return;
		}

		// Parse type
		var type = (typeCombo?.SelectedIndex ?? 2) switch
		{
			0 => WatchpointType.Read,
			1 => WatchpointType.Write,
			_ => WatchpointType.ReadWrite
		};

		try
		{
			await vm.AddWatchpointAsync(address, type);

			// Clear input on success
			if (addressInput is not null)
			{
				addressInput.Text = "";
			}
		}
		catch
		{
			// TODO: Show error message
		}
	}

	// Justification: Event handlers must be async void per Avalonia framework requirements
#pragma warning disable VSTHRD100
	private async void RemoveWatchpointButton_Click(object? sender, RoutedEventArgs e)
#pragma warning restore VSTHRD100
	{
		if (DataContext is not MainWindowViewModel vm)
		{
			return;
		}

		if (sender is Button { Tag: int id })
		{
			try
			{
				await vm.RemoveWatchpointAsync(id);
			}
			catch
			{
				// TODO: Show error message
			}
		}
	}
}
