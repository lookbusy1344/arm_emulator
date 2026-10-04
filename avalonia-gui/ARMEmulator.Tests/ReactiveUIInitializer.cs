using System.Reactive.Concurrency;
using System.Runtime.CompilerServices;
using ReactiveUI.Builder;

namespace ARMEmulator.Tests;

/// <summary>
/// ReactiveUI throws on first use until its builder has run. The app runs it via
/// <c>UseReactiveUI</c> in <c>Program</c>; tests bypass <c>Program</c>, so run it once per test assembly.
/// Both schedulers run on the current thread so view-model pipelines complete synchronously.
/// </summary>
internal static class ReactiveUIInitializer
{
	[ModuleInitializer]
	internal static void Initialize() =>
		RxAppBuilder.CreateReactiveUIBuilder()
			.WithMainThreadScheduler(CurrentThreadScheduler.Instance)
			.WithTaskPoolScheduler(CurrentThreadScheduler.Instance)
			.WithCoreServices()
			.BuildApp();
}
