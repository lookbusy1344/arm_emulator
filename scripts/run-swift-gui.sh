#!/usr/bin/env bash
# run-swift-gui.sh [program.s] — build the backend and the Swift GUI, then launch it.
#
# Quits a running copy first so the new build starts. A program path, relative
# or absolute, is loaded at startup.

set -euo pipefail

export PATH="$HOME/go/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"

readonly ROOT="$(cd "$(dirname "$0")/.." && pwd)"
readonly APP_NAME="ARMEmulator"
readonly XCODE_PROJECT="ARMEmulator.xcodeproj"
readonly XCODE_SCHEME="ARMEmulator"

program_args=()
if (($# > 0)); then
    program_args=(--args "$(cd "$(dirname "$1")" && pwd)/$(basename "$1")")
fi

make -C "$ROOT" build

cd "$ROOT/swift-gui"
xcodegen generate --quiet
xcodebuild -project "$XCODE_PROJECT" -scheme "$XCODE_SCHEME" build | xcbeautify --quiet

products_dir="$(xcodebuild -project "$XCODE_PROJECT" -scheme "$XCODE_SCHEME" -showBuildSettings 2>/dev/null |
    awk -F' = ' '/ BUILT_PRODUCTS_DIR /{print $2; exit}')"

if pgrep -x "$APP_NAME" >/dev/null; then
    osascript -e "tell application \"$APP_NAME\" to quit"
    while pgrep -x "$APP_NAME" >/dev/null; do sleep 0.2; done
fi

open "$products_dir/$APP_NAME.app" ${program_args[@]+"${program_args[@]}"}
