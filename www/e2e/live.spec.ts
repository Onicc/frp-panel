import { expect, test } from '@playwright/test'

const liveURL = process.env.FRP_PANEL_E2E_URL

test.skip(!liveURL, 'set FRP_PANEL_E2E_URL to test an empty live Master')

test('empty deployment supports setup, login, and deployment generation', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: '创建首个所有者账户' })).toBeVisible()

  await page.getByLabel('用户名').fill('owner')
  await page.getByLabel('邮箱').fill('owner@example.test')
  await page.getByLabel('密码', { exact: true }).fill('correct-horse-battery-staple')
  await page.getByLabel('确认密码').fill('correct-horse-battery-staple')
  await page.getByRole('button', { name: '创建所有者并进入控制台' }).click()
  await expect(page).toHaveURL(`${liveURL}/`)
  await expect(page.getByRole('heading', { name: '概览' })).toBeVisible()

  await page.getByRole('button', { name: '退出登录' }).click()
  await expect(page).toHaveURL(`${liveURL}/login`)
  await page.getByLabel('用户名或邮箱').fill('owner')
  await page.getByLabel('密码').fill('correct-horse-battery-staple')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).toHaveURL(`${liveURL}/`)

  await page.getByRole('link', { name: 'Servers' }).click()
  await page.getByRole('button', { name: /创建 Server/ }).click()
  await page.getByLabel('Server ID').fill('edge')
  await page.getByLabel('Server 公网地址').fill('127.0.0.1')
  await page.getByLabel('FRPS 绑定端口').fill('17000')
  await page.getByRole('button', { name: '创建并生成部署文件' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  const compose = await page.locator('pre').textContent()
  expect(compose).toContain(`PUBLIC_URL: ${JSON.stringify(liveURL)}`)
  expect(compose?.split(liveURL!).length).toBe(2)

  await page.getByRole('link', { name: 'Clients' }).click()
  await page.getByRole('button', { name: /添加 Client/ }).click()
  await page.getByLabel('Client ID').fill('mac')
  await page.getByRole('button', { name: '确认' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  await expect(page.getByText('安装 Client Agent')).toBeVisible()
  await page.getByRole('tab', { name: 'macOS' }).click()
  await expect(page.locator('pre')).toContainText('--client-id')
  await expect(page.locator('pre')).toContainText('--enrollment-token')
})
