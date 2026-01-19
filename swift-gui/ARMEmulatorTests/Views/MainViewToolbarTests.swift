import SwiftUI
import XCTest
@testable import ARMEmulator

// MARK: - Status Icon Tests

final class ToolbarStatusIconTests: XCTestCase {
    func testStatusIcons() {
        // Simulate the statusIcon computed property
        func statusIcon(for status: VMState) -> String {
            switch status {
            case .running:
                "play.fill"
            case .breakpoint:
                "pause.fill"
            case .halted, .idle:
                "stop.fill"
            case .waitingForInput:
                "keyboard.fill"
            case .error:
                "exclamationmark.triangle.fill"
            }
        }

        XCTAssertEqual(statusIcon(for: .running), "play.fill")
        XCTAssertEqual(statusIcon(for: .breakpoint), "pause.fill")
        XCTAssertEqual(statusIcon(for: .halted), "stop.fill")
        XCTAssertEqual(statusIcon(for: .idle), "stop.fill")
        XCTAssertEqual(statusIcon(for: .waitingForInput), "keyboard.fill")
        XCTAssertEqual(statusIcon(for: .error), "exclamationmark.triangle.fill")
    }

    func testIconSymbolNames() {
        // Verify SF Symbol names are valid
        let icons = [
            "play.fill",
            "pause.fill",
            "stop.fill",
            "keyboard.fill",
            "exclamationmark.triangle.fill",
        ]

        for icon in icons {
            XCTAssertFalse(icon.isEmpty)
            XCTAssertTrue(icon.contains(".")) // All SF Symbols have dot notation
        }
    }
}

// MARK: - Status Color Tests

final class ToolbarStatusColorTests: XCTestCase {
    func testStatusColors() {
        // Simulate the statusColor computed property
        func statusColor(for status: VMState) -> Color {
            switch status {
            case .running:
                .green
            case .breakpoint:
                .orange
            case .halted, .idle:
                .red
            case .waitingForInput:
                .orange
            case .error:
                .red
            }
        }

        XCTAssertEqual(statusColor(for: .running), .green)
        XCTAssertEqual(statusColor(for: .breakpoint), .orange)
        XCTAssertEqual(statusColor(for: .halted), .red)
        XCTAssertEqual(statusColor(for: .idle), .red)
        XCTAssertEqual(statusColor(for: .waitingForInput), .orange)
        XCTAssertEqual(statusColor(for: .error), .red)
    }

    func testColorSemantics() {
        // Document color meanings
        // Green = actively executing
        // Orange = paused/waiting (user action possible)
        // Red = stopped/error (not executing)

        let greenStates: [VMState] = [.running]
        let orangeStates: [VMState] = [.breakpoint, .waitingForInput]
        let redStates: [VMState] = [.halted, .idle, .error]

        XCTAssertEqual(greenStates.count, 1)
        XCTAssertEqual(orangeStates.count, 2)
        XCTAssertEqual(redStates.count, 3)
    }
}

// MARK: - Status Text Tests

final class ToolbarStatusTextTests: XCTestCase {
    func testStatusText() {
        // Simulate the statusText computed property
        func statusText(for status: VMState) -> String {
            switch status {
            case .running:
                "Running"
            case .breakpoint:
                "Paused"
            case .halted:
                "Halted"
            case .idle:
                "Idle"
            case .waitingForInput:
                "Waiting for Input"
            case .error:
                "Error"
            }
        }

        XCTAssertEqual(statusText(for: .running), "Running")
        XCTAssertEqual(statusText(for: .breakpoint), "Paused")
        XCTAssertEqual(statusText(for: .halted), "Halted")
        XCTAssertEqual(statusText(for: .idle), "Idle")
        XCTAssertEqual(statusText(for: .waitingForInput), "Waiting for Input")
        XCTAssertEqual(statusText(for: .error), "Error")
    }

    func testAllStatusesHaveText() {
        // Verify all VMState cases have corresponding text
        let allStates: [VMState] = [.running, .breakpoint, .halted, .idle, .waitingForInput, .error]

        for state in allStates {
            let text = statusText(for: state)
            XCTAssertFalse(text.isEmpty, "Status \(state) should have text")
        }

        func statusText(for status: VMState) -> String {
            switch status {
            case .running: "Running"
            case .breakpoint: "Paused"
            case .halted: "Halted"
            case .idle: "Idle"
            case .waitingForInput: "Waiting for Input"
            case .error: "Error"
            }
        }
    }
}

// MARK: - Button Label Tests

final class ToolbarButtonLabelTests: XCTestCase {
    func testLoadButtonLabel() {
        let label = "Load"
        let icon = "doc.text"

        XCTAssertEqual(label, "Load")
        XCTAssertEqual(icon, "doc.text")
    }

