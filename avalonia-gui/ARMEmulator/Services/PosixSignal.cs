using System.Runtime.InteropServices;
using System.Runtime.Versioning;

namespace ARMEmulator.Services;

/// <summary>Sends POSIX signals, which .NET exposes only for the current process.</summary>
[UnsupportedOSPlatform("windows")]
internal static partial class PosixSignal
{
	private const int SigTerm = 15;

	/// <summary>Asks <paramref name="processId"/> to terminate. Returns false when the process no longer exists.</summary>
	public static bool Terminate(int processId) => Kill(processId, SigTerm) == 0;

	[LibraryImport("libc", EntryPoint = "kill", SetLastError = true)]
	private static partial int Kill(int pid, int sig);
}
