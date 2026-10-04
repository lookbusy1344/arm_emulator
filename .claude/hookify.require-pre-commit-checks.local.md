---
name: require-pre-commit-checks
enabled: true
event: bash
pattern: (jj\s+(commit|describe|squash)|git\s+commit)
action: warn
---

**Pre-commit check required.**

Run `scripts/pre-commit.sh`. It reads the changed files from jj (`@`) or Git (against `HEAD`) and runs only the suites they need:

- Go (`*.go`, `go.mod`, `.golangci.yml`, `Makefile`, `examples/`, `include/`, `tests/`): gofmt, build, golangci-lint, tests
- Swift (`swift-gui/`): xcodegen, swiftformat, swiftlint --strict, xcodebuild build and test
- .NET (`avalonia-gui/`): restore, build, `dotnet format --verify-no-changes`, tests

Documentation-only changes (`*.md`, `docs/`) run nothing.

All suites must pass. Stop and fix on any failure.

In a jj repo, `jj describe` on an unfinished change does not need the checks; run them when finalising.

If you have already run the checks in this session on the same changes and they passed, you may proceed.
