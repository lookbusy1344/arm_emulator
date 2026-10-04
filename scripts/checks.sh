#!/usr/bin/env bash
# checks.sh [--all] — run the check suites selected by the changed paths on stdin.
#
# Suites: go (backend, examples, tests), swift (swift-gui/), dotnet
# (avalonia-gui/). Documentation and other paths select nothing. --all runs
# every suite and ignores stdin.
#
# The suites match CI (.github/workflows/ci.yml), plus the format and lint
# gates that AGENTS.md requires: gofmt, swiftlint --strict and dotnet format.
# Shared by pre-commit.sh and pre-push.sh.

set -euo pipefail

# GUI and hook launches omit the user-local dotnet, Go and Homebrew paths.
export PATH="$HOME/.dotnet:$HOME/go/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"

readonly SUITES=(go swift dotnet)
readonly GO_TEST_TIMEOUT="5m"
readonly DOTNET_TEST_TIMEOUT_SECONDS=120
readonly XCODE_PROJECT="ARMEmulator.xcodeproj"
readonly XCODE_SCHEME="ARMEmulator"
readonly XCODE_DESTINATION="platform=macOS"

suite_for() {
    case "$1" in
        *.md | docs/*) ;;
        swift-gui/*) echo swift ;;
        avalonia-gui/*) echo dotnet ;;
        *.go | go.mod | go.sum | .golangci.yml | Makefile | examples/* | include/* | tests/*) echo go ;;
    esac
}

run_go() {
    echo "==> Go: gofmt, build, golangci-lint, test"
    local unformatted
    unformatted="$(gofmt -l .)"
    if [[ -n "${unformatted}" ]]; then
        printf 'gofmt needed:\n%s\n' "${unformatted}" >&2
        return 1
    fi
    # tests/integration runs ./arm-emulator.
    make build
    golangci-lint run ./...
    go test -count=1 -timeout "${GO_TEST_TIMEOUT}" ./...
}

run_swift() {
    echo "==> Swift: xcodegen, swiftformat, swiftlint, build, test"
    # The app bundles ../arm-emulator in a post-build step.
    make build
    (
        cd swift-gui
        xcodegen generate --quiet
        swiftformat --lint . --quiet
        swiftlint --strict --quiet
        xcodebuild -project "${XCODE_PROJECT}" -scheme "${XCODE_SCHEME}" build | xcbeautify --quiet
        xcodebuild test -project "${XCODE_PROJECT}" -scheme "${XCODE_SCHEME}" -destination "${XCODE_DESTINATION}" | xcbeautify --quiet
    )
}

run_dotnet() {
    echo "==> .NET: restore, build, format, test"
    (
        cd avalonia-gui
        dotnet restore
        dotnet build --no-restore
        dotnet format --verify-no-changes --no-restore
        gtimeout "${DOTNET_TEST_TIMEOUT_SECONDS}" dotnet test --no-build
    )
}

selected_suites() {
    if [[ "${1:-}" == "--all" ]]; then
        printf '%s\n' "${SUITES[@]}"
        return
    fi
    local path
    while IFS= read -r path; do
        [[ -n "${path}" ]] && suite_for "${path}"
    done | sort -u
}

main() {
    local selected
    selected="$(selected_suites "$@")"
    if [[ -z "${selected}" ]]; then
        echo "==> No Go, Swift or .NET changes, skipping checks."
        return
    fi

    cd "$(dirname "${BASH_SOURCE[0]}")/.."
    local suite
    for suite in "${SUITES[@]}"; do
        if grep -qx "${suite}" <<< "${selected}"; then
            "run_${suite}"
        fi
    done
    echo "==> Checks passed: $(paste -sd ' ' - <<< "${selected}")"
}

main "$@"
