import XCTest
@testable import ARMEmulator

/// Test helpers for async operations and waiting for conditions
extension XCTestCase {
    /// Waits for the ViewModel to reach a specific status
    /// - Parameters:
    ///   - status: The expected VMState
    ///   - timeout: Maximum time to wait in seconds
    ///   - viewModel: The EmulatorViewModel to monitor
    /// - Throws: TestError.timeout if the status is not reached within timeout
    @MainActor
    func waitForStatus(
        _ status: VMState,
        timeout: TimeInterval,
        viewModel: EmulatorViewModel
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)

        while viewModel.status != status {
            if Date() > deadline {
                throw TestError.timeout("Status did not change to \(status) within \(timeout)s. Current: \(viewModel.status)")
            }
            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
    }

    /// Waits for the backend to reach a specific status
    /// - Parameters:
    ///   - expectedStatus: The expected BackendStatus
    ///   - timeout: Maximum time to wait in seconds
    ///   - manager: The BackendManager to monitor
    /// - Throws: TestError.timeout if the status is not reached within timeout
    @MainActor
    func waitForBackendStatus(
        _ expectedStatus: BackendStatus,
        timeout: TimeInterval,
        manager: BackendManager
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)

        while manager.backendStatus != expectedStatus {
            if Date() > deadline {
                throw TestError.timeout("Backend did not reach \(expectedStatus) within \(timeout)s. Current: \(manager.backendStatus)")
            }
            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
    }

    /// Waits for a condition to become true
    /// - Parameters:
    ///   - timeout: Maximum time to wait in seconds
    ///   - condition: Closure that returns true when condition is met
    /// - Throws: TestError.timeout if condition is not met within timeout
    func waitForCondition(
        timeout: TimeInterval,
        condition: @escaping () -> Bool
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)

        while !condition() {
            if Date() > deadline {
                throw TestError.timeout("Condition not met within \(timeout)s")
            }
            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
    }

    /// Waits for a condition with custom error message
    /// - Parameters:
    ///   - timeout: Maximum time to wait in seconds
    ///   - message: Custom error message if timeout occurs
    ///   - condition: Closure that returns true when condition is met
    /// - Throws: TestError.timeout if condition is not met within timeout
    func waitForCondition(
        timeout: TimeInterval,
        message: String,
        condition: @escaping () -> Bool
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)

        while !condition() {
            if Date() > deadline {
                throw TestError.timeout(message)
            }
            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
    }
}

/// Test-specific errors
enum TestError: Error, LocalizedError {
    case timeout(String)
    case setupFailure(String)
    case assertionFailure(String)

    var errorDescription: String? {
        switch self {
        case .timeout(let message):
            return "Timeout: \(message)"
        case .setupFailure(let message):
            return "Setup failed: \(message)"
        case .assertionFailure(let message):
            return "Assertion failed: \(message)"
        }
    }
}
