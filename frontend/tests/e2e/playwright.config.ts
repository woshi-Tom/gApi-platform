import { defineConfig, devices } from '@playwright/test'

const frontendUrl = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173'
const apiUrl = process.env.API_BASE_URL || 'http://localhost:8080'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: frontendUrl,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  globals: {
    apiUrl: apiUrl,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
