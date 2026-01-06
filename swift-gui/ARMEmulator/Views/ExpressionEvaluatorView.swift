import SwiftUI

struct EvaluationResult: Identifiable {
    let id = UUID()
    let expression: String
    let result: UInt32
    let timestamp: Date
}

struct ExpressionEvaluatorView: View {
    @ObservedObject var viewModel: EmulatorViewModel
    @State private var expression = ""
    @State private var history: [EvaluationResult] = []
    @State private var isEvaluating = false
    @State private var errorMessage: String?

    var body: some View {
        VStack(spacing: 0) {
            // Input field
            HStack {
                TextField("Enter expression (e.g., r0, r0+r1, [r0], 0x8000)", text: $expression)
                    .textFieldStyle(.roundedBorder)
                    .font(.system(.body, design: .monospaced))
                    .onSubmit {
                        Task { await evaluateExpression() }
                    }

                Button(action: { Task { await evaluateExpression() } }) {
                    Label("Evaluate", systemImage: "equal.circle.fill")
                }
                .disabled(expression.isEmpty || isEvaluating)
                .keyboardShortcut(.return, modifiers: [])
            }
            .padding()

            Divider()

            // Error message
            if let error = errorMessage {
                HStack {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundColor(.orange)
                    Text(error)
                        .font(.system(.caption, design: .monospaced))
                        .foregroundColor(.secondary)
                    Spacer()
                    Button("Dismiss") {
                        errorMessage = nil
                    }
                    .buttonStyle(.plain)
                    .font(.caption)
                }
                .padding(.horizontal)
                .padding(.vertical, 8)
                .background(Color.orange.opacity(0.1))
            }

            // History list
            if history.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "function")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary)
                    Text("No expressions evaluated yet")
                        .foregroundColor(.secondary)
                    Text("Try: r0, r0+r1, [r0], 0x8000")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 8) {
                        ForEach(history.reversed()) { result in
                            resultRow(result)
                        }
                    }
                    .padding()
                }
            }
        }
        .navigationTitle("Expression Evaluator")
    }

    @ViewBuilder
    private func resultRow(_ result: EvaluationResult) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(result.expression)
                    .font(.system(.body, design: .monospaced))
                    .fontWeight(.medium)
                Spacer()
                Text(result.timestamp, style: .time)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }

            HStack(spacing: 16) {
                resultValue(label: "Hex", value: String(format: "0x%08X", result.result))
                resultValue(label: "Dec", value: String(result.result))
                resultValue(
                    label: "Bin",
                    value: String(result.result, radix: 2).padding(toLength: 32, withPad: "0", startingAt: 0)
                )
            }
        }
        .padding()
        .background(Color.secondary.opacity(0.1))
        .cornerRadius(8)
    }

    @ViewBuilder
    private func resultValue(label: String, value: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(label)
                .font(.caption2)
                .foregroundColor(.secondary)
            Text(value)
                .font(.system(.caption, design: .monospaced))
                .textSelection(.enabled)
        }
    }

    private func evaluateExpression() async {
        guard !expression.isEmpty else { return }
        guard let sessionID = viewModel.sessionID else {
            errorMessage = "No active session"
            return
        }

        isEvaluating = true
        errorMessage = nil

        do {
            let result = try await viewModel.apiClient.evaluateExpression(
                sessionID: sessionID,
                expression: expression
            )

            let evaluation = EvaluationResult(
                expression: expression,
                result: result,
                timestamp: Date()
            )
            history.append(evaluation)

            // Clear input for next expression
            expression = ""
        } catch {
            errorMessage = "Evaluation failed: \(error.localizedDescription)"
        }

        isEvaluating = false
    }
}
