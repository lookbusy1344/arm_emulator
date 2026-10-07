using ARMEmulator.Services;
using Avalonia.Input;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Services;

public sealed class ShortcutsTests
{
	private static KeyGesture Gesture(Key key, KeyModifiers modifiers = KeyModifiers.None) => new(key, modifiers);

	private static IEnumerable<KeyGesture> GesturesOf(ShortcutId id, bool isMacOS) =>
		Shortcuts.For(isMacOS).Where(binding => binding.Id == id).Select(binding => binding.Gesture);

	[Fact]
	public void For_HasSixteenBindings_OnEachPlatform()
	{
		Shortcuts.For(isMacOS: false).Should().HaveCount(16);
		Shortcuts.For(isMacOS: true).Should().HaveCount(16);
	}

	[Fact]
	public void For_OffMacOS_UsesControl()
	{
		GesturesOf(ShortcutId.Run, isMacOS: false).Should().Equal(Gesture(Key.F5), Gesture(Key.R, KeyModifiers.Control));
		GesturesOf(ShortcutId.StepOut, isMacOS: false).Should().Equal(Gesture(Key.T, KeyModifiers.Control | KeyModifiers.Alt));
		GesturesOf(ShortcutId.Preferences, isMacOS: false).Should().Equal(Gesture(Key.OemComma, KeyModifiers.Control));
	}

	[Fact]
	public void For_OnMacOS_UsesCommand()
	{
		GesturesOf(ShortcutId.Run, isMacOS: true).Should().Equal(Gesture(Key.F5), Gesture(Key.R, KeyModifiers.Meta));
		GesturesOf(ShortcutId.StepOut, isMacOS: true).Should().Equal(Gesture(Key.T, KeyModifiers.Meta | KeyModifiers.Alt));
		GesturesOf(ShortcutId.Examples, isMacOS: true).Should().Equal(Gesture(Key.E, KeyModifiers.Meta | KeyModifiers.Shift));
	}

	[Fact]
	public void For_OnMacOS_UsesNoControlModifier() =>
		Shortcuts.For(isMacOS: true).Should().OnlyContain(binding => !binding.Gesture.KeyModifiers.HasFlag(KeyModifiers.Control));

	[Fact]
	public void For_OffMacOS_UsesNoCommandModifier() =>
		Shortcuts.For(isMacOS: false).Should().OnlyContain(binding => !binding.Gesture.KeyModifiers.HasFlag(KeyModifiers.Meta));

	[Fact]
	public void For_GivesNoGestureToTwoCommands() =>
		Shortcuts.For(isMacOS: true).GroupBy(binding => binding.Gesture).Should().OnlyContain(group => group.Count() == 1);

	[Theory]
	[InlineData(Key.R, KeyModifiers.Meta, true, "⌘R")]
	[InlineData(Key.E, KeyModifiers.Meta | KeyModifiers.Shift, true, "⇧⌘E")]
	[InlineData(Key.T, KeyModifiers.Meta | KeyModifiers.Alt, true, "⌥⌘T")]
	[InlineData(Key.OemComma, KeyModifiers.Meta, true, "⌘,")]
	[InlineData(Key.F5, KeyModifiers.None, true, "F5")]
	[InlineData(Key.T, KeyModifiers.Control | KeyModifiers.Shift, false, "Ctrl+Shift+T")]
	[InlineData(Key.T, KeyModifiers.Control | KeyModifiers.Alt, false, "Ctrl+Alt+T")]
	[InlineData(Key.OemPeriod, KeyModifiers.Control, false, "Ctrl+.")]
	[InlineData(Key.F10, KeyModifiers.None, false, "F10")]
	public void Display_WritesTheGestureTheWayThePlatformDoes(Key key, KeyModifiers modifiers, bool isMacOS, string expected) =>
		ShortcutText.Display(Gesture(key, modifiers), isMacOS).Should().Be(expected);

	[Fact]
	public void Hint_ListsEveryGestureOfTheCommand()
	{
		ShortcutText.Hint("Run", ShortcutId.Run, isMacOS: true).Should().Be("Run (F5, ⌘R)");
		ShortcutText.Hint("Step Out", ShortcutId.StepOut, isMacOS: false).Should().Be("Step Out (Ctrl+Alt+T)");
	}
}
