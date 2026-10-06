using System.Reactive;
using System.Reactive.Disposables;
using System.Reactive.Disposables.Fluent;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.Views;
using Avalonia.Controls;
using ReactiveUI.Reactive;

// ReactiveUI uses reflection for WhenAnyValue and RaiseAndSetIfChanged, which triggers IL2026 warnings
// This is acceptable since we don't use AOT compilation for this project
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

/// <summary>
/// Central ViewModel for the main window, managing emulator state and user interactions.
/// Uses ReactiveUI for MVVM with reactive state management.
/// </summary>
public class MainWindowViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient api;
	private readonly IWebSocketClient ws;
	private readonly IFileService fileService;
	private readonly CompositeDisposable disposables = [];
	private readonly Subject<string> registerHighlightTrigger = new();
	private static readonly TimeSpan HighlightDuration = TimeSpan.FromSeconds(1.5);

	// Window reference for showing dialogs (set by MainWindow)
	private Window? parentWindow;

	/// <summary>
	/// Initializes a new instance of the MainWindowViewModel with required services.
	/// </summary>
	public MainWindowViewModel(IApiClient api, IWebSocketClient ws, IFileService fileService)
	{
		this.api = api;
		this.ws = ws;
		this.fileService = fileService;

		// Initialize commands with can-execute observables
#pragma warning disable CA2000 // Commands are disposed via DisposeWith(disposables) in ReportFailures
		RunCommand = CreateCommand("Run", RunAsync, this.WhenAnyValue(x => x.Status).Select(s => !s.CanPause()));
		PauseCommand = CreateCommand("Pause", PauseAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanPause()));
		StepCommand = CreateCommand("Step", StepAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		StepOverCommand = CreateCommand("Step over", StepOverAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		StepOutCommand = CreateCommand("Step out", StepOutAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
		ResetCommand = CreateCommand("Reset", ResetAsync);
		LoadProgramCommand = CreateCommand("Load", LoadProgramAsync);
		ShowPcCommand = CreateCommand("Show PC", ShowPcAsync);
		SendInputCommand = CreateCommand("Send input", SendInputAsync, this.WhenAnyValue(x => x.InputText).Select(s => !string.IsNullOrWhiteSpace(s)));
		ToggleBreakpointCommand = ReportFailures("Toggle breakpoint", ReactiveCommand.CreateFromTask<int>(ToggleBreakpointAtLineAsync, outputScheduler: RxSchedulers.MainThreadScheduler));
		DismissErrorCommand = ReactiveCommand.Create(() => { ErrorMessage = null; }).DisposeWith(disposables);

		// File operation commands
		OpenFileCommand = CreateCommand("Open file", OpenFileAsync);
		SaveFileCommand = CreateCommand("Save", SaveFileAsync);
		SaveAsCommand = CreateCommand("Save as", SaveAsAsync);
		OpenExampleCommand = CreateCommand("Open example", OpenExampleAsync);
		ShowPreferencesCommand = CreateCommand("Preferences", ShowPreferencesAsync);
		ShowAboutCommand = CreateCommand("About", ShowAboutAsync);
		OpenRecentFileCommand = ReportFailures("Open recent file", ReactiveCommand.CreateFromTask<string>(OpenRecentFileAsync, outputScheduler: RxSchedulers.MainThreadScheduler));
#pragma warning restore CA2000

		// Set up computed properties using WhenAnyValue
		canPauseHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.CanPause())
			.ToProperty(this, x => x.CanPause)
			.DisposeWith(disposables);

		canStepHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.CanStep())
			.ToProperty(this, x => x.CanStep)
			.DisposeWith(disposables);

		isEditorEditableHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s.IsEditorEditable())
			.ToProperty(this, x => x.IsEditorEditable)
			.DisposeWith(disposables);

		isWaitingForInputHelper = this.WhenAnyValue(x => x.Status)
			.Select(s => s == VMState.WaitingForInput)
			.ToProperty(this, x => x.IsWaitingForInput)
			.DisposeWith(disposables);

		// Set up status indicator properties
		statusColorHelper = this.WhenAnyValue(x => x.Status, x => x.IsConnected, GetStatusColor)
			.ToProperty(this, x => x.StatusColor)
			.DisposeWith(disposables);

		statusTextHelper = this.WhenAnyValue(x => x.Status, x => x.IsConnected, GetStatusText)
			.ToProperty(this, x => x.StatusText)
			.DisposeWith(disposables);

		// Set up timed highlight removal pipeline
		SetupHighlightPipeline();

		// Initialize child ViewModels
		ExpressionEvaluator = new ExpressionEvaluatorViewModel(api);
