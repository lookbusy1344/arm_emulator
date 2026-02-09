using Avalonia.Controls;

namespace ARMEmulator.Views;

/// <summary>
/// View for displaying ARM register state with highlighting for changed values.
/// </summary>
public partial class RegistersView : UserControl
{
	public RegistersView()
	{
		InitializeComponent();
	}
}