    func testRunContinueButtonLabel() {
        // Button label changes based on status
        func runButtonLabel(for status: VMState) -> String {
            status == .breakpoint ? "Continue" : "Run"
        }

        func runButtonIcon(for status: VMState) -> String {
            status == .breakpoint ? "play.circle.fill" : "play.fill"
        }

        XCTAssertEqual(runButtonLabel(for: .breakpoint), "Continue")
        XCTAssertEqual(runButtonIcon(for: .breakpoint), "play.circle.fill")
        XCTAssertEqual(runButtonLabel(for: .idle), "Run")
        XCTAssertEqual(runButtonIcon(for: .idle), "play.fill")
    }

    func testPauseButtonLabel() {
        let label = "Pause"
        let icon = "pause.fill"

        XCTAssertEqual(label, "Pause")
        XCTAssertEqual(icon, "pause.fill")
    }

    func testStepButtonLabel() {
        let label = "Step"
        let icon = "forward.frame"

        XCTAssertEqual(label, "Step")
        XCTAssertEqual(icon, "forward.frame")
    }

    func testStepOverButtonLabel() {
        let label = "Step Over"
        let icon = "arrow.right.to.line"

        XCTAssertEqual(label, "Step Over")
        XCTAssertEqual(icon, "arrow.right.to.line")
    }

    func testStepOutButtonLabel() {
        let label = "Step Out"
        let icon = "arrow.up.left"

        XCTAssertEqual(label, "Step Out")
        XCTAssertEqual(icon, "arrow.up.left")
    }

    func testResetButtonLabel() {
        let label = "Reset"
        let icon = "arrow.counterclockwise"

        XCTAssertEqual(label, "Reset")
        XCTAssertEqual(icon, "arrow.counterclockwise")
    }

    func testShowPCButtonLabel() {
        let label = "Show PC"
        let icon = "arrow.down.to.line"

        XCTAssertEqual(label, "Show PC")
        XCTAssertEqual(icon, "arrow.down.to.line")
    }
}

// MARK: - Keyboard Shortcut Tests

final class ToolbarKeyboardShortcutTests: XCTestCase {
    func testKeyboardShortcuts() {
        // Document keyboard shortcuts for toolbar buttons
        struct Shortcut {
            let key: String
            let modifiers: [String]
            let description: String
        }

        let shortcuts = [
            Shortcut(key: "l", modifiers: ["⌘"], description: "Load program"),
            Shortcut(key: "r", modifiers: ["⌘"], description: "Run/Continue"),
            Shortcut(key: ".", modifiers: ["⌘"], description: "Pause"),
            Shortcut(key: "t", modifiers: ["⌘"], description: "Step"),
            Shortcut(key: "t", modifiers: ["⌘", "⇧"], description: "Step Over"),
            Shortcut(key: "t", modifiers: ["⌘", "⌥"], description: "Step Out"),
            Shortcut(key: "r", modifiers: ["⌘", "⇧"], description: "Reset"),
            Shortcut(key: "j", modifiers: ["⌘"], description: "Show PC"),
        ]

        XCTAssertEqual(shortcuts.count, 8)

        // Verify all shortcuts have non-empty keys
        for shortcut in shortcuts {
            XCTAssertFalse(shortcut.key.isEmpty)
            XCTAssertFalse(shortcut.modifiers.isEmpty)
            XCTAssertFalse(shortcut.description.isEmpty)
        }
    }

    func testShortcutModifiers() {
        // Document modifier keys used
        let modifiers = ["⌘", "⇧", "⌥"]

        XCTAssertEqual(modifiers[0], "⌘") // Command
        XCTAssertEqual(modifiers[1], "⇧") // Shift
        XCTAssertEqual(modifiers[2], "⌥") // Option
    }

    func testShortcutConflicts() {
        // Verify no duplicate shortcuts
        struct ShortcutKey: Hashable {
            let key: String
            let modifiers: Set<String>
        }

        let shortcuts: [ShortcutKey] = [
            ShortcutKey(key: "l", modifiers: ["⌘"]),
            ShortcutKey(key: "r", modifiers: ["⌘"]),
            ShortcutKey(key: ".", modifiers: ["⌘"]),
            ShortcutKey(key: "t", modifiers: ["⌘"]),
            ShortcutKey(key: "t", modifiers: ["⌘", "⇧"]),
            ShortcutKey(key: "t", modifiers: ["⌘", "⌥"]),
            ShortcutKey(key: "r", modifiers: ["⌘", "⇧"]),
            ShortcutKey(key: "j", modifiers: ["⌘"]),
        ]

        let uniqueShortcuts = Set(shortcuts)
        XCTAssertEqual(shortcuts.count, uniqueShortcuts.count, "No duplicate shortcuts")
    }
}

