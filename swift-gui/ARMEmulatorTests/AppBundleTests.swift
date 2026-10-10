import XCTest

/// The macOS menu bar, Dock and Finder take the app name from the bundle's Info.plist.
final class AppBundleTests: XCTestCase {
    private static let userVisibleName = "ARM Emulator"
    private static let executableName = "ARMEmulator"

    private func infoValue(_ key: String) -> String? {
        Bundle.main.object(forInfoDictionaryKey: key) as? String
    }

    func testBundleNameIsUserVisibleName() {
        XCTAssertEqual(infoValue("CFBundleName"), Self.userVisibleName)
    }

    func testBundleDisplayNameIsUserVisibleName() {
        XCTAssertEqual(infoValue("CFBundleDisplayName"), Self.userVisibleName)
    }

    func testExecutableNameHasNoSpace() {
        XCTAssertEqual(infoValue("CFBundleExecutable"), Self.executableName)
    }
}
