using ARMEmulator.Models;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

public partial class MainWindowViewModel
{
	/// <summary>The inspector panel on show. A change is saved with the settings.</summary>
	public InspectorPanel SelectedInspectorPanel
	{
		get => Settings.Layout.SelectedPanel;
		set
		{
			if (value == SelectedInspectorPanel) {
				return;
			}

			SaveSettings(Settings with { Layout = Settings.Layout with { SelectedPanel = value } });
		}
	}
}
