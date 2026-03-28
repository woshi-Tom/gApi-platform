import { test, expect } from '@playwright/test'

const API_BASE = 'http://localhost:8080'
const ADMIN_BASE = 'http://localhost:5174/admin.html'

test.describe('管理后台 - 页面功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')
  })

  test('页面加载成功', async ({ page }) => {
    const body = await page.locator('body')
    await expect(body).toBeVisible()
  })

  test('页面包含Vue应用', async ({ page }) => {
    const html = await page.content()
    expect(html.length).toBeGreaterThan(1000)
    const hasApp = await page.locator('#app').count() > 0
    expect(hasApp).toBe(true)
  })
})

test.describe('管理后台登录 - E2E', () => {
  test('登录页面可正常显示', async ({ page }) => {
    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.el-form', { timeout: 10000 })
    const usernameInput = page.locator('input[type="text"], input[placeholder*="用户"], input[placeholder*="admin"]').first()
    const passwordInput = page.locator('input[type="password"]').first()
    await expect(usernameInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
  })

  test('使用管理员凭据登录成功', async ({ page }) => {
    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.el-form', { timeout: 10000 })

    const loginData = await page.evaluate(async (apiBase) => {
      const response = await fetch(`${apiBase}/api/v1/admin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'admin123' })
      })
      return response.json()
    }, API_BASE)

    expect(loginData.success).toBe(true)
    expect(loginData.data.token).toBeTruthy()

    await page.evaluate((data) => {
      localStorage.setItem('admin_token', data.token)
      localStorage.setItem('admin_user', JSON.stringify({ username: data.username, role: data.role }))
    }, loginData.data)

    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.admin-dashboard', { timeout: 15000 })

    const statsCards = await page.locator('.stat-card').count()
    expect(statsCards).toBeGreaterThan(0)
  })
})

test.describe('管理后台仪表盘 - 已认证', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')

    const loginData = await page.evaluate(async (apiBase) => {
      const response = await fetch(`${apiBase}/api/v1/admin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'admin123' })
      })
      return response.json()
    }, API_BASE)

    await page.evaluate((data) => {
      localStorage.setItem('admin_token', data.token)
      localStorage.setItem('admin_user', JSON.stringify({ username: data.username, role: data.role }))
    }, loginData.data)

    await page.goto(ADMIN_BASE)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.admin-dashboard', { timeout: 15000 })
  })

  test('仪表盘显示统计数据卡片', async ({ page }) => {
    const statsCards = await page.locator('.stat-card').count()
    console.log('Admin stat cards found:', statsCards)
    expect(statsCards).toBeGreaterThanOrEqual(4)
  })

  test('图表区域已渲染', async ({ page }) => {
    await page.waitForSelector('.charts-grid', { timeout: 10000 })
    const chartContainers = await page.locator('.chart-container').count()
    console.log('Chart containers found:', chartContainers)
    expect(chartContainers).toBeGreaterThan(0)
  })

  test('ECharts实例已初始化', async ({ page }) => {
    await page.waitForSelector('.chart-container', { timeout: 10000 })
    await page.waitForTimeout(2000)
    const canvases = await page.locator('.chart-container canvas').count()
    console.log('ECharts canvases found:', canvases)
    expect(canvases).toBeGreaterThan(0)
  })

  test('显示API趋势图表', async ({ page }) => {
    await page.waitForSelector('.charts-section', { timeout: 10000 })
    const chartsSection = page.locator('.charts-section')
    await expect(chartsSection).toBeVisible()
  })
})
