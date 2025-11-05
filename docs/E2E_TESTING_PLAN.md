# End-to-End Testing Plan for Wails GUI

## Overview

This document outlines the plan for implementing comprehensive end-to-end (E2E) testing of the ARM Emulator's Wails-based GUI using Playwright. The testing infrastructure will enable automated browser-based testing of the entire application workflow, from loading programs to stepping through execution and examining state.

## Why Playwright?

Playwright is an optimal choice for E2E testing of Wails applications for several reasons:

1. **Browser-based testing**: Wails applications use a webview (WebKit/WKWebView on macOS, Edge WebView2 on Windows, WebKit on Linux), and Playwright can test the actual web content
2. **Cross-browser support**: Test on Chromium, Firefox, and WebKit (matching macOS behavior)
3. **Powerful selectors**: CSS, text, accessibility attributes, and custom data attributes
4. **Built-in waiting**: Auto-waits for elements to be ready before acting
5. **Screenshots and video**: Visual regression testing and debugging capabilities
6. **TypeScript native**: Seamless integration with existing TypeScript codebase
7. **Headless and headed modes**: Run tests with or without visible browser for debugging
8. **Network interception**: Mock backend responses if needed
9. **Parallel execution**: Run tests concurrently for faster feedback

## Architecture Overview

### Testing Strategy

The E2E tests will operate at three levels:

1. **Component-level E2E**: Test individual views (RegisterView, MemoryView, etc.) in isolation
2. **Feature-level E2E**: Test complete workflows (load program, step through, set breakpoints)
3. **Integration E2E**: Test the entire application with real ARM programs from examples/

### Test Environment Setup

```
gui/
├── frontend/
│   ├── e2e/
│   │   ├── fixtures/          # Test programs and data
│   │   ├── pages/             # Page Object Models
│   │   ├── tests/             # Test specifications
│   │   ├── utils/             # Helper functions
│   │   └── playwright.config.ts
│   ├── playwright-report/     # Test reports (gitignored)
│   └── test-results/          # Test artifacts (gitignored)
```

### How Playwright Tests Wails Applications

There are two approaches for testing Wails applications:

#### Approach 1: Test the Development Server (Recommended for CI)

During development, Wails runs a dev server (typically `http://localhost:34115`) that serves the frontend. Playwright can test against this server directly:

**Pros:**
- Fast test execution (no need to build native app)
- Easy to run in CI/CD pipelines
- Hot reload during test development
- Works cross-platform without platform-specific setup

**Cons:**
- Doesn't test the actual native webview
- Wails runtime APIs need mocking or a running dev backend

#### Approach 2: Test the Built Application (Recommended for Release)

Test the actual compiled Wails application by:
1. Building the native app with `wails build`
2. Using Playwright's `electronApp` or custom launcher to start the app
3. Testing the actual webview window

**Pros:**
- Tests the real application with native webview
- Tests actual Wails API bindings
- Catches platform-specific issues

**Cons:**
- Slower (requires full build)
- More complex setup per platform
- Harder to run in CI (needs platform-specific runners)

### Recommended Hybrid Approach

1. **Primary test suite**: Run against dev server for speed and developer experience
2. **Smoke tests**: Run against built application before releases
3. **Mock Wails APIs**: Create a test harness that mocks backend calls when testing dev server

## Implementation Plan

### Phase 1: Infrastructure Setup

#### 1.1 Install Dependencies

```bash
cd gui/frontend
npm install -D @playwright/test @playwright/experimental-ct-react
npm install -D playwright-core
```

#### 1.2 Initialize Playwright

```bash
npx playwright install chromium webkit firefox
npx playwright install-deps
```

#### 1.3 Create Configuration

Create `gui/frontend/playwright.config.ts`:

```typescript
import { defineConfig, devices } from '@playwright/test';

const PORT = process.env.PORT || 34115;
const BASE_URL = process.env.BASE_URL || `http://localhost:${PORT}`;

