import { defineConfig, devices } from '@playwright/test'
import dotenv from 'dotenv'

dotenv.config({ quiet: true })

const isBuildkite = !!process.env.BUILDKITE

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './e2e',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: isBuildkite,
  /* Retry on CI only */
  retries: isBuildkite ? 3 : 0,
  /* Opt out of parallel tests on CI. */
  workers: isBuildkite ? 1 : undefined,
  reporter: isBuildkite
    ? [
        ['html', { open: 'never' }],
        ['json', { outputFile: 'playwright-report/results.json' }],
      ]
    : [['html', { open: 'never' }]],
  outputDir: 'playwright-artifacts',
  /* Global test timeout - 2 minutes per test */
  timeout: 120_000,
  /* Timeout for expect assertions - 30 seconds to handle slow API responses */
  expect: {
    timeout: 30_000,
  },
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    /* Timeout for actions like click, fill, etc */
    actionTimeout: 30_000,
    /* Timeout for navigation */
    navigationTimeout: 60_000,
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=IsolateOrigins',
            '--disable-site-isolation-trials',
          ],
        },
      },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    /* Test against mobile viewports. */
    // {
    //   name: 'Mobile Chrome',
    //   use: { ...devices['Pixel 5'] },
    // },
    // {
    //   name: 'Mobile Safari',
    //   use: { ...devices['iPhone 12'] },
    // },

    /* Test against branded browsers. */
    // {
    //   name: 'Microsoft Edge',
    //   use: { ...devices['Desktop Edge'], channel: 'msedge' },
    // },
    // {
    //   name: 'Google Chrome',
    //   use: { ...devices['Desktop Chrome'], channel: 'chrome' },
    // },
  ],
})
