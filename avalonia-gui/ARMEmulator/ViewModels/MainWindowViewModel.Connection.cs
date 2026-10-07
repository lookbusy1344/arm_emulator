using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

/// <summary>
/// Connection state for the "starting" and "failed" view.
/// </summary>
public partial class MainWindowViewModel
{
	private string? connectionFailure;

	/// <summary>Why the last attempt to reach the backend failed. Null while connecting or connected. Dismissing the error bar leaves it in place.</summary>
	public string? ConnectionFailure
	{
		get => connectionFailure;
		private set
		{
			_ = this.RaiseAndSetIfChanged(ref connectionFailure, value);
			RaiseConnectionChanged();
		}
	}

	/// <summary>Where the application stands in reaching its backend.</summary>
	public ConnectionState Connection => (IsConnected, ConnectionFailure) switch {
		(true, _) => ConnectionState.Connected,
		(false, null) => ConnectionState.Connecting,
		_ => ConnectionState.Failed
	};

	/// <summary>True when there is no source text.</summary>
	public bool IsEditorEmpty => string.IsNullOrEmpty(SourceCode);

	private void RaiseConnectionChanged() => this.RaisePropertyChanged(nameof(Connection));

	private void ReportConnectionFailure(string message)
	{
		ErrorMessage = message;
		ConnectionFailure = message;
	}
}
