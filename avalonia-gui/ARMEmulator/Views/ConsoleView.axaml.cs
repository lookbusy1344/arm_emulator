using System.ComponentModel;
using ARMEmulator.ViewModels;
using Avalonia.Controls;

namespace ARMEmulator.Views;

public partial class ConsoleView : UserControl
{
	private MainWindowViewModel? observed;

	public ConsoleView()
	{
		InitializeComponent();
		DataContextChanged += OnDataContextChanged;
	}

	private void OnDataContextChanged(object? sender, EventArgs e)
	{
		if (observed is not null) {
			observed.PropertyChanged -= OnViewModelPropertyChanged;
		}

		observed = DataContext as MainWindowViewModel;
		if (observed is not null) {
			observed.PropertyChanged += OnViewModelPropertyChanged;
		}
	}

	private void OnViewModelPropertyChanged(object? sender, PropertyChangedEventArgs args)
	{
		if (args.PropertyName == nameof(MainWindowViewModel.ConsoleOutput)) {
			OutputScrollViewer.ScrollToEnd();
		}
	}
}
