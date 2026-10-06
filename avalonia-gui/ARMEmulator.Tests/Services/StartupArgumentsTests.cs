using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

public sealed class StartupArgumentsTests
{
	private const string WorkingDirectory = "/home/user/work";

	[Fact]
	public void FindSourceFile_WithNoArguments_ReturnsNull() =>
		StartupArguments.FindSourceFile([], WorkingDirectory).Should().BeNull();

	[Fact]
	public void FindSourceFile_ReturnsTheFirstAssemblyFile() =>
		StartupArguments.FindSourceFile(["/a/one.s", "/b/two.s"], WorkingDirectory).Should().Be("/a/one.s");

	[Fact]
	public void FindSourceFile_IgnoresFlagsAndOtherFiles() =>
		StartupArguments.FindSourceFile(["--verbose", "notes.txt", "-x.s", "prog.s"], WorkingDirectory)
			.Should().Be("/home/user/work/prog.s");

	[Fact]
	public void FindSourceFile_ResolvesRelativePathsAgainstTheWorkingDirectory() =>
		StartupArguments.FindSourceFile(["examples/../prog.s"], WorkingDirectory).Should().Be("/home/user/work/prog.s");

	[Fact]
	public void FindSourceFile_MatchesTheExtensionIgnoringCase() =>
		StartupArguments.FindSourceFile(["PROG.S"], WorkingDirectory).Should().Be("/home/user/work/PROG.S");

	[Fact]
	public void FindSourceFile_RequiresTheExtensionToEndTheName() =>
		StartupArguments.FindSourceFile(["prog.s.bak", "prog.sx", "s"], WorkingDirectory).Should().BeNull();
}
