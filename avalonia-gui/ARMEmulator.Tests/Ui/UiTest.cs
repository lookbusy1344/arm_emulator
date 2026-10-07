using ARMEmulator.Models;
using Avalonia.Headless;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Runs a test body on the headless UI thread. One session serves the whole test run, because Avalonia
/// allows a single application per process.
/// </summary>
internal static class UiTest
{
	private static readonly Lazy<HeadlessUnitTestSession> Session = new(() => HeadlessUnitTestSession.StartNew(typeof(TestAppBuilder)));

	/// <summary>
	/// Starts the session. Run once per test assembly, before any test touches Avalonia: the first thread to use the
	/// dispatcher owns it, and the session must be that thread.
	/// </summary>
	internal static void EnsureStarted() => _ = Session.Value;

	/// <summary>Runs <paramref name="body"/> on the UI thread. Avalonia objects such as brushes may only be read there.</summary>
	public static Task RunOnUiThread(Action body) =>
		Session.Value.Dispatch(
			() => {
				body();
				return true;
			},
			TestContext.Current.CancellationToken);

	/// <summary>Shows a main window over mocked services, runs <paramref name="body"/> against it, then closes it.</summary>
	public static Task RunAsync(Func<MainWindowHarness, Task> body) =>
		Session.Value.Dispatch(
			async () => {
				using var harness = new MainWindowHarness();
				await body(harness);
				return true;
			},
			TestContext.Current.CancellationToken);

	/// <summary>As <see cref="RunAsync(Func{MainWindowHarness, Task})"/>, with the settings applied before the window is built.</summary>
	public static Task RunAsync(AppSettings settings, Func<MainWindowHarness, Task> body) =>
		Session.Value.Dispatch(
			async () => {
				using var harness = new MainWindowHarness(settings);
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
