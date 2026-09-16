import { expect, test } from '@playwright/test'

test('unauthenticated users see the bilingual login screen', async ({ page }) => {
  await page.route('**/api/v2/bootstrap-status', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }),
  }))
  await page.goto('/login')
  await expect(page.getByRole('heading', { level: 1 })).toContainText('安全地管理你的 FRP 网络')
  await page.getByRole('button', { name: 'English' }).click()
  await expect(page.getByRole('heading', { level: 1 })).toContainText('Manage your FRP network securely')
})

async function mockOverview(page: import('@playwright/test').Page) {
  for (const path of ['client/list', 'server/list', 'proxy/list_configs']) {
    await page.route(`**/api/v1/${path}`, (route) => route.fulfill({
      status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: {} }),
    }))
  }
}

test('first-run setup uses password-manager semantics and enters the console', async ({ page }) => {
  await page.route('**/api/v2/bootstrap-status', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: true, ownerExists: false, canCreateOwner: true }),
  }))
  await page.route('**/api/v1/auth/register', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, msg: 'ok', body: { status: { code: 1, message: 'ok' } } }),
  }))
  await page.route('**/api/v1/auth/login', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, msg: 'ok', body: { status: { code: 1, message: 'ok' } } }),
  }))
  await mockOverview(page)
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: '创建首个所有者账户' })).toBeVisible()
  await expect(page.getByLabel('用户名')).toHaveAttribute('autocomplete', 'username')
  await expect(page.getByLabel('密码', { exact: true })).toHaveAttribute('autocomplete', 'new-password')
  await expect(page.getByLabel('密码', { exact: true })).toHaveAttribute('id', 'new-password')
  await expect(page.getByLabel('确认密码')).toHaveAttribute('autocomplete', 'new-password')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('邮箱').fill('owner@example.com')
  await page.getByLabel('密码', { exact: true }).fill('correct-horse-battery')
  await page.getByLabel('确认密码').fill('correct-horse-battery')
  await page.getByRole('button', { name: '创建所有者并进入控制台' }).click()
  await expect(page).toHaveURL('/')
  await expect(page.getByRole('heading', { name: '概览' })).toBeVisible()
  expect(await page.evaluate(() => localStorage.getItem('frp-panel.authenticated'))).toBe('1')
})

test('invalid login shows the real authentication error instead of ok', async ({ page }) => {
  await page.route('**/api/v2/bootstrap-status', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }),
  }))
  await page.route('**/api/v1/auth/login', (route) => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, msg: 'ok', body: { status: { code: 2, message: 'invalid username or password' } } }),
  }))
  await page.goto('/login')
  await expect(page.getByLabel('密码')).toHaveAttribute('id', 'current-password')
  await page.getByLabel('用户名或邮箱').fill('owner')
  await page.getByLabel('密码').fill('wrong-password')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page.getByRole('alert')).toContainText('invalid username or password')
  await expect(page.getByRole('alert')).not.toContainText(/^ok$/)
  await expect(page).toHaveURL('/login')
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
  await expect(page.locator('pre')).toContainText('PUBLIC_URL: "https://panel.example.com"')
})

test('a routed node can create and remove a TCP tunnel', async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('frp-panel.authenticated', '1'))
  const tunnel = { name: 'web', nodeId: 'owner.c.mac', clientId: 'owner.c.mac@1', serverId: 'owner.s.edge', type: 'tcp', localHost: '127.0.0.1', localPort: 3000, remotePort: 8300, stopped: false }
  let tunnels: typeof tunnel[] = []
  await page.route('**/api/v1/client/list', (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { clients: [{ id: 'owner.c.mac' }] } }) }))
  await page.route('**/api/v1/server/list', (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 200, body: { servers: [{ id: 'owner.s.edge', ip: 'edge.example.com' }] } }) }))
  await page.route('**/api/v2/node-routes', (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ routes: [{ nodeId: 'owner.c.mac', serverIds: ['owner.s.edge'] }] }) }))
  await page.route('**/api/v2/tunnels', async (route) => {
    if (route.request().method() === 'POST') { tunnels = [tunnel]; await route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify(tunnel) }); return }
    if (route.request().method() === 'DELETE') { tunnels = []; await route.fulfill({ status: 204, body: '' }); return }
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ tunnels }) })
  })
  page.on('dialog', (dialog) => void dialog.accept())
  await page.goto('/tunnels')
  await page.getByRole('button', { name: /创建隧道/ }).click()
  await page.getByLabel('隧道名称').fill('web')
  await page.getByLabel('本地端口').fill('3000')
  await page.getByLabel('公网端口').fill('8300')
  await page.getByRole('button', { name: '确认' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  await expect(page.getByText('edge.example.com:8300')).toBeVisible()
  await page.getByRole('button', { name: '删除' }).click()
  await expect(page.getByText('暂无隧道')).toBeVisible()
})
