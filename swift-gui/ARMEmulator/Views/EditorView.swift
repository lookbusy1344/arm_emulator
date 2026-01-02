import SwiftUI

struct EditorView: View {
    @Binding var text: String
    @State private var isEditing = false

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Assembly Editor")
                .font(.headline)
                .padding(.horizontal)
                .padding(.vertical, 8)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(NSColor.controlBackgroundColor))

            TextEditor(text: $text)
                .font(.system(.body, design: .monospaced))
                .padding(4)
                .background(Color(NSColor.textBackgroundColor))
                .border(Color.gray.opacity(0.3))
        }
    }
}

struct EditorView_Previews: PreviewProvider {
    static var previews: some View {
        EditorView(text: .constant("""
        .org 0x8000
        MOV R0, #42
        SWI #0
        """))
        .frame(width: 400, height: 300)
    }
}
