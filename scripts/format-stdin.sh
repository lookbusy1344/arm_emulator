#!/usr/bin/env bash
# format-stdin.sh <path> — `jj fix` tool: read source on stdin, write the
# formatted result to stdout. Dispatches on the file extension:
#   .go     gofmt
#   .swift  swiftformat; --stdin-path lets it find swift-gui/.swiftformat
#   .cs     dotnet format whitespace with avalonia-gui/.editorconfig
#
# `dotnet format` works on files, so C# input goes through a scratch folder.

set -euo pipefail

export PATH="$HOME/.dotnet:/opt/homebrew/bin:/usr/local/bin:$PATH"

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly EDITORCONFIG="${PROJECT_DIR}/avalonia-gui/.editorconfig"

path="${1:?usage: format-stdin.sh <path>}"

format_csharp() {
    local name
    scratch="$(mktemp -d)"
    trap 'rm -rf "${scratch}"' EXIT
    name="$(basename "${path}")"
    cp "${EDITORCONFIG}" "${scratch}/"
    cat > "${scratch}/${name}"
    dotnet format whitespace "${scratch}" --folder --include "${name}" > /dev/null 2>&1 || true
    cat "${scratch}/${name}"
}

case "${path}" in
    *.go) exec gofmt ;;
    *.swift) exec swiftformat stdin --stdin-path "${path}" --quiet ;;
    *.cs) format_csharp ;;
    *) exec cat ;;
esac
