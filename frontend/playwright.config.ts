import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  // E2E test location.
  testDir: './tests/e2e',

  // Prevent committed test.only calls from passing in CI.
  forbidOnly: !!process.env.CI,

  // Keep failures visible instead of masking flaky tests with retries.
  retries: 0,

  // Run tests sequentially for predictable local and CI behavior.
  workers: 1,

  // Human-readable and machine-readable reports.
  reporter: [
    ['html', { open: 'never', outputFolder: 'playwright-report/html' }],
    ['json', { outputFile: 'playwright-report/report.json' }],
    ['list'],
  ],

  // Shared browser settings.
  use: {
    baseURL: 'http://127.0.0.1:5173',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },

  // Cover desktop and mobile Chrome experiences.
  projects: [
    {
      name: 'desktop-chrome',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
    },
  ],

  // Start the Vite development server before running E2E tests.
  webServer: {
    command: 'pnpm dev --host 127.0.0.1',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
})
