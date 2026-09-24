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
    baseURL: 'http://localhost:5173',
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

  // Exercise production assets against an isolated real Go service.
  webServer: [
    {
      command:
        'pnpm build && pnpm preview --host 127.0.0.1 --port 5173 --strictPort',
      url: 'http://127.0.0.1:5173',
      reuseExistingServer: true,
      timeout: 120_000,
    },
    {
      command: 'make -C ../backend run/api ARGS="-limiter-enabled=false"',
      url: 'http://127.0.0.1:4000/healthcheck',
      reuseExistingServer: true,
      timeout: 120_000,
      gracefulShutdown: { signal: 'SIGTERM', timeout: 6_000 },
    },
  ],
})