export default defineConfig({
  testDir: './e2e/tests',

  // Run tests in files in parallel
  fullyParallel: true,

  // Fail the build on CI if you accidentally left test.only
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,

  // Reporter to use
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['list'] // Console output
  ],

  use: {
    // Base URL to use in actions like `await page.goto('/')`
    baseURL: BASE_URL,

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Default timeout for each action
    actionTimeout: 10000,
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    // Test against mobile viewports (optional)
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 13'] },
    },
  ],

  // Run dev server before starting tests
  webServer: {
    command: 'npm run dev',
    port: PORT,
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
});
```

#### 1.4 Update package.json Scripts

```json
{
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:debug": "playwright test --debug",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:report": "playwright show-report"
  }
}
```

### Phase 2: Page Object Models

Implement Page Object Models (POM) for maintainability and reusability.

#### 2.1 Base Page Object

Create `gui/frontend/e2e/pages/base.page.ts`:

```typescript
import { Page, Locator } from '@playwright/test';

export class BasePage {
  constructor(protected page: Page) {}

  async goto() {
    await this.page.goto('/');
  }

  async waitForLoad() {
    await this.page.waitForLoadState('networkidle');
  }

  async takeScreenshot(name: string) {
    await this.page.screenshot({ path: `test-results/${name}.png` });
  }
}
```

#### 2.2 Main Application Page Object

Create `gui/frontend/e2e/pages/app.page.ts`:

```typescript
import { Page, Locator } from '@playwright/test';
import { BasePage } from './base.page';

export class AppPage extends BasePage {
  // Toolbar buttons
  readonly loadButton: Locator;
  readonly stepButton: Locator;
  readonly stepOverButton: Locator;
  readonly stepOutButton: Locator;
  readonly runButton: Locator;
  readonly pauseButton: Locator;
  readonly resetButton: Locator;

  // Tab selectors
  readonly sourceTab: Locator;
  readonly disassemblyTab: Locator;
  readonly outputTab: Locator;
  readonly breakpointsTab: Locator;
  readonly statusTab: Locator;
  readonly expressionsTab: Locator;

  // Content areas
  readonly sourceView: Locator;
  readonly disassemblyView: Locator;
  readonly registerView: Locator;
  readonly memoryView: Locator;
  readonly stackView: Locator;
  readonly outputView: Locator;
  readonly commandInput: Locator;

  constructor(page: Page) {
    super(page);

    // Initialize toolbar buttons
    this.loadButton = page.getByRole('button', { name: 'Load' });
    this.stepButton = page.getByRole('button', { name: 'Step' });
    this.stepOverButton = page.getByRole('button', { name: 'Step Over' });
    this.stepOutButton = page.getByRole('button', { name: 'Step Out' });
    this.runButton = page.getByRole('button', { name: 'Run' });
    this.pauseButton = page.getByRole('button', { name: 'Pause' });
    this.resetButton = page.getByRole('button', { name: 'Reset' });

    // Initialize tabs
    this.sourceTab = page.getByRole('button', { name: 'Source' });
    this.disassemblyTab = page.getByRole('button', { name: 'Disassembly' });
    this.outputTab = page.getByRole('button', { name: 'Output' });
    this.breakpointsTab = page.getByRole('button', { name: 'Breakpoints' });
    this.statusTab = page.getByRole('button', { name: 'Status' });
    this.expressionsTab = page.getByRole('button', { name: 'Expressions' });

    // Initialize content areas (using test IDs - need to add these to components)
    this.sourceView = page.locator('[data-testid="source-view"]');
    this.disassemblyView = page.locator('[data-testid="disassembly-view"]');
    this.registerView = page.locator('[data-testid="register-view"]');
    this.memoryView = page.locator('[data-testid="memory-view"]');
    this.stackView = page.locator('[data-testid="stack-view"]');
    this.outputView = page.locator('[data-testid="output-view"]');
    this.commandInput = page.locator('[data-testid="command-input"]');
  }

  // Actions
  async clickLoad() {
    await this.loadButton.click();
  }

  async clickStep() {
    await this.stepButton.click();
  }

  async clickStepOver() {
    await this.stepOverButton.click();
  }

  async clickStepOut() {
    await this.stepOutButton.click();
  }

  async clickRun() {
    await this.runButton.click();
  }

  async clickPause() {
    await this.pauseButton.click();
  }

  async clickReset() {
    await this.resetButton.click();
  }

  async switchToSourceView() {
    await this.sourceTab.click();
  }

