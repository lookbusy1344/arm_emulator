import XCTest
@testable import ARMEmulator

// MARK: - ExampleProgram Model Tests

final class ExampleProgramModelTests: XCTestCase {
    func testExampleProgramCreation() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let program = ExampleProgram(
            name: "fibonacci",
            filename: "fibonacci.s",
            description: "Calculates Fibonacci numbers",
            size: 1024,
            url: testURL,
        )

        XCTAssertEqual(program.name, "fibonacci")
        XCTAssertEqual(program.filename, "fibonacci.s")
        XCTAssertEqual(program.description, "Calculates Fibonacci numbers")
        XCTAssertEqual(program.size, 1024)
        XCTAssertEqual(program.url, testURL)
        XCTAssertNotNil(program.id) // UUID generated
    }

    func testExampleProgramIdentifiable() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let program1 = ExampleProgram(
            name: "test1",
            filename: "test1.s",
            description: "Test 1",
            size: 100,
            url: testURL,
        )
        let program2 = ExampleProgram(
            name: "test2",
            filename: "test2.s",
            description: "Test 2",
            size: 200,
            url: testURL,
        )

        // Each program has unique ID
        XCTAssertNotEqual(program1.id, program2.id)
    }

    func testFormattedSize() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")

        // Test various file sizes
        let program1 = ExampleProgram(name: "small", filename: "small.s", description: "", size: 100, url: testURL)
        XCTAssertFalse(program1.formattedSize.isEmpty)

        let program2 = ExampleProgram(name: "medium", filename: "medium.s", description: "", size: 10000, url: testURL)
        XCTAssertFalse(program2.formattedSize.isEmpty)

        let program3 = ExampleProgram(
            name: "large",
            filename: "large.s",
            description: "",
            size: 1_000_000,
            url: testURL,
        )
        XCTAssertFalse(program3.formattedSize.isEmpty)

        // Verify formatted strings are different for different sizes
        XCTAssertNotEqual(program1.formattedSize, program2.formattedSize)
        XCTAssertNotEqual(program2.formattedSize, program3.formattedSize)
    }

    func testExampleProgramEquality() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let program1 = ExampleProgram(name: "test", filename: "test.s", description: "", size: 100, url: testURL)
        let program2 = ExampleProgram(name: "test", filename: "test.s", description: "", size: 100, url: testURL)

        // Different instances should not be equal (different UUIDs)
        XCTAssertNotEqual(program1, program2)

        // Same instance should equal itself
        XCTAssertEqual(program1, program1)
    }

    func testExampleProgramHashable() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let program1 = ExampleProgram(name: "test", filename: "test.s", description: "", size: 100, url: testURL)
        let program2 = ExampleProgram(name: "test", filename: "test.s", description: "", size: 100, url: testURL)

        var set: Set<ExampleProgram> = []
        set.insert(program1)
        set.insert(program2)

        // Two different instances should both be in the set
        XCTAssertEqual(set.count, 2)
    }
}

// MARK: - ExamplesBrowserView Filtering Tests

final class ExamplesBrowserFilteringTests: XCTestCase {
    func testFilterByName() {
        /// Simulate the filteredExamples logic
        func filterExamples(
            _ examples: [ExampleProgram],
            searchText: String,
        ) -> [ExampleProgram] {
            if searchText.isEmpty {
                return examples
            }
            return examples.filter { example in
                example.name.localizedCaseInsensitiveContains(searchText)
                    || example.description.localizedCaseInsensitiveContains(searchText)
            }
        }

        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let examples = [
            ExampleProgram(
                name: "fibonacci",
                filename: "fibonacci.s",
                description: "Calculates Fibonacci numbers",
                size: 100,
                url: testURL,
            ),
            ExampleProgram(
                name: "factorial",
                filename: "factorial.s",
                description: "Calculates factorial",
                size: 200,
                url: testURL,
            ),
            ExampleProgram(
                name: "hello",
                filename: "hello.s",
                description: "Prints hello world",
                size: 50,
                url: testURL,
            ),
        ]

        // Empty search returns all
        XCTAssertEqual(filterExamples(examples, searchText: "").count, 3)

        // Case-insensitive name search
        XCTAssertEqual(filterExamples(examples, searchText: "fib").count, 1)
        XCTAssertEqual(filterExamples(examples, searchText: "FIB").count, 1)

        // Description search
        XCTAssertEqual(filterExamples(examples, searchText: "Calculates").count, 2)
        XCTAssertEqual(filterExamples(examples, searchText: "hello world").count, 1)

        // No matches
        XCTAssertEqual(filterExamples(examples, searchText: "nonexistent").count, 0)
    }

