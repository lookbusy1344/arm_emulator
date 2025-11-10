import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram } from '../utils/helpers';
import { TIMEOUTS } from '../utils/test-constants';

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

  test('should match memory view screenshot', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Take screenshot of memory view
    await expect(appPage.memoryView).toHaveScreenshot('memory-view.png');
  });

  test('should match source view with program loaded', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Switch to source view
    await appPage.switchToSourceView();

    // Take screenshot of source view
    await expect(appPage.sourceView).toHaveScreenshot('source-view-with-program.png');
  });

  test('should match disassembly view', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Switch to disassembly view
    await appPage.switchToDisassemblyView();

    // Take screenshot of disassembly view
    await expect(appPage.disassemblyView).toHaveScreenshot('disassembly-view.png');
  });

  test('should match output view with program output', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run program to generate output
    await appPage.clickRun();

    // Wait for execution to complete
    await page.waitForFunction(() => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return false;
      const status = statusElement.textContent?.toLowerCase() || '';
      return status === 'halted' || status === 'exited';
    }, { timeout: TIMEOUTS.EXECUTION_NORMAL });

    // Switch to output tab
    await appPage.switchToOutputTab();

    // Take screenshot of output view
    await expect(appPage.outputView).toHaveScreenshot('output-view-with-text.png');
  });

  test('should match breakpoints tab', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set some breakpoints
    await appPage.pressF9();
    await appPage.clickStep();
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();

    // Take screenshot of breakpoints tab
    const breakpointsTab = page.locator('[data-testid="breakpoints-tab-content"]');
    await expect(breakpointsTab).toHaveScreenshot('breakpoints-tab.png');
  });

  test('should match status tab', async ({ page }) => {
    // Skip in CI: Status tab renders with 2px height difference between local macOS
    // and GitHub Actions macOS runners (145px vs 143px) due to font rendering variations.
    test.skip(!!process.env.CI, 'Skipped in CI due to cross-environment rendering differences');

    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run program
    await appPage.clickRun();

    // Wait for execution to complete
    await page.waitForFunction(() => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return false;
      const status = statusElement.textContent?.toLowerCase() || '';
      return status === 'halted' || status === 'exited';
    }, { timeout: TIMEOUTS.EXECUTION_NORMAL });

    // Switch to status tab
    await appPage.switchToStatusTab();

    // Take screenshot of status tab
    const statusTab = page.locator('[data-testid="status-tab-content"]');
    await expect(statusTab).toHaveScreenshot('status-tab.png');
  });
});

test.describe('Visual Regression - Toolbar', () => {
  test('should match toolbar in initial state', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();

    // Take screenshot of toolbar
    const toolbar = page.locator('[data-testid="toolbar"]');
    await expect(toolbar).toHaveScreenshot('toolbar-initial.png');
  });

  test('should match toolbar with program loaded', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Take screenshot of toolbar (buttons may have different states)
    const toolbar = page.locator('[data-testid="toolbar"]');
    await expect(toolbar).toHaveScreenshot('toolbar-program-loaded.png');
  });

  test('should match toolbar during execution', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.infinite_loop);

    // Start execution
    await appPage.clickRun();

    // Wait for cycles to increase (more reliable than status text)
    await page.waitForFunction(() => {
      const cyclesElement = document.querySelector('.status-cycles');
      if (!cyclesElement) return false;
      const match = cyclesElement.textContent?.match(/Cycles: (\d+)/);
      return match && parseInt(match[1]) > 10;
    }, { timeout: TIMEOUTS.WAIT_FOR_STATE });

    // Take screenshot of toolbar during execution
    const toolbar = page.locator('[data-testid="toolbar"]');
    await expect(toolbar).toHaveScreenshot('toolbar-executing.png');

    // Pause to clean up
    await appPage.clickPause();
  });
});

test.describe('Visual Regression - Responsive Layout', () => {
  test('should match layout on tablet viewport', async ({ page }) => {
    const appPage = new AppPage(page);

    // Set tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });

    await appPage.goto();
    await appPage.waitForLoad();

    // Take full page screenshot
    await expect(page).toHaveScreenshot('layout-tablet.png', {
      fullPage: true,
    });
  });

  test('should match layout on mobile viewport', async ({ page }) => {
    const appPage = new AppPage(page);

    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await appPage.goto();
    await appPage.waitForLoad();

    // Take full page screenshot
    await expect(page).toHaveScreenshot('layout-mobile.png', {
      fullPage: true,
    });
  });

  test('should match layout on large desktop', async ({ page }) => {
    const appPage = new AppPage(page);

    // Set large desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });

    await appPage.goto();
    await appPage.waitForLoad();

    // Take full page screenshot
    await expect(page).toHaveScreenshot('layout-large-desktop.png', {
      fullPage: true,
    });
  });
});

test.describe('Visual Regression - Execution States', () => {
  test('should match UI in paused state', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute some steps
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Take screenshot in paused state
    await expect(page).toHaveScreenshot('state-paused.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('should match UI after program completion', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run to completion
    await appPage.clickRun();

    // Wait for execution to complete
    await page.waitForFunction(() => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return false;
      const status = statusElement.textContent?.toLowerCase() || '';
      return status === 'halted' || status === 'exited';
    }, { timeout: TIMEOUTS.EXECUTION_NORMAL });

    // Take screenshot after completion
    await expect(page).toHaveScreenshot('state-completed.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('should match UI at breakpoint', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.pressF9();

    // Restart and run to breakpoint (Restart preserves program and breakpoints)
    await appPage.clickRestart();
    await appPage.clickRun();

    // Wait for execution to complete (fibonacci completes quickly)
    await page.waitForFunction(() => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return false;
      const status = statusElement.textContent?.toLowerCase() || '';
      return status === 'paused' || status === 'halted';
    }, { timeout: TIMEOUTS.EXECUTION_NORMAL });

    // Take screenshot at breakpoint
    await expect(page).toHaveScreenshot('state-at-breakpoint.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });
});

test.describe('Visual Regression - Themes', () => {
  test.skip('should match dark mode', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();

    // Toggle dark mode (if available)
    const themeToggle = page.locator('[data-testid="theme-toggle"]');
    if (await themeToggle.isVisible()) {
      await themeToggle.click();

      // Take screenshot in dark mode
      await expect(page).toHaveScreenshot('dark-mode.png', {
        fullPage: true,
      });
    } else {
      test.skip();
    }
  });

  test.skip('should match light mode', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();

    // Ensure light mode is active
    // Take screenshot in light mode
    await expect(page).toHaveScreenshot('light-mode.png', {
      fullPage: true,
    });
  });
});

test.describe('Visual Regression - Component States', () => {
  test('should match register view with changed values', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Execute to change register values
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
    }

    // Take screenshot of register view with changes
    await expect(appPage.registerView).toHaveScreenshot('register-view-changed.png');
  });

  test('should match memory view with data', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Execute some instructions
    await appPage.clickStep();
    await appPage.clickStep();

    // Take screenshot of memory view
    await expect(appPage.memoryView).toHaveScreenshot('memory-view-with-data.png');
  });

  test('should match stack view during execution', async ({ page }) => {
    const appPage = new AppPage(page);
    await appPage.goto();
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute some instructions that use stack
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
    }

    // Take screenshot of stack view
    await expect(appPage.stackView).toHaveScreenshot('stack-view-during-execution.png');
  });
});
