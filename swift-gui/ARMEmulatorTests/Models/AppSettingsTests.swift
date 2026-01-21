import SwiftUI
import XCTest
@testable import ARMEmulator

@MainActor
final class AppSettingsTests: XCTestCase {
    var settings: AppSettings!
    var testDefaults: UserDefaults!

    override func setUp() async throws {
        try await super.setUp()

        // Create isolated UserDefaults for testing
        // Use a unique suite name to avoid polluting app defaults
        let suiteName = "test.ARMEmulator.AppSettingsTests.\(UUID().uuidString)"
        testDefaults = UserDefaults(suiteName: suiteName)!

        // Clear any existing data
        testDefaults.removePersistentDomain(forName: suiteName)

        // Note: AppSettings uses @AppStorage which internally uses UserDefaults.standard
        // We cannot inject a custom UserDefaults into @AppStorage directly
        // So we'll test via the actual UserDefaults.standard but clean up after
        settings = AppSettings.shared
    }

    override func tearDown() async throws {
        // Clean up test data from UserDefaults.standard
        UserDefaults.standard.removeObject(forKey: "backendURL")
        UserDefaults.standard.removeObject(forKey: "editorFontSize")
        UserDefaults.standard.removeObject(forKey: "colorScheme")
        UserDefaults.standard.removeObject(forKey: "maxRecentFiles")
        UserDefaults.standard.removeObject(forKey: "selectedTab")

        settings = nil
        testDefaults = nil

        try await super.tearDown()
    }

    // MARK: - Default Values Tests

    func testDefaultBackendURL() {
        // Create fresh settings instance
        UserDefaults.standard.removeObject(forKey: "backendURL")
        let freshSettings = AppSettings()

        XCTAssertEqual(freshSettings.backendURL, "http://localhost:8080")
    }

    func testDefaultEditorFontSize() {
        UserDefaults.standard.removeObject(forKey: "editorFontSize")
        let freshSettings = AppSettings()

        XCTAssertEqual(freshSettings.editorFontSize, 14)
    }

    func testDefaultColorScheme() {
        UserDefaults.standard.removeObject(forKey: "colorScheme")
        let freshSettings = AppSettings()

        XCTAssertEqual(freshSettings.colorScheme, "auto")
    }

    func testDefaultMaxRecentFiles() {
        UserDefaults.standard.removeObject(forKey: "maxRecentFiles")
        let freshSettings = AppSettings()

        XCTAssertEqual(freshSettings.maxRecentFiles, 10)
    }

    func testDefaultSelectedTab() {
        UserDefaults.standard.removeObject(forKey: "selectedTab")
        let freshSettings = AppSettings()

        XCTAssertEqual(freshSettings.selectedTab, 0)
    }

    // MARK: - UserDefaults Persistence Tests

