#!/usr/bin/env swift

import Cocoa
import CoreGraphics

// Generate app icon with 'arm' text
// Creates multiple sizes for macOS app bundle

let sizes: [CGFloat] = [16, 32, 64, 128, 256, 512, 1024]

func generateIcon(size: CGFloat, outputPath: String) {
    let rect = CGRect(x: 0, y: 0, width: size, height: size)

    // Create bitmap context
    let colorSpace = CGColorSpaceCreateDeviceRGB()
    guard let context = CGContext(
        data: nil,
        width: Int(size),
        height: Int(size),
        bitsPerComponent: 8,
        bytesPerRow: 0,
        space: colorSpace,
        bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue,
    ) else {
        print("Failed to create context for size \(size)")
        return
    }

    // Enable antialiasing
    context.setAllowsAntialiasing(true)
    context.setShouldAntialias(true)

    // Background - gradient from deep blue to purple
    let gradient = CGGradient(
        colorsSpace: colorSpace,
        colors: [
            CGColor(red: 0.2, green: 0.3, blue: 0.6, alpha: 1.0), // Deep blue
            CGColor(red: 0.4, green: 0.2, blue: 0.6, alpha: 1.0), // Purple
        ] as CFArray,
        locations: [0.0, 1.0],
    )!

    // Draw rounded rectangle background
    let cornerRadius = size * 0.2
    let backgroundPath = CGPath(
        roundedRect: rect,
        cornerWidth: cornerRadius,
        cornerHeight: cornerRadius,
        transform: nil,
    )
    context.addPath(backgroundPath)
    context.clip()
    context.drawLinearGradient(
        gradient,
        start: CGPoint(x: size / 2, y: 0),
        end: CGPoint(x: size / 2, y: size),
        options: [],
    )

    // Add subtle circuit board pattern in background
    context.setStrokeColor(CGColor(red: 1, green: 1, blue: 1, alpha: 0.1))
    context.setLineWidth(size * 0.01)
    let gridSpacing = size / 8
    for i in stride(from: gridSpacing, to: size, by: gridSpacing) {
        // Vertical lines
        context.move(to: CGPoint(x: i, y: 0))
        context.addLine(to: CGPoint(x: i, y: size))
        // Horizontal lines
        context.move(to: CGPoint(x: 0, y: i))
        context.addLine(to: CGPoint(x: size, y: i))
    }
    context.strokePath()

    // Draw 'arm' text using Core Text for proper rendering
    let fontSize = size * 0.35
    let text = "arm"

    // Create attributed string with font
    let font = CTFontCreateWithName("Helvetica-Bold" as CFString, fontSize, nil)
    let attributes: [NSAttributedString.Key: Any] = [
        .font: font,
        .foregroundColor: CGColor(red: 1, green: 1, blue: 1, alpha: 1),
    ]

    let attributedString = CFAttributedStringCreate(
        kCFAllocatorDefault,
        text as CFString,
        attributes as CFDictionary,
    )!

    let line = CTLineCreateWithAttributedString(attributedString)
    let textBounds = CTLineGetBoundsWithOptions(line, .useOpticalBounds)

    // Center the text
    let textX = (size - textBounds.width) / 2 - textBounds.origin.x
    let textY = (size - textBounds.height) / 2 - textBounds.origin.y

    // Save graphics state
    context.saveGState()

    // Set up shadow
    context.setShadow(
        offset: CGSize(width: 0, height: -size * 0.02),
        blur: size * 0.05,
        color: CGColor(red: 0, green: 0, blue: 0, alpha: 0.5),
    )

    // Position and draw text
    context.textPosition = CGPoint(x: textX, y: textY)
    CTLineDraw(line, context)

    // Restore graphics state
    context.restoreGState()

    // Add subtle highlight on top edge
    context.setStrokeColor(CGColor(red: 1, green: 1, blue: 1, alpha: 0.3))
    context.setLineWidth(size * 0.02)
    context.addArc(
        center: CGPoint(x: size / 2, y: size / 2),
        radius: size * 0.45,
        startAngle: .pi * 1.1,
        endAngle: .pi * 1.9,
        clockwise: false,
    )
    context.strokePath()

    // Create image from context
    guard let cgImage = context.makeImage() else {
        print("Failed to create image for size \(size)")
        return
    }

    // Save as PNG
    let bitmapRep = NSBitmapImageRep(cgImage: cgImage)
    guard let pngData = bitmapRep.representation(using: .png, properties: [:]) else {
        print("Failed to create PNG data for size \(size)")
        return
    }

    do {
        try pngData.write(to: URL(fileURLWithPath: outputPath))
        print("✓ Generated icon: \(size)x\(size) -> \(outputPath)")
    } catch {
        print("Failed to write icon for size \(size): \(error)")
    }
}

// Create output directory
let outputDir = "ARMEmulator/Assets.xcassets/AppIcon.appiconset"
let fileManager = FileManager.default
try? fileManager.createDirectory(atPath: outputDir, withIntermediateDirectories: true)

// Generate all sizes
for size in sizes {
    let filename = "\(outputDir)/icon_\(Int(size))x\(Int(size)).png"
    generateIcon(size: size, outputPath: filename)

    // Generate @2x version for retina
    if size <= 512 {
        let filename2x = "\(outputDir)/icon_\(Int(size))x\(Int(size))@2x.png"
        generateIcon(size: size * 2, outputPath: filename2x)
    }
}

// Generate Contents.json for Asset Catalog
let contentsJSON = """
{
  "images" : [
    {
      "filename" : "icon_16x16.png",
      "idiom" : "mac",
      "scale" : "1x",
      "size" : "16x16"
    },
    {
      "filename" : "icon_16x16@2x.png",
      "idiom" : "mac",
      "scale" : "2x",
      "size" : "16x16"
    },
    {
      "filename" : "icon_32x32.png",
      "idiom" : "mac",
      "scale" : "1x",
      "size" : "32x32"
    },
    {
      "filename" : "icon_32x32@2x.png",
      "idiom" : "mac",
      "scale" : "2x",
      "size" : "32x32"
    },
    {
      "filename" : "icon_128x128.png",
      "idiom" : "mac",
      "scale" : "1x",
      "size" : "128x128"
    },
    {
      "filename" : "icon_128x128@2x.png",
      "idiom" : "mac",
      "scale" : "2x",
      "size" : "128x128"
    },
    {
      "filename" : "icon_256x256.png",
      "idiom" : "mac",
      "scale" : "1x",
      "size" : "256x256"
    },
    {
      "filename" : "icon_256x256@2x.png",
      "idiom" : "mac",
      "scale" : "2x",
      "size" : "256x256"
    },
    {
      "filename" : "icon_512x512.png",
      "idiom" : "mac",
      "scale" : "1x",
      "size" : "512x512"
    },
    {
      "filename" : "icon_512x512@2x.png",
      "idiom" : "mac",
      "scale" : "2x",
      "size" : "512x512"
    }
  ],
  "info" : {
    "author" : "xcode",
    "version" : 1
  }
}
"""

let contentsPath = "\(outputDir)/Contents.json"
try? contentsJSON.write(toFile: contentsPath, atomically: true, encoding: .utf8)
print("✓ Generated Contents.json")

print("\n✅ Icon generation complete!")
print("The icons are ready in: \(outputDir)")
print("\nNext steps:")
print("1. Open swift-gui/ARMEmulator.xcodeproj in Xcode")
print("2. The AppIcon asset should now contain all generated icons")
print("3. Build and run to see the new icon!")
