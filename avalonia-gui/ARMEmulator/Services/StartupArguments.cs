namespace ARMEmulator.Services;

/// <summary>
/// Interprets the command line.
/// </summary>
public static class StartupArguments
{
	private const string SourceExtension = ".s";
	private const string FlagPrefix = "-";

	/// <summary>
	/// Returns the full path of the first argument that names an assembly file (<c>.s</c>), or null.
	/// Arguments that start with <c>-</c> are flags and are skipped. Relative paths resolve against <paramref name="workingDirectory"/>.
	/// </summary>
	public static string? FindSourceFile(IEnumerable<string> arguments, string workingDirectory) =>
		arguments
			.Where(argument => !argument.StartsWith(FlagPrefix, StringComparison.Ordinal)
				&& argument.EndsWith(SourceExtension, StringComparison.OrdinalIgnoreCase))
			.Select(argument => Path.GetFullPath(argument, workingDirectory))
			.FirstOrDefault();
}
