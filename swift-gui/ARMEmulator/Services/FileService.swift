import AppKit
import Foundation

@MainActor
class FileService: ObservableObject {
    @Published var recentFiles: [URL] = []
    @Published var currentFileURL: URL?

    private let recentFilesKey = "recentFiles"
    private let settings = AppSettings.shared

    static let shared = FileService()

    private init() {
        loadRecentFiles()
    }

    // MARK: - File Operations

    func openFile() async -> String? {
        let panel = NSOpenPanel()
        panel.allowsMultipleSelection = false
        panel.canChooseDirectories = false
        panel.canChooseFiles = true
        panel.allowedContentTypes = [.init(filenameExtension: "s")].compactMap(\.self)
        panel.message = "Select an ARM assembly file"

        let response = await panel.begin()
        guard response == .OK, let url = panel.url else {
            return nil
        }

        do {
            let content = try String(contentsOf: url, encoding: .utf8)
            currentFileURL = url
            addToRecentFiles(url)
            return content
        } catch {
            print("Error reading file: \(error)")
            return nil
        }
    }

    func saveFile(content: String, url: URL? = nil) async -> Bool {
        let targetURL: URL?

        if let url {
            targetURL = url
        } else {
            let panel = NSSavePanel()
            panel.allowedContentTypes = [.init(filenameExtension: "s")].compactMap(\.self)
            panel.nameFieldStringValue = currentFileURL?.lastPathComponent ?? "program.s"
            panel.message = "Save ARM assembly file"

            let response = await panel.begin()
            guard response == .OK else {
                return false
            }
            targetURL = panel.url
        }

        guard let url = targetURL else {
            return false
        }

        do {
            try content.write(to: url, atomically: true, encoding: .utf8)
            currentFileURL = url
            addToRecentFiles(url)
            return true
        } catch {
            print("Error saving file: \(error)")
            return false
        }
    }

    func saveFileAs(content: String) async -> Bool {
        await saveFile(content: content, url: nil)
    }

    // MARK: - Recent Files

    func addToRecentFiles(_ url: URL) {
        // Remove if already exists
        recentFiles.removeAll { $0 == url }

        // Add to front
        recentFiles.insert(url, at: 0)

        // Trim to max
        if recentFiles.count > settings.maxRecentFiles {
            recentFiles = Array(recentFiles.prefix(settings.maxRecentFiles))
        }

        persistRecentFiles()
    }

    func clearRecentFiles() {
        recentFiles.removeAll()
        persistRecentFiles()
    }

    private func persistRecentFiles() {
        let bookmarks = recentFiles.compactMap { url -> Data? in
            try? url.bookmarkData(options: .withSecurityScope, includingResourceValuesForKeys: nil, relativeTo: nil)
        }
        UserDefaults.standard.set(bookmarks, forKey: recentFilesKey)
    }

    private func loadRecentFiles() {
        guard let bookmarks = UserDefaults.standard.array(forKey: recentFilesKey) as? [Data] else {
            return
        }

        recentFiles = bookmarks.compactMap { data -> URL? in
            var isStale = false
            return try? URL(
                resolvingBookmarkData: data,
                options: .withSecurityScope,
                relativeTo: nil,
                bookmarkDataIsStale: &isStale,
            )
        }
    }

    // MARK: - Examples

    func loadExamples() -> [ExampleProgram] {
        // For now, we'll load from the local examples directory
        // In production, this could call the API endpoint
        let fileManager = FileManager.default
        let examplesPath = findExamplesDirectory()

        guard let examplesURL = examplesPath,
              let files = try? fileManager.contentsOfDirectory(
                  at: examplesURL,
                  includingPropertiesForKeys: [.fileSizeKey],
                  options: [.skipsHiddenFiles],
              )
        else {
            return []
        }

        return files
            .filter { $0.pathExtension == "s" }
            .compactMap { url -> ExampleProgram? in
                guard let content = try? String(contentsOf: url, encoding: .utf8),
                      let size = try? url.resourceValues(forKeys: [.fileSizeKey]).fileSize
                else {
                    return nil
                }

                let description = extractDescription(from: content)
                return ExampleProgram(
                    name: url.deletingPathExtension().lastPathComponent,
                    filename: url.lastPathComponent,
                    description: description,
                    size: size,
                    url: url,
                )
            }
            .sorted { $0.name < $1.name }
    }

    private func findExamplesDirectory() -> URL? {
        // Try several possible locations
        let fileManager = FileManager.default
        let currentDir = URL(fileURLWithPath: fileManager.currentDirectoryPath)

        // Try ../examples (when running from Xcode)
        let examplesRelative = currentDir.deletingLastPathComponent().appendingPathComponent("examples")
        if fileManager.fileExists(atPath: examplesRelative.path) {
            return examplesRelative
        }

        // Try ../../examples (when running from DerivedData)
        let examplesUp2 = currentDir.deletingLastPathComponent().deletingLastPathComponent()
            .appendingPathComponent("examples")
        if fileManager.fileExists(atPath: examplesUp2.path) {
            return examplesUp2
        }

        // Try in bundle resources (if embedded)
        if let bundlePath = Bundle.main.resourceURL?.appendingPathComponent("examples"),
           fileManager.fileExists(atPath: bundlePath.path)
        {
            return bundlePath
        }

        return nil
    }

    private func extractDescription(from content: String) -> String {
        // Extract first comment block as description
        let lines = content.components(separatedBy: .newlines)
        var description = ""

        for line in lines {
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            if trimmed.hasPrefix(";") {
                let comment = trimmed.dropFirst().trimmingCharacters(in: .whitespaces)
                if !comment.isEmpty {
                    if !description.isEmpty {
                        description += " "
                    }
                    description += comment
                }
            } else if !trimmed.isEmpty {
                // Stop at first non-comment, non-empty line
                break
            }
        }

        return description.isEmpty ? "No description" : description
    }
}

struct ExampleProgram: Identifiable, Hashable {
    let id = UUID()
    let name: String
    let filename: String
    let description: String
    let size: Int
    let url: URL

    var formattedSize: String {
        ByteCountFormatter.string(fromByteCount: Int64(size), countStyle: .file)
    }

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }

    static func == (lhs: ExampleProgram, rhs: ExampleProgram) -> Bool {
        lhs.id == rhs.id
    }
}
