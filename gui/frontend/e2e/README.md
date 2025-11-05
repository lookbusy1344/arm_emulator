# E2E Testing for ARM Emulator GUI

End-to-end testing infrastructure for the Wails-based GUI using Playwright.

## Overview

The E2E tests validate the complete GUI application workflow, from loading programs to stepping through execution and examining state. Tests are written in TypeScript using Playwright and follow the Page Object Model pattern for maintainability.

## Directory Structure

```
e2e/
├── fixtures/          # Test programs and data
├── mocks/             # Wails backend mocks for dev server testing
├── pages/             # Page Object Models
├── tests/             # Test specifications
├── utils/             # Helper functions
└── README.md          # This file
```

## Running Tests

### Prerequisites

Ensure dependencies are installed:

```bash
cd gui/frontend
npm install
npx playwright install
```

### Local Test Execution

**IMPORTANT:** E2E tests require the Wails backend server to be running. You need **two terminal windows**:

**Terminal 1 - Start Wails Backend:**
```bash
cd gui
wails dev -nocolour
```

Wait for the server to fully start (you'll see "Wails v2" and build completion messages). The backend must be running at `http://localhost:34115`.

**Terminal 2 - Run E2E Tests (after backend is ready):**
```bash
cd gui/frontend
npm run test:e2e -- --project=chromium
```

### Available Commands

```bash
# Run all tests headless (default)
npm run test:e2e

# Run tests with visible browser window
npm run test:e2e:headed

# Run tests in debug mode (interactive stepping)
npm run test:e2e:debug

# Open Playwright UI mode (visual test runner)
npm run test:e2e:ui

# View test report
npm run test:e2e:report
```

### Running Specific Tests

```bash
# Run only smoke tests
npm run test:e2e -- smoke.spec.ts

# Run on specific browser
npm run test:e2e -- --project=webkit

# Run with grep filter
npm run test:e2e -- --grep "should load"
```

## Test Structure

### Page Object Models

Located in `e2e/pages/`:

- **BasePage** - Common functionality for all pages
- **AppPage** - Main application interface (toolbar, tabs, views)
- **RegisterViewPage** - Register inspection and interaction
- **MemoryViewPage** - Memory navigation and inspection

### Test Specifications

Located in `e2e/tests/`:

- **smoke.spec.ts** - Basic functionality and UI presence
- **execution.spec.ts** - Program execution workflows
- **breakpoints.spec.ts** - Breakpoint management
- **memory.spec.ts** - Memory and stack inspection
- **examples.spec.ts** - Integration tests with real programs
- **visual.spec.ts** - Visual regression tests

### Test Fixtures

Located in `e2e/fixtures/`:

- **programs.ts** - Test ARM assembly programs

### Utilities

Located in `e2e/utils/`:

- **helpers.ts** - Common test utilities (loadProgram, waitForExecution, etc.)

### Mocks

Located in `e2e/mocks/`:

- **wails-mock.ts** - Mock Wails backend for dev server testing

## Writing Tests

### Example Test

```typescript
import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';

test.describe('Feature Tests', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should perform action', async () => {
    await appPage.clickStep();
    const pc = await appPage.getRegisterValue('PC');
    expect(pc).not.toBe('0x00000000');
  });
});
```

### Best Practices

1. **Use Page Object Models** - Encapsulate page interactions
2. **Prefer semantic selectors** - Use data-testid, roles, labels
3. **Avoid hardcoded waits** - Use auto-waiting and waitForFunction
4. **Keep tests independent** - Each test should run in isolation
5. **Use descriptive names** - Clear test names aid debugging

## Test Data

Test programs are defined in `e2e/fixtures/programs.ts` and include:

- `hello` - Simple hello world program
- `fibonacci` - Fibonacci sequence calculation
- `infinite_loop` - Infinite loop for pause testing
- `arithmetic` - Basic arithmetic operations

## CI/CD Integration

E2E tests run automatically on:

- Push to main or e2e branches
- Pull requests to main

Tests run on:
- Ubuntu and macOS
- Chromium, WebKit, and Firefox

See `.github/workflows/e2e-tests.yml` for configuration.

## Debugging Tests

### Interactive Debugging

```bash
# Run with Playwright Inspector
npm run test:e2e:debug

# Run with headed browser and slow motion
HEADED=1 SLOW_MO=1000 npm run test:e2e
```

### View Test Artifacts

Test artifacts are saved on failure:

- `playwright-report/` - HTML report with traces
- `test-results/` - Screenshots and videos

```bash
# Open HTML report
npm run test:e2e:report
```

### Common Issues

**Tests hang indefinitely or timeout:**
- The Wails backend is not running - start `wails dev -nocolour` in a separate terminal first
- Check that `http://localhost:34115` is accessible before running tests

**Tests fail to connect to server:**
- Ensure port 34115 is not in use by another process
- Verify `wails dev` works standalone before running tests
- Wait for the Wails server to fully initialize before running tests

**Selectors not found:**
- Verify data-testid attributes are added to components
- Check element visibility with Playwright Inspector

**Flaky tests:**
- Add explicit waits for dynamic content
- Use waitForFunction for custom conditions
- Check for race conditions
- Some timing-sensitive tests may fail on slower CI runners but pass on retry

## Adding New Tests

1. Create test spec in `e2e/tests/`
2. Use existing Page Objects or create new ones
3. Add test fixtures if needed
4. Run locally to verify
5. Update this README if adding new patterns

## Data-testid Attributes

Tests rely on data-testid attributes for stable selectors. Required attributes:

- `data-testid="register-view"` - Register view container
- `data-testid="memory-view"` - Memory view container
- `data-testid="source-view"` - Source code view
- `data-testid="output-view"` - Output console
- `data-register="R0"` - Individual registers
- `data-address="0x00008000"` - Memory addresses

See plan document for complete list.

## Resources

- [Playwright Documentation](https://playwright.dev/)
- [E2E Testing Plan](../../../docs/E2E_TESTING_PLAN.md)
- [Wails Documentation](https://wails.io/)
