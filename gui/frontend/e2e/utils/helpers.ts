import { Page } from '@playwright/test';
import { AppPage } from '../pages/app.page';

export async function loadProgram(page: AppPage, program: string) {
  // This would interact with the Load dialog
  // Implementation depends on how the load dialog works
  await page.clickLoad();
  // TODO: Fill in program or select file
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
