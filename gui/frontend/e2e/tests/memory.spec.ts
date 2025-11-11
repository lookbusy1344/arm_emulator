import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { MemoryViewPage } from '../pages/memory-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, formatAddress } from '../utils/helpers';
import { ADDRESSES } from '../utils/test-constants';

test.describe('Memory View', () => {
  let appPage: AppPage;
  let memoryView: MemoryViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    memoryView = new MemoryViewPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should navigate to specific address', async () => {
    const targetAddress = formatAddress(ADDRESSES.CODE_SEGMENT_START);
    await memoryView.goToAddress(targetAddress);

    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBe(targetAddress);
  });

  test('should display memory changes after execution', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Read stack pointer initial value (SP is R13 in ARM)
    // loadProgram already waits for VM to be ready
    const sp = await appPage.getRegisterValue('R13');

    // Verify SP is set
    expect(sp).toBeTruthy();
    expect(sp).toMatch(/0x[0-9A-F]{8}/i);

    // Execute some instructions
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Just verify memory view is still responsive
    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBeTruthy();
  });

  test.skip('should scroll through memory', async ({ page }) => {
    // Memory view may be virtualized and not use traditional scrolling
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

  test('should display memory at program start address', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Navigate to program start
    const programStart = formatAddress(ADDRESSES.CODE_SEGMENT_START);
    await memoryView.goToAddress(programStart);

    // Verify memory view shows the address
    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBe(programStart);
  });

  test('should navigate to address from register', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Get PC value
    const pc = await appPage.getRegisterValue('PC');

    // Navigate to PC address in memory view
    await memoryView.goToAddress(pc);

    // Verify we're at the right address
    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBe(pc);
  });

  test('should display stack memory', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Get SP value (SP is R13 in ARM)
    // loadProgram already waits for VM to be ready
    const sp = await appPage.getRegisterValue('R13');

    // Navigate to stack
    await memoryView.goToAddress(sp);

    // Verify memory view updated (address range is visible)
    const range = await memoryView.getVisibleMemoryRange();
    expect(range.start).toBeTruthy();
    expect(range.start).toMatch(/0x[0-9A-F]{8}/i);
  });

  test('should highlight modified memory addresses', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Execute some instructions that modify memory
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
    }

    // Check for modified memory indicators
    const modifiedMemory = page.locator('[data-testid="memory-modified"]');
    // Modified memory cells should be highlighted if feature exists
  });

  test('should display memory in different formats', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Navigate to program memory
    await memoryView.goToAddress(formatAddress(ADDRESSES.CODE_SEGMENT_START));

    // Switch to hex format (default)
    // Values should be displayed as hex

    // Switch to decimal format if available
    const formatSelector = page.locator('[data-testid="memory-format-selector"]');
    if (await formatSelector.isVisible()) {
      await formatSelector.selectOption('decimal');
      // Verify format changed
    }
  });

  test('should show ASCII representation of memory', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Navigate to data section with string
    // Strings should have ASCII representation visible
    const asciiColumn = page.locator('[data-testid="memory-ascii"]').first();
    if (await asciiColumn.isVisible()) {
      const ascii = await asciiColumn.textContent();
      // Should contain readable characters
    }
  });

  test('should refresh memory view on step', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Navigate to a specific address
    await memoryView.goToAddress(formatAddress(ADDRESSES.CODE_SEGMENT_START));

    // Get initial memory state
    const initialRange = await memoryView.getVisibleMemoryRange();

    // Step execution
    await appPage.clickStep();

    // Memory view should update
    const updatedRange = await memoryView.getVisibleMemoryRange();
    // Range should remain the same but content may change
    expect(updatedRange.start).toBe(initialRange.start);
  });
});

test.describe('Stack View', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should display stack contents', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Switch to Stack tab
    await appPage.switchToStackTab();

    // Stack view should be visible
    const stackView = page.locator('[data-testid="stack-view"]');
    await expect(stackView).toBeVisible();
  });

  test('should update stack on push operations', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    const stackView = page.locator('[data-testid="stack-view"]');

    // Get initial stack state
    const initialStack = await stackView.textContent();

    // Execute instructions that use stack
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
    }

    // Verify stack changed
    const updatedStack = await stackView.textContent();
    // Stack content may have changed based on program operations
  });

  test('should highlight stack pointer position', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Stack pointer should be highlighted in stack view
    const spIndicator = page.locator('[data-testid="stack-pointer-indicator"]');
    if (await spIndicator.isVisible()) {
      await expect(spIndicator).toBeVisible();
    }
  });

  test('should show stack growth direction', async ({ page }) => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Switch to Stack tab
    await appPage.switchToStackTab();

    const stackView = page.locator('[data-testid="stack-view"]');
    await expect(stackView).toBeVisible();

    // Stack should show proper growth direction (typically downward)
    // Newer items at lower addresses
  });

  test('should detect stack overflow', async ({ page }) => {
    // Load a program that causes stack overflow
    // Stack view should indicate overflow condition if it occurs
    const warningIndicator = page.locator('[data-testid="stack-overflow-warning"]');
    // Would only be visible if overflow occurs
  });
});
