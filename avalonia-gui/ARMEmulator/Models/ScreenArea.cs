namespace ARMEmulator.Models;

/// <summary>A screen's area in device pixels.</summary>
public readonly record struct ScreenArea(int X, int Y, int Width, int Height);
