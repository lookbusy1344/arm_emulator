import { Page, Locator } from '@playwright/test';

export class BasePage {
  constructor(protected page: Page) {}

  async goto() {
    await this.page.goto('/');
  }

  async waitForLoad() {
    await this.page.waitForLoadState('networkidle');

    // Wait for Wails runtime to be fully initialized
    // This ensures window.go.main.App is available before tests try to use it
    await this.page.waitForFunction(
      () => {
        // @ts-ignore - Wails runtime injects window.go
        return typeof window.go !== 'undefined'
          && typeof window.go.main !== 'undefined'
          && typeof window.go.main.App !== 'undefined';
      },
      { timeout: 30000 } // 30 second timeout for Wails runtime initialization
    );
  }

  async takeScreenshot(name: string) {
    await this.page.screenshot({ path: `test-results/${name}.png` });
  }
}
