using ARMEmulator.Models;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

public partial class MainWindowViewModel
{
	/// <summary>The inspector panel on show. It is saved with the window geometry when the window closes.</summary>
	public InspectorPanel SelectedInspectorPanel
	{
		get => Settings.Layout.SelectedPanel;
		set
		{
			if (value == SelectedInspectorPanel) {
				return;
			}

			Settings = Settings with { Layout = Settings.Layout with { SelectedPanel = value } };
			this.RaisePropertyChanged(nameof(SelectedInspectorPanel));
		}
	}

	/// <summary>Saves where the window and its splitters were left.</summary>
	public void SaveWindowGeometry(WindowGeometry geometry) =>
		SaveSettings(Settings with { Layout = Settings.Layout with { Geometry = geometry } });
}