#pragma warning disable CA2000 // Child view models are disposed via DisposeWith(disposables)
		Memory = new MemoryViewModel(api).DisposeWith(disposables);
		Stack = new StackViewModel(api).DisposeWith(disposables);
		Disassembly = new DisassemblyViewModel(api).DisposeWith(disposables);
#pragma warning restore CA2000
		WireChildViewModels();

		// Subscribe to WebSocket events
		_ = this.ws.Events
			.ObserveOn(RxSchedulers.MainThreadScheduler)
			.Subscribe(HandleEvent)
			.DisposeWith(disposables);
	}

	// Reactive properties (manual implementation)
	private RegisterState registers = RegisterState.Create();

	public RegisterState Registers
	{
		get => registers;
		set => this.RaiseAndSetIfChanged(ref registers, value);
	}

	private RegisterState? previousRegisters;

	public RegisterState? PreviousRegisters
	{
		get => previousRegisters;
		set => this.RaiseAndSetIfChanged(ref previousRegisters, value);
	}

	private ImmutableHashSet<string> changedRegisters = [];

	public ImmutableHashSet<string> ChangedRegisters
	{
		get => changedRegisters;
		set => this.RaiseAndSetIfChanged(ref changedRegisters, value);
	}

	private VMState status = VMState.Idle;

	public VMState Status
	{
		get => status;
		set => this.RaiseAndSetIfChanged(ref status, value);
	}

	private string consoleOutput = "";

	public string ConsoleOutput
	{
		get => consoleOutput;
		set => this.RaiseAndSetIfChanged(ref consoleOutput, value);
	}

	private string inputText = "";

	public string InputText
	{
		get => inputText;
		set => this.RaiseAndSetIfChanged(ref inputText, value);
	}

	private string? errorMessage;

	public string? ErrorMessage
	{
		get => errorMessage;
		set => this.RaiseAndSetIfChanged(ref errorMessage, value);
	}

	// Debugging state
	private ImmutableHashSet<uint> breakpoints = [];

	public ImmutableHashSet<uint> Breakpoints
	{
		get => breakpoints;
		set => this.RaiseAndSetIfChanged(ref breakpoints, value);
	}

	private ImmutableArray<Watchpoint> watchpoints = [];

	public ImmutableArray<Watchpoint> Watchpoints
	{
		get => watchpoints;
		set => this.RaiseAndSetIfChanged(ref watchpoints, value);
	}

	// Source mapping
	private string sourceCode = "";

	public string SourceCode
	{
		get => sourceCode;
		set => this.RaiseAndSetIfChanged(ref sourceCode, value);
	}

	private ImmutableDictionary<uint, int> addressToLine = ImmutableDictionary<uint, int>.Empty;

	public ImmutableDictionary<uint, int> AddressToLine
	{
		get => addressToLine;
		set => this.RaiseAndSetIfChanged(ref addressToLine, value);
	}

	private ImmutableDictionary<int, uint> lineToAddress = ImmutableDictionary<int, uint>.Empty;

	public ImmutableDictionary<int, uint> LineToAddress
	{
		get => lineToAddress;
		set => this.RaiseAndSetIfChanged(ref lineToAddress, value);
	}

	private ImmutableHashSet<int> validBreakpointLines = [];

	public ImmutableHashSet<int> ValidBreakpointLines
	{
		get => validBreakpointLines;
		set => this.RaiseAndSetIfChanged(ref validBreakpointLines, value);
	}

	// Memory state
	private MemoryWrite? lastMemoryWrite;

	public MemoryWrite? LastMemoryWrite
	{
		get => lastMemoryWrite;
		set => this.RaiseAndSetIfChanged(ref lastMemoryWrite, value);
	}

	// Connection state
	private bool isConnected;

	public bool IsConnected
	{
		get => isConnected;
		set => this.RaiseAndSetIfChanged(ref isConnected, value);
	}

	private string? sessionId;

	public string? SessionId
	{
		get => sessionId;
		internal set => this.RaiseAndSetIfChanged(ref sessionId, value);
	}

	// Computed properties via ObservableAsPropertyHelper
	// These are disposed via _disposables.Dispose()
