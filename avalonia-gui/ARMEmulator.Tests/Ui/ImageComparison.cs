namespace ARMEmulator.Tests.Ui;

/// <summary>
/// Compares rendered images, allowing for anti-aliasing noise.
/// </summary>
internal static class ImageComparison
{
	private const int BytesPerPixel = 4;

	/// <summary>A channel may differ by this much before the pixel counts as changed.</summary>
	public const int ChannelTolerance = 8;

	/// <summary>The share of pixels that may differ for two images to match.</summary>
	public const double MaxDifferentFraction = 0.005;

	/// <summary>The share of pixels in which any channel differs by more than <paramref name="channelTolerance"/>. Images of different sizes give 1.</summary>
	public static double FractionDifferent(PixelImage left, PixelImage right, int channelTolerance)
	{
		if (left.Width != right.Width || left.Height != right.Height) {
			return 1;
		}

		// A plain loop: images hold about a million pixels.
		var pixels = left.Width * left.Height;
		var different = 0;
		for (var pixel = 0; pixel < pixels; ++pixel) {
			if (Differs(left.Bgra, right.Bgra, pixel * BytesPerPixel, channelTolerance)) {
				++different;
			}
		}

		return (double)different / pixels;
	}

	/// <summary>True when the images differ in no more than <see cref="MaxDifferentFraction"/> of their pixels.</summary>
	public static bool Matches(PixelImage baseline, PixelImage actual) =>
		FractionDifferent(baseline, actual, ChannelTolerance) <= MaxDifferentFraction;

	private static bool Differs(byte[] left, byte[] right, int offset, int tolerance)
	{
		for (var channel = 0; channel < BytesPerPixel; ++channel) {
			if (Math.Abs(left[offset + channel] - right[offset + channel]) > tolerance) {
				return true;
			}
		}

		return false;
	}
}
