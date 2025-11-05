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
    // Wait for memory view to update after navigation
    await this.container.page().waitForTimeout(300);
  }

  async readMemoryAt(address: string): Promise<string> {
    await this.goToAddress(address);
    // Memory rows are aligned to 16-byte boundaries
    // Look for the row that contains this address
    const value = await this.container
      .locator('[data-address]')
      .first()
      .locator('.memory-hex .memory-byte')
      .first()
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
