import { Page, Locator } from '@playwright/test';
import { BasePage } from './base.page';

export class AppPage extends BasePage {
  // Toolbar
  readonly toolbar: Locator;

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
  readonly memoryTab: Locator;
  readonly stackTab: Locator;
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

    // Initialize toolbar
    this.toolbar = page.locator('[data-testid="toolbar"]');

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
    this.memoryTab = page.getByTestId('memory-tab');
    this.stackTab = page.getByTestId('stack-tab');
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
    // Call Reset via Wails binding (clears program, breakpoints, resets VM to PC=0)
    await this.page.evaluate(() => {
      // @ts-ignore - Wails runtime
      return window.go.main.App.Reset();
    });

    // Wait for RegisterView to load (it might have been stuck in "Loading..." if backend was dirty)
    // The Reset() call above triggers state-changed event which should unstick RegisterView
    await this.page.waitForFunction(() => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      return pcElement !== null;
    }, { timeout: 10000 });

    // Now wait for PC to actually be reset to zero
    await this.page.waitForFunction(() => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      if (!pcElement) return false;
      const pcValue = pcElement.textContent?.trim() || '';
      return pcValue === '0x00000000';
    }, { timeout: 10000 });
  }

  async clickRestart() {
    // Call Restart via Wails binding (preserves program and breakpoints)
    await this.page.evaluate(() => {
      // @ts-ignore - Wails runtime
      return window.go.main.App.Restart();
    });

    // Wait for frontend to process the state-changed event and update UI
    // Check that PC has been reset to entry point (0x00008000)
    await this.page.waitForFunction(() => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      if (!pcElement) return false;
      const pcValue = pcElement.textContent?.trim() || '';
      return pcValue === '0x00008000';
    }, { timeout: 10000 });
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

  async switchToMemoryTab() {
    await this.memoryTab.click();
  }

  async switchToStackTab() {
    await this.stackTab.click();
  }

  async enterCommand(command: string) {
    await this.commandInput.fill(command);
    await this.commandInput.press('Enter');
  }

  async getRegisterValue(register: string): Promise<string> {
    const regLocator = this.registerView.locator(`[data-register="${register}"]`);
    // Get just the value span, not the entire row text
    const valueSpan = regLocator.locator('.register-value');
    return await valueSpan.textContent() || '';
  }

  async getOutputText(): Promise<string> {
    return await this.outputView.textContent() || '';
  }

  // Keyboard shortcuts
  async pressF5() {
    await this.page.keyboard.press('F5');
  }

  async pressF9() {
    // Get current breakpoint count before pressing F9
    const countBefore = await this.page.evaluate(() => {
      return document.querySelectorAll('.breakpoint-item').length;
    });

    await this.page.keyboard.press('F9');

    // Wait for breakpoint count to change (either add or remove)
    await this.page.waitForFunction(
      (before) => {
        const countNow = document.querySelectorAll('.breakpoint-item').length;
        return countNow !== before;
      },
      countBefore,
      { timeout: 2000 }
    );
  }

  async pressF10() {
    await this.page.keyboard.press('F10');
  }

  async pressF11() {
    await this.page.keyboard.press('F11');
  }
}
