using System.Reactive;
using System.Reactive.Disposables.Fluent;
using ARMEmulator.Services;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

/// <summary>
/// Backend lifecycle: start, status display and restart.
/// </summary>
public partial class MainWindowViewModel
{
	private IBackendManager? backendManager;

	/// <summary>Status of the backend process, as last reported by the backend manager.</summary>
	public BackendStatus BackendStatus
	{
		get => backendStatus;
		private set => this.RaiseAndSetIfChanged(ref backendStatus, value);
	}

	private BackendStatus backendStatus = BackendStatus.Unknown;

	/// <summary>Stops the backend, starts it again, opens a new session and reloads the current program.</summary>
	public ReactiveCommand<Unit, Unit> RestartBackendCommand { get; private set; } = null!;

	/// <summary>
	/// Starts the backend, then creates a session and connects the WebSocket.
	/// Failures are reported through <see cref="ErrorMessage"/>.
	/// </summary>
	public async Task StartAsync(IBackendManager backend, CancellationToken ct = default)
	{
		ArgumentNullException.ThrowIfNull(backend);

		backendManager = backend;
		BackendStatus = backend.Status;
		_ = backend.StatusChanged
			.ObserveOn(RxSchedulers.MainThreadScheduler)
			.Subscribe(status => BackendStatus = status)
			.DisposeWith(disposables);

		_ = await StartBackendAndSessionAsync(backend, ct);
	}

	private async Task<bool> StartBackendAndSessionAsync(IBackendManager backend, CancellationToken ct)
	{
		try {
			await backend.StartAsync(ct);
		}
		catch (BackendStartException ex) {
			ErrorMessage = $"Failed to start backend: {ex.Message}";
			return false;
		}

		try {
			await CreateSessionAsync(ct);
			ErrorMessage = null;
			return true;
		}
		catch (ApiException ex) {
			ErrorMessage = $"Failed to connect to backend: {ex.Message}";
			return false;
		}
	}

	private async Task RestartBackendAsync(CancellationToken ct)
	{
		var backend = backendManager ?? throw new InvalidOperationException("no backend to restart");
		var reloadSource = isProgramLoaded;

		await ws.DisconnectAsync();
		try {
			await backend.StopAsync();
		}
		catch (Exception ex) when (ex is not OperationCanceledException) {
			ErrorMessage = $"Failed to stop backend: {ex.Message}";
			return;
		}

		// The old session died with the process
		SessionId = null;
		IsConnected = false;
		isProgramLoaded = false;

		if (await StartBackendAndSessionAsync(backend, ct) && reloadSource) {
			await LoadProgramAsync(ct);
		}
	}
}
