import { expect, test } from '@playwright/test'

const user = { username: 'owner', email: 'owner@example.test', role: 'owner' }
const client = { id: 'owner.c.mac', comment: '', configurationState: 'unconfigured', status: 'pending', enabled: true, tunnelCount: 0 }
const server = { id: 'owner.s.edge', address: 'edge.example.test', bindPort: 7000, comment: '', configurationState: 'unconfigured', status: 'pending', tunnelCount: 0 }
const strictCSP = "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:"

async function mockAPI(page: import('@playwright/test').Page) {
  await page.route('**/api/v2/bootstrap-status', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ registrationEnabled: false, ownerExists: true, canCreateOwner: false }) }))
  await page.route('**/api/v2/account', route => route.fulfill({ status: 401, contentType: 'application/problem+json', body: JSON.stringify({ title: 'Unauthorized', detail: 'sign in required' }) }))
  await page.route('**/api/v2/auth/login', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user }) }))
  await page.route('**/api/v2/overview', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ clients: 1, servers: 1, tunnels: 0 }) }))
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
  await page.getByRole('link', { name: 'Clients', exact: true }).click()
  await expect(page.getByText('owner.c.mac')).toBeVisible()
  await page.getByRole('button', { name: '添加 Client' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.getByRole('button', { name: '取消' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
})
