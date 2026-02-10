using Avalonia.Controls;
using Avalonia.Interactivity;
using ARMEmulator.ViewModels;

namespace ARMEmulator.Views;

public partial class BreakpointsListView : UserControl
{
	public BreakpointsListView()
	{
		InitializeComponent();
	}

	// Justification: Event handlers must be async void per Avalonia framework requirements
#pragma warning disable VSTHRD100
	private async void RemoveBreakpointButton_Click(object? sender, RoutedEventArgs e)
#pragma warning restore VSTHRD100
	{
		if (DataContext is not MainWindowViewModel vm)
		{
			return;
		}

		if (sender is Button { Tag: uint address })
		{
			try
			{
				await vm.RemoveBreakpointAsync(address);
			}
			catch
			{
				// TODO: Show error message
			}
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
