#!/usr/bin/env bash
# pre-push.sh — format and check unpushed jj revisions.
#
# Default: format the tip (the newest non-empty mutable revision in ::@) with
# `jj fix`, then run scripts/checks.sh in the working copy for the files
# changed across the stack.
# --full: format every non-empty mutable revision in ::@, then check each in
# its own checkout via `jj run` (cold Xcode and .NET builds per revision).
#
# In a plain Git checkout there are no revisions to select: it checks the
# working tree for the files changed since the upstream, or every suite when
# there is no upstream. --full makes no difference there. Install as a hook:
#   ln -sf ../../scripts/pre-push.sh .git/hooks/pre-push
#
# Run before `jj git push` or moving a shared bookmark.

set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"

readonly STACK='mutable() & ::@ ~ empty()'
readonly TIP="heads(${STACK})"
# Runs in each `jj run` checkout, which has no .jj of its own: the diff runs in
# the repo at PRE_PUSH_REPO so paths are repo-relative. jj sets JJ_CHANGE_ID.
# shellcheck disable=SC2016
readonly REVISION_CHECKS='set -euo pipefail; (cd "${PRE_PUSH_REPO}" && jj --ignore-working-copy diff --name-only -r "${JJ_CHANGE_ID}") | scripts/checks.sh'

# Inline so formatting does not depend on unversioned .jj/repo/config.toml.
# The full style checks run in scripts/checks.sh.
readonly FORMAT_CONFIG=(
    --config 'fix.tools.gofmt.command=["scripts/format-stdin.sh", "$path"]'
    --config 'fix.tools.gofmt.patterns=["glob:\"**/*.go\""]'
    --config 'fix.tools.swiftformat.command=["scripts/format-stdin.sh", "$path"]'
    --config 'fix.tools.swiftformat.patterns=["glob:\"swift-gui/**/*.swift\""]'
    --config 'fix.tools.dotnet-format.command=["scripts/format-stdin.sh", "$path"]'
    --config 'fix.tools.dotnet-format.patterns=["glob:\"avalonia-gui/**/*.cs\""]'
)

usage() {
    echo "Usage: $0 [--full|-f]"
    echo "  --full, -f  Format and check every mutable revision, not only the tip."
}

full=false
while [[ $# -gt 0 ]]; do
    case "$1" in
        --full | -f) full=true; shift ;;
        -h | --help) usage; exit 0 ;;
        # Git hooks receive <remote> <url>; they select nothing here.
        [^-]*) shift ;;
        *) echo "Error: unknown argument '$1'" >&2; usage >&2; exit 1 ;;
    esac
done

if root="$(jj --ignore-working-copy root 2> /dev/null)"; then
    use_jj=true
elif root="$(git rev-parse --show-toplevel 2> /dev/null)"; then
    use_jj=false
else
    echo "==> Not a jj or Git repository, skipping pre-push checks."
    exit 0
fi
cd "${root}"

if [[ "${use_jj}" == false ]]; then
    echo "==> Checking working tree"
    if upstream="$(git rev-parse --abbrev-ref '@{upstream}' 2> /dev/null)"; then
        git diff --name-only "${upstream}" | scripts/checks.sh
    else
        scripts/checks.sh --all
    fi
    exit 0
fi

stack="$(jj log --no-graph -r "${STACK}" -T 'change_id ++ "\n"')"
if [[ -z "${stack}" ]]; then
    echo "==> No non-empty mutable revisions in ${STACK}, nothing to check."
    exit 0
fi

if [[ "${full}" == true ]]; then
    echo "==> Formatting ${STACK}"
    jj "${FORMAT_CONFIG[@]}" fix -s "roots(${STACK})"
    echo "==> Checking each revision in ${STACK}"
    PRE_PUSH_REPO="${root}" jj run --root --ignore-changes -r "${STACK}" -- bash -c "${REVISION_CHECKS}"
else
    # Checks run in the working copy; when @ is empty its tree is the tip's.
    echo "==> Formatting tip ${TIP}"
    jj "${FORMAT_CONFIG[@]}" fix -s "${TIP}"
    echo "==> Checking tip $(jj log --no-graph -r "${TIP}" -T 'change_id.short() ++ " " ++ coalesce(description.first_line(), "(no description)")')"
    jj diff --name-only -r "${STACK}" | scripts/checks.sh
fi

echo "==> All checks passed."
