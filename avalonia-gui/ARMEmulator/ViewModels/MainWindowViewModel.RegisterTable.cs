using ARMEmulator.Collections;
using ARMEmulator.Models;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

public partial class MainWindowViewModel
{
	/// <summary>The register table lines, with recently changed registers marked.</summary>
	public EquatableArray<RegisterRow> RegisterRows => RegisterTable.Rows(Registers, ChangedRegisters);

	/// <summary>The N, Z, C and V flags.</summary>
	public EquatableArray<FlagItem> RegisterFlags => RegisterTable.Flags(Registers);

	/// <summary>True while the flags are marked as recently changed.</summary>
	public bool AreFlagsChanged => ChangedRegisters.Contains("CPSR");

	private void RaiseRegisterTableChanged()
	{
		this.RaisePropertyChanged(nameof(RegisterRows));
		this.RaisePropertyChanged(nameof(RegisterFlags));
		this.RaisePropertyChanged(nameof(AreFlagsChanged));
	}
}
