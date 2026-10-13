import AppKit
import XCTest
@testable import ARMEmulator

final class WindowTitleTests: XCTestCase {
    private let file = URL(fileURLWithPath: "/tmp/examples/hello.s")

    func testNoFileAndNoEditsIsTheApplicationName() {
        XCTAssertEqual(WindowTitle.format(fileURL: nil, isDirty: false), "ARM Emulator")
    }

    func testEditsWithoutAFileAreUntitledAndMarked() {
        XCTAssertEqual(WindowTitle.format(fileURL: nil, isDirty: true), "ARM Emulator — Untitled •")
    }

    func testSavedFileShowsItsNameUnmarked() {
        XCTAssertEqual(WindowTitle.format(fileURL: file, isDirty: false), "ARM Emulator — hello.s")
    }

    func testEditedFileShowsItsNameMarked() {
        XCTAssertEqual(WindowTitle.format(fileURL: file, isDirty: true), "ARM Emulator — hello.s •")
    }
}

@MainActor
final class WindowCloseGuardTests: XCTestCase {
    /// Records which delegate calls reach it
    private final class RecordingDelegate: NSObject, NSWindowDelegate {
        var shouldCloseAnswer = true
        var shouldCloseCalls = 0
        var resizeCalls = 0

        func windowShouldClose(_: NSWindow) -> Bool {
            shouldCloseCalls += 1
            return shouldCloseAnswer
        }

        func windowDidResize(_: Notification) {
            resizeCalls += 1
        }
    }

    private var window: NSWindow!
    private var original: RecordingDelegate!

    override func setUp() async throws {
        window = NSWindow(contentRect: .zero, styleMask: [.titled, .closable], backing: .buffered, defer: true)
        window.isReleasedWhenClosed = false
        original = RecordingDelegate()
        window.delegate = original
    }

    override func tearDown() async throws {
        window = nil
        original = nil
    }

    private func install(isDirty: Bool, confirm: @escaping () async -> Bool) -> WindowCloseGuard {
        let closeGuard = WindowCloseGuard(hasUnsavedChanges: { isDirty }, confirm: confirm)
        closeGuard.install(on: window)
        return closeGuard
    }

    func testInstallingWrapsTheExistingDelegate() {
        let closeGuard = install(isDirty: false) { true }

        XCTAssertTrue(window.delegate === closeGuard)
    }

    func testInstallingTwiceKeepsTheOriginalDelegate() {
        let closeGuard = install(isDirty: false) { true }
        closeGuard.install(on: window)

        XCTAssertTrue(closeGuard.windowShouldClose(window))
        XCTAssertEqual(original.shouldCloseCalls, 1)
    }

    func testCleanWindowDefersToTheOriginalDelegate() {
        let closeGuard = install(isDirty: false) { XCTFail("asked without edits")
            return true
        }
        original.shouldCloseAnswer = false

        XCTAssertFalse(closeGuard.windowShouldClose(window))
        XCTAssertEqual(original.shouldCloseCalls, 1)
    }

    func testDirtyWindowStaysOpenAndAsks() async {
        let asked = expectation(description: "asked")
        let closeGuard = install(isDirty: true) {
            asked.fulfill()
            return false
        }

        XCTAssertFalse(closeGuard.windowShouldClose(window))
        await fulfillment(of: [asked])
        XCTAssertEqual(original.shouldCloseCalls, 0)
    }

    func testDirtyWindowClosesWhenTheUserProceeds() async {
        window.makeKeyAndOrderFront(nil)
        let closeGuard = install(isDirty: true) { true }
        let closed = expectation(forNotification: NSWindow.willCloseNotification, object: window)

        XCTAssertFalse(closeGuard.windowShouldClose(window))

        await fulfillment(of: [closed])
    }

    func testOtherDelegateCallsReachTheOriginalDelegate() {
        _ = install(isDirty: false) { true }

        window.delegate?.windowDidResize?(Notification(name: NSWindow.didResizeNotification, object: window))

        XCTAssertEqual(original.resizeCalls, 1)
    }
}
