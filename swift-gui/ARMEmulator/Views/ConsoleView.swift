import SwiftUI

struct ConsoleView: View {
    let output: String
    @State private var inputText = ""
    var onSendInput: ((String) -> Void)?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Console Output")
                .font(.system(size: 11, weight: .semibold))
                .padding(.horizontal)
                .padding(.vertical, 8)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(NSColor.controlBackgroundColor))

            ScrollView {
                ScrollViewReader { proxy in
                    Text(output.isEmpty ? "Program output will appear here..." : output)
                        .font(.system(size: 10, design: .monospaced))
                        .padding(8)
                        .frame(maxWidth: .infinity, alignment: .topLeading)
                        .id("bottom")
                        .onChange(of: output) { _ in
                            proxy.scrollTo("bottom", anchor: .bottom)
                        }
                }
            }
            .background(Color(NSColor.textBackgroundColor))

            if onSendInput != nil {
                Divider()

                HStack {
                    TextField("Input...", text: $inputText)
                        .textFieldStyle(.roundedBorder)
                        .font(.system(size: 10, design: .monospaced))
                        .onSubmit {
                            sendInput()
                        }

                    Button("Send") {
                        sendInput()
                    }
                    .keyboardShortcut(.return, modifiers: [])
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 12)
                .background(Color(NSColor.controlBackgroundColor))
            }
        }
    }

    private func sendInput() {
        guard !inputText.isEmpty else { return }
        onSendInput?(inputText + "\n")
        inputText = ""
    }
}

struct ConsoleView_Previews: PreviewProvider {
    static var previews: some View {
        ConsoleView(
            output: "Hello, World!\nProgram exited with code 0\n",
            onSendInput: { input in
                print("Input: \(input)")
            }
        )
        .frame(width: 600, height: 200)
    }
}
