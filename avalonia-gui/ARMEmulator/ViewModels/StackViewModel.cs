using System.Collections.Immutable;
using System.Reactive;
using System.Reactive.Disposables;
using System.Reactive.Disposables.Fluent;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI.Reactive;

// ReactiveUI uses reflection which triggers IL2026 warnings
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

public sealed class StackViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient api;
	private readonly CompositeDisposable disposables = [];

	// Stack typically starts at 0x50000 in ARM emulator
	private const uint StackTop = 0x50000;
	private const uint MaxStackBytes = 0x10000;
	private const int GeneralRegisterCount = 13;

	// Address ranges for annotation detection
	private const uint CodeStart = 0x8000;
	private const uint CodeEnd = 0x10000;
	private const uint StackStart = 0x40000;
	private const uint StackEnd = 0x60000;

	public StackViewModel(IApiClient apiClient)
	{
		api = apiClient;

		RefreshCommand = ReactiveCommand.CreateFromTask(RefreshStackAsync);

		// Compute stack entries whenever memory data or registers change
		stackEntriesHelper = this.WhenAnyValue(
				x => x.StackMemory,
				x => x.StackPointer,
				x => x.LinkRegister,
				x => x.GeneralRegisters)
			.Select(_ => FormatStackEntries())
			.ToProperty(this, x => x.StackEntries)
			.DisposeWith(disposables);

		// Compute stack size whenever SP changes
		stackSizeHelper = this.WhenAnyValue(x => x.StackPointer)
			.Select(sp => CalculateStackSize(sp))
			.ToProperty(this, x => x.StackSize)
			.DisposeWith(disposables);
	}

	private uint stackPointer;
	public uint StackPointer
	{
		get => stackPointer;
		set => this.RaiseAndSetIfChanged(ref stackPointer, value);
	}

	private ImmutableArray<uint> generalRegisters = [];
	public ImmutableArray<uint> GeneralRegisters
	{
		get => generalRegisters;
		private set => this.RaiseAndSetIfChanged(ref generalRegisters, value);
	}

	private uint linkRegister;
	public uint LinkRegister
	{
		get => linkRegister;
		set => this.RaiseAndSetIfChanged(ref linkRegister, value);
	}

	private ImmutableArray<byte> stackMemory = [];
	public ImmutableArray<byte> StackMemory
	{
		get => stackMemory;
		set => this.RaiseAndSetIfChanged(ref stackMemory, value);
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

	// Justification: stackEntriesHelper is disposed via DisposeWith(disposables) in constructor
#pragma warning disable CA2213
	private readonly ObservableAsPropertyHelper<ImmutableList<StackEntry>> stackEntriesHelper;
#pragma warning restore CA2213
	public ImmutableList<StackEntry> StackEntries => stackEntriesHelper.Value;

	// Justification: stackSizeHelper is disposed via DisposeWith(disposables) in constructor
#pragma warning disable CA2213
	private readonly ObservableAsPropertyHelper<uint> stackSizeHelper;
#pragma warning restore CA2213
	public uint StackSize => stackSizeHelper.Value;

	public ReactiveCommand<Unit, Unit> RefreshCommand { get; }

	public void UpdateRegisters(RegisterState registers)
	{
		StackPointer = registers.SP;
		LinkRegister = registers.LR;
		GeneralRegisters = registers.Registers.Take(GeneralRegisterCount).ToImmutableArray();
	}

	public async Task RefreshStackAsync()
	{
		if (SessionId is null) {
			return;
		}

		try {
			ErrorMessage = null;

			var used = CalculateStackSize(StackPointer);
			if (used > MaxStackBytes) {
				StackMemory = [];
				ErrorMessage = $"Stack of {used} bytes exceeds the {MaxStackBytes} byte limit; SP may be corrupt";
				return;
			}

			if (used == 0) {
				StackMemory = [];
				return;
			}

			// Load the used stack from SP upward (the stack grows downward)
			var data = await api.GetMemoryAsync(SessionId, StackPointer, (int)used, CancellationToken.None);
			StackMemory = data;
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to load stack: {ex.Message}";
		}
	}

	private static uint CalculateStackSize(uint sp)
	{
		// Stack size is difference from stack top
		// Return 0 if SP is uninitialized (0) or above stack top
		return sp == 0 || sp > StackTop ? 0 : StackTop - sp;
	}

	private ImmutableList<StackEntry> FormatStackEntries()
	{
		if (StackMemory.IsEmpty) {
			return [];
		}

		const int bytesPerWord = 4;
		var wordCount = StackMemory.Length / bytesPerWord;

		return Enumerable.Range(0, wordCount)
			.Select(i => {
				var offset = i * bytesPerWord;
				var address = StackPointer + (uint)offset;
				var wordBytes = StackMemory.AsSpan(offset, bytesPerWord);

				// ARM is little-endian
				var value = BitConverter.ToUInt32(wordBytes);
				var hexValue = value.ToString("X8", System.Globalization.CultureInfo.InvariantCulture);
				var asciiText = new string(wordBytes.ToArray().Select(b => b >= 0x20 && b < 0x7F ? (char)b : '.').ToArray());

				// Calculate offset from SP
				var offsetStr = i == 0 ? "SP+0" : $"SP+{offset}";

				// Determine annotation
				var annotation = GetAnnotation(value);

				return new StackEntry(
					Address: address.ToString("X8", System.Globalization.CultureInfo.InvariantCulture),
					Value: hexValue,
					Ascii: asciiText,
					Offset: offsetStr,
					Annotation: annotation,
					IsCurrentSP: address == StackPointer
				);
			})
			.ToImmutableList();
	}

	private string? GetAnnotation(uint value)
	{
		// Check if value matches LR
		if (value != 0 && value == LinkRegister) {
			return "LR (return address)";
		}

		// Check if value is in code range
		if (value >= CodeStart && value < CodeEnd) {
			return "Return addr?";
		}

		// Check if value is in stack range
		if (value >= StackStart && value < StackEnd) {
			return "stack address";
		}

		var registerIndex = GeneralRegisters.IndexOf(value);
		if (value != 0 && registerIndex >= 0) {
			return $"R{registerIndex}";
		}

		return null;
	}

	public void Dispose() => disposables.Dispose();
}

public sealed record StackEntry(
	string Address,
	string Value,
	string Ascii,
	string Offset,
	string? Annotation,
	bool IsCurrentSP
);
