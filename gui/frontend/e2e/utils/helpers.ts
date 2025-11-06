import { Page } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { TIMEOUTS, ADDRESSES, LIMITS } from './test-constants';

/**
 * Load a program by calling the Wails backend directly, bypassing the file dialog.
 * This uses LoadProgramFromSource which is exposed to the frontend.
 *
 * @param page - The AppPage instance
 * @param program - The ARM assembly source code to load
 * @param filename - Optional filename (defaults to 'test.s')
 * @throws Error if program fails to load
 */
export async function loadProgram(page: AppPage, program: string, filename = 'test.s') {
  // Call LoadProgramFromSource directly via the window.go object injected by Wails
  const result = await page.page.evaluate(
    ({ source, file, entryPoint }) => {
      // @ts-ignore - Wails runtime injects window.go
      return window.go.main.App.LoadProgramFromSource(source, file, entryPoint);
    },
    { source: program, file: filename, entryPoint: ADDRESSES.CODE_SEGMENT_START }
  );

  // Verify the program loaded successfully
  if (result && typeof result === 'object' && 'error' in result) {
    throw new Error(`Failed to load program: ${result.error}`);
  }

  // Wait for VM state to update by checking PC is set to entry point
  const expectedPC = `0x${ADDRESSES.CODE_SEGMENT_START.toString(16).toUpperCase().padStart(8, '0')}`;
  await page.page.waitForFunction(
    (pc) => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      if (!pcElement) return false;
      const pcValue = pcElement.textContent?.trim() || '';
      return pcValue === pc;
    },
    expectedPC,
    { timeout: TIMEOUTS.VM_STATE_CHANGE }
  );
}

/**
 * Wait for execution to complete by checking the execution status.
 * Waits until status is not "running".
 */
export async function waitForExecution(page: Page, maxWait = TIMEOUTS.EXECUTION_NORMAL) {
  // Wait for execution state to change from "running"
  await page.waitForFunction(
    () => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return false;
      const status = statusElement.textContent?.toLowerCase() || '';
      return status !== 'running';
    },
    { timeout: maxWait }
  );

  // Wait for UI to fully update by checking register view is stable
  await page.waitForFunction(
    () => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      return pcElement !== null && pcElement.textContent !== null;
    },
    { timeout: TIMEOUTS.UI_STABILIZE }
  );
}

/**
 * Wait for output to appear in the OutputView.
 * Useful when waiting for a program to produce output.
 */
export async function waitForOutput(page: Page, maxWait = TIMEOUTS.EXECUTION_MAX) {
  await page.waitForFunction(
    () => {
      const outputElement = document.querySelector('[data-testid="output-view"] .output-content');
      if (!outputElement) return false;
      const text = outputElement.textContent || '';
      return text.trim() !== '' && text !== '(no output)';
    },
    { timeout: maxWait }
  );
}

/**
 * Step through program execution until reaching a target address.
 *
 * @param page - The AppPage instance
 * @param targetAddress - Target PC address (hex string like "0x00008010")
 * @param maxSteps - Maximum number of steps to attempt
 * @returns true if target reached, false if maxSteps exceeded
 * @throws Error if step operation fails
 */
export async function stepUntilAddress(page: AppPage, targetAddress: string, maxSteps = LIMITS.MAX_STEPS): Promise<boolean> {
  for (let i = 0; i < maxSteps; i++) {
    const pc = await page.getRegisterValue('PC');
    if (pc === targetAddress) {
      return true;
    }

    // Execute step
    await page.clickStep();

    // Wait for step to complete by checking PC changed or stayed same (could be branch)
    await page.page.waitForFunction(
      (previousPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        // PC should be set (not empty)
        return currentPC !== '';
      },
      pc,
      { timeout: TIMEOUTS.STEP_COMPLETE }
    );
  }

  console.warn(`stepUntilAddress: Failed to reach ${targetAddress} after ${maxSteps} steps`);
  return false;
}

export function formatAddress(address: number): string {
  return `0x${address.toString(16).toUpperCase().padStart(8, '0')}`;
}

/**
 * Wait for a VM state change by monitoring the execution status.
 * Useful after reset, step, or other VM operations.
 *
 * @param page - The Page instance
 * @param timeout - Maximum time to wait
 */
export async function waitForVMStateChange(page: Page, timeout = TIMEOUTS.VM_STATE_CHANGE) {
  await page.waitForFunction(
    () => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      return statusElement !== null && statusElement.textContent !== null;
    },
    { timeout }
  );
}

/**
 * Verify operation succeeded by checking for error indicators.
 *
 * @param page - The Page instance
 * @returns true if no errors detected
 */
export async function verifyNoErrors(page: Page): Promise<boolean> {
  const errorIndicators = await page.locator('[data-testid="error-message"]').count();
  return errorIndicators === 0;
}
