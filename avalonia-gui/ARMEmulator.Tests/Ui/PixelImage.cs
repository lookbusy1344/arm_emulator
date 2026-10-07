namespace ARMEmulator.Tests.Ui;

/// <summary>A decoded image: four bytes per pixel, rows top to bottom.</summary>
internal sealed record PixelImage(int Width, int Height, byte[] Bgra);
