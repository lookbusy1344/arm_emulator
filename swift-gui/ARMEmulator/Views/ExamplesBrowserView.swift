import SwiftUI

struct ExamplesBrowserView: View {
    @EnvironmentObject var fileService: FileService
    @State private var examples: [ExampleProgram] = []
    @State private var selectedExample: ExampleProgram?
    @State private var searchText = ""
    @State private var previewContent = ""
    @Environment(\.dismiss) var dismiss

    let onSelect: (ExampleProgram) -> Void

    var filteredExamples: [ExampleProgram] {
        if searchText.isEmpty {
            return examples
        }
        return examples.filter { example in
            example.name.localizedCaseInsensitiveContains(searchText)
                || example.description.localizedCaseInsensitiveContains(searchText)
        }
    }

    var body: some View {
        HSplitView {
            // Left: List of examples
            VStack(spacing: 0) {
                // Search field
                HStack {
                    Image(systemName: "magnifyingglass")
                        .foregroundColor(.secondary)
                    TextField("Search examples...", text: $searchText)
                        .textFieldStyle(.plain)
                    if !searchText.isEmpty {
                        Button(
                            action: { searchText = "" },
                            label: {
                                Image(systemName: "xmark.circle.fill")
                                    .foregroundColor(.secondary)
                            }
                        )
                        .buttonStyle(.plain)
                    }
                }
                .padding(8)
                .background(Color(NSColor.controlBackgroundColor))

                Divider()

                // Examples list
                List(filteredExamples, selection: $selectedExample) { example in
                    ExampleRow(example: example)
                        .tag(example)
                }
                .listStyle(.sidebar)
                .onChange(of: selectedExample) { newValue in
                    if let example = newValue {
                        loadPreview(example)
                    }
                }

                Divider()

                // Bottom info
                HStack {
                    Text("\(filteredExamples.count) example(s)")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                    Spacer()
                }
                .padding(8)
                .background(Color(NSColor.controlBackgroundColor))
            }
            .frame(minWidth: 300, maxWidth: 400)

            // Right: Preview pane
            VStack(alignment: .leading, spacing: 8) {
                if let example = selectedExample {
                    Text(example.name)
                        .font(.system(size: 14, weight: .semibold))

                    Text(example.description)
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                    HStack {
                        Label(example.formattedSize, systemImage: "doc.text")
                        Spacer()
                    }
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)

                    Divider()

                    Text("Preview:")
                        .font(.system(size: 11, weight: .semibold))

                    ScrollView {
                        Text(previewContent)
                            .font(.system(size: 10, design: .monospaced))
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(8)
                            .background(Color(NSColor.textBackgroundColor))
                            .cornerRadius(4)
                    }
                } else {
                    VStack {
                        Spacer()
                        Text("Select an example to preview")
                            .foregroundColor(.secondary)
                        Spacer()
                    }
                    .frame(maxWidth: .infinity)
                }
            }
            .padding()
            .frame(minWidth: 400)
        }
        .frame(width: 900, height: 600)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") {
                    dismiss()
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Open") {
                    if let example = selectedExample {
                        onSelect(example)
                    }
                }
                .disabled(selectedExample == nil)
                .keyboardShortcut(.defaultAction)
            }
        }
        .onAppear {
            loadExamples()
        }
    }

    private func loadExamples() {
        examples = fileService.loadExamples()
        if !examples.isEmpty {
            selectedExample = examples[0]
        }
    }

    private func loadPreview(_ example: ExampleProgram) {
        do {
            let content = try String(contentsOf: example.url, encoding: .utf8)
            let lines = content.components(separatedBy: .newlines)
            previewContent = lines.prefix(15).joined(separator: "\n")
            if lines.count > 15 {
                previewContent += "\n..."
            }
        } catch {
            previewContent = "Error loading preview: \(error.localizedDescription)"
        }
    }
}

struct ExampleRow: View {
    let example: ExampleProgram

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(example.name)
                .font(.system(size: 11))
                .fontWeight(.medium)

            Text(example.description)
                .font(.system(size: 11))
                .foregroundColor(.secondary)
                .lineLimit(2)
        }
        .padding(.vertical, 2)
    }
}
