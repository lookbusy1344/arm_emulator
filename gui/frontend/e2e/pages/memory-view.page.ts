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
