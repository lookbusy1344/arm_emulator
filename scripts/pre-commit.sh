#!/usr/bin/env bash
# pre-commit.sh — run the checks selected by the files changed in @.
#
# Works in jj and plain Git checkouts.
#
# jj: no commit hooks, so run it directly before `jj commit`, `jj describe`
# (finalising) and `jj squash`. It checks the files changed in `@`.
#
# Git: install as a hook, it checks the files changed against HEAD:
#   ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit
#
# scripts/checks.sh maps paths to suites; documentation-only changes run none.

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly PROJECT_DIR

changed_files() {
    if [[ -z "${GIT_INDEX_FILE:-}" ]] && jj --ignore-working-copy root > /dev/null 2>&1; then
        jj diff --name-only -r @ --no-pager
    else
        git diff HEAD --name-only --diff-filter=ACMR
    fi
}

cd "${PROJECT_DIR}"
changed_files | scripts/checks.sh
