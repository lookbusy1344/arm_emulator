using System.Collections.Immutable;
using System.Reactive;
using System.Reactive.Disposables;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection which triggers IL2026 warnings
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

public sealed class DisassemblyViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient api;
	private readonly CompositeDisposable disposables = [];

	private const int WindowSize = 64; // ±32 instructions around PC
	private const int InstructionsBeforePC = 32;

	public DisassemblyViewModel(IApiClient apiClient)
	{
		api = apiClient;

		RefreshCommand = ReactiveCommand.CreateFromTask(RefreshDisassemblyAsync);

		// Compute formatted instructions whenever instructions, PC, or breakpoints change
		formattedInstructionsHelper = this.WhenAnyValue(
				x => x.Instructions,
				x => x.ProgramCounter,
				x => x.Breakpoints)
			.Select(_ => FormatInstructions())
			.ToProperty(this, x => x.FormattedInstructions)
			.DisposeWith(disposables);
	}

	private uint programCounter;
	public uint ProgramCounter
	{
		get => programCounter;
		set => this.RaiseAndSetIfChanged(ref programCounter, value);
	}

	private ImmutableArray<DisassemblyInstruction> instructions = [];
	public ImmutableArray<DisassemblyInstruction> Instructions
	{
		get => instructions;
		set => this.RaiseAndSetIfChanged(ref instructions, value);
	}

	private ImmutableHashSet<uint> breakpoints = [];
	public ImmutableHashSet<uint> Breakpoints
	{
		get => breakpoints;
		set => this.RaiseAndSetIfChanged(ref breakpoints, value);
	}

	private string? sessionId;
	public string? SessionId
	{
		get => sessionId;
		set => this.RaiseAndSetIfChanged(ref sessionId, value);
	}

	private string? errorMessage;
	public string? ErrorMessage
	{
		get => errorMessage;
		set => this.RaiseAndSetIfChanged(ref errorMessage, value);
	}

	// Justification: formattedInstructionsHelper is disposed via DisposeWith(disposables) in constructor
#pragma warning disable CA2213
	private readonly ObservableAsPropertyHelper<ImmutableList<FormattedInstruction>> formattedInstructionsHelper;
#pragma warning restore CA2213
	public ImmutableList<FormattedInstruction> FormattedInstructions => formattedInstructionsHelper.Value;

	public ReactiveCommand<Unit, Unit> RefreshCommand { get; }

	public void UpdateRegisters(RegisterState registers)
	{
		ProgramCounter = registers.PC;
	}

	public void UpdateBreakpoints(ImmutableHashSet<uint> newBreakpoints)
	{
		Breakpoints = newBreakpoints;
	}

	public async Task RefreshDisassemblyAsync()
	{
		if (SessionId is null) {
			return;
		}

		try {
			ErrorMessage = null;

			// Calculate window centered around PC
			// Each instruction is 4 bytes, so 32 instructions before PC = 128 bytes
			var startAddress = ProgramCounter >= InstructionsBeforePC * 4
				? ProgramCounter - (InstructionsBeforePC * 4)
				: 0;

			var disasm = await api.GetDisassemblyAsync(SessionId, startAddress, WindowSize, CancellationToken.None);
			Instructions = disasm;
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to load disassembly: {ex.Message}";
		}
	}

	private ImmutableList<FormattedInstruction> FormatInstructions()
	{
		if (Instructions.IsEmpty) {
			return [];
		}

		return Instructions
			.Select(instr => new FormattedInstruction(
				Address: instr.Address.ToString("X8", System.Globalization.CultureInfo.InvariantCulture),
				MachineCode: instr.MachineCode.ToString("X8", System.Globalization.CultureInfo.InvariantCulture),
				Mnemonic: instr.Mnemonic,
				Symbol: instr.Symbol,
				IsCurrentPC: instr.Address == ProgramCounter,
				HasBreakpoint: Breakpoints.Contains(instr.Address)
			))
			.ToImmutableList();
	}

	public void Dispose() => disposables.Dispose();
}

public sealed record FormattedInstruction(
	string Address,
	string MachineCode,
	string Mnemonic,
	string? Symbol,
	bool IsCurrentPC,
	bool HasBreakpoint
);
