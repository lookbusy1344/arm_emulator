using ARMEmulator.Models;
using ARMEmulator.Services;
using ReactiveUI.Reactive;

namespace ARMEmulator.ViewModels;

/// <summary>
/// Assembling the editor source into the session. Run, Step, Reset and breakpoint toggles assemble edited source first.
/// </summary>
public partial class MainWindowViewModel
{
	private string? assembledSource;

	// Source lines of the breakpoints. Kept across a failed assembly, which clears the line maps.
	private ImmutableArray<int> breakpointLines = [];

	/// <summary>The source in the session, or null before the first successful assembly and after a failed one.</summary>
	public string? AssembledSource
	{
		get => assembledSource;
		private set
		{
			_ = this.RaiseAndSetIfChanged(ref assembledSource, value);
			this.RaisePropertyChanged(nameof(NeedsAssembly));
		}
	}

	/// <summary>True when the editor holds source that is not in the session. An empty editor needs nothing.</summary>
	public bool NeedsAssembly => SourceDiffers(SourceCode, AssembledSource);

	private static bool SourceDiffers(string source, string? assembled) => source != (assembled ?? "");

	/// <summary>Step is available while paused, and while stopped with edits that stepping will assemble.</summary>
	private static bool CanStepWith(VMState status, string source, string? assembled) =>
		status.CanStep() || (status.IsEditorEditable() && SourceDiffers(source, assembled));

	private IObservable<bool> CanStepChanges() =>
		this.WhenAnyValue(x => x.Status, x => x.SourceCode, x => x.AssembledSource, CanStepWith);

	/// <summary>Assembles when the editor holds source that is not in the session. Returns false when assembly failed.</summary>
	private async Task<bool> EnsureAssembledAsync(CancellationToken ct) => !NeedsAssembly || await AssembleAsync(ct);

	/// <summary>
	/// Assembles <see cref="SourceCode"/> into the session, then rebuilds the source line maps and register state.
	/// Breakpoints stay on their source lines. Returns false when assembly failed.
	/// </summary>
	private async Task<bool> AssembleAsync(CancellationToken ct)
	{
		if (SessionId is null) {
			ErrorMessage = "No active session";
			return false;
		}

		var source = SourceCode;
		if (!AddressToLine.IsEmpty) {
			breakpointLines = [.. Breakpoints.Where(AddressToLine.ContainsKey).Select(address => AddressToLine[address])];
		}

		ClearSourceMap();

		try {
			_ = await api.LoadProgramAsync(SessionId, source, ct);
			var sourceMap = await api.GetSourceMapAsync(SessionId, ct);
			var registers = await api.GetRegistersAsync(SessionId, ct);

			AddressToLine = sourceMap.ToImmutableDictionary(e => e.Address, e => e.LineNumber);
			LineToAddress = sourceMap.ToImmutableDictionary(e => e.LineNumber, e => e.Address);
			ValidBreakpointLines = [.. sourceMap.Select(e => e.LineNumber)];
			AssembledSource = source;

			// A new program starts without output or highlights carried over from the previous one
			ConsoleOutput = "";
			ResetRegisterBaseline(registers);

			// The backend reports "halted" for a loaded program that has not run; the GUI treats it as ready
			Status = VMState.Idle;
			ErrorMessage = null;

			await MoveBreakpointsToLinesAsync(ct);
			return true;
		}
		catch (ProgramLoadException ex) {
			ErrorMessage = $"Failed to assemble program:\n{string.Join('\n', ex.Errors)}";
			return false;
		}
		catch (ApiException ex) {
			ClearSourceMap();
			ErrorMessage = $"Failed to assemble program: {ex.Message}";
			return false;
		}
	}

	/// <summary>
	/// The session keeps breakpoints by address across assembly. Replaces them with the new addresses of
	/// <see cref="breakpointLines"/>; lines without an instruction lose their breakpoint.
	/// </summary>
	private async Task MoveBreakpointsToLinesAsync(CancellationToken ct)
	{
		foreach (var address in Breakpoints) {
			await RemoveBreakpointAsync(address, ct);
		}

		foreach (var address in breakpointLines.Where(LineToAddress.ContainsKey).Select(line => LineToAddress[line]).Distinct()) {
			await AddBreakpointAsync(address, ct);
		}
	}

	private void ClearSourceMap()
	{
		AssembledSource = null;
		AddressToLine = ImmutableDictionary<uint, int>.Empty;
		LineToAddress = ImmutableDictionary<int, uint>.Empty;
		ValidBreakpointLines = [];
	}
}
