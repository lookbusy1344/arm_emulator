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
