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
