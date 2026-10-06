using ARMEmulator;
using ARMEmulator.Tests.Ui;
using Avalonia;
using Avalonia.Headless;
using ReactiveUI.Avalonia.Reactive;

[assembly: AvaloniaTestApplication(typeof(TestAppBuilder))]

namespace ARMEmulator.Tests.Ui;

/// <summary>Builds the application on the headless platform for [AvaloniaFact] tests.</summary>
public static class TestAppBuilder
{
	public static AppBuilder BuildAvaloniaApp() =>
		AppBuilder.Configure<App>()
			.UseHeadless(new AvaloniaHeadlessPlatformOptions())
			.UseReactiveUI(static _ => { });
}
