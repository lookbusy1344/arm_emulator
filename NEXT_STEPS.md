# Next Steps for ARM Emulator

**Last Updated:** 2025-10-09  
**Current Status:** All 10 phases complete, ready for production hardening

This document provides a prioritized action plan based on the comprehensive project review in [PROJECT_REVIEW.md](PROJECT_REVIEW.md).

---

## Quick Wins ✅ COMPLETED (2025-10-10)

All quick wins have been completed:

### 1. Fix Go Vet Warnings ✅ COMPLETE
**Status:** Fixed
**Files:** `vm/memory.go` and all call sites

- Renamed `ReadByte` → `ReadByteAt` to avoid conflict with `io.ByteReader`
- Renamed `WriteByte` → `WriteByteAt` to avoid conflict with `io.ByteWriter`
- Updated all call sites across the codebase (14 files)
- Go vet now passes with no warnings

### 2. Update CI Go Version ✅ COMPLETE
**Status:** Updated
**Files:** `.github/workflows/ci.yml`

- Updated from Go 1.21 to Go 1.25
- CI now matches project requirements

### 3. Add .gitignore Entry for Build Artifacts ✅ COMPLETE
**Status:** Added

Added the following entries to `.gitignore`:
- `/tmp/`
- `*.prof`
- `coverage.out`
- `*.log`

---

## Phase 11: Production Hardening (Week 18-19)

**Total Effort:** 14-20 hours  
**Priority:** HIGH

### Task 1: Fix Code Quality Issues (30 minutes)
- [x] Fix go vet warnings (✅ COMPLETED 2025-10-10)
- [ ] Run golangci-lint and address issues
- [x] Ensure all code is gofmt clean (✅ COMPLETED)

### Task 2: Enhance CI/CD Pipeline (4-6 hours)

**File:** `.github/workflows/ci.yml`

**Requirements:**
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.25']

steps:
  - name: Run tests with coverage
    run: go test -v -race -coverprofile=coverage.out ./...
  
  - name: Upload coverage
    uses: codecov/codecov-action@v3
    with:
      files: ./coverage.out
  
  - name: Check coverage threshold
    run: |
      coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
      if (( $(echo "$coverage < 70" | bc -l) )); then
        echo "Coverage $coverage% is below threshold 70%"
        exit 1
      fi
```

**Deliverables:**
- Cross-platform testing (macOS, Linux, Windows)
- Coverage reporting with codecov
- Coverage threshold enforcement (70% minimum)
- Test results as CI artifacts

### Task 3: Cross-Platform Manual Testing (3-4 hours)

Create testing checklist in `docs/testing_checklist.md`:

**Windows Testing:**
- [ ] Binary builds successfully
- [ ] All tests pass
- [ ] TUI renders correctly
- [ ] File paths work (backslash handling)
- [ ] Config file loads from correct location
- [ ] Example programs execute correctly
- [ ] Command-line flags work

**macOS Testing:**
- [ ] Builds on Intel and Apple Silicon
- [ ] TUI colors render correctly
- [ ] File I/O works
- [ ] All examples run

**Linux Testing:**
- [ ] Test on Ubuntu, Fedora, Arch
- [ ] Terminal compatibility (various emulators)
- [ ] Package dependencies documented

### Task 4: Increase Code Coverage to 75%+ (4-6 hours)

**Target Coverage by Package:**
- `vm/`: 75%+ (currently no direct tests)
- `debugger/`: 65%+ (currently 47.9%)
- `parser/`: 75%+
- `tools/`: 90%+ (currently 86.6%)
- `config/`: 85%+

**Focus Areas:**

**A. VM Package Tests (2 hours)**
Create `vm/vm_test.go`:
```go
func TestVM_Initialization(t *testing.T) {
    vm := NewVM()
    // Test initial state
}

func TestVM_Reset(t *testing.T) {
    vm := setupVM()
    vm.Reset()
    // Verify state cleared
}

func TestVM_ResetRegisters(t *testing.T) {
    vm := setupVM()
    vm.ResetRegisters()
    // Verify memory preserved
}
```

**B. Debugger Expression Tests (1-2 hours)**
Enhance `debugger/expressions_test.go`:
```go
func TestEvaluateExpression_ComplexHex(t *testing.T) {
    // Test hex arithmetic
}

func TestEvaluateExpression_BitwiseOps(t *testing.T) {
    // Test AND, OR, XOR
}

func TestEvaluateExpression_ErrorHandling(t *testing.T) {
    // Test invalid expressions
}
```

**C. Parser Error Paths (1-2 hours)**
Add tests for invalid input:
```go
func TestParser_InvalidSyntax(t *testing.T) {
    // Test malformed instructions
}

func TestParser_UndefinedLabels(t *testing.T) {
    // Test forward references
}
```

### Task 5: Connect Tracing to Execution (2-3 hours)

**Files:** `vm/executor.go`, `vm/trace.go`

**Implementation:**
```go
// In vm/executor.go, modify Step():
func (vm *VM) Step() error {
    if vm.traceExecution {
        vm.recordExecutionTrace()
    }
    
    // ... existing execution code ...
    
    if vm.memTrace {
        // Already implemented in memory.go
    }
    
    return nil
}

