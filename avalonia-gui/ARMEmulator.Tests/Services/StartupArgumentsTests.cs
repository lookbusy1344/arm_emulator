using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

public sealed class StartupArgumentsTests
{
	private static readonly string WorkingDirectory = Path.Combine(Path.GetTempPath(), "work");

	private static string InWorkingDirectory(string name) => Path.Combine(WorkingDirectory, name);
	private static readonly string AbsoluteOne = Path.Combine(Path.GetTempPath(), "a", "one.s");
	private static readonly string AbsoluteTwo = Path.Combine(Path.GetTempPath(), "b", "two.s");

	[Fact]
	public void FindSourceFile_WithNoArguments_ReturnsNull() =>
		StartupArguments.FindSourceFile([], WorkingDirectory).Should().BeNull();

	[Fact]
	public void FindSourceFile_ReturnsTheFirstAssemblyFile() =>
		StartupArguments.FindSourceFile([AbsoluteOne, AbsoluteTwo], WorkingDirectory).Should().Be(AbsoluteOne);

	[Fact]
	public void FindSourceFile_IgnoresFlagsAndOtherFiles() =>
		StartupArguments.FindSourceFile(["--verbose", "notes.txt", "-x.s", "prog.s"], WorkingDirectory)
			.Should().Be(InWorkingDirectory("prog.s"));

	[Fact]
	public void FindSourceFile_ResolvesRelativePathsAgainstTheWorkingDirectory() =>
		StartupArguments.FindSourceFile(["examples/../prog.s"], WorkingDirectory).Should().Be(InWorkingDirectory("prog.s"));

	[Fact]
	public void FindSourceFile_MatchesTheExtensionIgnoringCase() =>
		StartupArguments.FindSourceFile(["PROG.S"], WorkingDirectory).Should().Be(InWorkingDirectory("PROG.S"));

	[Fact]
	public void FindSourceFile_RequiresTheExtensionToEndTheName() =>
		StartupArguments.FindSourceFile(["prog.s.bak", "prog.sx", "s"], WorkingDirectory).Should().BeNull();
}
