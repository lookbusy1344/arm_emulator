using System.Collections.Immutable;
using System.Reactive;
using System.Reactive.Disposables;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged, which triggers IL2026 warnings
// This is acceptable since we don't use AOT compilation for this project
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// Central ViewModel for the main window, managing emulator state and user interactions.
/// Uses ReactiveUI for MVVM with reactive state management.
/// </summary>
public partial class MainWindowViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient _api;
	private readonly IWebSocketClient _ws;
	private readonly CompositeDisposable _disposables = new();
	private readonly Subject<string> _registerHighlightTrigger = new();
	private static readonly TimeSpan HighlightDuration = TimeSpan.FromSeconds(1.5);

	/// <summary>
	/// Initializes a new instance of the MainWindowViewModel with required services.
	/// </summary>
	public MainWindowViewModel(IApiClient api, IWebSocketClient ws)
	{
		_api = api;
		_ws = ws;

		// Initialize commands with can-execute observables
		RunCommand = CreateCommand(RunAsync, this.WhenAnyValue(x => x.Status).Select(s => !s.CanPause()));
		PauseCommand = CreateCommand(PauseAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanPause()));
		StepCommand = CreateCommand(StepAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		StepOverCommand = CreateCommand(StepOverAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		StepOutCommand = CreateCommand(StepOutAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		ResetCommand = CreateCommand(ResetAsync);
		LoadProgramCommand = CreateCommand(LoadProgramAsync);

		// Set up computed properties using WhenAnyValue
		_canPauseHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.CanPause())
			.ToProperty(this, x => x.CanPause)
			.DisposeWith(_disposables);

		_canStepHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.CanStep())
			.ToProperty(this, x => x.CanStep)
			.DisposeWith(_disposables);

		_isEditorEditableHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.IsEditorEditable())
			.ToProperty(this, x => x.IsEditorEditable)
			.DisposeWith(_disposables);

		// Set up timed highlight removal pipeline
		SetupHighlightPipeline();

		// Subscribe to WebSocket events
		_ = _ws.Events
			.ObserveOn(RxApp.MainThreadScheduler)
			.Subscribe(HandleEvent)
			.DisposeWith(_disposables);
	}

	// Reactive properties (manual implementation)
	private RegisterState _registers = RegisterState.Create();
	public RegisterState Registers
	{
		get => _registers;
		set => this.RaiseAndSetIfChanged(ref _registers, value);
	}

	private RegisterState? _previousRegisters;
	public RegisterState? PreviousRegisters
	{
		get => _previousRegisters;
		set => this.RaiseAndSetIfChanged(ref _previousRegisters, value);
	}

	private ImmutableHashSet<string> _changedRegisters = [];
	public ImmutableHashSet<string> ChangedRegisters
	{
		get => _changedRegisters;
		set => this.RaiseAndSetIfChanged(ref _changedRegisters, value);
	}

	private VMState _status = VMState.Idle;
	public VMState Status
	{
		get => _status;
		set => this.RaiseAndSetIfChanged(ref _status, value);
	}

	private string _consoleOutput = "";
	public string ConsoleOutput
	{
		get => _consoleOutput;
		set => this.RaiseAndSetIfChanged(ref _consoleOutput, value);
	}

	private string? _errorMessage;
	public string? ErrorMessage
	{
		get => _errorMessage;
		set => this.RaiseAndSetIfChanged(ref _errorMessage, value);
	}

	// Debugging state
	private ImmutableHashSet<uint> _breakpoints = [];
	public ImmutableHashSet<uint> Breakpoints
	{
		get => _breakpoints;
		set => this.RaiseAndSetIfChanged(ref _breakpoints, value);
	}

	private ImmutableArray<Watchpoint> _watchpoints = [];
	public ImmutableArray<Watchpoint> Watchpoints
	{
		get => _watchpoints;
		set => this.RaiseAndSetIfChanged(ref _watchpoints, value);
	}

	// Source mapping
	private string _sourceCode = "";
	public string SourceCode
	{
		get => _sourceCode;
		set => this.RaiseAndSetIfChanged(ref _sourceCode, value);
	}

	private ImmutableDictionary<uint, int> _addressToLine = ImmutableDictionary<uint, int>.Empty;
	public ImmutableDictionary<uint, int> AddressToLine
	{
		get => _addressToLine;
		set => this.RaiseAndSetIfChanged(ref _addressToLine, value);
	}

	private ImmutableDictionary<int, uint> _lineToAddress = ImmutableDictionary<int, uint>.Empty;
	public ImmutableDictionary<int, uint> LineToAddress
	{
		get => _lineToAddress;
		set => this.RaiseAndSetIfChanged(ref _lineToAddress, value);
	}

	private ImmutableHashSet<int> _validBreakpointLines = [];
	public ImmutableHashSet<int> ValidBreakpointLines
	{
		get => _validBreakpointLines;
		set => this.RaiseAndSetIfChanged(ref _validBreakpointLines, value);
	}

	// Memory state
	private ImmutableArray<byte> _memoryData = [];
	public ImmutableArray<byte> MemoryData
	{
		get => _memoryData;
		set => this.RaiseAndSetIfChanged(ref _memoryData, value);
	}

	private uint _memoryAddress;
	public uint MemoryAddress
	{
		get => _memoryAddress;
		set => this.RaiseAndSetIfChanged(ref _memoryAddress, value);
	}

	private MemoryWrite? _lastMemoryWrite;
	public MemoryWrite? LastMemoryWrite
	{
		get => _lastMemoryWrite;
		set => this.RaiseAndSetIfChanged(ref _lastMemoryWrite, value);
	}

	// Disassembly
	private ImmutableArray<DisassemblyInstruction> _disassembly = [];
	public ImmutableArray<DisassemblyInstruction> Disassembly
	{
		get => _disassembly;
		set => this.RaiseAndSetIfChanged(ref _disassembly, value);
	}

	// Connection state
	private bool _isConnected;
	public bool IsConnected
	{
		get => _isConnected;
		set => this.RaiseAndSetIfChanged(ref _isConnected, value);
	}

	private string? _sessionId;
	public string? SessionId
	{
		get => _sessionId;
		private set => this.RaiseAndSetIfChanged(ref _sessionId, value);
	}

	// Computed properties via ObservableAsPropertyHelper
	// These are disposed via _disposables.Dispose()
