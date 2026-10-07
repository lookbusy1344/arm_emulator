using ARMEmulator.Controls;
using ARMEmulator.Tests.Ui;
using Avalonia.Controls;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Controls;

public sealed class DialogButtonOrderTests
{
	private static StackPanel RowOf(params string[] labels)
	{
		var row = new StackPanel();
		row.Children.AddRange(labels.Select(label => new Button { Content = label }));
		return row;
	}

	private static IEnumerable<object?> Labels(Panel row) => row.Children.OfType<Button>().Select(button => button.Content);

	[Fact]
	public Task Apply_OnWindows_PutsThePrimaryButtonFirst() =>
		UiTest.RunOnUiThread(() => {
			var row = RowOf("Cancel", "Discard", "Save");

			DialogButtonOrder.Apply(row, isWindows: true);

			Labels(row).Should().Equal("Save", "Discard", "Cancel");
		});

	[Fact]
	public Task Apply_OffWindows_KeepsThePrimaryButtonLast() =>
		UiTest.RunOnUiThread(() => {
			var row = RowOf("Cancel", "Discard", "Save");

			DialogButtonOrder.Apply(row, isWindows: false);

			Labels(row).Should().Equal("Cancel", "Discard", "Save");
		});

	[Fact]
	public Task Apply_WithTwoButtons_SwapsThemOnWindows() =>
		UiTest.RunOnUiThread(() => {
			var row = RowOf("Cancel", "OK");

			DialogButtonOrder.Apply(row, isWindows: true);

			Labels(row).Should().Equal("OK", "Cancel");
		});

	[Fact]
	public Task Apply_WithOneButton_ChangesNothing() =>
		UiTest.RunOnUiThread(() => {
			var row = RowOf("Close");

			DialogButtonOrder.Apply(row, isWindows: true);

			Labels(row).Should().Equal("Close");
		});
}
