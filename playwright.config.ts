import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? 'http://127.0.0.1:4174';
const useExistingServer = process.env.PLAYWRIGHT_USE_EXISTING_SERVER === '1';

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30_000,
  expect: {
    timeout: 5_000
  },
  use: {
    baseURL,
    screenshot: 'only-on-failure',
    trace: 'on-first-retry'
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Pixel 7']
      }
    }
  ],
  webServer: useExistingServer
    ? undefined
    : [
        {
          command: "sh -c 'API_ADDR=127.0.0.1:8080 pnpm dev:api'",
          url: 'http://127.0.0.1:8080/healthz',
          reuseExistingServer: true,
          timeout: 30_000
        },
        {
          command: "sh -c 'VITE_API_BASE_URL=http://127.0.0.1:8080 pnpm --filter web dev -- --host 127.0.0.1 --port 4174'",
          url: 'http://127.0.0.1:4174',
          reuseExistingServer: true,
          timeout: 30_000
        }
      ]
});