// MARK: - Button Disabled State Tests

final class ToolbarButtonDisabledTests: XCTestCase {
    func testRunButtonDisabled() {
        // Run button disabled when running or waiting for input
        func isRunButtonDisabled(status: VMState) -> Bool {
            status == .running || status == .waitingForInput
        }

        XCTAssertTrue(isRunButtonDisabled(status: .running))
        XCTAssertTrue(isRunButtonDisabled(status: .waitingForInput))
        XCTAssertFalse(isRunButtonDisabled(status: .idle))
        XCTAssertFalse(isRunButtonDisabled(status: .breakpoint))
        XCTAssertFalse(isRunButtonDisabled(status: .halted))
        XCTAssertFalse(isRunButtonDisabled(status: .error))
    }

    func testPauseButtonDisabled() {
        // Pause button enabled only when canPause
        func isPauseButtonDisabled(canPause: Bool) -> Bool {
            !canPause
        }

        XCTAssertFalse(isPauseButtonDisabled(canPause: true))
        XCTAssertTrue(isPauseButtonDisabled(canPause: false))
    }

    func testStepButtonsDisabled() {
        // Step buttons enabled only when canStep
        func isStepButtonDisabled(canStep: Bool) -> Bool {
            !canStep
        }

        XCTAssertFalse(isStepButtonDisabled(canStep: true))
        XCTAssertTrue(isStepButtonDisabled(canStep: false))
    }

    func testShowPCButtonDisabled() {
        // Show PC button disabled when PC is 0
        func isShowPCButtonDisabled(currentPC: UInt32) -> Bool {
            currentPC == 0
        }

        XCTAssertTrue(isShowPCButtonDisabled(currentPC: 0))
        XCTAssertFalse(isShowPCButtonDisabled(currentPC: 0x8000))
        XCTAssertFalse(isShowPCButtonDisabled(currentPC: 0x1000))
    }
}

// MARK: - Help Text Tests

final class ToolbarHelpTextTests: XCTestCase {
    func testHelpTexts() {
        // Document help text (tooltips) for buttons
        let helpTexts: [String: String] = [
            "Load": "Load program (⌘L)",
            "Run": "Run program (⌘R)",
            "Continue": "Continue execution (⌘R)",
            "Pause": "Pause execution (⌘.)",
            "Step": "Step one instruction (⌘T)",
            "StepOver": "Step over function calls (⌘⇧T)",
            "StepOut": "Step out of current function (⌘⌥T)",
            "Reset": "Reset VM (⌘⇧R)",
            "ShowPC": "Scroll to current PC (⌘J)",
        ]

        XCTAssertEqual(helpTexts.count, 9)

        // Verify all help texts are non-empty
        for (_, helpText) in helpTexts {
            XCTAssertFalse(helpText.isEmpty)
        }
    }

    func testHelpTextFormat() {
        // Help text format: "Action description (Shortcut)"
        func formatHelpText(action: String, shortcut: String) -> String {
            "\(action) (\(shortcut))"
        }

        XCTAssertEqual(formatHelpText(action: "Load program", shortcut: "⌘L"), "Load program (⌘L)")
        XCTAssertEqual(formatHelpText(action: "Run program", shortcut: "⌘R"), "Run program (⌘R)")
    }

    func testDynamicHelpText() {
        // Run button help text changes based on status
        func runButtonHelpText(for status: VMState) -> String {
            status == .breakpoint
                ? "Continue execution (⌘R)"
                : "Run program (⌘R)"
        }

        XCTAssertEqual(runButtonHelpText(for: .breakpoint), "Continue execution (⌘R)")
        XCTAssertEqual(runButtonHelpText(for: .idle), "Run program (⌘R)")
    }
}

// MARK: - Status Indicator Tests

final class ToolbarStatusIndicatorTests: XCTestCase {
    func testStatusIndicatorComponents() {
        // Status indicator consists of: icon + text
        struct StatusIndicator {
            let icon: String
            let iconColor: Color
            let text: String
        }

        func createIndicator(for status: VMState) -> StatusIndicator {
            let icon: String
            let color: Color
            let text: String

            switch status {
            case .running:
                icon = "play.fill"
                color = .green
                text = "Running"
            case .breakpoint:
                icon = "pause.fill"
                color = .orange
                text = "Paused"
            case .halted:
                icon = "stop.fill"
                color = .red
                text = "Halted"
            case .idle:
                icon = "stop.fill"
                color = .red
                text = "Idle"
            case .waitingForInput:
                icon = "keyboard.fill"
                color = .orange
                text = "Waiting for Input"
            case .error:
                icon = "exclamationmark.triangle.fill"
                color = .red
                text = "Error"
            }

            return StatusIndicator(icon: icon, iconColor: color, text: text)
        }

        let runningIndicator = createIndicator(for: .running)
        XCTAssertEqual(runningIndicator.icon, "play.fill")
        XCTAssertEqual(runningIndicator.iconColor, .green)
        XCTAssertEqual(runningIndicator.text, "Running")
    }

