using ARMEmulator.Services;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Services;

public sealed class ShortcutTextTests
{
	[Fact]
	public void Format_OnMacOS_UsesCommandSymbol() =>
		ShortcutText.Format("O", shift: false, isMacOS: true).Should().Be("⌘O");

	[Fact]
	public void Format_OnMacOSWithShift_PutsShiftBeforeCommand() =>
		ShortcutText.Format("E", shift: true, isMacOS: true).Should().Be("⇧⌘E");

	[Fact]
	public void Format_ElsewhereWithoutShift_UsesCtrl() =>
		ShortcutText.Format("O", shift: false, isMacOS: false).Should().Be("Ctrl+O");

	[Fact]
	public void Format_ElsewhereWithShift_UsesCtrlShift() =>
		ShortcutText.Format("E", shift: true, isMacOS: false).Should().Be("Ctrl+Shift+E");

	[Fact]
	public void EmptyEditorHint_NamesTheOpenAndExampleShortcuts()
	{
		ShortcutText.EmptyEditorHint(isMacOS: true).Should().Be("Open a file (⌘O) or choose an example (⇧⌘E)");
		ShortcutText.EmptyEditorHint(isMacOS: false).Should().Be("Open a file (Ctrl+O) or choose an example (Ctrl+Shift+E)");
	}
}