  async switchToDisassemblyView() {
    await this.disassemblyTab.click();
  }

  async switchToOutputTab() {
    await this.outputTab.click();
  }

  async switchToBreakpointsTab() {
    await this.breakpointsTab.click();
  }

  async switchToStatusTab() {
    await this.statusTab.click();
  }

  async switchToExpressionsTab() {
    await this.expressionsTab.click();
  }

  async enterCommand(command: string) {
    await this.commandInput.fill(command);
    await this.commandInput.press('Enter');
  }

  async getRegisterValue(register: string): Promise<string> {
    const regLocator = this.registerView.locator(`[data-register="${register}"]`);
    return await regLocator.textContent() || '';
  }

  async getOutputText(): Promise<string> {
    return await this.outputView.textContent() || '';
  }

  // Keyboard shortcuts
  async pressF5() {
    await this.page.keyboard.press('F5');
  }

  async pressF9() {
    await this.page.keyboard.press('F9');
  }

  async pressF10() {
    await this.page.keyboard.press('F10');
  }

  async pressF11() {
    await this.page.keyboard.press('F11');
  }
}
```

#### 2.3 Component-Specific Page Objects

Create page objects for complex components:

**RegisterView** (`gui/frontend/e2e/pages/register-view.page.ts`):
```typescript
import { Locator, Page } from '@playwright/test';

export class RegisterViewPage {
  private readonly container: Locator;

  constructor(page: Page) {
    this.container = page.locator('[data-testid="register-view"]');
  }

  async getRegisterValue(register: string): Promise<string> {
    const value = await this.container
      .locator(`[data-register="${register}"]`)
      .locator('.register-value')
      .textContent();
    return value?.trim() || '';
  }

  async getAllRegisters(): Promise<Record<string, string>> {
    const registers: Record<string, string> = {};
    const regElements = await this.container.locator('[data-register]').all();

    for (const element of regElements) {
      const name = await element.getAttribute('data-register');
      const value = await element.locator('.register-value').textContent();
      if (name && value) {
        registers[name] = value.trim();
      }
    }

    return registers;
  }

  async getCPSRFlags(): Promise<{ N: boolean; Z: boolean; C: boolean; V: boolean }> {
    const flags = await this.container.locator('[data-testid="cpsr-flags"]').textContent();
    return {
      N: flags?.includes('N') || false,
      Z: flags?.includes('Z') || false,
      C: flags?.includes('C') || false,
      V: flags?.includes('V') || false,
    };
  }

  async scrollToRegister(register: string) {
    await this.container.locator(`[data-register="${register}"]`).scrollIntoViewIfNeeded();
  }
}
```

**MemoryView** (`gui/frontend/e2e/pages/memory-view.page.ts`):
```typescript
import { Locator, Page } from '@playwright/test';

export class MemoryViewPage {
  private readonly container: Locator;
  private readonly addressInput: Locator;
  private readonly goButton: Locator;

  constructor(page: Page) {
    this.container = page.locator('[data-testid="memory-view"]');
    this.addressInput = this.container.locator('[data-testid="address-input"]');
    this.goButton = this.container.locator('[data-testid="go-button"]');
  }

  async goToAddress(address: string) {
    await this.addressInput.fill(address);
    await this.goButton.click();
  }

  async readMemoryAt(address: string): Promise<string> {
    await this.goToAddress(address);
    const value = await this.container
      .locator(`[data-address="${address}"]`)
      .locator('.memory-value')
      .textContent();
    return value?.trim() || '';
  }

  async scrollToAddress(address: string) {
    await this.container.locator(`[data-address="${address}"]`).scrollIntoViewIfNeeded();
  }

