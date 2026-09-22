import { expect, test } from '@playwright/test'

const user = { username: 'owner', email: 'owner@example.test', role: 'owner' }
const oldVersion = { gitVersion: 'v1.0.0', gitCommit: 'a'.repeat(40), buildDate: '2026-01-01T00:00:00Z', platform: 'darwin/arm64' }
const client = { id: 'owner.c.mac', comment: '', configurationState: 'unconfigured', status: 'pending', enabled: true, tunnelCount: 0, version: oldVersion }
const server = { id: 'owner.s.edge', address: 'edge.example.test', bindPort: 7000, serverApiPort: 8999, comment: '', configurationState: 'unconfigured', status: 'pending', tunnelCount: 0, version: { ...oldVersion, platform: 'linux/amd64' }, autoUpdate: false, updateZone: 'Asia/Shanghai', updateStart: '03:00', updateEnd: '04:00' }
const strictCSP = "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data: blob: https://basemaps.cartocdn.com https://*.basemaps.cartocdn.com; style-src 'self' 'unsafe-inline'; worker-src 'self' blob:; connect-src 'self' ws: wss: https://basemaps.cartocdn.com https://*.basemaps.cartocdn.com https://get.geojs.io https://ipwho.is"

async function mockAPI(page: import('@playwright/test').Page) {
  await page.route('**/api/v2/bootstrap-status', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }) }))
  await page.route('**/api/v2/account', route => route.fulfill({ status: 401, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Unauthorized', detail: 'sign in required' }) }))
  await page.route('**/api/v2/auth/login', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user }) }))
  await page.route('**/api/v2/overview', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ clients: 1, servers: 1, tunnels: 0 }) }))
  await page.route('**/api/v2/updates/release**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ supported: true, release: { channel: 'stable', currentVersion: 'v1.0.0', currentCommit: 'a'.repeat(40), latestVersion: 'v1.1.0', latestCommit: 'b'.repeat(40), available: true, checkedAt: new Date().toISOString() } }) }))
  await page.route('**/api/v2/topology', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ nodes: [client, server].map((item) => ({ ...item, kind: item.id.includes('.c.') ? 'client' : 'server', label: item.id, enabled: true })), links: [], locatedCount: 0, totalCount: 2, generatedAt: new Date().toISOString() }) }))
  await page.route('**/api/v2/clients**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [client], total: 1, page: 1, pageSize: 25 }) }))
  await page.route('**/api/v2/servers**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [server], total: 1, page: 1, pageSize: 25 }) }))
  await page.route('**/api/v2/tunnels**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ items: [], total: 0, page: 1, pageSize: 25 }) }))
}

test('login renders with the production strict CSP', async ({ page }) => {
  const pageErrors: string[] = []
  let unauthenticatedReleaseChecks = 0
  page.on('request', request => { if (request.url().includes('/api/v2/updates/release')) unauthenticatedReleaseChecks += 1 })
  page.on('pageerror', error => pageErrors.push(error.stack || String(error)))
  await page.route('**/*', async route => {
    const response = await route.fetch()
    await route.fulfill({ response, headers: { ...response.headers(), 'content-security-policy': strictCSP } })
  })
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: '登录 Master' })).toBeVisible()
  await expect(page.locator('link[rel="icon"][type="image/svg+xml"]')).toHaveAttribute('href', '/frppanel-logo.svg')
  expect(pageErrors).toEqual([])
  expect(unauthenticatedReleaseChecks).toBe(0)
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
  await expect(page.locator('.sidebar-footer')).toContainText('FRP PANEL')
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Clients', exact: true }).click()
  await expect(page.getByText('owner.c.mac')).toBeVisible()
  await page.getByRole('button', { name: '添加 Client' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.getByRole('button', { name: '取消' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
})

test('version controls show a local Client command and editable Server maintenance window', async ({ page }) => {
  await mockAPI(page)
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await page.getByRole('button', { name: 'v1.0.0', exact: false }).first().click()
  await expect(page.getByText('Master 版本')).toBeVisible()
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Clients', exact: true }).click()
  await page.getByRole('button', { name: '有新版本' }).click()
  await expect(page.getByRole('dialog')).toContainText('sudo /usr/local/libexec/frp-panel/frp-panel-agent update --version v1.1.0')
  await page.getByRole('dialog').getByRole('button', { name: '关闭' }).last().click()
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Servers', exact: true }).click()
  await page.getByRole('button', { name: '更新设置' }).click()
  await expect(page.getByRole('dialog').getByLabel('时区（IANA）')).toHaveValue('Asia/Shanghai')
  await page.getByRole('dialog').getByRole('button', { name: '取消' }).click()
  await page.getByRole('button', { name: '有新版本' }).click()
  await expect(page.getByRole('dialog')).toContainText('隧道可能中断')
})

test('Owner confirms Master update and Server policy saves through scoped API', async ({ page }) => {
  await mockAPI(page)
  let masterRequested = false
  let policyBody: Record<string, unknown> | undefined
  await page.route('**/api/v2/updates/master', route => {
    masterRequested = true
    return route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ operationId: 'test-operation' }) })
  })
  await page.route('**/api/v2/servers/*/update-policy', route => {
    policyBody = route.request().postDataJSON() as Record<string, unknown>
    return route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  })
  await page.route('**/api/v2/updates/operations/test-operation', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ operation: { id: 'test-operation', kind: 'master', targetId: 'master', targetCommit: 'b'.repeat(40), version: 'v1.1.0', state: 'succeeded', startedAt: new Date().toISOString(), updatedAt: new Date().toISOString() } }) }))
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await page.getByRole('button', { name: 'v1.0.0', exact: false }).first().click()
  await page.getByRole('button', { name: '立即更新' }).click()
  expect(masterRequested).toBe(false)
  await page.getByRole('dialog').getByRole('button', { name: '确认' }).click()
  await expect.poll(() => masterRequested).toBe(true)
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Servers', exact: true }).click()
  await page.getByRole('button', { name: '更新设置' }).click()
  await page.getByRole('dialog').getByLabel('窗口开始').fill('02:00')
  await page.getByRole('dialog').getByRole('button', { name: '保存' }).click()
  await expect.poll(() => policyBody).toMatchObject({ enabled: false, timeZone: 'Asia/Shanghai', windowStart: '02:00', windowEnd: '04:00' })
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

