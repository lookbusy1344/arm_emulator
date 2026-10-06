using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for the unsaved-changes flag, window title and the discard prompt.
/// </summary>
public sealed class MainWindowViewModelFileTests : IDisposable
{
	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IUnsavedChangesPrompt prompt = Substitute.For<IUnsavedChangesPrompt>();

	private readonly Subject<EmulatorEvent> events = new();

	public MainWindowViewModelFileTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
	}

	public void Dispose() => events.Dispose();

	private MainWindowViewModel CreateViewModel() => new(api, ws, files, unsavedChangesPrompt: prompt);

	// Title formatting

	[Theory]
	[InlineData(null, false, "ARM Emulator")]
	[InlineData(null, true, "ARM Emulator — Untitled •")]
	[InlineData("/work/file.s", false, "ARM Emulator — file.s")]
	[InlineData("/work/file.s", true, "ARM Emulator — file.s •")]
	public void WindowTitle_Format(string? path, bool dirty, string expected) =>
		WindowTitle.Format(path, dirty).Should().Be(expected);

	// Dirty flag

	[Fact]
	public void NewViewModel_IsClean()
	{
		using var vm = CreateViewModel();

		vm.IsDirty.Should().BeFalse();
		vm.WindowTitle.Should().Be("ARM Emulator");
	}

	[Fact]
	public void EditingTheSource_MarksItDirtyAndAddsTheMarkerToTheTitle()
	{
		using var vm = CreateViewModel();

		vm.SourceCode = "MOV R0, #1";

		vm.IsDirty.Should().BeTrue();
		vm.WindowTitle.Should().Be("ARM Emulator — Untitled •");
	}

	[Fact]
	public void AssigningTheSameSource_LeavesItClean()
	{
		using var vm = CreateViewModel();

		vm.SourceCode = "";

		vm.IsDirty.Should().BeFalse();
	}

	[Fact]
	public async Task OpeningAFile_ClearsTheFlagAndShowsTheFileNameInTheTitle()
	{
		var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.s");
		try {
			await File.WriteAllTextAsync(path, "MOV R0, #1", TestContext.Current.CancellationToken);
			using var vm = CreateViewModel();
			vm.SourceCode = "edited";

			await vm.OpenSourceFileAsync(path, TestContext.Current.CancellationToken);

			vm.IsDirty.Should().BeFalse();
			vm.WindowTitle.Should().Be($"ARM Emulator — {Path.GetFileName(path)}");
			files.Received().CurrentFilePath = path;
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task EditingAfterOpening_MarksTheTitleDirty()
	{
		var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.s");
		try {
			await File.WriteAllTextAsync(path, "MOV R0, #1", TestContext.Current.CancellationToken);
			using var vm = CreateViewModel();
			await vm.OpenSourceFileAsync(path, TestContext.Current.CancellationToken);

			vm.SourceCode += "\nSWI #0";

			vm.WindowTitle.Should().Be($"ARM Emulator — {Path.GetFileName(path)} •");
		}
		finally {
			File.Delete(path);
		}
	}

	// Discard prompt

	[Fact]
	public async Task ConfirmDiscard_WhenClean_AllowsWithoutPrompting()
	{
		using var vm = CreateViewModel();

		var allowed = await vm.ConfirmDiscardAsync();

		allowed.Should().BeTrue();
		await prompt.DidNotReceiveWithAnyArgs().AskAsync(default!);
	}

	[Fact]
	public async Task ConfirmDiscard_WhenDirty_NamesTheDocumentInThePrompt()
	{
		var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.s");
		try {
			await File.WriteAllTextAsync(path, "MOV R0, #1", TestContext.Current.CancellationToken);
			prompt.AskAsync(Arg.Any<string>()).Returns(UnsavedChangesChoice.Cancel);
			using var vm = CreateViewModel();
			await vm.OpenSourceFileAsync(path, TestContext.Current.CancellationToken);
			vm.SourceCode = "edited";

			_ = await vm.ConfirmDiscardAsync();

			await prompt.Received(1).AskAsync(Path.GetFileName(path));
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task ConfirmDiscard_ForAnUnsavedNewDocument_UsesTheUntitledName()
	{
		prompt.AskAsync(Arg.Any<string>()).Returns(UnsavedChangesChoice.Cancel);
		using var vm = CreateViewModel();
		vm.SourceCode = "edited";

		_ = await vm.ConfirmDiscardAsync();

		await prompt.Received(1).AskAsync("Untitled");
	}

	[Theory]
	[InlineData(UnsavedChangesChoice.Discard, true)]
	[InlineData(UnsavedChangesChoice.Cancel, false)]
	public async Task ConfirmDiscard_WhenDirty_FollowsTheChoice(UnsavedChangesChoice choice, bool expected)
	{
		prompt.AskAsync(Arg.Any<string>()).Returns(choice);
		using var vm = CreateViewModel();
		vm.SourceCode = "edited";

		var allowed = await vm.ConfirmDiscardAsync();

		allowed.Should().Be(expected);
		vm.IsDirty.Should().BeTrue();
	}

	[Fact]
	public async Task ConfirmDiscard_WhenSaveDoesNotComplete_KeepsTheEdits()
	{
		// Without a parent window the save dialog cannot open, so the document stays unsaved
		prompt.AskAsync(Arg.Any<string>()).Returns(UnsavedChangesChoice.Save);
		using var vm = CreateViewModel();
		vm.SourceCode = "edited";

		var allowed = await vm.ConfirmDiscardAsync();

		allowed.Should().BeFalse();
		vm.IsDirty.Should().BeTrue();
	}

	[Fact]
	public async Task OpenRecentFile_WhenDirtyAndCancelled_KeepsTheEditorContent()
	{
		var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.s");
		try {
			await File.WriteAllTextAsync(path, "other", TestContext.Current.CancellationToken);
			prompt.AskAsync(Arg.Any<string>()).Returns(UnsavedChangesChoice.Cancel);
			using var vm = CreateViewModel();
			vm.SourceCode = "edited";

			await vm.OpenRecentFileCommand.Execute(path);

			vm.SourceCode.Should().Be("edited");
			vm.IsDirty.Should().BeTrue();
		}
		finally {
			File.Delete(path);
		}
	}

	[Fact]
	public async Task OpenRecentFile_WhenDirtyAndDiscarded_ReplacesTheContent()
	{
		var path = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.s");
		try {
			await File.WriteAllTextAsync(path, "other", TestContext.Current.CancellationToken);
			prompt.AskAsync(Arg.Any<string>()).Returns(UnsavedChangesChoice.Discard);
			using var vm = CreateViewModel();
			vm.SourceCode = "edited";

			await vm.OpenRecentFileCommand.Execute(path);

			vm.SourceCode.Should().Be("other");
			vm.IsDirty.Should().BeFalse();
		}
		finally {
			File.Delete(path);
		}
	}
}