#pragma warning disable CA2213 // Disposable fields should be disposed - disposed via DisposeWith(_disposables)
	private readonly ObservableAsPropertyHelper<bool> _canPauseHelper;
	private readonly ObservableAsPropertyHelper<bool> _canStepHelper;
	private readonly ObservableAsPropertyHelper<bool> _isEditorEditableHelper;
#pragma warning restore CA2213

	public bool CanPause => _canPauseHelper.Value;
	public bool CanStep => _canStepHelper.Value;
	public bool IsEditorEditable => _isEditorEditableHelper.Value;

	// Commands
	public ReactiveCommand<Unit, Unit> RunCommand { get; }
	public ReactiveCommand<Unit, Unit> PauseCommand { get; }
	public ReactiveCommand<Unit, Unit> StepCommand { get; }
	public ReactiveCommand<Unit, Unit> StepOverCommand { get; }
	public ReactiveCommand<Unit, Unit> StepOutCommand { get; }
	public ReactiveCommand<Unit, Unit> ResetCommand { get; }
	public ReactiveCommand<Unit, Unit> LoadProgramCommand { get; }

	/// <summary>
	/// Helper to create commands with consistent error handling and scheduling.
	/// </summary>
#pragma warning disable CA2000 // Commands are disposed via DisposeWith(_disposables)
	private ReactiveCommand<Unit, Unit> CreateCommand(
		Func<CancellationToken, Task> execute,
		IObservable<bool>? canExecute = null
	) => ReactiveCommand.CreateFromTask(
		execute,
		canExecute,
		outputScheduler: RxApp.MainThreadScheduler
	).DisposeWith(_disposables);