  async getVisibleMemoryRange(): Promise<{ start: string; end: string }> {
    const firstVisible = await this.container.locator('[data-address]').first().getAttribute('data-address');
    const lastVisible = await this.container.locator('[data-address]').last().getAttribute('data-address');
    return {
      start: firstVisible || '0x00000000',
      end: lastVisible || '0x00000000',
    };
  }
}
```

### Phase 3: Test Infrastructure

#### 3.1 Fixtures and Test Data

Create `gui/frontend/e2e/fixtures/programs.ts`:

```typescript
export const TEST_PROGRAMS = {
  hello: `
    .text
    .global _start
_start:
    MOV R0, #msg
    SWI #0x02          ; WRITE_STRING
    SWI #0x00          ; EXIT

.data
msg: .ascii "Hello, World!\\n"
     .byte 0
`,

  fibonacci: `
    .text
    .global _start
_start:
    MOV R0, #10        ; Calculate 10 Fibonacci numbers
    MOV R1, #0         ; First number
    MOV R2, #1         ; Second number
loop:
    CMP R0, #0
    BEQ done
    MOV R3, R1
    ADD R1, R1, R2
    MOV R2, R3
    SUB R0, R0, #1
    B loop
done:
    SWI #0x00          ; EXIT
`,

  infinite_loop: `
    .text
    .global _start
_start:
    MOV R0, #0
loop:
    ADD R0, R0, #1
    B loop
`,
};
```

#### 3.2 Test Utilities

Create `gui/frontend/e2e/utils/helpers.ts`:

```typescript
import { Page } from '@playwright/test';
import { AppPage } from '../pages/app.page';

export async function loadProgram(page: AppPage, program: string) {
  // This would interact with the Load dialog
  // Implementation depends on how the load dialog works
  await page.clickLoad();
  // TODO: Fill in program or select file
}

export async function waitForExecution(page: Page, maxWait = 5000) {
  // Wait for execution state to change
  await page.waitForFunction(
    () => {
      // Check for running state indicator
      return document.querySelector('[data-execution-state="running"]') === null;
    },
    { timeout: maxWait }
  );
}

export async function stepUntilAddress(page: AppPage, targetAddress: string, maxSteps = 100) {
  for (let i = 0; i < maxSteps; i++) {
    const pc = await page.getRegisterValue('PC');
    if (pc === targetAddress) {
      return true;
    }
    await page.clickStep();
  }
  return false;
}

export function formatAddress(address: number): string {
  return `0x${address.toString(16).toUpperCase().padStart(8, '0')}`;
}
```

### Phase 4: Test Cases

#### 4.1 Basic Smoke Tests

Create `gui/frontend/e2e/tests/smoke.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';

test.describe('Smoke Tests', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should load the application', async ({ page }) => {
    await expect(page).toHaveTitle(/ARM Emulator/);
  });

  test('should display all main UI elements', async () => {
    // Check toolbar buttons
    await expect(appPage.loadButton).toBeVisible();
    await expect(appPage.stepButton).toBeVisible();
    await expect(appPage.stepOverButton).toBeVisible();
    await expect(appPage.runButton).toBeVisible();
    await expect(appPage.resetButton).toBeVisible();

    // Check tabs
    await expect(appPage.sourceTab).toBeVisible();
    await expect(appPage.disassemblyTab).toBeVisible();
    await expect(appPage.outputTab).toBeVisible();
    await expect(appPage.breakpointsTab).toBeVisible();

    // Check views
    await expect(appPage.registerView).toBeVisible();
    await expect(appPage.memoryView).toBeVisible();
    await expect(appPage.stackView).toBeVisible();
  });

  test('should switch between tabs', async () => {
    await appPage.switchToDisassemblyView();
    await expect(appPage.disassemblyView).toBeVisible();

    await appPage.switchToSourceView();
    await expect(appPage.sourceView).toBeVisible();

    await appPage.switchToBreakpointsTab();
    await expect(appPage.breakpointsTab).toHaveClass(/active/);
  });

  test('should respond to keyboard shortcuts', async () => {
    // F11 = Step
    await appPage.pressF11();
    // Verify step occurred (would need to check state change)

    // F10 = Step Over
    await appPage.pressF10();
    // Verify step over occurred

    // F5 = Run
    await appPage.pressF5();
    // Verify execution started
  });
});
```

#### 4.2 Program Execution Tests

Create `gui/frontend/e2e/tests/execution.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';

