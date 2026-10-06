using System.Diagnostics.CodeAnalysis;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

public sealed class MainWindowViewModelSettingsTests : IDisposable
{
	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly ISettingsStore store = Substitute.For<ISettingsStore>();

	private readonly Subject<EmulatorEvent> events = new();

	private static readonly AppSettings Changed = AppSettings.Default with {
		EditorFontSize = 20,
		Theme = AppTheme.Dark,
		AutoScrollToMemoryWrites = false
	};

	public MainWindowViewModelSettingsTests() => ws.Events.Returns(events);

	public void Dispose() => events.Dispose();

	private MainWindowViewModel CreateViewModel() => new(api, ws, files, store);

	[Fact]
	public void Settings_DefaultToAppDefaults()
	{
		using var vm = CreateViewModel();

		vm.Settings.Should().Be(AppSettings.Default);
	}

	[Fact]
	public void SaveSettings_AppliesThenPersists()
	{
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		vm.Settings.Should().Be(Changed);
		vm.Memory.AutoScrollToWrites.Should().BeFalse();
		store.Received(1).Save(Changed);
		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public void SaveSettings_WhenStoreFails_ReportsErrorAndKeepsAppliedSettings()
	{
		store.When(s => s.Save(Arg.Any<AppSettings>())).Throw(new IOException("disk full"));
		using var vm = CreateViewModel();

		vm.SaveSettings(Changed);

		vm.ErrorMessage.Should().Be("Failed to save settings: disk full");
		vm.Settings.Should().Be(Changed);
	}

	[Fact]
	public void ApplySettings_DoesNotPersist()
	{
		using var vm = CreateViewModel();

		vm.ApplySettings(Changed);

		vm.Settings.Should().Be(Changed);
		store.DidNotReceiveWithAnyArgs().Save(default!);
	}
}
