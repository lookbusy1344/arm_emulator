# Expected Outputs

This directory contains the expected stdout output for example programs.

## Convention

- Each file is named `<basename>.txt` where `<basename>` matches the example program name without the `.s` extension
- For example, `quicksort.txt` contains the expected output for `examples/quicksort.s`
- The content should be the exact output produced by running the example program
- Files should include trailing newlines as produced by the actual program

## Adding New Tests

To add a test for a new example program:

1. Run the program and capture its output:
   ```bash
   ./arm-emulator examples/yourprogram.s > tests/integration/expected_outputs/yourprogram.txt
   ```

2. Add a test function in `example_programs_test.go`:
   ```go
   func TestExampleProgram_YourProgram(t *testing.T) {
       stdout, _, exitCode, err := runExampleProgram(t, "yourprogram.s")
       if err != nil {
           t.Fatalf("execution failed: %v", err)
       }
       
       if exitCode != 0 {
           t.Errorf("expected exit code 0, got %d", exitCode)
       }
       
       expected := loadExpectedOutput(t, "yourprogram")
       if stdout != expected {
           t.Errorf("output mismatch\nExpected (%d bytes):\n%q\nGot (%d bytes):\n%q", 
               len(expected), expected, len(stdout), stdout)
       }
   }
   ```

3. Run the test to verify:
   ```bash
   go test ./tests/integration -run TestExampleProgram_YourProgram -v
   ```