test.describe('Program Execution', () => {
  let appPage: AppPage;
  let registerView: RegisterViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    registerView = new RegisterViewPage(page);
    await appPage.goto();
  });

  test('should execute hello world program', async () => {
    // Load program
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run program
    await appPage.clickRun();

    // Switch to output tab
    await appPage.switchToOutputTab();

    // Verify output
    const output = await appPage.getOutputText();
    expect(output).toContain('Hello, World!');
  });

  test('should step through fibonacci program', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Get initial PC
    const initialPC = await registerView.getRegisterValue('PC');

    // Step once
    await appPage.clickStep();

    // Verify PC changed
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);

    // Step through several instructions
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
    }

    // Verify registers changed
    const r0 = await registerView.getRegisterValue('R0');
    expect(r0).not.toBe('0x00000000');
  });

  test('should pause infinite loop', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.infinite_loop);

    // Start execution
    await appPage.clickRun();

    // Wait a bit
    await new Promise(resolve => setTimeout(resolve, 500));

    // Pause
    await appPage.clickPause();

    // Verify we can step after pause
    const pc = await registerView.getRegisterValue('PC');
    await appPage.clickStep();
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(pc);
  });

  test('should reset program state', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute several steps
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
    }

    // Get current register state
    const beforeReset = await registerView.getAllRegisters();

    // Reset
    await appPage.clickReset();

    // Verify registers reset
    const afterReset = await registerView.getAllRegisters();
    expect(afterReset['R0']).toBe('0x00000000');
    expect(afterReset['R1']).toBe('0x00000000');
  });
});
```

#### 4.3 Breakpoint Tests

Create `gui/frontend/e2e/tests/breakpoints.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';

test.describe('Breakpoints', () => {
  let appPage: AppPage;
  let registerView: RegisterViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    registerView = new RegisterViewPage(page);
    await appPage.goto();
  });

  test('should set breakpoint via F9', async () => {
    // Load program
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Step to a specific line
    await appPage.clickStep();
    await appPage.clickStep();

    // Set breakpoint with F9
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();

    // Verify breakpoint appears in list
    const breakpointsList = await appPage.page.locator('[data-testid="breakpoints-list"]');
    await expect(breakpointsList.locator('.breakpoint-item')).toHaveCount(1);
  });

  test('should stop at breakpoint during run', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Step to address in loop
    await stepUntilAddress(appPage, '0x00008008', 20);

    // Set breakpoint
    await appPage.pressF9();

    // Reset and run
    await appPage.clickReset();
    await appPage.clickRun();

    // Should stop at breakpoint
    await waitForExecution(appPage.page);
    const pc = await registerView.getRegisterValue('PC');
    expect(pc).toBe('0x00008008');
  });

  test('should toggle breakpoint on/off', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(1);

    // Toggle off (press F9 again)
    await appPage.pressF9();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });
});
```

#### 4.4 Memory and Stack Tests

Create `gui/frontend/e2e/tests/memory.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { MemoryViewPage } from '../pages/memory-view.page';

test.describe('Memory View', () => {
  let appPage: AppPage;
  let memoryView: MemoryViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    memoryView = new MemoryViewPage(page);
    await appPage.goto();
  });

  test('should navigate to specific address', async () => {
    const targetAddress = '0x00008000';
    await memoryView.goToAddress(targetAddress);

    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBe(targetAddress);
  });

  test('should display memory changes after execution', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Read stack pointer initial value
    const sp = await appPage.getRegisterValue('SP');

    // Check initial stack memory
    const initialMem = await memoryView.readMemoryAt(sp);

    // Execute some instructions
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Check if memory changed (if instructions wrote to stack)
    const newMem = await memoryView.readMemoryAt(sp);
    // Memory might have changed depending on program
  });

  test('should scroll through memory', async ({ page }) => {
    await memoryView.goToAddress('0x00000000');

    // Scroll down in memory view
    const container = page.locator('[data-testid="memory-view"]');
    await container.evaluate(node => {
      node.scrollTop = node.scrollHeight / 2;
    });

    // Verify scroll position changed
    const scrollTop = await container.evaluate(node => node.scrollTop);
    expect(scrollTop).toBeGreaterThan(0);
  });
});
```

#### 4.5 Integration Tests with Example Programs

Create `gui/frontend/e2e/tests/examples.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import * as fs from 'fs';
import * as path from 'path';