#pragma warning disable CA2213 // Disposable fields should be disposed - disposed via DisposeWith(_disposables)
	private readonly ObservableAsPropertyHelper<bool> canPauseHelper;
	private readonly ObservableAsPropertyHelper<bool> canStepHelper;
	private readonly ObservableAsPropertyHelper<bool> isEditorEditableHelper;
	private readonly ObservableAsPropertyHelper<bool> isWaitingForInputHelper;
	private readonly ObservableAsPropertyHelper<string> statusColorHelper;
	private readonly ObservableAsPropertyHelper<string> statusTextHelper;
#pragma warning restore CA2213

	public bool CanPause => canPauseHelper.Value;
	public bool CanStep => canStepHelper.Value;
	public bool IsEditorEditable => isEditorEditableHelper.Value;
	public bool IsWaitingForInput => isWaitingForInputHelper.Value;
	public string StatusColor => statusColorHelper.Value;
	public string StatusText => statusTextHelper.Value;

	// Child ViewModels
	public ExpressionEvaluatorViewModel ExpressionEvaluator { get; }
	public MemoryViewModel Memory { get; }
	public StackViewModel Stack { get; }
	public DisassemblyViewModel Disassembly { get; }

	// Commands
	public ReactiveCommand<Unit, Unit> RunCommand { get; }
	public ReactiveCommand<Unit, Unit> PauseCommand { get; }
	public ReactiveCommand<Unit, Unit> StepCommand { get; }
	public ReactiveCommand<Unit, Unit> StepOverCommand { get; }
	public ReactiveCommand<Unit, Unit> StepOutCommand { get; }
	public ReactiveCommand<Unit, Unit> ResetCommand { get; }
	public ReactiveCommand<Unit, Unit> LoadProgramCommand { get; }
	public ReactiveCommand<Unit, Unit> ShowPcCommand { get; }
	public ReactiveCommand<Unit, Unit> SendInputCommand { get; }

	// File operation commands
	public ReactiveCommand<Unit, Unit> OpenFileCommand { get; }
	public ReactiveCommand<Unit, Unit> SaveFileCommand { get; }
	public ReactiveCommand<Unit, Unit> SaveAsCommand { get; }
	public ReactiveCommand<Unit, Unit> OpenExampleCommand { get; }
	public ReactiveCommand<Unit, Unit> ShowPreferencesCommand { get; }
	public ReactiveCommand<Unit, Unit> ShowAboutCommand { get; }
	public ReactiveCommand<string, Unit> OpenRecentFileCommand { get; }

	/// <summary>Toggles a breakpoint at the instruction on the given 1-based source line.</summary>
	public ReactiveCommand<int, Unit> ToggleBreakpointCommand { get; }

	/// <summary>Clears <see cref="ErrorMessage"/>.</summary>
	public ReactiveCommand<Unit, Unit> DismissErrorCommand { get; }

	// Recent files from FileService
	public IReadOnlyList<RecentFile> RecentFiles => fileService.RecentFiles;

	/// <summary>
	/// Sets the parent window for showing dialogs.
	/// Called by MainWindow after construction.
	/// </summary>
	public void SetParentWindow(Window window) => parentWindow = window;

	/// <summary>
	/// Creates a command whose failures are reported as "<paramref name="operation"/> failed: …".
	/// </summary>
