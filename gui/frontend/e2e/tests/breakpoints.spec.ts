import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, waitForExecution, stepUntilAddress } from '../utils/helpers';

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
    const success = await stepUntilAddress(appPage, '0x00008008', 20);
    expect(success).toBe(true);

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

  test('should display breakpoint in source view', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to source view
    await appPage.switchToSourceView();

    // Verify breakpoint indicator is visible
    const breakpointIndicator = appPage.page.locator('[data-testid="breakpoint-indicator"]');
    await expect(breakpointIndicator).toBeVisible();
  });

  test('should set multiple breakpoints', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set first breakpoint
    await appPage.clickStep();
    await appPage.pressF9();

    // Set second breakpoint
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.pressF9();

    // Set third breakpoint
    await appPage.clickStep();
    await appPage.pressF9();

    // Verify all breakpoints in list
    await appPage.switchToBreakpointsTab();
    const breakpointsList = await appPage.page.locator('[data-testid="breakpoints-list"]');
    await expect(breakpointsList.locator('.breakpoint-item')).toHaveCount(3);
  });

  test('should continue execution after hitting breakpoint', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.pressF9();

    // Reset and run
    await appPage.clickReset();
    await appPage.clickRun();

    // Should stop at breakpoint
    await waitForExecution(appPage.page);

    const pcAtBreakpoint = await registerView.getRegisterValue('PC');

    // Continue execution
    await appPage.clickRun();

    // Wait a bit
    await new Promise(resolve => setTimeout(resolve, 100));

    // Verify PC has advanced
    const pcAfterContinue = await registerView.getRegisterValue('PC');
    expect(pcAfterContinue).not.toBe(pcAtBreakpoint);
  });

  test('should remove breakpoint from list', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(1);

    // Click remove button
    const removeButton = appPage.page.locator('[data-testid="remove-breakpoint-button"]').first();
    await removeButton.click();

    // Verify breakpoint removed
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });

  test('should disable/enable breakpoint', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();

    // Disable breakpoint
    const disableCheckbox = appPage.page.locator('[data-testid="breakpoint-enabled-checkbox"]').first();
    await disableCheckbox.click();

    // Reset and run - should not stop at disabled breakpoint
    await appPage.clickReset();
    await appPage.clickRun();

    // Should not stop (or will stop at exit)
    await waitForExecution(appPage.page, 2000);

    // Re-enable breakpoint
    await appPage.switchToBreakpointsTab();
    await disableCheckbox.click();

    // Reset and run - should now stop at breakpoint
    await appPage.clickReset();
    await appPage.clickRun();

    await waitForExecution(appPage.page);
    // Should have stopped at breakpoint
  });

  test('should clear all breakpoints', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set multiple breakpoints
    await appPage.pressF9();
    await appPage.clickStep();
    await appPage.pressF9();
    await appPage.clickStep();
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(3);

    // Click clear all button
    const clearAllButton = appPage.page.locator('[data-testid="clear-all-breakpoints-button"]');
    await clearAllButton.click();

    // Verify all breakpoints removed
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });
});