const EXAMPLES_DIR = path.join(__dirname, '../../../examples');

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
      const program = fs.readFileSync(programPath, 'utf-8');
      await loadProgram(appPage, program);

      // Run program
      await appPage.clickRun();

      // Wait for completion
      await waitForExecution(page, 10000);

      // Verify program completed (check for EXIT)
      await appPage.switchToStatusTab();
      const status = await appPage.page.locator('[data-testid="execution-status"]').textContent();
      expect(status).toContain('Exited');
    });
  }
});
```

### Phase 5: Visual Regression Testing

#### 5.1 Screenshot Comparisons

Create `gui/frontend/e2e/tests/visual.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';

test.describe('Visual Regression', () => {
  test('should match initial state screenshot', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();

    // Take screenshot and compare with baseline
    await expect(page).toHaveScreenshot('initial-state.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('should match register view after execution', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Screenshot just the register view
    await expect(appPage.registerView).toHaveScreenshot('register-view-after-steps.png');
  });

  test('should match dark mode (if implemented)', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();

    // Toggle dark mode (if available)
    // await page.locator('[data-testid="theme-toggle"]').click();

    await expect(page).toHaveScreenshot('dark-mode.png', {
      fullPage: true,
    });
  });
});
```

### Phase 6: CI/CD Integration

#### 6.1 GitHub Actions Workflow

Create `.github/workflows/e2e-tests.yml`:

```yaml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    timeout-minutes: 60
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        browser: [chromium, webkit, firefox]

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: gui/frontend/package-lock.json

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install frontend dependencies
        working-directory: gui/frontend
        run: npm ci

      - name: Install Playwright Browsers
        working-directory: gui/frontend
        run: npx playwright install --with-deps ${{ matrix.browser }}

      - name: Build Go backend
        working-directory: gui
        run: go build -o arm-emulator-gui

      - name: Run E2E tests
        working-directory: gui/frontend
        run: npm run test:e2e -- --project=${{ matrix.browser }}

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: playwright-report-${{ matrix.os }}-${{ matrix.browser }}
          path: gui/frontend/playwright-report/
          retention-days: 30

      - name: Upload test videos
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: test-videos-${{ matrix.os }}-${{ matrix.browser }}
          path: gui/frontend/test-results/
          retention-days: 7
```

### Phase 7: Headless vs Headed Mode

#### 7.1 Configuration Options

The Playwright config already supports both modes. Control via command line:

**Headless mode (default):**
```bash
npm run test:e2e
```

**Headed mode (see browser window):**
```bash
npm run test:e2e:headed
```

**Debug mode (interactive stepping):**
```bash
npm run test:e2e:debug
```

**UI mode (visual test runner):**
```bash
npm run test:e2e:ui
```

#### 7.2 Environment-Specific Configuration

Create `gui/frontend/e2e/playwright.env.ts`:

```typescript
export const testConfig = {
  // Use headed mode for local development
  headless: process.env.CI ? true : process.env.HEADED !== '1',

  // Slow down for debugging
  slowMo: process.env.SLOW_MO ? parseInt(process.env.SLOW_MO) : 0,

  // Video recording
  video: process.env.CI ? 'retain-on-failure' : 'off',

  // Screenshot on failure
  screenshot: process.env.CI ? 'only-on-failure' : 'off',
};
```

### Phase 8: Mocking Wails Backend

For testing against the dev server without a running Wails backend:

#### 8.1 Backend Mock Service

Create `gui/frontend/e2e/mocks/wails-mock.ts`:

```typescript
import { Page } from '@playwright/test';

