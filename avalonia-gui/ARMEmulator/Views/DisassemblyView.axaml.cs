using System.ComponentModel;
using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Threading;

namespace ARMEmulator.Views;

public partial class DisassemblyView : UserControl
{
	private DisassemblyViewModel? observed;

	public DisassemblyView()
	{
		InitializeComponent();
		DataContextChanged += OnDataContextChanged;
	}

	private void OnDataContextChanged(object? sender, EventArgs e)
	{
		if (observed is not null) {
			observed.PropertyChanged -= OnViewModelPropertyChanged;
		}

		observed = DataContext as DisassemblyViewModel;
		if (observed is not null) {
			observed.PropertyChanged += OnViewModelPropertyChanged;
		}
	}

	private void OnViewModelPropertyChanged(object? sender, PropertyChangedEventArgs args)
	{
		if (args.PropertyName == nameof(DisassemblyViewModel.FormattedInstructions)) {
			// Wait for the rows to be laid out before scrolling to one
			Dispatcher.UIThread.Post(ScrollCurrentRowIntoView, DispatcherPriority.Loaded);
		}
	}

	private void ScrollCurrentRowIntoView()
	{
		var rows = observed?.FormattedInstructions;
		var index = rows?.FindIndex(row => row.IsCurrentPC) ?? -1;
		DisassemblyRows.ContainerFromIndex(index)?.BringIntoView();
	}
}
