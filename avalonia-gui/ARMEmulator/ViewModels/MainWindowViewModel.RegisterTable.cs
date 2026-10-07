using ARMEmulator.Collections;
using ARMEmulator.Models;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

public partial class MainWindowViewModel
{
	/// <summary>The register table lines, with recently changed registers marked.</summary>
	public EquatableArray<RegisterRow> RegisterRows { get; private set; } = RegisterTable.ToRows(RegisterState.Create(), ImmutableHashSet<string>.Empty);

	/// <summary>The N, Z, C and V flags.</summary>
	public EquatableArray<FlagItem> RegisterFlags { get; private set; } = RegisterTable.ToFlags(RegisterState.Create());

	/// <summary>True while the flags are marked as recently changed.</summary>
	public bool AreFlagsChanged => ChangedRegisters.Contains("CPSR");

	private void RaiseRegisterTableChanged()
	{
		RegisterRows = RegisterTable.ToRows(Registers, ChangedRegisters);
		RegisterFlags = RegisterTable.ToFlags(Registers);
		this.RaisePropertyChanged(nameof(RegisterRows));
		this.RaisePropertyChanged(nameof(RegisterFlags));
		this.RaisePropertyChanged(nameof(AreFlagsChanged));
	}
}
