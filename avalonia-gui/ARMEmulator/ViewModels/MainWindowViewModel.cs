using System.Collections.Immutable;
using System.Reactive;
using System.Reactive.Disposables;
using System.Reactive.Linq;
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

	public string? SessionId { get; private set; }

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

	// Command implementations (stubs for now)
	private Task RunAsync(CancellationToken ct) => Task.CompletedTask;
	private Task PauseAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepOverAsync(CancellationToken ct) => Task.CompletedTask;
	private Task StepOutAsync(CancellationToken ct) => Task.CompletedTask;
	private Task ResetAsync(CancellationToken ct) => Task.CompletedTask;
	private Task LoadProgramAsync(CancellationToken ct) => Task.CompletedTask;

	/// <summary>
	/// Updates the register state and tracks which registers changed for highlighting.
	/// </summary>
	public void UpdateRegisters(RegisterState newRegisters)
	{
		if (PreviousRegisters is not null) {
			// Compute diff between new and CURRENT registers (not previous!)
			var changes = newRegisters.Diff(Registers);
			ChangedRegisters = changes;
		}

		PreviousRegisters = Registers;
		Registers = newRegisters;
	}

	public void Dispose()
	{
		_disposables.Dispose();
		GC.SuppressFinalize(this);
	}
}