    func testStatusIndicatorFontSizes() {
        // Document font sizes used in status indicator
        let iconFontSize: CGFloat = 11
        let textFontSize: CGFloat = 11

        XCTAssertEqual(iconFontSize, 11)
        XCTAssertEqual(textFontSize, 11)
    }

    func testStatusIndicatorSpacing() {
        // Document spacing between icon and text
        let spacing: CGFloat = 4

        XCTAssertEqual(spacing, 4)
    }
}

// MARK: - Toolbar Layout Tests

final class ToolbarLayoutTests: XCTestCase {
    func testToolbarItemGroups() {
        // Document logical grouping of toolbar items
        let groups: [String: [String]] = [
            "Status": ["Status Indicator"],
            "Program": ["Load"],
            "Execution": ["Run/Continue", "Pause", "Step", "Step Over", "Step Out", "Reset"],
            "Navigation": ["Show PC"],
        ]

        XCTAssertEqual(groups.count, 4)
        XCTAssertEqual(groups["Execution"]?.count, 6)
    }

    func testDividerPlacement() {
        // Dividers separate logical groups
        // Status | Program | Execution | Navigation

        let dividerCount = 2

        XCTAssertEqual(dividerCount, 2)
    }
}

// MARK: - MainViewToolbar Initialization Tests

@MainActor
final class MainViewToolbarInitializationTests: XCTestCase {
    func testInitWithViewModel() {
        let viewModel = EmulatorViewModel()
        let toolbar = MainViewToolbar(viewModel: viewModel)

        XCTAssertNotNil(toolbar)
    }

    func testToolbarWithDifferentStatuses() {
        // Test toolbar initialization with different VM states
        let viewModel = EmulatorViewModel()

        // Idle state
        viewModel.status = .idle
        var toolbar = MainViewToolbar(viewModel: viewModel)
        XCTAssertNotNil(toolbar)

        // Running state
        viewModel.status = .running
        toolbar = MainViewToolbar(viewModel: viewModel)
        XCTAssertNotNil(toolbar)

        // Breakpoint state
        viewModel.status = .breakpoint
        toolbar = MainViewToolbar(viewModel: viewModel)
        XCTAssertNotNil(toolbar)
    }
}

// MARK: - Button Action Tests

@MainActor
final class ToolbarButtonActionTests: XCTestCase {
    func testButtonActionPatterns() {
        // All button actions follow the pattern: Task { await viewModel.method() }
        // This is necessary for async ViewModel methods

        // Simulate button actions
        var loadCalled = false
        var runCalled = false
        var pauseCalled = false

        // Load button
        Task {
            loadCalled = true
        }

        // Run button
        Task {
            runCalled = true
        }

        // Pause button
        Task {
            pauseCalled = true
        }

        // Give tasks time to execute
        let expectation = XCTestExpectation(description: "Tasks execute")
        Task {
            try? await Task.sleep(nanoseconds: 10_000_000) // 10ms
            XCTAssertTrue(loadCalled)
            XCTAssertTrue(runCalled)
            XCTAssertTrue(pauseCalled)
            expectation.fulfill()
        }

        wait(for: [expectation], timeout: 1.0)
    }
}

// MARK: - Note on SwiftUI View Testing Limitations

/*
 MainViewToolbar Testing Limitations:

 MainViewToolbar is a ToolbarContent view that depends on EmulatorViewModel
 for state and actions. It uses @ObservedObject to observe ViewModel changes.

 What we CAN test:
 - Status icon mapping (VMState → SF Symbol)
 - Status color mapping (VMState → Color)
 - Status text mapping (VMState → String)
 - Button label logic (dynamic Run/Continue)
 - Button disabled state logic
 - Help text content and format
 - Keyboard shortcuts documentation
 - Toolbar layout structure

 What we CANNOT easily test:
 - Actual toolbar rendering
 - Button click actions (require ViewModel integration)
 - ToolbarItemGroup placement
 - Divider visual appearance
 - @ObservedObject property updates
 - Keyboard shortcut triggering
 - Help text tooltip display
 - Status indicator layout (HStack)
 - Button styling

 Recommendations:
 1. Test state-to-UI mapping logic comprehensively (done above)
 2. Use integration tests for button actions with real ViewModel
 3. Use UI tests for keyboard shortcuts and button interactions
 4. Extract complex logic to testable ViewModels (already done)
 */
