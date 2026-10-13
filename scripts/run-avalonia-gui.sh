#!/usr/bin/env bash
# run-avalonia-gui.sh [program.s] — build the backend and the Avalonia GUI, then run it.
#
# Runs in the foreground; quit the app or press Ctrl+C to stop. A program path,
# relative or absolute, is loaded at startup.

set -euo pipefail

export PATH="$HOME/.dotnet:$HOME/go/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"

readonly ROOT="$(cd "$(dirname "$0")/.." && pwd)"
readonly PROJECT="$ROOT/avalonia-gui/ARMEmulator"

program_args=()
if (($# > 0)); then
    program_args=("$(cd "$(dirname "$1")" && pwd)/$(basename "$1")")
fi

# The build copies the backend binary next to the app
make -C "$ROOT" build
dotnet build "$PROJECT" --verbosity quiet
dotnet run --project "$PROJECT" --no-build -- ${program_args[@]+"${program_args[@]}"}
