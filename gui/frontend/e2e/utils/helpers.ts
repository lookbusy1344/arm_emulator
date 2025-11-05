import { Page } from '@playwright/test';
import { AppPage } from '../pages/app.page';

/**
 * Load a program by calling the Wails backend directly, bypassing the file dialog.
 * This uses LoadProgramFromSource which is exposed to the frontend.
 *
 * @param page - The AppPage instance
 * @param program - The ARM assembly source code to load
 * @param filename - Optional filename (defaults to 'test.s')
 */
export async function loadProgram(page: AppPage, program: string, filename = 'test.s') {
  // CodeSegmentStart constant from vm/constants.go
  const CODE_SEGMENT_START = 0x00008000;

  // Call LoadProgramFromSource directly via the window.go object injected by Wails
  await page.page.evaluate(
    ({ source, file, entryPoint }) => {
      // @ts-ignore - Wails runtime injects window.go
      return window.go.main.App.LoadProgramFromSource(source, file, entryPoint);
    },
    { source: program, file: filename, entryPoint: CODE_SEGMENT_START }
  );

  // Wait for the vm:state-changed event or a short timeout
  // The backend emits 'vm:state-changed' after loading
  await page.page.waitForTimeout(200);
}

export async function waitForExecution(page: Page, maxWait = 5000) {
  // Wait for execution state to change
  await page.waitForFunction(
    () => {
      // Check for running state indicator
      return document.querySelector('[data-execution-state="running"]') === null;
    },
    { timeout: maxWait }
  );
}

export async function stepUntilAddress(page: AppPage, targetAddress: string, maxSteps = 100) {
  for (let i = 0; i < maxSteps; i++) {
    const pc = await page.getRegisterValue('PC');
    if (pc === targetAddress) {
      return true;
    }
    await page.clickStep();
  }
  return false;
}

export function formatAddress(address: number): string {
  return `0x${address.toString(16).toUpperCase().padStart(8, '0')}`;
}
