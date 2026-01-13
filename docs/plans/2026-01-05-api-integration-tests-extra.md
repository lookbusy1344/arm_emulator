# API Integration Tests - Follow-up Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Address remaining gaps from the 2026-01-04 API integration tests plan.

**Issues Identified:**
1. Missing 4 example program tests (47 of 49 implemented)
2. Noisy WebSocket log output during test runs

---

## Task 1: Add test_store_highlight.s Test

The `test_store_highlight.s` program has deterministic output and should be tested.

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`
- Create: `tests/integration/expected_outputs/test_store_highlight.txt`

**Step 1: Generate expected output**

```bash
./arm-emulator examples/test_store_highlight.s > tests/integration/expected_outputs/test_store_highlight.txt
```

**Step 2: Add test case to TestAPIExamplePrograms**

Add to tests slice:

```go
{
    name:           "TestStoreHighlight_API",
    programFile:    "test_store_highlight.s",
    expectedOutput: "test_store_highlight.txt",
    stdinMode:      "",
},
```

**Step 3: Verify test passes**

```bash
go test ./tests/integration -run TestAPIExamplePrograms/TestStoreHighlight_API -v
```

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go tests/integration/expected_outputs/test_store_highlight.txt
git commit -m "test: add test_store_highlight.s API integration test"
```

---

## Task 2: Add test_debug_syscalls.s Test

The `test_debug_syscalls.s` program tests debug output syscalls. The output includes formatted register/memory dumps which should be deterministic.

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`
- Create: `tests/integration/expected_outputs/test_debug_syscalls.txt`

**Step 1: Generate expected output**

```bash
./arm-emulator examples/test_debug_syscalls.s > tests/integration/expected_outputs/test_debug_syscalls.txt
```

**Step 2: Add test case**

```go
{
    name:           "TestDebugSyscalls_API",
    programFile:    "test_debug_syscalls.s",
    expectedOutput: "test_debug_syscalls.txt",
    stdinMode:      "",
},
```

**Step 3: Verify and commit**

```bash
go test ./tests/integration -run TestAPIExamplePrograms/TestDebugSyscalls_API -v
git add tests/integration/api_example_programs_test.go tests/integration/expected_outputs/test_debug_syscalls.txt
git commit -m "test: add test_debug_syscalls.s API integration test"
```

---

## Task 3: Document Non-Deterministic Tests as Intentionally Skipped

The following programs have non-deterministic output and cannot have fixed expected output files:
- `test_get_random.s` - Uses GET_RANDOM syscall
- `test_get_time.s` - Uses GET_TIME syscall

**Files:**
- Modify: `tests/integration/api_example_programs_test.go` (add skip comments)
- Modify: `CLAUDE.md` (document skipped tests)

**Step 1: Add documentation comment to test file**

Add at the end of the tests slice definition:

```go
// Note: The following example programs are intentionally NOT tested:
// - test_get_random.s: Uses GET_RANDOM syscall (non-deterministic output)
// - test_get_time.s: Uses GET_TIME syscall (non-deterministic output)
```

**Step 2: Update CLAUDE.md API Integration Tests section**

Add under the existing API Integration Tests section:

```markdown
**Excluded Tests:**
- `test_get_random.s` - Non-deterministic (random number generation)
- `test_get_time.s` - Non-deterministic (system timestamp)
```

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go CLAUDE.md
git commit -m "docs: document intentionally skipped non-deterministic tests"
```

---

## Task 4: Suppress Benign WebSocket Close Logs (Optional)

The test output shows "WebSocket error: websocket: close 1000 (normal)" which is a benign normal closure. Consider suppressing this to reduce test noise.

**Files:**
- Modify: `api/websocket.go` or `tests/integration/api_example_programs_test.go`

**Step 1: Identify log source**

Search for the log statement:

```bash
grep -r "WebSocket error" api/ tests/
```

**Step 2: Suppress normal closure logs**

If logging is in the API server, check for close code 1000 and skip logging:

```go
if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
    return // Don't log normal closures
}
```

**Step 3: Verify and commit**

```bash
go test ./tests/integration -run TestAPIExamplePrograms -v 2>&1 | grep -c "WebSocket error"
# Should be 0 or much lower
git add .
git commit -m "fix: suppress benign WebSocket normal closure logs"
```

---

## Completion Checklist

- [ ] Task 1: test_store_highlight.s test added
- [ ] Task 2: test_debug_syscalls.s test added
- [ ] Task 3: Non-deterministic tests documented
- [ ] Task 4: (Optional) WebSocket log noise reduced

---

## Success Criteria

- [ ] 49 example programs accounted for (47 tested + 2 documented as skipped)
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Test output is clean (no unexpected warnings)
