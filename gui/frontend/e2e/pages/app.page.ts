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
    this.loadButton = page.getByRole('button', { name: 'Load', exact: true });
    this.stepButton = page.getByRole('button', { name: 'Step', exact: true });
    this.stepOverButton = page.getByRole('button', { name: 'Step Over', exact: true });
    this.stepOutButton = page.getByRole('button', { name: 'Step Out', exact: true });
    this.runButton = page.getByRole('button', { name: 'Run', exact: true });
    this.pauseButton = page.getByRole('button', { name: 'Pause', exact: true });
    this.resetButton = page.getByRole('button', { name: 'Reset', exact: true });

    // Initialize tabs
    this.sourceTab = page.getByRole('button', { name: 'Source', exact: true });
    this.disassemblyTab = page.getByRole('button', { name: 'Disassembly', exact: true });
    this.outputTab = page.getByRole('button', { name: 'Output', exact: true });
    this.breakpointsTab = page.getByRole('button', { name: 'Breakpoints', exact: true });
    this.statusTab = page.getByRole('button', { name: 'Status', exact: true });
    this.expressionsTab = page.getByRole('button', { name: 'Expressions', exact: true });

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