// Add helper methods:
func (vm *VM) recordExecutionTrace() {
    if vm.executionTrace != nil {
        entry := vm.executionTrace.RecordInstruction(vm, /* params */)
        vm.executionTrace.Entries = append(vm.executionTrace.Entries, entry)
    }
}
```

**Testing:**
```go
func TestVM_TracingEnabled(t *testing.T) {
    vm := NewVM()
    vm.EnableTrace(true)
    
    // Execute some instructions
    vm.Step()
    vm.Step()
    
    // Verify trace entries created
    if len(vm.executionTrace.Entries) != 2 {
        t.Error("Expected 2 trace entries")
    }
}
```

---

## Phase 12: Performance & Benchmarking (Week 20-21)

**Total Effort:** 14-20 hours  
**Priority:** MEDIUM

### Task 1: Create Benchmark Suite (4-6 hours)

**File:** `vm/vm_bench_test.go`

```go
func BenchmarkExecutionLoop(b *testing.B) {
    vm := setupBenchmarkVM()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        vm.Step()
    }
}

func BenchmarkMemoryRead(b *testing.B) {
    vm := setupBenchmarkVM()
    addr := uint32(0x8000)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        vm.Memory.ReadWord(addr)
    }
}

func BenchmarkInstructionDecode(b *testing.B) {
    vm := setupBenchmarkVM()
    instruction := uint32(0xE0812003) // ADD R2, R1, R3
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        decodeInstruction(instruction)
    }
}
```

**Parser Benchmarks:**

**File:** `parser/parser_bench_test.go`

```go
func BenchmarkParseLargeFile(b *testing.B) {
    source := loadTestFile("examples/bubble_sort.s")
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        parser := NewParser()
        parser.Parse(source)
    }
}
```

### Task 2: Profile Performance (2-3 hours)

```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./vm
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./vm
go tool pprof mem.prof

# Generate profile reports
go tool pprof -http=:8080 cpu.prof
```

**Create:** `docs/performance_analysis.md` with findings.

### Task 3: Implement Optimizations (6-8 hours)

Based on profiling, implement targeted optimizations:

**A. Instruction Decode Cache**
```go
type VM struct {
    // ... existing fields ...
    decodeCache map[uint32]*DecodedInstruction
}

func (vm *VM) Step() error {
    pc := vm.Reg[PC]
    
    decoded, cached := vm.decodeCache[pc]
    if !cached {
        instruction := vm.Memory.ReadWord(pc)
        decoded = decodeInstruction(instruction)
        vm.decodeCache[pc] = decoded
    }
    
    return vm.execute(decoded)
}
```

**B. Memory Access Optimization**
```go
// Pre-allocate common memory regions
type Memory struct {
    codeSection []byte        // Flat array for code (fast)
    dataSection []byte        // Flat array for data (fast)
    dynamicPages map[uint32][]byte  // Sparse for heap (flexible)
}
```

**C. Flag Calculation Optimization**
```go
// Only calculate flags when needed
func (vm *VM) executeADD(dest, src1, src2 uint32, setFlags bool) {
    result := vm.Reg[src1] + vm.Reg[src2]
    vm.Reg[dest] = result
    
    if setFlags {
        vm.updateFlags(result, ...)
    }
}
```

### Task 4: Document Performance (2-3 hours)

Create `docs/performance_characteristics.md`:
- Benchmark results
- Performance comparison to similar tools
- Optimization opportunities
- Scalability limits

---

## Phase 13: Release Engineering (Week 22-23)

**Total Effort:** 16-22 hours  
**Priority:** MEDIUM-HIGH

### Task 1: Create Release Pipeline (4-6 hours)

**File:** `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
            output: arm-emulator-linux-amd64
          - os: macos-latest
            goos: darwin
            goarch: amd64
            output: arm-emulator-darwin-amd64
          - os: macos-latest
            goos: darwin
            goarch: arm64
            output: arm-emulator-darwin-arm64
          - os: windows-latest
            goos: windows
            goarch: amd64
            output: arm-emulator-windows-amd64.exe

    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: go build -o ${{ matrix.output }} -ldflags "-s -w"
      
      - name: Create archive
        run: tar -czf ${{ matrix.output }}.tar.gz ${{ matrix.output }} examples/ docs/ README.md
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.output }}
          path: ${{ matrix.output }}.tar.gz

  release:
    needs: build
    runs-on: ubuntu-latest
    
    steps:
      - name: Download artifacts
        uses: actions/download-artifact@v4
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: '**/*.tar.gz'
          draft: false
          prerelease: false
