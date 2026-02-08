using ARMEmulator.Services;
using FluentAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Tests for FileService.
/// Note: File dialog tests require integration testing with actual UI.
/// These tests verify recent file tracking and basic functionality.
/// </summary>
public sealed class FileServiceTests
{
	[Fact]
	public void Constructor_InitializesEmptyRecentFiles()
	{
		var service = new FileService();
		service.RecentFiles.Should().BeEmpty();
		service.CurrentFilePath.Should().BeNull();
	}

	[Fact]
	public void AddRecentFile_AddsToList()
	{
		var service = new FileService();
		service.AddRecentFile("/path/to/test.s");

		service.RecentFiles.Should().HaveCount(1);
		service.RecentFiles[0].Path.Should().Be("/path/to/test.s");
		service.RecentFiles[0].FileName.Should().Be("test.s");
	}

	[Fact]
	public void AddRecentFile_MostRecentFirst()
	{
		var service = new FileService();
		service.AddRecentFile("/path/one.s");
		service.AddRecentFile("/path/two.s");

		service.RecentFiles.Should().HaveCount(2);
		service.RecentFiles[0].Path.Should().Be("/path/two.s"); // Most recent
		service.RecentFiles[1].Path.Should().Be("/path/one.s");
	}

	[Fact]
	public void AddRecentFile_RemovesDuplicates()
	{
		var service = new FileService();
		service.AddRecentFile("/path/test.s");
		service.AddRecentFile("/path/other.s");
		service.AddRecentFile("/path/test.s"); // Duplicate

		service.RecentFiles.Should().HaveCount(2);
		service.RecentFiles[0].Path.Should().Be("/path/test.s"); // Moved to top
	}

	[Fact]
	public void AddRecentFile_LimitsToMaximum()
	{
		var service = new FileService();

		// Add more than the maximum
		for (int i = 0; i < 15; i++)
		{
			service.AddRecentFile($"/path/file{i}.s");
		}

		service.RecentFiles.Should().HaveCountLessOrEqualTo(10); // Default max
		service.RecentFiles[0].Path.Should().Be("/path/file14.s"); // Most recent
	}

	[Fact]
	public void ClearRecentFiles_RemovesAll()
	{
		var service = new FileService();
		service.AddRecentFile("/path/one.s");
		service.AddRecentFile("/path/two.s");

		service.ClearRecentFiles();

		service.RecentFiles.Should().BeEmpty();
	}

	[Fact]
	public void CurrentFilePath_CanBeSetAndRead()
	{
		var service = new FileService();
		service.CurrentFilePath.Should().BeNull();

		service.CurrentFilePath = "/path/test.s";
		service.CurrentFilePath.Should().Be("/path/test.s");
	}

	[Fact]
	public void RecentFile_FileName_ExtractsCorrectly()
	{
		var recent = new RecentFile("/some/long/path/example.s", DateTime.Now);
		recent.FileName.Should().Be("example.s");
	}
}
