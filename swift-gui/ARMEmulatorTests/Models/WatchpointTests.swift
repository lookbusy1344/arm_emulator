import XCTest
@testable import ARMEmulator

final class WatchpointTests: XCTestCase {
    // MARK: - Watchpoint Decoding Tests

    func testWatchpointDecoding() throws {
        let json = """
        {
            "id": 1,
            "address": 32768,
            "type": "write"
        }
        """

        let data = json.data(using: .utf8)!
        let watchpoint = try JSONDecoder().decode(Watchpoint.self, from: data)

        XCTAssertEqual(watchpoint.id, 1)
        XCTAssertEqual(watchpoint.address, 32768)
        XCTAssertEqual(watchpoint.type, "write")
    }

    func testWatchpointReadType() throws {
        let json = """
        {
            "id": 2,
            "address": 327680,
            "type": "read"
        }
        """

        let data = json.data(using: .utf8)!
        let watchpoint = try JSONDecoder().decode(Watchpoint.self, from: data)

        XCTAssertEqual(watchpoint.id, 2)
        XCTAssertEqual(watchpoint.address, 327680)
        XCTAssertEqual(watchpoint.type, "read")
    }

    func testWatchpointReadWriteType() throws {
        let json = """
        {
            "id": 3,
            "address": 40960,
            "type": "readwrite"
        }
        """

        let data = json.data(using: .utf8)!
        let watchpoint = try JSONDecoder().decode(Watchpoint.self, from: data)

        XCTAssertEqual(watchpoint.id, 3)
        XCTAssertEqual(watchpoint.address, 40960)
        XCTAssertEqual(watchpoint.type, "readwrite")
    }

    // MARK: - Watchpoint Encoding Tests

    func testWatchpointEncoding() throws {
        let watchpoint = Watchpoint(id: 1, address: 32768, type: "write")

        let data = try JSONEncoder().encode(watchpoint)
        let decoded = try JSONDecoder().decode(Watchpoint.self, from: data)

        XCTAssertEqual(decoded.id, 1)
        XCTAssertEqual(decoded.address, 32768)
        XCTAssertEqual(decoded.type, "write")
    }

    // MARK: - Watchpoint Identifiable Tests

    func testWatchpointIdentifiable() {
        let watchpoint = Watchpoint(id: 42, address: 32768, type: "write")
        XCTAssertEqual(watchpoint.id, 42)
    }

    // MARK: - Watchpoint Hashable Tests

    func testWatchpointHashable() {
        let watchpoint1 = Watchpoint(id: 1, address: 32768, type: "write")
        let watchpoint2 = Watchpoint(id: 1, address: 32768, type: "write")
        let watchpoint3 = Watchpoint(id: 2, address: 32768, type: "read")

        XCTAssertEqual(watchpoint1, watchpoint2) // Same ID
        XCTAssertNotEqual(watchpoint1, watchpoint3) // Different ID

        var set: Set<Watchpoint> = [watchpoint1]
        set.insert(watchpoint2)
        set.insert(watchpoint3)

        XCTAssertEqual(set.count, 2) // watchpoint1 and watchpoint2 are the same
    }

    func testWatchpointHashFunction() {
        let watchpoint1 = Watchpoint(id: 1, address: 32768, type: "write")
        let watchpoint2 = Watchpoint(id: 1, address: 32768, type: "write")
        let watchpoint3 = Watchpoint(id: 1, address: 40960, type: "read") // Different address, same ID

        var hasher1 = Hasher()
        watchpoint1.hash(into: &hasher1)
        let hash1 = hasher1.finalize()

        var hasher2 = Hasher()
        watchpoint2.hash(into: &hasher2)
        let hash2 = hasher2.finalize()

        var hasher3 = Hasher()
        watchpoint3.hash(into: &hasher3)
        let hash3 = hasher3.finalize()

        XCTAssertEqual(hash1, hash2) // Same ID produces same hash
        XCTAssertEqual(hash1, hash3) // Hash only depends on ID, not address or type
    }

    // MARK: - Watchpoint Equality Tests

    func testWatchpointEqualityBasedOnIDOnly() {
        let watchpoint1 = Watchpoint(id: 1, address: 32768, type: "write")
        let watchpoint2 = Watchpoint(id: 1, address: 40960, type: "read") // Different address and type
        let watchpoint3 = Watchpoint(id: 2, address: 32768, type: "write") // Different ID

        XCTAssertEqual(watchpoint1, watchpoint2) // Equality based on ID only
        XCTAssertNotEqual(watchpoint1, watchpoint3)
    }

    // MARK: - Watchpoint Type Validation Tests

    func testWatchpointAllValidTypes() {
        let validTypes = ["read", "write", "readwrite"]

        for type in validTypes {
            let watchpoint = Watchpoint(id: 1, address: 32768, type: type)
            XCTAssertEqual(watchpoint.type, type)
        }
    }
}
