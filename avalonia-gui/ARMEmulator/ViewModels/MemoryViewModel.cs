using System.Collections.Immutable;
using System.Reactive;
using System.Reactive.Disposables;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI;

// ReactiveUI uses reflection which triggers IL2026 warnings
#pragma warning disable IL2026

namespace ARMEmulator.ViewModels;

public sealed class MemoryViewModel : ReactiveObject, IDisposable
{
	private readonly IApiClient api;
	private readonly CompositeDisposable disposables = [];

	public MemoryViewModel(IApiClient apiClient)
	{
		api = apiClient;

		NavigateToAddressCommand = ReactiveCommand.CreateFromTask(NavigateToAddressAsync);
		JumpToPCCommand = ReactiveCommand.CreateFromTask(JumpToPCAsync);
		JumpToSPCommand = ReactiveCommand.CreateFromTask(JumpToSPAsync);
		JumpToRegisterCommand = ReactiveCommand.CreateFromTask<int>(JumpToRegisterAsync);

		// Compute formatted rows whenever memory data or address changes
		formattedRowsHelper = this.WhenAnyValue(x => x.MemoryData, x => x.CurrentAddress, x => x.LastWriteAddress)
			.Select(_ => FormatMemoryRows())
			.ToProperty(this, x => x.FormattedRows)
			.DisposeWith(disposables);
	}

	private uint currentAddress;
	public uint CurrentAddress
	{
		get => currentAddress;
		set => this.RaiseAndSetIfChanged(ref currentAddress, value);
	}

	private bool autoScrollToWrites = true;
	public bool AutoScrollToWrites
	{
		get => autoScrollToWrites;
		set => this.RaiseAndSetIfChanged(ref autoScrollToWrites, value);
	}

	private ImmutableArray<byte> memoryData = [];
	public ImmutableArray<byte> MemoryData
	{
		get => memoryData;
		set => this.RaiseAndSetIfChanged(ref memoryData, value);
	}

	private uint? lastWriteAddress;
	public uint? LastWriteAddress
	{
		get => lastWriteAddress;
		set => this.RaiseAndSetIfChanged(ref lastWriteAddress, value);
	}

	private string addressInput = "";
	public string AddressInput
	{
		get => addressInput;
		set => this.RaiseAndSetIfChanged(ref addressInput, value);
	}

	private string? errorMessage;
	public string? ErrorMessage
	{
		get => errorMessage;
		set => this.RaiseAndSetIfChanged(ref errorMessage, value);
	}

	private string? sessionId;
	public string? SessionId
	{
		get => sessionId;
		set => this.RaiseAndSetIfChanged(ref sessionId, value);
	}

	private RegisterState? currentRegisters;

	// Justification: formattedRowsHelper is disposed via DisposeWith(disposables) in constructor
#pragma warning disable CA2213
	private readonly ObservableAsPropertyHelper<ImmutableList<MemoryRow>> formattedRowsHelper;
#pragma warning restore CA2213
	public ImmutableList<MemoryRow> FormattedRows => formattedRowsHelper.Value;

	public ReactiveCommand<Unit, Unit> NavigateToAddressCommand { get; }
	public ReactiveCommand<Unit, Unit> JumpToPCCommand { get; }
	public ReactiveCommand<Unit, Unit> JumpToSPCommand { get; }
	public ReactiveCommand<int, Unit> JumpToRegisterCommand { get; }

	public void UpdateRegisters(RegisterState registers)
	{
		currentRegisters = registers;
	}

	public void UpdateMemoryWrite(MemoryWrite? write)
	{
		if (write is null) {
			return;
		}

		LastWriteAddress = write.Address;

		if (AutoScrollToWrites && SessionId is not null) {
			// Navigate to write address asynchronously
			_ = LoadMemoryAsync(write.Address);
		}
	}

	public async Task LoadMemoryAsync(uint address)
	{
		if (SessionId is null) {
			return;
		}

		try {
			ErrorMessage = null;
			var data = await api.GetMemoryAsync(SessionId, address, 256, CancellationToken.None);
			MemoryData = data;
			CurrentAddress = address;
		}
		catch (Exception ex) {
			ErrorMessage = $"Failed to load memory: {ex.Message}";
		}
	}

	private async Task NavigateToAddressAsync()
	{
		try {
			ErrorMessage = null;
			var address = ParseAddress(AddressInput);
			await LoadMemoryAsync(address);
		}
		catch (FormatException) {
			ErrorMessage = $"Invalid address format: {AddressInput}";
		}
	}

	private async Task JumpToPCAsync()
	{
		if (currentRegisters is null) {
			return;
		}

		await LoadMemoryAsync(currentRegisters.PC);
	}

	private async Task JumpToSPAsync()
	{
		if (currentRegisters is null) {
			return;
		}

		await LoadMemoryAsync(currentRegisters.SP);
	}

	private async Task JumpToRegisterAsync(int registerIndex)
	{
		if (currentRegisters is null) {
			return;
		}

		var address = currentRegisters.Registers[registerIndex];
		await LoadMemoryAsync(address);
	}

	private static uint ParseAddress(string input)
	{
		var trimmed = input.Trim();
		if (trimmed.StartsWith("0x", StringComparison.OrdinalIgnoreCase)) {
			return Convert.ToUInt32(trimmed[2..], 16);
		}

		return uint.Parse(trimmed, System.Globalization.CultureInfo.InvariantCulture);
	}

	private ImmutableList<MemoryRow> FormatMemoryRows()
	{
		if (MemoryData.IsEmpty) {
			return [];
		}

		const int bytesPerRow = 16;
		var rowCount = (MemoryData.Length + bytesPerRow - 1) / bytesPerRow;

		return Enumerable.Range(0, rowCount)
			.Select(i => {
				var offset = i * bytesPerRow;
				var rowAddress = CurrentAddress + (uint)offset;
				var rowLength = Math.Min(bytesPerRow, MemoryData.Length - offset);
				var rowBytes = MemoryData.AsSpan(offset, rowLength);

				var hexBytes = string.Join(" ", rowBytes.ToArray().Select(b => b.ToString("X2", System.Globalization.CultureInfo.InvariantCulture)));
				var asciiText = new string(rowBytes.ToArray().Select(b => b >= 0x20 && b < 0x7F ? (char)b : '.').ToArray());

				return new MemoryRow(
					Address: rowAddress.ToString("X8", System.Globalization.CultureInfo.InvariantCulture),
					HexBytes: hexBytes,
					AsciiText: asciiText,
					IsHighlighted: LastWriteAddress.HasValue &&
								   rowAddress <= LastWriteAddress.Value &&
								   LastWriteAddress.Value < rowAddress + rowLength
				);
			})
			.ToImmutableList();
	}

	public void Dispose() => disposables.Dispose();
}

public sealed record MemoryRow(
	string Address,
	string HexBytes,
	string AsciiText,
	bool IsHighlighted
);
