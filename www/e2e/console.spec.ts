import { expect, test } from '@playwright/test'

const user = { username: 'owner', email: 'owner@example.test', role: 'owner' }
const client = { id: 'owner.c.mac', comment: '', configurationState: 'unconfigured', status: 'pending', enabled: true, tunnelCount: 0 }
const server = { id: 'owner.s.edge', address: 'edge.example.test', bindPort: 7000, serverApiPort: 8999, comment: '', configurationState: 'unconfigured', status: 'pending', tunnelCount: 0 }
const strictCSP = "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data: blob: https://basemaps.cartocdn.com https://*.basemaps.cartocdn.com; style-src 'self' 'unsafe-inline'; worker-src 'self' blob:; connect-src 'self' ws: wss: https://basemaps.cartocdn.com https://*.basemaps.cartocdn.com https://get.geojs.io https://ipwho.is"

async function mockAPI(page: import('@playwright/test').Page) {
  await page.route('**/api/v2/bootstrap-status', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }) }))
  await page.route('**/api/v2/account', route => route.fulfill({ status: 401, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Unauthorized', detail: 'sign in required' }) }))
  await page.route('**/api/v2/auth/login', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user }) }))
  await page.route('**/api/v2/overview', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ clients: 1, servers: 1, tunnels: 0 }) }))
  await page.route('**/api/v2/topology', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ nodes: [client, server].map((item) => ({ ...item, kind: item.id.includes('.c.') ? 'client' : 'server', label: item.id, enabled: true })), links: [], locatedCount: 0, totalCount: 2, generatedAt: new Date().toISOString() }) }))
  await page.route('**/api/v2/clients**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [client], total: 1, page: 1, pageSize: 25 }) }))
  await page.route('**/api/v2/servers**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [server], total: 1, page: 1, pageSize: 25 }) }))
  await page.route('**/api/v2/tunnels**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [], total: 0, page: 1, pageSize: 25 }) }))
}

test('login renders with the production strict CSP', async ({ page }) => {
  const pageErrors: string[] = []
  page.on('pageerror', error => pageErrors.push(error.stack || String(error)))
  await page.route('**/*', async route => {
    const response = await route.fetch()
    await route.fulfill({ response, headers: { ...response.headers(), 'content-security-policy': strictCSP } })
  })
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: '登录 Master' })).toBeVisible()
  expect(pageErrors).toEqual([])
})

test('invalid sign-in shows a real error and no misleading ok response', async ({ page }) => {
  await page.route('**/api/v2/bootstrap-status', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }) }))
  await page.route('**/api/v2/account', route => route.fulfill({ status: 401, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Unauthorized', detail: 'sign in required' }) }))
  await page.route('**/api/v2/auth/login', route => route.fulfill({ status: 401, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Sign-in failed', detail: 'the username or password is incorrect' }) }))
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('wrong-password')
  await page.getByRole('button', { name: '继续' }).click()
  await expect(page.getByRole('alert')).toContainText('the username or password is incorrect')
  await expect(page.getByText('ok', { exact: true })).toHaveCount(0)
})

test('authenticated console exposes resource states and controlled dialogs', async ({ page }) => {
  await mockAPI(page)
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await expect(page).toHaveURL('/')
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Clients', exact: true }).click()
  await expect(page.getByText('owner.c.mac')).toBeVisible()
  await page.getByRole('button', { name: '添加 Client' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.getByRole('button', { name: '取消' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
})

test('Server creation exposes a configurable API port and blocks equal ports', async ({ page }) => {
  await mockAPI(page)
  let createBody: Record<string, unknown> | undefined
  await page.route('**/api/v2/servers', async route => {
    if (route.request().method() !== 'POST') {
      await route.fallback()
      return
    }
    createBody = route.request().postDataJSON() as Record<string, unknown>
    await route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({
        server: { ...server, id: 'owner.s.custom', bindPort: 7001, serverApiPort: 8998 },
        enrollment: { token: 'token', expiresAt: new Date(Date.now() + 60000).toISOString(), apiUrl: 'https://panel.example.test', rpcUrl: 'wss://panel.example.test', composeYaml: 'SERVER_API_PORT: "8998"' },
      }),
    })
  })
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Servers', exact: true }).click()
  await page.getByRole('button', { name: '添加 Server' }).click()

  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Server ID').fill('custom')
  await dialog.getByLabel('公网地址').fill('edge.example.test')
  await dialog.getByLabel('Bind port').fill('8999')
  await dialog.getByLabel('Server port').fill('8999')
  await expect(dialog.getByRole('alert')).toContainText('不能相同')
  await dialog.getByRole('button', { name: '保存' }).click()
  expect(createBody).toBeUndefined()

  const portInputs = dialog.locator('input[type="number"]')
  await portInputs.nth(0).fill('7001')
  await portInputs.nth(1).fill('8998')
  await dialog.getByRole('button', { name: '保存' }).click()
  await expect(page.getByText('SERVER_API_PORT: "8998"')).toBeVisible()
  expect(createBody).toMatchObject({ bindPort: 7001, serverApiPort: 8998 })
})

test('overview renders geolocated Client to Server Tunnel topology', async ({ page }) => {
  const pageErrors: string[] = []
  const mapFailures: string[] = []
  page.on('pageerror', error => pageErrors.push(error.stack || String(error)))
  page.on('requestfailed', request => {
    if (/(cartocdn|maplibre-gl-worker|geojs|ipwho)/.test(request.url())) mapFailures.push(`${request.url()} :: ${request.failure()?.errorText}`)
  })
  await page.route('**/*', async route => {
    if (!route.request().url().startsWith('http://127.0.0.1')) {
      await route.continue()
      return
    }
    const response = await route.fetch()
    await route.fulfill({ response, headers: { ...response.headers(), 'content-security-policy': strictCSP } })
  })
  await mockAPI(page)
  await page.route('**/api/v2/topology', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      nodes: [
        { ...client, kind: 'client', label: client.id, enabled: true, locationIp: '8.8.8.8' },
        { ...server, kind: 'server', label: server.id, enabled: true, locationIp: '1.1.1.1' },
      ],
      links: [{ id: 'tunnel-1', name: 'ssh', sourceClientId: client.id, targetServerId: server.id, type: 'tcp', remotePort: 60000, enabled: true, status: 'offline' }],
      locatedCount: 2,
      totalCount: 2,
      generatedAt: new Date().toISOString(),
    }),
  }))
  await page.route('**/v1/ip/geo/**', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ country_code: 'US', region: 'California', city: 'Mountain View', latitude: '37.4056', longitude: '-122.0775' }),
  }))
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await expect(page.getByRole('heading', { name: '网络拓扑' })).toBeVisible()
  await expect(page.locator('.topology-map-shell')).toBeVisible()
  await expect(page.locator('.maplibregl-canvas')).toBeVisible()
  await expect(page.getByText('ssh · TCP :60000')).toBeVisible()
  await page.waitForTimeout(1500)
  expect(mapFailures).toEqual([])
  expect(pageErrors).toEqual([])
})
