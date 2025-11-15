import { defineConfig, devices } from '@playwright/test';

const PORT = process.env.PORT || 34115;
const BASE_URL = process.env.BASE_URL || `http://localhost:${PORT}`;

export default defineConfig({
  testDir: './e2e/tests',

  // Run tests serially (not in parallel) because Wails backend has single VM instance
  fullyParallel: false,

  // Fail the build on CI if you accidentally left test.only
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Always run tests serially to avoid VM state conflicts
  workers: 1,

  // Reporter to use
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/junit.xml' }],
    ['list'] // Console output
  ],

  use: {
    // Base URL to use in actions like `await page.goto('/')`
    baseURL: BASE_URL,

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Default timeout for each action
    actionTimeout: 10000,
  },

  // Visual comparison settings
  expect: {
    toHaveScreenshot: {
      // Tightened to 3% pixel difference to catch more regressions while still allowing for
      // minor font rendering variations across CI environments (was 6%, reduced for better detection)
      maxDiffPixelRatio: 0.03,
      // Per-pixel color difference threshold (0-1, where 0.1 = 10% color difference per pixel)
      // Tightened from 0.2 (20%) to 0.15 (15%) for better regression detection
      threshold: 0.15,
    },
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    // Mobile Safari removed - Wails desktop apps don't make sense on mobile viewports
    // and this device is not tested in CI
    // {
    //   name: 'Mobile Safari',
    //   use: { ...devices['iPhone 13'] },
    // },
  ],

  // Wails dev server must be started manually with "wails dev" before running tests
  // We don't use webServer.command because:
  // 1. "npm run dev" starts Vite (port 5173), not Wails backend (port 34115)
  // 2. Wails needs "wails dev" which integrates backend + frontend
  // 3. Tests require the full Wails stack to be running
  //
  // To run tests:
  //   Terminal 1: cd gui && wails dev -nocolour
  //   Terminal 2: cd gui/frontend && npm run test:e2e
});
