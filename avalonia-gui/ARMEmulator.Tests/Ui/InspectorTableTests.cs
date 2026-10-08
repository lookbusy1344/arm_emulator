using ARMEmulator.Models;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Threading;
using Avalonia.VisualTree;
using AwesomeAssertions;
using NSubstitute;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Row markers and column headers of the memory, stack and disassembly tables.
/// </summary>
public sealed class InspectorTableTests
{
	private const string SessionId = MainWindowHarness.SessionId;
	private const uint FirstAddress = 0x8000;
	private const uint SecondAddress = 0x8004;
	private const string CurrentClass = "current";
	private const string HighlightClass = "highlight";

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	private static async Task ShowPanelAsync(MainWindowHarness ui, InspectorPanel panel)
	{
		await ui.ViewModel.StartAsync(ui.Backend, Ct);
		ui.Find<ComboBox>("InspectorSelector").SelectedValue = panel;
		Settle(ui);
	}

	private static void Settle(MainWindowHarness ui)
	{
		Dispatcher.UIThread.RunJobs();
		ui.Window.UpdateLayout();
		Dispatcher.UIThread.RunJobs();
	}

	private static Border Row(MainWindowHarness ui, string rowsName, int index) =>
		ui.Find<ItemsControl>(rowsName).ContainerFromIndex(index)!
			.GetVisualDescendants().OfType<Border>().First(border => border.Classes.Contains("dataRow"));

	private static bool IsMarkerVisible(Border row, string marker) =>
		row.GetVisualDescendants().OfType<Control>().Single(control => control.Classes.Contains(marker)).IsVisible;

	private static IEnumerable<string?> HeaderTexts(MainWindowHarness ui, string tableName) =>
		ui.Find<Border>(tableName).GetVisualDescendants().OfType<TextBlock>().Select(text => text.Text);

	[Fact]
	public Task DisassemblyRow_AtThePc_IsCurrentAndShowsOnlyThePcMarker() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Disassembly);
			ui.Api.GetDisassemblyAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Ct)
				.ReturnsForAnyArgs(ImmutableArray.Create(
					new DisassemblyInstruction(FirstAddress, 0xE3A00005, "MOV R0, #5", null),
					new DisassemblyInstruction(SecondAddress, 0xE3A01006, "MOV R1, #6", null)));

			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: FirstAddress));
			Settle(ui);

			Row(ui, "DisassemblyRows", 0).Classes.Should().Contain(CurrentClass);
			Row(ui, "DisassemblyRows", 1).Classes.Should().NotContain(CurrentClass);
			IsMarkerVisible(Row(ui, "DisassemblyRows", 0), "pcMarker").Should().BeTrue();
			IsMarkerVisible(Row(ui, "DisassemblyRows", 1), "pcMarker").Should().BeFalse();
		});

	[Fact]
	public Task DisassemblyRow_WithABreakpoint_ShowsOnlyTheBreakpointMarker() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Disassembly);
			ui.Api.GetDisassemblyAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Ct)
				.ReturnsForAnyArgs(ImmutableArray.Create(
					new DisassemblyInstruction(FirstAddress, 0xE3A00005, "MOV R0, #5", null),
					new DisassemblyInstruction(SecondAddress, 0xE3A01006, "MOV R1, #6", null)));
			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: FirstAddress));

			ui.ViewModel.Breakpoints = [SecondAddress];
			Settle(ui);

			IsMarkerVisible(Row(ui, "DisassemblyRows", 0), "breakpointMarker").Should().BeFalse();
			IsMarkerVisible(Row(ui, "DisassemblyRows", 1), "breakpointMarker").Should().BeTrue();
		});

	[Fact]
	public Task DisassemblyPanel_ScrollsTheCurrentRowIntoView() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Disassembly);
			const int pcIndex = 48;
			const uint pc = FirstAddress + (pcIndex * 4);
			ui.Api.GetDisassemblyAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Ct)
				.ReturnsForAnyArgs(Enumerable.Range(0, 64)
					.Select(i => new DisassemblyInstruction(FirstAddress + (uint)(i * 4), 0xE3A00005, "MOV R0, #5", null))
					.ToImmutableArray());

			ui.ViewModel.UpdateRegisters(RegisterState.Create(pc: pc));
			Settle(ui);
			Settle(ui);

			var scroll = ui.Find<ScrollViewer>("DisassemblyScroll");
			var row = ui.Find<ItemsControl>("DisassemblyRows").ContainerFromIndex(pcIndex)!;
			scroll.Offset.Y.Should().BeGreaterThan(0);
			var rowTop = ((Visual)row).TranslatePoint(default, scroll)!.Value.Y;
			rowTop.Should().BeGreaterThanOrEqualTo(0);
			(rowTop + row.Bounds.Height).Should().BeLessThanOrEqualTo(scroll.Viewport.Height);
		});

	[Fact]
	public Task DisassemblyTable_HasColumnHeaders() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Disassembly);

			HeaderTexts(ui, "DisassemblyHeader").Should().Equal("Address", "Code", "Instruction");
		});

	[Fact]
	public Task StackRow_AtTheStackPointer_IsCurrent() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Stack);
			const uint stackPointer = 0x4FFF0;
			ui.Api.GetMemoryAsync(SessionId, stackPointer, Arg.Any<int>(), Ct)
				.ReturnsForAnyArgs(ImmutableArray.Create<byte>(1, 0, 0, 0, 2, 0, 0, 0));

			ui.ViewModel.UpdateRegisters(RegisterState.Create(sp: stackPointer));
			Settle(ui);

			Row(ui, "StackRows", 0).Classes.Should().Contain(CurrentClass);
			Row(ui, "StackRows", 1).Classes.Should().NotContain(CurrentClass);
		});

	[Fact]
	public Task StackTable_HasColumnHeaders() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Stack);

			HeaderTexts(ui, "StackHeader").Should().Equal("Offset", "Address", "Value", "ASCII", "Note");
		});

	[Fact]
	public Task MemoryRow_ContainingTheLastWrite_IsHighlighted() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Memory);
			ui.Api.GetMemoryAsync(SessionId, Arg.Any<uint>(), Arg.Any<int>(), Ct)
				.ReturnsForAnyArgs(Enumerable.Range(0, 32).Select(i => (byte)i).ToImmutableArray());
			await ui.ViewModel.Memory.LoadMemoryAsync(FirstAddress);

			ui.ViewModel.Memory.LastWriteAddress = FirstAddress + 3;
			Settle(ui);

			Row(ui, "MemoryRows", 0).Classes.Should().Contain(HighlightClass);
			Row(ui, "MemoryRows", 1).Classes.Should().NotContain(HighlightClass);
		});

	[Fact]
	public Task MemoryTable_HasColumnHeaders() =>
		UiTest.RunAsync(async ui => {
			await ShowPanelAsync(ui, InspectorPanel.Memory);

			HeaderTexts(ui, "MemoryHeader").Should().Equal("Address", "Bytes", "ASCII");
		});
}