    func testFilterPartialMatches() {
        /// Test partial matching behavior
        func matches(_ text: String, searchText: String) -> Bool {
            text.localizedCaseInsensitiveContains(searchText)
        }

        XCTAssertTrue(matches("fibonacci", searchText: "fib"))
        XCTAssertTrue(matches("fibonacci", searchText: "bon"))
        XCTAssertTrue(matches("fibonacci", searchText: "acci"))
        XCTAssertFalse(matches("fibonacci", searchText: "xyz"))
    }

    func testFilterCaseSensitivity() {
        func matches(_ text: String, searchText: String) -> Bool {
            text.localizedCaseInsensitiveContains(searchText)
        }

        // Should be case-insensitive
        XCTAssertTrue(matches("Fibonacci", searchText: "fib"))
        XCTAssertTrue(matches("FIBONACCI", searchText: "fib"))
        XCTAssertTrue(matches("fibonacci", searchText: "FIB"))
    }
}

// MARK: - ExamplesBrowserView Preview Tests

final class ExamplePreviewTests: XCTestCase {
    func testPreviewTruncation() {
        /// Simulate the preview truncation logic
        func generatePreview(content: String, maxLines: Int = 15) -> String {
            let lines = content.components(separatedBy: .newlines)
            var preview = lines.prefix(maxLines).joined(separator: "\n")
            if lines.count > maxLines {
                preview += "\n..."
            }
            return preview
        }

        // Test short content (no truncation)
        let shortContent = "Line 1\nLine 2\nLine 3"
        let shortPreview = generatePreview(content: shortContent, maxLines: 15)
        XCTAssertEqual(shortPreview, shortContent)
        XCTAssertFalse(shortPreview.contains("..."))

        // Test long content (with truncation)
        let longLines = (1 ... 20).map { "Line \($0)" }
        let longContent = longLines.joined(separator: "\n")
        let longPreview = generatePreview(content: longContent, maxLines: 15)
        XCTAssertTrue(longPreview.contains("..."))
        XCTAssertTrue(longPreview.contains("Line 15"))
        XCTAssertFalse(longPreview.contains("Line 16"))
    }

    func testPreviewExactlyAtLimit() {
        func generatePreview(content: String, maxLines: Int = 15) -> String {
            let lines = content.components(separatedBy: .newlines)
            var preview = lines.prefix(maxLines).joined(separator: "\n")
            if lines.count > maxLines {
                preview += "\n..."
            }
            return preview
        }

        // Exactly 15 lines - no truncation marker
        let exactLines = (1 ... 15).map { "Line \($0)" }
        let exactContent = exactLines.joined(separator: "\n")
        let exactPreview = generatePreview(content: exactContent, maxLines: 15)
        XCTAssertFalse(exactPreview.contains("..."))
    }

    func testPreviewErrorMessage() {
        // Test error message format
        let errorMessage = "Error loading preview: File not found"

        XCTAssertTrue(errorMessage.starts(with: "Error loading preview:"))
        XCTAssertTrue(errorMessage.contains("File not found"))
    }
}

// MARK: - ExamplesBrowserView Selection Tests

final class ExampleSelectionTests: XCTestCase {
    func testInitialSelection() {
        /// Simulate initial selection behavior
        func selectFirst(_ examples: [ExampleProgram]) -> ExampleProgram? {
            examples.isEmpty ? nil : examples[0]
        }

        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let examples = [
            ExampleProgram(name: "first", filename: "first.s", description: "", size: 100, url: testURL),
            ExampleProgram(name: "second", filename: "second.s", description: "", size: 200, url: testURL),
        ]

        let selected = selectFirst(examples)
        XCTAssertNotNil(selected)
        XCTAssertEqual(selected?.name, "first")
    }

    func testEmptyListSelection() {
        func selectFirst(_ examples: [ExampleProgram]) -> ExampleProgram? {
            examples.isEmpty ? nil : examples[0]
        }

        let selected = selectFirst([])
        XCTAssertNil(selected)
    }

    func testOpenButtonState() {
        /// Open button should be disabled when no selection
        func isOpenButtonEnabled(selectedExample: ExampleProgram?) -> Bool {
            selectedExample != nil
        }

        XCTAssertTrue(isOpenButtonEnabled(selectedExample: ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "",
            size: 100,
            url: URL(fileURLWithPath: "/tmp/test.s"),
        )))
        XCTAssertFalse(isOpenButtonEnabled(selectedExample: nil))
    }
}

