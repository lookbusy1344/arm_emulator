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
      // Allow up to 6% pixel difference to account for font rendering variations across CI environments
      maxDiffPixelRatio: 0.06,
      // Per-pixel color difference threshold (0-1, where 0.1 = 10% color difference per pixel)
      threshold: 0.2,
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

    // Test against mobile viewports (optional)
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 13'] },
    },
  ],

  // Run dev server before starting tests (only in local development)
  // In CI, the workflow manually starts Wails dev server
  webServer: process.env.CI ? undefined : {
    command: 'npm run dev',
    port: PORT as number,
    reuseExistingServer: true,
    timeout: 120000,
  },
});