#pragma warning disable CA2000 // The command is disposed via DisposeWith(disposables) in ReportFailures
	private ReactiveCommand<Unit, Unit> CreateCommand(
		string operation,
		Func<CancellationToken, Task> execute,
		IObservable<bool>? canExecute = null
	) => ReportFailures(operation, ReactiveCommand.CreateFromTask(
		execute,
		canExecute,
		outputScheduler: RxSchedulers.MainThreadScheduler));
#pragma warning restore CA2000

	/// <summary>
	/// Routes a command's exceptions to <see cref="ErrorMessage"/>. Without a subscriber, ReactiveUI rethrows
	/// command exceptions on the main thread and the process ends. Cancellation is not a failure.
	/// </summary>
	private ReactiveCommand<TParam, Unit> ReportFailures<TParam>(string operation, ReactiveCommand<TParam, Unit> command)
	{
		_ = command.ThrownExceptions
			.Where(static ex => ex is not OperationCanceledException)
			.Subscribe(ex => ErrorMessage = $"{operation} failed: {ex.Message}")
			.DisposeWith(disposables);
		return command.DisposeWith(disposables);
	}

	/// <summary>
	/// Starts the backend, then creates a session and connects the WebSocket.
	/// Failures are reported through <see cref="ErrorMessage"/>.
	/// </summary>
	public async Task StartAsync(IBackendManager backend, CancellationToken ct = default)
	{
		ArgumentNullException.ThrowIfNull(backend);

		try {
			await backend.StartAsync(ct);
		}
		catch (BackendStartException ex) {
			ErrorMessage = $"Failed to start backend: {ex.Message}";
			return;
		}

		try {
			await CreateSessionAsync(ct);
			ErrorMessage = null;
		}
		catch (ApiException ex) {
			ErrorMessage = $"Failed to connect to backend: {ex.Message}";
		}
	}

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
		var sessionInfo = await api.CreateSessionAsync(ct);
		SessionId = sessionInfo.SessionId;

		// Connect WebSocket
		await ws.ConnectAsync(sessionInfo.SessionId, ct);
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
			await ws.DisconnectAsync();

			// Destroy session on backend
			await api.DestroySessionAsync(SessionId, ct);
		}
		finally {
			// Always clear local state
			SessionId = null;
			IsConnected = false;
		}
	}

	// Command implementations
	private async Task RunAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		await api.RunAsync(SessionId, ct);
	}

	private async Task PauseAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		await api.StopAsync(SessionId, ct);
	}

	private async Task StepAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		var newRegisters = await api.StepAsync(SessionId, ct);
		UpdateRegisters(newRegisters);
	}

	private async Task StepOverAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		var newRegisters = await api.StepOverAsync(SessionId, ct);
		UpdateRegisters(newRegisters);
	}

	private async Task StepOutAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		var newRegisters = await api.StepOutAsync(SessionId, ct);
		UpdateRegisters(newRegisters);
	}

	private async Task ResetAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			return;
		}

		try {
			await api.RestartAsync(SessionId, ct);
			var registers = await api.GetRegistersAsync(SessionId, ct);

			ConsoleOutput = "";
			ClearRegisterHighlights();
			Registers = registers;
			Status = VMState.Idle;
			ErrorMessage = null;
		}
		catch (ApiException ex) {
			ErrorMessage = $"Failed to restart: {ex.Message}";
		}
	}

	private void ClearRegisterHighlights()
	{
		PreviousRegisters = null;
		ChangedRegisters = [];
		LastMemoryWrite = null;
	}

	/// <summary>
	/// Assembles <see cref="SourceCode"/> into the session, then rebuilds the source line maps and register state.
	/// </summary>
	private async Task LoadProgramAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			ErrorMessage = "No active session";
			return;
		}

		ClearSourceMap();

		try {
			_ = await api.LoadProgramAsync(SessionId, SourceCode, ct);
			var sourceMap = await api.GetSourceMapAsync(SessionId, ct);
			var registers = await api.GetRegistersAsync(SessionId, ct);

			AddressToLine = sourceMap.ToImmutableDictionary(e => e.Address, e => e.LineNumber);
			LineToAddress = sourceMap.ToImmutableDictionary(e => e.LineNumber, e => e.Address);
			ValidBreakpointLines = [.. sourceMap.Select(e => e.LineNumber)];

			// A new program starts without highlights carried over from the previous one
			ClearRegisterHighlights();
			Registers = registers;

			// The backend reports "halted" for a loaded program that has not run; the GUI treats it as ready
			Status = VMState.Idle;
			ErrorMessage = null;
		}
		catch (ProgramLoadException ex) {
			ErrorMessage = $"Failed to load program:\n{string.Join('\n', ex.Errors)}";
		}
		catch (ApiException ex) {
			ClearSourceMap();
			ErrorMessage = $"Failed to load program: {ex.Message}";
		}
	}

	private void ClearSourceMap()
	{
		AddressToLine = ImmutableDictionary<uint, int>.Empty;
		LineToAddress = ImmutableDictionary<int, uint>.Empty;
		ValidBreakpointLines = [];
	}

	/// <summary>
	/// Keeps the memory, stack and disassembly views in step with the session, registers, breakpoints and memory writes.
	/// Stack and disassembly reload when the registers change while the VM is stopped.
	/// </summary>
	private void WireChildViewModels()
	{
		_ = this.WhenAnyValue(x => x.SessionId)
			.Subscribe(id => {
				ExpressionEvaluator.SessionId = id;
				Memory.SessionId = id;
				Stack.SessionId = id;
				Disassembly.SessionId = id;
			})
			.DisposeWith(disposables);

		_ = this.WhenAnyValue(x => x.Registers)
			.Subscribe(regs => {
				Memory.UpdateRegisters(regs);
				Stack.UpdateRegisters(regs);
				Disassembly.UpdateRegisters(regs);
			})
			.DisposeWith(disposables);

		_ = this.WhenAnyValue(x => x.Breakpoints)
			.Subscribe(Disassembly.UpdateBreakpoints)
			.DisposeWith(disposables);

		_ = this.WhenAnyValue(x => x.LastMemoryWrite)
			.Subscribe(Memory.UpdateMemoryWrite)
			.DisposeWith(disposables);

		_ = this.WhenAnyValue(x => x.Registers, x => x.Status, static (registers, status) => (registers, status))
			.Where(x => x.status != VMState.Running && SessionId is not null)
			.Select(static x => x.registers)
			.DistinctUntilChanged()
			.SelectMany(async _ => {
				await Stack.RefreshStackAsync();
				await Disassembly.RefreshDisassemblyAsync();
				return Unit.Default;
			})
			.Subscribe()
			.DisposeWith(disposables);
	}

	private Task ShowPcAsync(CancellationToken ct)
	{
		// TODO: Implement scroll-to-PC logic (will be handled by EditorView)
		return Task.CompletedTask;
	}

	/// <summary>
	/// Sends input to the emulator with smart logic:
	/// - If VM is waiting for input, the input unblocks a pending step() call.
	/// - If VM is not waiting, input is buffered and we must call step() to consume it.
	/// This matches the Swift EmulatorViewModel+Input.swift implementation.
	/// </summary>
	private async Task SendInputAsync(CancellationToken ct)
	{
		if (SessionId is null || string.IsNullOrWhiteSpace(InputText)) {
			return;
		}

		// Capture current status before sending input
		// If VM is waiting for input, the input will unblock an existing step() call
		// If VM is NOT waiting, the backend buffers input and we need to step() to consume it
		var wasWaitingForInput = Status == VMState.WaitingForInput;

		var inputData = InputText + "\n"; // Auto-append newline as per plan
		var sentInput = InputText; // Save for clearing after success

		try {
			// Send input to backend
			await api.SendStdinAsync(SessionId, inputData, ct);
			ErrorMessage = null;

			if (wasWaitingForInput) {
				// VM was waiting for input - the step() that triggered the input request
				// is still in progress and will complete now that we've provided input.
				// DO NOT call step() again or we'll execute an extra instruction!
				// Just refresh state to get the updated status
				var status = await api.GetStatusAsync(SessionId, ct);
				var registers = await api.GetRegistersAsync(SessionId, ct);
				UpdateRegisters(registers);
				Status = status.State;
				LastMemoryWrite = status.LastWrite;
			} else {
				// VM was not waiting - the backend buffered the input for later.
				// Call step() to consume the buffered input.
				var newRegisters = await api.StepAsync(SessionId, ct);
				UpdateRegisters(newRegisters);
			}

			// Clear input field after successful send
			InputText = "";
		}
		catch (SessionNotFoundException) {
			ErrorMessage = "Session not found. Please create a new session.";
		}
		catch (ApiException ex) {
			ErrorMessage = $"Failed to send input: {ex.Message}";
		}
	}

	private async Task ToggleBreakpointAtLineAsync(int line, CancellationToken ct)
	{
		if (!LineToAddress.TryGetValue(line, out var address)) {
			ErrorMessage = $"Line {line} has no instruction for a breakpoint";
			return;
		}

		if (Breakpoints.Contains(address)) {
			await RemoveBreakpointAsync(address, ct);
		} else {
			await AddBreakpointAsync(address, ct);
		}
	}

	/// <summary>
	/// Adds a breakpoint at the specified address.
	/// </summary>
	public async Task AddBreakpointAsync(uint address, CancellationToken ct = default)
	{
		if (SessionId is null) {
			return;
		}

		await api.AddBreakpointAsync(SessionId, address, ct);
		Breakpoints = Breakpoints.Add(address);
	}

	/// <summary>
	/// Removes a breakpoint at the specified address.
	/// </summary>
	public async Task RemoveBreakpointAsync(uint address, CancellationToken ct = default)
	{
		if (SessionId is null) {
			return;
		}

		await api.RemoveBreakpointAsync(SessionId, address, ct);
		Breakpoints = Breakpoints.Remove(address);
	}

	/// <summary>
	/// Adds a watchpoint at the specified address with the given type.
	/// </summary>
	public async Task AddWatchpointAsync(uint address, WatchpointType type, CancellationToken ct = default)
	{
		if (SessionId is null) {
			return;
		}

		var watchpoint = await api.AddWatchpointAsync(SessionId, address, type, ct);
		Watchpoints = [.. Watchpoints, watchpoint];
	}

	/// <summary>
	/// Removes a watchpoint by its ID.
	/// </summary>
	public async Task RemoveWatchpointAsync(int watchpointId, CancellationToken ct = default)
	{
		if (SessionId is null) {
			return;
		}

		await api.RemoveWatchpointAsync(SessionId, watchpointId, ct);
		Watchpoints = [.. Watchpoints.Where(w => w.Id != watchpointId)];
	}

	/// <summary>
	/// Gets the status indicator color based on VM state and connection status.
	/// </summary>
	private static string GetStatusColor(VMState status, bool isConnected)
	{
		if (!isConnected) {
			return "Gray";
		}

		return status switch {
			VMState.Idle => "Green",
			VMState.Running => "DodgerBlue",
			VMState.Breakpoint => "Orange",
			VMState.Halted => "Purple",
			VMState.Error => "Red",
			VMState.WaitingForInput => "Orange",
			_ => "Gray"
		};
	}

	/// <summary>
	/// Gets the status indicator tooltip text based on VM state and connection status.
	/// </summary>
	private static string GetStatusText(VMState status, bool isConnected)
	{
		if (!isConnected) {
			return "Disconnected";
		}

		return status switch {
			VMState.Idle => "Idle",
			VMState.Running => "Running",
			VMState.Breakpoint => "Breakpoint Hit",
			VMState.Halted => "Halted",
			VMState.Error => "Error",
			VMState.WaitingForInput => "Waiting for Input",
			_ => "Unknown"
		};
	}

	/// <summary>
	/// Sets up the Rx.NET pipeline for timed register highlight removal.
	/// Each register gets its own debounced removal stream using GroupBy and Throttle.
	/// </summary>
	private void SetupHighlightPipeline()
	{
		_ = registerHighlightTrigger
			.GroupBy(register => register)
			.SelectMany(group =>
				group.Select(register => (register, action: "add"))
					.Merge(group
						.Throttle(HighlightDuration)
						.Select(register => (register, action: "remove"))
					)
			)
			.ObserveOn(RxSchedulers.MainThreadScheduler)
			.Subscribe(x => {
				ChangedRegisters = x.action == "add"
					? ChangedRegisters.Add(x.register)
					: ChangedRegisters.Remove(x.register);
			})
			.DisposeWith(disposables);
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
				registerHighlightTrigger.OnNext(register);
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

	private bool ApplyStateUpdate(VMStatus status, RegisterState? registers)
	{
		if (registers is not null) {
			UpdateRegisters(registers);
		}

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

	// File operations
	private async Task OpenFileAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var result = await fileService.OpenFileAsync(parentWindow);
		if (result is null) {
			return;
		}

		var (path, content) = result.Value;
		SourceCode = content;
		// Auto-load program after opening
		await LoadProgramAsync(ct);
	}

	private async Task SaveFileAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var path = await fileService.SaveFileAsync(parentWindow, SourceCode, fileService.CurrentFilePath);
		if (path is not null) {
			fileService.CurrentFilePath = path;
		}
	}

	private async Task SaveAsAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var path = await fileService.SaveFileAsync(parentWindow, SourceCode, null);
		if (path is not null) {
			fileService.CurrentFilePath = path;
		}
	}

	private async Task OpenExampleAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var vm = new ExamplesBrowserViewModel(api);
		var window = new ExamplesBrowserWindow(vm);
		await window.ShowDialog(parentWindow);

		if (window.SelectedExampleContent is not null) {
			SourceCode = window.SelectedExampleContent;
			await LoadProgramAsync(ct);
		}

		vm.Dispose();
	}

	private async Task ShowPreferencesAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		// TODO: Load settings from persistent storage
		var window = new PreferencesWindow(AppSettings.Default);
		await window.ShowDialog(parentWindow);

		if (window.UpdatedSettings is not null) {
			// TODO: Save settings to persistent storage
			// For now, just apply font size to editor
		}
	}

	private async Task ShowAboutAsync(CancellationToken ct)
	{
		if (parentWindow is null) {
			return;
		}

		var vm = new AboutWindowViewModel(api);
		var window = new AboutWindow(vm);
		await window.ShowDialog(parentWindow);
	}

	private async Task OpenRecentFileAsync(string path, CancellationToken ct)
	{
		try {
			var content = await File.ReadAllTextAsync(path, ct);
			SourceCode = content;
			fileService.CurrentFilePath = path;
			await LoadProgramAsync(ct);
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to open {path}: {ex.Message}";
		}
	}

	public void Dispose()
	{
		registerHighlightTrigger.Dispose();
		disposables.Dispose();
		GC.SuppressFinalize(this);
	}
}