test('Tunnels waits for initial resources and shows the Server remote connection', async ({ page }) => {
  await mockAPI(page)
  const tunnel = {
    id: 'tunnel-ssh', name: 'ssh-primary', clientId: client.id, serverId: server.id,
    serverAddress: '43.134.184.42', type: 'tcp', localHost: '127.0.0.1', localPort: 22,
    remotePort: 60000, enabled: true, status: 'online', updatedAt: new Date().toISOString(),
  }
  let release!: () => void
  const gate = new Promise<void>((resolve) => { release = resolve })
  const delayedResponses: Record<string, unknown> = {
    clients: { items: [client], total: 1, page: 1, pageSize: 100 },
    servers: { items: [server], total: 1, page: 1, pageSize: 100 },
    tunnels: { items: [tunnel], total: 1, page: 1, pageSize: 25 },
  }
  for (const [resource, payload] of Object.entries(delayedResponses)) {
    await page.route(`**/api/v2/${resource}**`, async route => {
      if (route.request().method() !== 'GET') {
        await route.fallback()
        return
      }
      await gate
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(payload) })
    })
  }
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await page.getByRole('complementary', { name: 'Primary navigation' }).getByRole('link', { name: 'Tunnels', exact: true }).click()

  await expect(page.locator('.notice-card')).toHaveCount(0)
  release()
  await expect(page.getByRole('columnheader', { name: '远程连接' })).toBeVisible()
  await expect(page.getByText('43.134.184.42:60000', { exact: true })).toBeVisible()
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
        { ...client, id: 'owner.c.sfo', kind: 'client', label: 'owner.c.sfo', enabled: true, locationIp: '8.8.8.8' },
        { ...client, id: 'owner.c.shanghai', kind: 'client', label: 'owner.c.shanghai', enabled: true, locationIp: '118.25.94.27' },
        { ...server, id: 'owner.s.singapore', kind: 'server', label: 'owner.s.singapore', enabled: true, locationIp: '43.134.184.42' },
        { ...server, id: 'owner.s.dns', kind: 'server', label: 'owner.s.dns', enabled: true, locationIp: '1.1.1.1' },
      ],
      links: [
        { id: 'tunnel-1', name: 'ssh-primary', sourceClientId: 'owner.c.sfo', targetServerId: 'owner.s.singapore', type: 'tcp', remotePort: 60000, enabled: true, status: 'online' },
        { id: 'tunnel-2', name: 'ssh-backup', sourceClientId: 'owner.c.sfo', targetServerId: 'owner.s.singapore', type: 'tcp', remotePort: 60001, enabled: true, status: 'pending' },
        { id: 'tunnel-3', name: 'web', sourceClientId: 'owner.c.shanghai', targetServerId: 'owner.s.singapore', type: 'tcp', remotePort: 61000, enabled: true, status: 'offline' },
        { id: 'tunnel-4', name: 'dns', sourceClientId: 'owner.c.shanghai', targetServerId: 'owner.s.dns', type: 'udp', remotePort: 53000, enabled: true, status: 'error' },
      ],
      locatedCount: 4,
      totalCount: 4,
      generatedAt: new Date().toISOString(),
    }),
  }))
  await page.route('https://get.geojs.io/v1/ip/geo/8.8.8.8.json', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ country_code: 'US', region: 'California', city: 'Mountain View', latitude: '37.4056', longitude: '-122.0775' }) }))
  await page.route('https://get.geojs.io/v1/ip/geo/118.25.94.27.json', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ country_code: 'CN', region: 'Shanghai', city: 'Shanghai', latitude: '31.2222', longitude: '121.4581' }) }))
  await page.route('https://get.geojs.io/v1/ip/geo/43.134.184.42.json', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ country_code: 'SG', region: 'Singapore', city: 'Singapore', latitude: '1.2872', longitude: '103.8507' }) }))
  await page.route('https://get.geojs.io/v1/ip/geo/1.1.1.1.json', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ country_code: 'AU', region: 'Queensland', city: 'Brisbane', latitude: '-27.4679', longitude: '153.0281' }) }))
  await page.goto('/login')
  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('密码').fill('correct-password')
  await page.getByRole('button', { name: '继续' }).click()
  await expect(page.getByRole('heading', { name: '网络拓扑' })).toBeVisible()
  await expect(page.locator('.topology-map-shell')).toBeVisible()
  await expect(page.locator('.maplibregl-canvas')).toBeVisible()
  await expect(page.getByText('ssh-primary · TCP :60000')).toBeVisible()
  await expect(page.getByText('dns · UDP :53000')).toBeVisible()
  await expect(page.locator('.topology-link-row')).toHaveCount(4)
  await expect.poll(() => page.locator('canvas').count(), { timeout: 10000 }).toBeGreaterThan(1)
  await page.waitForTimeout(500)
  expect(mapFailures).toEqual([])
  expect(pageErrors).toEqual([])
})
