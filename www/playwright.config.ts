import { defineConfig } from '@playwright/test'

const liveURL = process.env.FRP_PANEL_E2E_URL

export default defineConfig({
  testDir: './e2e',
  use: { baseURL: liveURL || 'http://127.0.0.1:4173' },
  webServer: liveURL ? undefined : {
    command: 'npm run build && npm run preview -- --host 127.0.0.1',
    port: 4173,
    reuseExistingServer: false,
  },
})
