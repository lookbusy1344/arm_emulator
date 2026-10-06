namespace ARMEmulator.Tests;

/// <summary>
/// Shared test locations. <see cref="RepoRoot"/> is the Avalonia solution directory (the one holding
/// <c>ARMEmulator.slnx</c>), which the CodeStyle_* guards scan.
/// </summary>
internal static class Harness
{
	private const string SolutionFileName = "ARMEmulator.slnx";

	public static string RepoRoot { get; } = FindRepoRoot();

	private static string FindRepoRoot()
	{
		for (var directory = new DirectoryInfo(AppContext.BaseDirectory); directory is not null; directory = directory.Parent) {
			if (File.Exists(Path.Combine(directory.FullName, SolutionFileName))) {
				return directory.FullName;
			}
		}

		throw new InvalidOperationException($"No {SolutionFileName} found above {AppContext.BaseDirectory}.");
	}
}