#pragma warning restore CA2000

	/// <summary>
	/// Creates a new emulator session and connects the WebSocket.
	/// If a session already exists, destroys it first.
	/// </summary>
	public async Task CreateSessionAsync(CancellationToken ct = default)
	{
		// Destroy existing session if present
		if (SessionId is not null) {
			await DestroySessionAsync(ct);
		}

		// Create new session
		var sessionInfo = await _api.CreateSessionAsync(ct);
		SessionId = sessionInfo.SessionId;

		// Connect WebSocket
		await _ws.ConnectAsync(sessionInfo.SessionId, ct);
		IsConnected = true;
	}

	/// <summary>
	/// Destroys the current session and disconnects the WebSocket.
	/// </summary>
	public async Task DestroySessionAsync(CancellationToken ct = default)
	{
		if (SessionId is null) {
			return;
		}

		try {
			// Disconnect WebSocket first
			await _ws.DisconnectAsync();

			// Destroy session on backend
			await _api.DestroySessionAsync(SessionId, ct);
		}
		finally {
			// Always clear local state
			SessionId = null;
			IsConnected = false;
		}
	}

	// Command implementations (stubs for now)
	private Task RunAsync(CancellationToken ct) => Task.CompletedTask;
	private Task PauseAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepOverAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepOutAsync(CancellationToken ct) => Task.CompletedTask;
	private Task ResetAsync(CancellationToken ct) => Task.CompletedTask;
	private Task LoadProgramAsync(CancellationToken ct) => Task.CompletedTask;

	/// <summary>
	/// Sets up the Rx.NET pipeline for timed register highlight removal.
	/// Each register gets its own debounced removal stream using GroupBy and Throttle.
	/// </summary>
	private void SetupHighlightPipeline()
	{
		_ = _registerHighlightTrigger
			.GroupBy(register => register)
			.SelectMany(group =>
				group.Select(register => (register, action: "add"))
					.Merge(group
						.Throttle(HighlightDuration)
						.Select(register => (register, action: "remove"))
					)
			)
			.ObserveOn(RxApp.MainThreadScheduler)
			.Subscribe(x => {
				ChangedRegisters = x.action == "add"
					? ChangedRegisters.Add(x.register)
					: ChangedRegisters.Remove(x.register);
			})
			.DisposeWith(_disposables);
	}

	/// <summary>
	/// Updates the register state and tracks which registers changed for highlighting.
	/// Changes trigger the timed highlight removal pipeline.
	/// </summary>
	public void UpdateRegisters(RegisterState newRegisters)
	{
		if (PreviousRegisters is not null) {
			// Compute diff and trigger highlights for each changed register
			var changes = newRegisters.Diff(Registers);
			foreach (var register in changes) {
				_registerHighlightTrigger.OnNext(register);
			}
		}

		PreviousRegisters = Registers;
		Registers = newRegisters;
	}

	/// <summary>
	/// Handles WebSocket events with exhaustive pattern matching.
	/// Guards against stale state updates when already halted.
	/// </summary>
	private void HandleEvent(EmulatorEvent evt)
	{
		// Guard against stale events when already halted
		if (Status == VMState.Halted && evt is StateEvent) {
			return;
		}

		_ = evt switch {
			StateEvent { Status: var status, Registers: var regs } =>
				ApplyStateUpdate(status, regs),

			OutputEvent { Content: var content } =>
				AppendOutput(content),

			ExecutionEvent { EventType: var type, Message: var msg } =>
				ApplyExecutionEvent(type, msg),

			_ => false // Unreachable with sealed record hierarchy
		};
	}

	private bool ApplyStateUpdate(VMStatus status, RegisterState registers)
	{
		UpdateRegisters(registers);
		Status = status.State;
		LastMemoryWrite = status.LastWrite;
		return true;
	}

	private bool AppendOutput(string content)
	{
		ConsoleOutput += content;
		return true;
	}

	private bool ApplyExecutionEvent(ExecutionEventType type, string? message) =>
		type switch {
			ExecutionEventType.BreakpointHit => SetStatus(VMState.Breakpoint),
			ExecutionEventType.Halted => SetStatus(VMState.Halted),
			ExecutionEventType.Error => SetStatusWithError(VMState.Error, message),
			_ => throw new ArgumentOutOfRangeException(nameof(type), type, "Unknown execution event type")
		};

	private bool SetStatus(VMState state)
	{
		Status = state;
		return true;
	}

	private bool SetStatusWithError(VMState state, string? msg)
	{
		Status = state;
		ErrorMessage = msg;
		return true;
	}

	public void Dispose()
	{
		_registerHighlightTrigger.Dispose();
		_disposables.Dispose();
		GC.SuppressFinalize(this);
	}
}
