using System.Reactive;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

public sealed class FileServiceRecentFilesTests : IDisposable
{
	private readonly DirectoryInfo directory = Directory.CreateTempSubdirectory("arm-recent-tests");

	public void Dispose() => directory.Delete(recursive: true);

	private string CreateFile(string name)
	{
		var path = Path.Combine(directory.FullName, name);
		File.WriteAllText(path, "");
		return path;
	}

	private static string[] Paths(FileService service) => [.. service.RecentFiles.Select(f => f.Path)];

	[Fact]
	public void AddRecentFile_HonoursTheConfiguredLimit()
	{
		using var service = new FileService { RecentFilesLimit = 3 };

		foreach (var name in new[] { "a", "b", "c", "d", "e" }) {
			service.AddRecentFile($"/p/{name}.s");
		}

		Paths(service).Should().Equal("/p/e.s", "/p/d.s", "/p/c.s");
	}

	[Fact]
	public void LoweringTheLimit_DropsTheOldestEntries()
	{
		using var service = new FileService();
		service.LoadRecentFiles(["/p/a.s", "/p/b.s", "/p/c.s"]);

		service.RecentFilesLimit = 2;

		Paths(service).Should().Equal("/p/a.s", "/p/b.s");
	}

	[Fact]
	public void LoadRecentFiles_KeepsOrderDropsDuplicatesAndAppliesLimit()
	{
		using var service = new FileService { RecentFilesLimit = 2 };

		service.LoadRecentFiles(["/p/a.s", "/p/b.s", "/p/A.s", "/p/c.s"]);

		Paths(service).Should().Equal("/p/a.s", "/p/b.s");
	}

	[Fact]
	public void LoadRecentFiles_DoesNotNotify()
	{
		using var service = new FileService();
		var notifications = 0;
		using var subscription = service.RecentFilesChanged.Subscribe(_ => notifications++);

		service.LoadRecentFiles(["/p/a.s"]);

		notifications.Should().Be(0);
	}

	[Fact]
	public void AddRemoveAndClear_EachNotifyOnce()
	{
		using var service = new FileService();
		var notifications = new List<Unit>();
		using var subscription = service.RecentFilesChanged.Subscribe(notifications.Add);

		service.AddRecentFile("/p/a.s");
		service.RemoveRecentFile("/p/a.s");
		service.AddRecentFile("/p/b.s");
		service.ClearRecentFiles();

		notifications.Should().HaveCount(4);
	}

	[Fact]
	public void RemoveRecentFile_OfUnknownPath_DoesNotNotify()
	{
		using var service = new FileService();
		var notifications = 0;
		using var subscription = service.RecentFilesChanged.Subscribe(_ => notifications++);

		service.RemoveRecentFile("/p/none.s");

		notifications.Should().Be(0);
	}

	[Fact]
	public void RemoveMissingRecentFiles_DropsOnlyFilesThatNoLongerExist()
	{
		var kept = CreateFile("kept.s");
		var missing = Path.Combine(directory.FullName, "missing.s");
		using var service = new FileService();
		service.LoadRecentFiles([missing, kept]);
		var notifications = 0;
		using var subscription = service.RecentFilesChanged.Subscribe(_ => notifications++);

		service.RemoveMissingRecentFiles();

		Paths(service).Should().Equal(kept);
		notifications.Should().Be(1);
	}

	[Fact]
	public void RemoveMissingRecentFiles_WhenAllExist_DoesNotNotify()
	{
		using var service = new FileService();
		service.LoadRecentFiles([CreateFile("one.s")]);
		var notifications = 0;
		using var subscription = service.RecentFilesChanged.Subscribe(_ => notifications++);

		service.RemoveMissingRecentFiles();

		notifications.Should().Be(0);
	}

	[Fact]
	public void RecentFiles_IsASnapshotThatLaterChangesDoNotAlter()
	{
		using var service = new FileService();
		service.AddRecentFile("/p/a.s");
		var snapshot = service.RecentFiles;

		service.AddRecentFile("/p/b.s");

		snapshot.Select(f => f.Path).Should().Equal("/p/a.s");
	}
}
