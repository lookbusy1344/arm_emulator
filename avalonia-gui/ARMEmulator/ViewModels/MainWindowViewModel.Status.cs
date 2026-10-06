using System.Reactive.Disposables.Fluent;
using ARMEmulator.Models;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

/// <summary>
/// The status pill in the toolbar.
/// </summary>
public partial class MainWindowViewModel
{
	private const string DisconnectedKey = "Disconnected";

	/// <summary>Names the pill's state: "Disconnected", or the <see cref="VMState"/> name. Styles select the pill colour from it.</summary>
	public string StatusKey => IsConnected ? Status.ToString() : DisconnectedKey;

	/// <summary>Short text for the pill.</summary>
	public string StatusLabel => !IsConnected
		? DisconnectedKey
		: Status switch {
			VMState.WaitingForInput => "Waiting for input",
			_ => Status.ToString()
		};

	private void InitializeStatusPill() =>
		this.WhenAnyValue(x => x.Status, x => x.IsConnected, static (status, connected) => (status, connected))
			.Subscribe(_ => {
				this.RaisePropertyChanged(nameof(StatusKey));
				this.RaisePropertyChanged(nameof(StatusLabel));
			})
			.DisposeWith(disposables);
}
