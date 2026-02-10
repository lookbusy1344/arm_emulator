using ARMEmulator.ViewModels;
using Avalonia.Controls;
using Avalonia.Interactivity;

namespace ARMEmulator.Views;

public partial class ConsoleView : UserControl
{
	public ConsoleView()
	{
		InitializeComponent();
		DataContextChanged += OnDataContextChanged;
	}

	private void OnDataContextChanged(object? sender, EventArgs e)
	{
		if (DataContext is MainWindowViewModel viewModel) {
			// Subscribe to ConsoleOutput changes for auto-scroll
			viewModel.PropertyChanged += (_, args) => {
				if (args.PropertyName == nameof(MainWindowViewModel.ConsoleOutput)) {
					ScrollToBottom();
				}
			};
		}
	}

	private void ScrollToBottom()
	{
		var scrollViewer = this.FindControl<ScrollViewer>("OutputScrollViewer");
		scrollViewer?.ScrollToEnd();
	}
}