```

### Task 2: Create Installation Packages (6-8 hours)

**A. Homebrew Formula (2 hours)**

Create `homebrew-arm-emulator` repository with formula:
```ruby
class ArmEmulator < Formula
  desc "ARM2 assembly emulator with TUI debugger"
  homepage "https://github.com/lookbusy1344/arm_emulator"
  url "https://github.com/lookbusy1344/arm_emulator/archive/v1.0.0.tar.gz"
  sha256 "..."
  
  depends_on "go" => :build
  
  def install
    system "go", "build", "-o", bin/"arm-emulator"
  end
  
  test do
    assert_match "ARM2 Emulator", shell_output("#{bin}/arm-emulator --version")
  end
end
```

**B. Chocolatey Package (2 hours)**

Create `chocolatey/arm-emulator.nuspec`:
```xml
<?xml version="1.0"?>
<package>
  <metadata>
    <id>arm-emulator</id>
    <version>1.0.0</version>
    <title>ARM2 Emulator</title>
    <authors>lookbusy1344</authors>
    <description>ARM2 assembly language emulator with TUI debugger</description>
    <projectUrl>https://github.com/lookbusy1344/arm_emulator</projectUrl>
    <tags>arm emulator debugger assembly</tags>
  </metadata>
</package>
```

**C. DEB Package (2 hours)**

Create Debian package structure:
```
debian/
├── control
├── changelog
├── rules
└── install
```

**D. RPM Package (2 hours)**

Create RPM spec file for Fedora/RHEL.

### Task 3: Release Documentation (3-4 hours)

**A. Create CHANGELOG.md**
```markdown
# Changelog

## [1.0.0] - 2025-XX-XX

### Added
- Complete ARM2 instruction set implementation
- Full TUI debugger with breakpoints and watchpoints
- Cross-platform configuration management
- Execution tracing and performance statistics
- Development tools (linter, formatter, xref)
- 17 example programs
- Comprehensive documentation

### Fixed
- Go vet interface warnings
- Memory method naming conflicts

### Performance
- Achieved 61% code coverage (target: 75%)
- 493 tests passing (99.4% pass rate)
```

**B. Update README.md**
Add installation section:
```markdown
## Installation

### Homebrew (macOS/Linux)
```bash
brew install lookbusy1344/tap/arm-emulator
```

### Chocolatey (Windows)
```powershell
choco install arm-emulator
```

### From Source
```bash
git clone https://github.com/lookbusy1344/arm_emulator
cd arm_emulator
go build -o arm-emulator
```
```

**C. Create CONTRIBUTING.md**
Guidelines for contributors.

### Task 4: Release Testing (3-4 hours)

Create `docs/release_checklist.md`:

```markdown
# Release Checklist

## Pre-Release
- [ ] All tests passing
- [ ] Code coverage >70%
- [ ] Documentation up to date
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated

## Build Testing
- [ ] Linux build works
- [ ] macOS Intel build works
- [ ] macOS ARM build works
- [ ] Windows build works

## Installation Testing
- [ ] Homebrew formula works
- [ ] Chocolatey package works
- [ ] DEB package installs
- [ ] RPM package installs

## Functional Testing
- [ ] All examples run successfully
- [ ] TUI debugger works on all platforms
- [ ] Command-line arguments work
- [ ] Configuration files load correctly

## Release
- [ ] Tag version in git
- [ ] GitHub Release created
- [ ] Release notes published
- [ ] Social media announcement
```

---

## Phase 14: Advanced Features (Optional, Future)

**Total Effort:** 80-120 hours  
**Priority:** LOW (Post-release)

### 1. JIT Compilation (40-60 hours)
Translate ARM to native x86-64 for 10-100x speedup.

### 2. Web-Based Debugger (30-40 hours)
React frontend with WebSocket backend for browser-based debugging.

### 3. GDB Protocol Support (20-30 hours)
Implement GDB remote serial protocol for IDE integration.

### 4. Additional Architectures (40+ hours)
Extend to ARM3, ARM7, or other RISC architectures.

---

## Getting Started

To begin Phase 11 immediately:

```bash
# 1. Fix go vet warnings
cd /home/runner/work/arm_emulator/arm_emulator
vim vm/memory.go  # Rename ReadByte/WriteByte methods

# 2. Update CI
vim .github/workflows/ci.yml  # Change Go version to 1.25

# 3. Run tests to verify
go test ./...

# 4. Commit changes
git add -A
git commit -m "Phase 11: Fix go vet warnings and update CI"
git push
```

---

## Success Metrics

**Phase 11 Complete When:**
- ✅ Zero go vet warnings
- ✅ CI runs on all 3 platforms
- ✅ Code coverage >75%
- ✅ Tracing connected to execution

**Phase 12 Complete When:**
- ✅ Benchmark suite created
- ✅ Performance profiling done
- ✅ 2+ optimizations implemented
- ✅ Performance documented

**Phase 13 Complete When:**
- ✅ Automated releases work
- ✅ Installation packages available
- ✅ v1.0.0 released on GitHub

---

## Questions or Issues?

If you encounter problems or have questions during implementation:

1. Review the detailed [PROJECT_REVIEW.md](PROJECT_REVIEW.md)
2. Check existing documentation in `docs/`
3. Look at similar test files for patterns
4. Create an issue on GitHub for discussion

Happy coding! 🚀
