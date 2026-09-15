import { expect, test } from '@playwright/test'

test('unauthenticated users see the bilingual login screen', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByRole('heading', { level: 1 })).toContainText('安全地管理你的 FRP 网络')
  await page.getByRole('button', { name: 'English' }).click()
  await expect(page.getByRole('heading', { level: 1 })).toContainText('Manage your FRP network securely')
})

async function mockNodePage(page: import('@playwright/test').Page, enrollmentStatus = 201) {
  await page.addInitScript(() => localStorage.setItem('frp-panel.authenticated', '1'))
  await page.route('**/api/v1/client/list', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { clients: [] } }),
  }))
  await page.route('**/api/v1/platform/baseinfo', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { clientApiUrl: 'https://panel.example.com', clientRpcUrl: 'wss://panel.example.com' } }),
  }))
  await page.route('**/api/v1/server/list', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { servers: [] } }),
  }))
  await page.route('**/api/v2/node-routes', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ routes: [] }),
  }))
  await page.route('**/api/v2/enrollments', (route) => route.fulfill(enrollmentStatus === 201 ? {
    status: 201, contentType: 'application/json', body: JSON.stringify({ nodeId: 'owner.c.edge', token: 'one-use-token', expiresAt: '2026-09-15T12:00:00Z' }),
  } : {
    status: 500, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Enrollment failed', status: 500, detail: 'database unavailable' }),
  }))
}

async function mockServerPage(page: import('@playwright/test').Page) {
  await page.addInitScript(() => localStorage.setItem('frp-panel.authenticated', '1'))
  await page.route('**/api/v1/server/list', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { servers: [] } }),
  }))
  await page.route('**/api/v1/platform/baseinfo', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { client_api_url: 'https://panel.example.com', client_rpc_url: 'wss://panel.example.com' } }),
  }))
  await page.route('**/api/v2/server-enrollments', (route) => route.fulfill({
    status: 201, contentType: 'application/json', body: JSON.stringify({ serverId: 'owner.s.edge', token: 'one-use-token', expiresAt: '2026-09-15T12:00:00Z' }),
  }))
}

test('successful node creation closes the dialog and presents OS install tabs', async ({ page }) => {
  await mockNodePage(page)
  await page.goto('/nodes')
  await page.getByRole('button', { name: /添加节点/ }).click()
  await page.getByLabel('节点 ID').fill('edge')
  await page.getByRole('button', { name: '确认' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  await expect(page.getByText('安装 Agent')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Windows' })).toBeVisible()
})

test('failed node creation keeps the dialog open with an actionable error', async ({ page }) => {
  await mockNodePage(page, 500)
  await page.goto('/nodes')
  await page.getByRole('button', { name: /添加节点/ }).click()
  await page.getByLabel('节点 ID').fill('edge')
  await page.getByRole('button', { name: '确认' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText('database unavailable')
})

test('successful server creation closes the dialog and shows standalone Compose', async ({ page }) => {
  await mockServerPage(page)
  await page.goto('/servers')
  await page.getByRole('button', { name: /创建服务端/ }).click()
  await page.getByLabel('服务端 ID').fill('edge')
  await page.getByLabel('服务端 IP').fill('edge.example.com')
  await page.getByRole('button', { name: '创建并生成部署文件' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  await expect(page.getByText('部署 FRPS')).toBeVisible()
  await expect(page.locator('pre')).toContainText('network_mode: host')
  await expect(page.locator('pre')).toContainText('one-use-token')
})