export async function mockWailsBackend(page: Page) {
  // Intercept Wails runtime calls
  await page.addInitScript(() => {
    // Mock the Wails runtime
    (window as any).go = {
      main: {
        App: {
          Step: async () => ({ success: true }),
          StepOver: async () => ({ success: true }),
          StepOut: async () => ({ success: true }),
          Continue: async () => ({ success: true }),
          Pause: async () => ({ success: true }),
          Reset: async () => ({ success: true }),
          GetRegisters: async () => ({
            R0: 0, R1: 0, R2: 0, R3: 0,
            R4: 0, R5: 0, R6: 0, R7: 0,
            R8: 0, R9: 0, R10: 0, R11: 0,
            R12: 0, SP: 0x50000, LR: 0, PC: 0x8000,
            CPSR: 0,
          }),
          GetMemory: async (address: number, length: number) => {
            return new Array(length).fill(0);
          },
          LoadProgramFromFile: async () => ({ success: true }),
          ToggleBreakpoint: async (address: number) => ({ success: true }),
        },
      },
    };
  });
}
```

Use in tests:

```typescript
test.beforeEach(async ({ page }) => {
  await mockWailsBackend(page);
  appPage = new AppPage(page);
  await appPage.goto();
});
```

## Best Practices

### 1. Test Data Management

- Store test programs in `e2e/fixtures/`
- Use actual example programs from `examples/` directory
- Create minimal reproducible test cases for specific scenarios

### 2. Test Organization

- Group related tests with `test.describe()`
- Use clear, descriptive test names
- Follow AAA pattern: Arrange, Act, Assert
- Keep tests independent and idempotent

### 3. Waiting Strategies

- Prefer auto-waiting over manual waits
- Use `waitForLoadState`, `waitForSelector` when needed
- Avoid `waitForTimeout` except for race conditions
- Use `waitForFunction` for custom conditions

### 4. Selectors

- Prefer semantic selectors (role, label, text)
- Add `data-testid` attributes for unique elements
- Avoid CSS class selectors (they change)
- Use `data-*` attributes for dynamic content

### 5. Page Object Model

- Encapsulate page interactions
- Keep tests readable and maintainable
- Reuse common actions across tests
- Update POM when UI changes

### 6. Performance

- Run tests in parallel when possible
- Use `test.describe.parallel()` for independent test groups
- Minimize network requests in tests
- Reuse browser contexts when safe

### 7. Debugging

- Use `page.pause()` to pause test execution
- Enable trace viewer for failed tests
- Capture screenshots and videos on failure
- Use `--debug` flag for interactive debugging

### 8. Maintenance

- Keep Playwright updated
- Review and update baselines for visual tests
- Refactor tests when UI changes significantly
- Monitor test flakiness and investigate root causes

## Implementation Timeline

### Week 1: Setup and Infrastructure
- Install Playwright and dependencies
- Create configuration and directory structure
- Set up basic Page Object Models
- Create initial smoke tests

### Week 2: Core Test Cases
- Implement execution tests
- Add breakpoint tests
- Create memory/stack tests
- Develop test utilities

### Week 3: Advanced Features
- Add visual regression tests
- Implement example program tests
- Create backend mocking layer
- Add keyboard shortcut tests

### Week 4: CI/CD and Polish
- Set up GitHub Actions workflow
- Configure test reporting
- Add documentation
- Train team on writing E2E tests

## Success Metrics

- **Coverage**: Test all major user workflows
- **Reliability**: < 1% flaky tests
- **Speed**: Full suite completes in < 10 minutes
- **Maintainability**: Tests remain stable across UI changes
- **CI Integration**: Automated testing on all PRs

## Required Code Changes

To support E2E testing, add `data-testid` attributes to components:

### RegisterView.tsx
```typescript
<div data-testid="register-view">
  <div data-register="R0" data-testid="register-r0">
    <span className="register-name">R0</span>
    <span className="register-value">{registers.R0}</span>
  </div>
  {/* ... */}
</div>
```

### MemoryView.tsx
```typescript
<div data-testid="memory-view">
  <input data-testid="address-input" />
  <button data-testid="go-button">Go</button>
  <div className="memory-content">
    <div data-address="0x00008000" className="memory-row">
      <span className="memory-address">0x00008000</span>
      <span className="memory-value">0x00000000</span>
    </div>
  </div>
</div>
```

### App.tsx
```typescript
<div data-testid="output-view" data-execution-state={executionState}>
  {/* content */}
</div>
```

## Conclusion

This comprehensive E2E testing plan provides:

1. **Robust infrastructure** using Playwright with support for multiple browsers
2. **Maintainable tests** using Page Object Model pattern
3. **Flexible execution** with headless and headed modes
4. **CI/CD integration** for automated testing
5. **Visual regression** capabilities for UI consistency
6. **Complete coverage** of GUI functionality

Implementation of this plan will ensure the Wails GUI is thoroughly tested, reliable, and maintainable as the project evolves.
