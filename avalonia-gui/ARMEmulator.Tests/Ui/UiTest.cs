using Avalonia.Headless;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Runs a test body on the headless UI thread. One session serves the whole test run, because Avalonia
/// allows a single application per process.
/// </summary>
internal static class UiTest
{
	private static readonly Lazy<HeadlessUnitTestSession> Session = new(() => HeadlessUnitTestSession.StartNew(typeof(TestAppBuilder)));

	/// <summary>Shows a main window over mocked services, runs <paramref name="body"/> against it, then closes it.</summary>
	public static Task RunAsync(Func<MainWindowHarness, Task> body) =>
		Session.Value.Dispatch(
			async () => {
				using var harness = new MainWindowHarness();
				await body(harness);
				return true;
			},
			TestContext.Current.CancellationToken);

	/// <summary>As <see cref="RunAsync(Func{MainWindowHarness, Task})"/> for a synchronous body.</summary>
	public static Task Run(Action<MainWindowHarness> body) =>
		RunAsync(harness => {
			body(harness);
			return Task.CompletedTask;
		});
}
