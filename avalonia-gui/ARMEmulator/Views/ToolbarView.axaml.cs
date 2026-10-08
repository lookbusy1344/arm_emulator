using ARMEmulator.Services;
using Avalonia.Automation;
using Avalonia.Controls;

namespace ARMEmulator.Views;

public partial class ToolbarView : UserControl
{
	public ToolbarView()
	{
		InitializeComponent();
		ApplyShortcutHints(OperatingSystem.IsMacOS());
	}

	/// <summary>Shows each command's shortcuts, in the platform's notation, as the button's tooltip and accelerator key.</summary>
	private void ApplyShortcutHints(bool isMacOS)
	{
		(Button Button, string Label, ShortcutId Id)[] buttons = [
			(RunButton, "Run", ShortcutId.Run),
			(PauseButton, "Pause", ShortcutId.Pause),
			(StepButton, "Step", ShortcutId.Step),
			(StepOverButton, "Step Over", ShortcutId.StepOver),
			(StepOutButton, "Step Out", ShortcutId.StepOut),
			(ResetButton, "Reset", ShortcutId.Reset),
			(ShowPcButton, "Show PC", ShortcutId.ShowPc)
		];

		foreach (var (button, label, id) in buttons) {
			ToolTip.SetTip(button, ShortcutText.Hint(label, id, isMacOS));
			AutomationProperties.SetAcceleratorKey(button, ShortcutText.Display(ShortcutBindings.FirstGesture(id, isMacOS)!, isMacOS));
		}
	}
}
