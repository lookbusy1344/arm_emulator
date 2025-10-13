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

2. Add an entry to the test table in `example_programs_test.go`:
   ```go
   {
       name:           "YourProgram",
       programFile:    "yourprogram.s",
       expectedOutput: "yourprogram.txt",
   },
   ```

3. Run the test to verify:
   ```bash
   go test ./tests/integration -run TestExamplePrograms/YourProgram -v
   ```

That's it! The test framework automatically handles loading, running, and comparing outputs.