    func testBackendURLPersistence() {
        // Set value
        settings.backendURL = "http://localhost:9000"

        // Verify persisted to UserDefaults
        let persisted = UserDefaults.standard.string(forKey: "backendURL")
        XCTAssertEqual(persisted, "http://localhost:9000")

        // Create new instance and verify it reads the persisted value
        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.backendURL, "http://localhost:9000")
    }

    func testEditorFontSizePersistence() {
        settings.editorFontSize = 18

        let persisted = UserDefaults.standard.integer(forKey: "editorFontSize")
        XCTAssertEqual(persisted, 18)

        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.editorFontSize, 18)
    }

    func testColorSchemePersistence() {
        settings.colorScheme = "dark"

        let persisted = UserDefaults.standard.string(forKey: "colorScheme")
        XCTAssertEqual(persisted, "dark")

        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.colorScheme, "dark")
    }

    func testMaxRecentFilesPersistence() {
        settings.maxRecentFiles = 20

        let persisted = UserDefaults.standard.integer(forKey: "maxRecentFiles")
        XCTAssertEqual(persisted, 20)

        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.maxRecentFiles, 20)
    }

    func testSelectedTabPersistence() {
        settings.selectedTab = 2

        let persisted = UserDefaults.standard.integer(forKey: "selectedTab")
        XCTAssertEqual(persisted, 2)

        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.selectedTab, 2)
    }

    // MARK: - PreferredColorScheme Computed Property Tests

    func testPreferredColorSchemeLight() {
        settings.colorScheme = "light"
        XCTAssertEqual(settings.preferredColorScheme, .light)
    }

    func testPreferredColorSchemeDark() {
        settings.colorScheme = "dark"
        XCTAssertEqual(settings.preferredColorScheme, .dark)
    }

    func testPreferredColorSchemeAuto() {
        settings.colorScheme = "auto"
        XCTAssertNil(settings.preferredColorScheme) // nil means use system default
    }

    func testPreferredColorSchemeInvalidDefaultsToAuto() {
        settings.colorScheme = "invalid"
        XCTAssertNil(settings.preferredColorScheme) // Unknown values default to auto (nil)
    }

    func testPreferredColorSchemeEmptyStringDefaultsToAuto() {
        settings.colorScheme = ""
        XCTAssertNil(settings.preferredColorScheme)
    }

    // MARK: - Shared Instance Tests

    func testSharedInstanceIsSingleton() {
        let instance1 = AppSettings.shared
        let instance2 = AppSettings.shared

        // Both references should point to the same object
        XCTAssertTrue(instance1 === instance2)
    }

    func testSharedInstanceModificationReflectsInAllReferences() {
        let instance1 = AppSettings.shared
        let instance2 = AppSettings.shared

        instance1.editorFontSize = 16
        XCTAssertEqual(instance2.editorFontSize, 16)
    }

    // MARK: - Edge Case Tests

    func testBackendURLWithIPAddress() {
        settings.backendURL = "http://192.168.1.100:8080"
        XCTAssertEqual(settings.backendURL, "http://192.168.1.100:8080")

        let persisted = UserDefaults.standard.string(forKey: "backendURL")
        XCTAssertEqual(persisted, "http://192.168.1.100:8080")
    }

    func testBackendURLWithHTTPS() {
        settings.backendURL = "https://localhost:8443"
        XCTAssertEqual(settings.backendURL, "https://localhost:8443")
    }

    func testEditorFontSizeMinimumValue() {
        settings.editorFontSize = 8 // Small but valid
        XCTAssertEqual(settings.editorFontSize, 8)
    }

    func testEditorFontSizeMaximumValue() {
        settings.editorFontSize = 72 // Large but valid
        XCTAssertEqual(settings.editorFontSize, 72)
    }

    func testMaxRecentFilesZero() {
        settings.maxRecentFiles = 0
        XCTAssertEqual(settings.maxRecentFiles, 0)
    }

    func testMaxRecentFilesLargeNumber() {
        settings.maxRecentFiles = 100
        XCTAssertEqual(settings.maxRecentFiles, 100)
    }

    func testSelectedTabNegativeValue() {
        // @AppStorage with Int will store negative values
        settings.selectedTab = -1
        XCTAssertEqual(settings.selectedTab, -1)
    }

    func testSelectedTabLargeValue() {
        settings.selectedTab = 99
        XCTAssertEqual(settings.selectedTab, 99)
    }

    // MARK: - Migration Tests (for future schema changes)

    func testMigrationFromMissingKey() {
        // Simulate missing key (fresh install)
        UserDefaults.standard.removeObject(forKey: "backendURL")

        let newSettings = AppSettings()
        XCTAssertEqual(newSettings.backendURL, "http://localhost:8080") // Should use default
    }

    func testMigrationPreservesExistingValues() {
        // Simulate existing settings from previous version
        UserDefaults.standard.set("http://custom:8080", forKey: "backendURL")
        UserDefaults.standard.set(16, forKey: "editorFontSize")
        UserDefaults.standard.set("dark", forKey: "colorScheme")

        let newSettings = AppSettings()

        // Existing values should be preserved
        XCTAssertEqual(newSettings.backendURL, "http://custom:8080")
        XCTAssertEqual(newSettings.editorFontSize, 16)
        XCTAssertEqual(newSettings.colorScheme, "dark")
    }

    func testMigrationWithPartialSettings() {
        // Simulate scenario where some settings exist but not others
        UserDefaults.standard.set("http://custom:8080", forKey: "backendURL")
        UserDefaults.standard.removeObject(forKey: "editorFontSize")

        let newSettings = AppSettings()

        // Existing setting preserved
        XCTAssertEqual(newSettings.backendURL, "http://custom:8080")
        // Missing setting uses default
        XCTAssertEqual(newSettings.editorFontSize, 14)
    }

    // MARK: - ObservableObject Tests

    func testAppSettingsIsObservableObject() {
        // AppSettings conforms to ObservableObject
        // Changes should trigger @Published updates
        // Note: We can't easily test @Published in unit tests without SwiftUI views
        // This test just verifies the type conforms to ObservableObject
        XCTAssertTrue(settings is ObservableObject)
    }

    // MARK: - MainActor Isolation Tests

    func testAppSettingsRequiresMainActor() async {
        // This test verifies that AppSettings is @MainActor isolated
        // Creating an instance should work on MainActor
        await MainActor.run {
            let testSettings = AppSettings()
            XCTAssertNotNil(testSettings)
        }
    }

    // MARK: - Concurrent Access Tests

    func testConcurrentReadAccess() async {
        // Multiple concurrent reads should work
        await withTaskGroup(of: String.self) { group in
            for _ in 0 ..< 10 {
                group.addTask { @MainActor in
                    self.settings.backendURL
                }
            }

            var results: [String] = []
            for await result in group {
                results.append(result)
            }

            XCTAssertEqual(results.count, 10)
        }
    }

    func testSequentialWriteAccess() async {
        // Sequential writes should persist correctly
        for i in 0 ..< 5 {
            await MainActor.run {
                settings.editorFontSize = 14 + i
            }
        }

        await MainActor.run {
            XCTAssertEqual(settings.editorFontSize, 18) // 14 + 4
        }
    }

    // MARK: - ColorScheme Validation Tests

    func testAllValidColorSchemeValues() {
        let validSchemes = ["light", "dark", "auto"]

        for scheme in validSchemes {
            settings.colorScheme = scheme

            // Verify persisted
            let persisted = UserDefaults.standard.string(forKey: "colorScheme")
            XCTAssertEqual(persisted, scheme)

            // Verify read back
            let newSettings = AppSettings()
            XCTAssertEqual(newSettings.colorScheme, scheme)
        }
    }

    func testColorSchemePreferredMappingConsistency() {
        let mappings: [(String, ColorScheme?)] = [
            ("light", .light),
            ("dark", .dark),
            ("auto", nil),
            ("invalid", nil),
            ("", nil),
        ]

        for (input, expected) in mappings {
            settings.colorScheme = input
            XCTAssertEqual(
                settings.preferredColorScheme, expected,
                "colorScheme '\(input)' should map to \(String(describing: expected))",
            )
        }
    }
}
