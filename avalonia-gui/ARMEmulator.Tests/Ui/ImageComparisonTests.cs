using AwesomeAssertions;

namespace ARMEmulator.Tests.Ui;

public sealed class ImageComparisonTests
{
	private const int Tolerance = 8;

	private static PixelImage Image(int width, int height, params byte[][] pixels) =>
		new(width, height, [.. pixels.SelectMany(pixel => pixel)]);

	private static byte[] Pixel(byte b, byte g, byte r, byte a = 255) => [b, g, r, a];

	private static PixelImage FourPixels(byte[] last) =>
		Image(2, 2, Pixel(10, 20, 30), Pixel(40, 50, 60), Pixel(70, 80, 90), last);

	[Fact]
	public void FractionDifferent_OfIdenticalImages_IsZero() =>
		ImageComparison.FractionDifferent(FourPixels(Pixel(1, 2, 3)), FourPixels(Pixel(1, 2, 3)), Tolerance).Should().Be(0);

	[Fact]
	public void FractionDifferent_WithOneOfFourPixelsChanged_IsAQuarter() =>
		ImageComparison.FractionDifferent(FourPixels(Pixel(1, 2, 3)), FourPixels(Pixel(1, 2, 200)), Tolerance).Should().Be(0.25);

	[Fact]
	public void FractionDifferent_TreatsADifferenceAtTheToleranceAsEqual() =>
		ImageComparison.FractionDifferent(FourPixels(Pixel(100, 100, 100)), FourPixels(Pixel(108, 92, 100)), Tolerance).Should().Be(0);

	[Fact]
	public void FractionDifferent_TreatsADifferenceOverTheToleranceAsDifferent() =>
		ImageComparison.FractionDifferent(FourPixels(Pixel(100, 100, 100)), FourPixels(Pixel(109, 100, 100)), Tolerance).Should().Be(0.25);

	[Fact]
	public void FractionDifferent_ComparesTheAlphaChannelToo() =>
		ImageComparison.FractionDifferent(FourPixels(Pixel(1, 2, 3, 255)), FourPixels(Pixel(1, 2, 3, 0)), Tolerance).Should().Be(0.25);

	[Fact]
	public void FractionDifferent_OfDifferentSizes_IsOne() =>
		ImageComparison.FractionDifferent(Image(1, 1, Pixel(1, 2, 3)), FourPixels(Pixel(1, 2, 3)), Tolerance).Should().Be(1);

	[Fact]
	public void Matches_AcceptsAFewStrayPixels()
	{
		var width = 100;
		var height = 100;
		var baseline = new PixelImage(width, height, new byte[width * height * 4]);
		var changed = new PixelImage(width, height, [.. baseline.Bgra]);
		changed.Bgra[0] = 255;
		changed.Bgra[4] = 255;

		ImageComparison.Matches(baseline, changed).Should().BeTrue();
	}

	[Fact]
	public void Matches_RejectsAVisibleRegionChange()
	{
		var width = 100;
		var height = 100;
		var baseline = new PixelImage(width, height, new byte[width * height * 4]);
		var changed = new PixelImage(width, height, [.. baseline.Bgra]);
		for (var offset = 0; offset < 100 * 4; offset += 4) {
			changed.Bgra[offset] = 255;
		}

		ImageComparison.Matches(baseline, changed).Should().BeFalse();
	}
}
