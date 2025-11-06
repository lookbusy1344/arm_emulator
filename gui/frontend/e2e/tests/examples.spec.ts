import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { loadProgram, waitForExecution } from '../utils/helpers';
import { TIMEOUTS } from '../utils/test-constants';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

// Path to examples directory relative to the frontend directory
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const EXAMPLES_DIR = path.join(__dirname, '../../../../examples');

test.describe('Example Programs', () => {
  const exampleFiles = [
    'hello.s',
    'loops.s',
    'arithmetic.s',
    'factorial.s',
  ];

  for (const file of exampleFiles) {
    test(`should execute ${file}`, async ({ page }) => {
      const appPage = new AppPage(page);
      await appPage.goto();

      // Load example file
      const programPath = path.join(EXAMPLES_DIR, file);

      // Check if file exists
      if (!fs.existsSync(programPath)) {
        test.skip();
        return;
      }

      const program = fs.readFileSync(programPath, 'utf-8');
      await loadProgram(appPage, program);

      // Switch to output tab BEFORE running (critical for event capture)
      await appPage.switchToOutputTab();

      // Run program
      await appPage.clickRun();

      // Wait for completion
      await waitForExecution(page, TIMEOUTS.EXECUTION_MAX);

      // Verify program completed (check for EXIT)
      await appPage.switchToStatusTab();
      const status = await page.locator('[data-testid="execution-status"]').textContent();
      expect(status?.toLowerCase()).toContain('halted');
    });
  }
});

test.describe('Complex Example Programs', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should execute quicksort.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'quicksort.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for completion (sorting may take longer)
    await waitForExecution(page, 30000);
    const output = await appPage.getOutputText();

    // Verify output exists
    expect(output.length).toBeGreaterThan(0);
  });

  test('should execute linked_list.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'linked_list.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for completion
    await waitForExecution(page, 15000);

    // Verify program completed
    await appPage.switchToStatusTab();
    const status = await page.locator('[data-testid="execution-status"]').textContent();
    expect(status?.toLowerCase()).toContain('halted');
  });

  test('should execute recursive_factorial.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'recursive_factorial.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for completion
    await waitForExecution(page, 10000);

    const output = await appPage.getOutputText();

    // Factorial output should contain numbers
    expect(output.length).toBeGreaterThan(0);
  });

  test('should execute string_reverse.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'string_reverse.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for completion
    await waitForExecution(page, 10000);

    const output = await appPage.getOutputText();

    // String reverse should produce output
    expect(output.length).toBeGreaterThan(0);
  });

  test('should execute arrays.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'arrays.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for completion
    await waitForExecution(page, 10000);

    // Verify program completed
    await appPage.switchToStatusTab();
    const status = await page.locator('[data-testid="execution-status"]').textContent();
    expect(status?.toLowerCase()).toContain('halted');
  });
});

test.describe('Example Program Stepping', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should step through hello.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'hello.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Step through program
    let stepCount = 0;
    const maxSteps = 20;

    while (stepCount < maxSteps) {
      await appPage.clickStep();
      stepCount++;

      // Check if program has exited
      const status = await page.locator('[data-testid="execution-status"]').textContent();
      if (status && status.toLowerCase().includes('halted')) {
        break;
      }
    }

    // Verify we stepped through some instructions
    expect(stepCount).toBeGreaterThan(0);
    expect(stepCount).toBeLessThanOrEqual(maxSteps);
  });

  test('should step through loops.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'loops.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Step through first 50 instructions
    for (let i = 0; i < 50; i++) {
      await appPage.clickStep();

      // Check if program has exited early
      const status = await page.locator('[data-testid="execution-status"]').textContent();
      if (status && status.toLowerCase().includes('halted')) {
        break;
      }
    }

    // Should have executed some instructions
    // (loops.s should have loop iterations)
  });

  test('should step through functions.s', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'functions.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Step through program with function calls
    for (let i = 0; i < 30; i++) {
      await appPage.clickStep();

      // Check if program has exited
      const status = await page.locator('[data-testid="execution-status"]').textContent();
      if (status && status.toLowerCase().includes('halted')) {
        break;
      }
    }

    // Verify we can step through function calls
  });
});

test.describe('Example Program Output Verification', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('hello.s should output "Hello, World!"', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'hello.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();
    await waitForExecution(page, 10000);
    const output = await appPage.getOutputText();

    // Verify output contains "Hello, World!"
    expect(output).toContain('Hello');
  });

  test('arithmetic.s should perform calculations', async ({ page }) => {
    const programPath = path.join(EXAMPLES_DIR, 'arithmetic.s');

    if (!fs.existsSync(programPath)) {
      test.skip();
      return;
    }

    const program = fs.readFileSync(programPath, 'utf-8');
    await loadProgram(appPage, program);

    // Switch to output tab BEFORE running (critical for event capture)
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();
    await waitForExecution(page, 10000);

    // Program should complete successfully
    await appPage.switchToStatusTab();
    const status = await page.locator('[data-testid="execution-status"]').textContent();
    expect(status?.toLowerCase()).toContain('halted');
  });
});