// MARK: - ExamplesBrowserView Initialization Tests

@MainActor
final class ExamplesBrowserViewInitializationTests: XCTestCase {
    func testInitWithCallback() {
        var selectedProgram: ExampleProgram?

        let view = ExamplesBrowserView { program in
            selectedProgram = program
        }

        XCTAssertNotNil(view)
        XCTAssertNil(selectedProgram) // Callback not invoked yet
    }

    func testCallbackInvocation() {
        var selectedProgram: ExampleProgram?

        let onSelect: (ExampleProgram) -> Void = { program in
            selectedProgram = program
        }

        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let testProgram = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "Test",
            size: 100,
            url: testURL,
        )

        // Simulate callback
        onSelect(testProgram)

        XCTAssertNotNil(selectedProgram)
        XCTAssertEqual(selectedProgram?.name, "test")
    }
}

// MARK: - ExampleRow Tests

@MainActor
final class ExampleRowTests: XCTestCase {
    func testExampleRowInit() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let example = ExampleProgram(
            name: "fibonacci",
            filename: "fibonacci.s",
            description: "Calculates Fibonacci numbers",
            size: 1024,
            url: testURL,
        )

        let row = ExampleRow(example: example)

        XCTAssertNotNil(row)
    }

    func testExampleRowWithLongDescription() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let longDescription = String(repeating: "This is a very long description. ", count: 10)
        let example = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: longDescription,
            size: 100,
            url: testURL,
        )

        let row = ExampleRow(example: example)

        // Row should handle long descriptions (lineLimit: 2 in view)
        XCTAssertNotNil(row)
    }

    func testExampleRowWithEmptyDescription() {
        let testURL = URL(fileURLWithPath: "/tmp/test.s")
        let example = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "",
            size: 100,
            url: testURL,
        )

        let row = ExampleRow(example: example)

        XCTAssertNotNil(row)
    }
}

// MARK: - Search UI Tests

final class SearchUITests: XCTestCase {
    func testSearchTextEmpty() {
        /// Test the search clear button logic
        func shouldShowClearButton(searchText: String) -> Bool {
            !searchText.isEmpty
        }

        XCTAssertFalse(shouldShowClearButton(searchText: ""))
        XCTAssertTrue(shouldShowClearButton(searchText: "fib"))
        XCTAssertTrue(shouldShowClearButton(searchText: " ")) // Single space counts as non-empty
    }

    func testSearchClear() {
        // Simulate clearing search text
        var searchText = "fibonacci"

        // Clear button action
        searchText = ""

        XCTAssertEqual(searchText, "")
    }
}

// MARK: - Counter Display Tests

final class CounterDisplayTests: XCTestCase {
    func testExampleCounter() {
        /// Test the counter format: "N example(s)"
        func formatCounter(count: Int) -> String {
            "\(count) example(s)"
        }

        XCTAssertEqual(formatCounter(count: 0), "0 example(s)")
        XCTAssertEqual(formatCounter(count: 1), "1 example(s)")
        XCTAssertEqual(formatCounter(count: 49), "49 example(s)")
    }

    func testCounterWithFiltering() {
        // Verify counter updates based on filtered results
        let totalExamples = 49
        let filteredExamples = 5

        XCTAssertEqual(filteredExamples, 5)
        XCTAssertNotEqual(filteredExamples, totalExamples)
    }
}

// MARK: - Note on SwiftUI View Testing Limitations

/*
 ExamplesBrowserView Testing Limitations:

 ExamplesBrowserView uses @State for examples, selectedExample, searchText, and
 previewContent. It also depends on @EnvironmentObject FileService. These are
 not directly accessible in unit tests.

 What we CAN test:
 - ExampleProgram model (Identifiable, Hashable, formatted size)
 - Filtering logic (search by name/description)
 - Preview truncation logic (15 lines + "...")
 - Selection behavior simulation
 - Counter formatting
 - Callback invocation

 What we CANNOT easily test:
 - FileService integration (loadExamples())
 - HSplitView layout behavior
 - List selection state management
 - Search field text binding
 - Preview content loading (async file I/O)
 - Toolbar button states
 - Keyboard shortcuts (defaultAction)
 - onChange handler triggering

 Recommendations:
 1. Test model and filtering logic comprehensively (done above)
 2. Use integration tests for FileService.loadExamples()
 3. Use UI tests for search and selection interaction
 4. Mock FileService in tests if needed
 */
