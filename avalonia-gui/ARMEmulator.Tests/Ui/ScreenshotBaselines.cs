#pragma warning disable CS0618 // Bitmap.Save(string) is the only overload that needs no encoder options
using System.Runtime.InteropServices;
using Avalonia.Media.Imaging;

namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Reviewed screenshots per operating system, compared with a pixel tolerance. Text rasterisation differs between
/// platforms, so each platform keeps its own set, and CI compares only when <c>ARM_VERIFY_SCREENSHOTS=1</c>.
/// Write or refresh the set with <c>ARM_UPDATE_BASELINES=1</c>.
/// </summary>
internal static class ScreenshotBaselines
{
	private const string UpdateVariable = "ARM_UPDATE_BASELINES";
	private const string VerifyVariable = "ARM_VERIFY_SCREENSHOTS";
	private const string CiVariable = "CI";
	private const string ReviewDirectoryVariable = "ARM_SCREENSHOT_DIR";
	private const string SetEnabled = "1";

	private static string Platform
	{
		get
		{
			if (OperatingSystem.IsMacOS()) {
				return "macos";
			}

			return OperatingSystem.IsWindows() ? "windows" : "linux";
		}
	}

	private static string Directory => Path.Combine(Harness.RepoRoot, "ARMEmulator.Tests", "Ui", "Baselines", Platform);

	private static bool IsSet(string variable) => !string.IsNullOrEmpty(Environment.GetEnvironmentVariable(variable));

	/// <summary>True when baselines are being written.</summary>
	public static bool UpdateRequested => Environment.GetEnvironmentVariable(UpdateVariable) == SetEnabled;

	/// <summary>True when screenshots are written to a folder for review.</summary>
	public static bool ReviewRequested => IsSet(ReviewDirectoryVariable);

	/// <summary>True when this platform has baselines and the run should compare against them.</summary>
	public static bool CanVerify =>
		System.IO.Directory.Exists(Directory) && (!IsSet(CiVariable) || Environment.GetEnvironmentVariable(VerifyVariable) == SetEnabled);

	/// <summary>Whether the screenshot tests have anything to do in this run.</summary>
	public static bool IsActive => UpdateRequested || ReviewRequested || CanVerify;

	/// <summary>Writes, reviews or compares one screenshot, as the environment asks.</summary>
	public static void Check(string name, WriteableBitmap actual)
	{
		if (ReviewRequested) {
			actual.Save(Path.Combine(Environment.GetEnvironmentVariable(ReviewDirectoryVariable)!, $"{name}.png"));
		}

		var baselinePath = Path.Combine(Directory, $"{name}.png");
		if (UpdateRequested) {
			_ = System.IO.Directory.CreateDirectory(Directory);
			actual.Save(baselinePath);
			return;
		}

		if (!CanVerify) {
			return;
		}

		Assert.True(File.Exists(baselinePath), $"No baseline for {name}. Run with {UpdateVariable}=1 to create it.");
		using var baselineBitmap = new Bitmap(baselinePath);
		var baseline = Decode(baselineBitmap);
		var current = Decode(actual);
		var different = ImageComparison.FractionDifferent(baseline, current, ImageComparison.ChannelTolerance);
		if (different > ImageComparison.MaxDifferentFraction) {
			var actualPath = Path.Combine(Path.GetTempPath(), $"{name}.actual.png");
			actual.Save(actualPath);
			Assert.Fail($"{name} differs from its baseline in {different:P2} of pixels (limit {ImageComparison.MaxDifferentFraction:P2}). Actual image: {actualPath}");
		}
	}

	private static PixelImage Decode(Bitmap bitmap)
	{
		var size = bitmap.PixelSize;
		var stride = size.Width * 4;
		var bytes = new byte[stride * size.Height];
		var handle = GCHandle.Alloc(bytes, GCHandleType.Pinned);
		try {
			bitmap.CopyPixels(new Avalonia.PixelRect(size), handle.AddrOfPinnedObject(), bytes.Length, stride);
		}
		finally {
			handle.Free();
		}

		return new PixelImage(size.Width, size.Height, bytes);
	}
}
