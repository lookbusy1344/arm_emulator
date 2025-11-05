import { Page } from '@playwright/test';

export async function mockWailsBackend(page: Page) {
  // Intercept Wails runtime calls
  await page.addInitScript(() => {
    // Mock the Wails runtime
    (window as any).go = {
      main: {
        App: {
          Step: async () => ({ success: true }),
          StepOver: async () => ({ success: true }),
          StepOut: async () => ({ success: true }),
          Continue: async () => ({ success: true }),
          Pause: async () => ({ success: true }),
          Reset: async () => ({ success: true }),
          GetRegisters: async () => ({
            R0: 0, R1: 0, R2: 0, R3: 0,
            R4: 0, R5: 0, R6: 0, R7: 0,
            R8: 0, R9: 0, R10: 0, R11: 0,
            R12: 0, SP: 0x50000, LR: 0, PC: 0x8000,
            CPSR: 0,
          }),
          GetMemory: async (address: number, length: number) => {
            return new Array(length).fill(0);
          },
          LoadProgramFromFile: async () => ({ success: true }),
          ToggleBreakpoint: async (address: number) => ({ success: true }),
        },
      },
    };
  });
}
