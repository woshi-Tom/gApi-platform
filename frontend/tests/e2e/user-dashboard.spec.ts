import { test, expect } from '@playwright/test'

const FRONTEND_BASE = 'http://localhost:5173'

async function loginUser(page: any) {
  return await page.evaluate(async () => {
    const response = await fetch('/api/v1/user/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'test@test.com', password: 'password' })
    })
    return response.json()
  })
}

test.describe('用户仪表盘 - 页面功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/`)
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

test.describe('用户登录 - E2E', () => {
  test('登录页面可正常显示', async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.el-form', { timeout: 10000 })
    const emailInput = page.locator('input[type="text"], input[placeholder*="邮箱"], input[placeholder*="email"]').first()
    const passwordInput = page.locator('input[type="password"]').first()
    await expect(emailInput).toBeVisible()
    await expect(passwordInput).toBeVisible()
  })

  test('使用测试凭据登录成功', async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.el-form', { timeout: 10000 })

    const loginData = await loginUser(page)
    expect(loginData.success).toBe(true)
    expect(loginData.data.token).toBeTruthy()

    await page.evaluate((data: any) => {
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
    }, loginData.data)

    await page.goto(`${FRONTEND_BASE}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.dashboard', { timeout: 15000 })

    const statsCards = await page.locator('.stat-card').count()
    expect(statsCards).toBeGreaterThan(0)
  })
})

test.describe('用户仪表盘 - 已认证', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/`)
    await page.waitForLoadState('networkidle')

    const loginData = await loginUser(page)
    await page.evaluate((data: any) => {
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
    }, loginData.data)

    await page.goto(`${FRONTEND_BASE}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.dashboard', { timeout: 15000 })
  })

  test('仪表盘显示统计数据卡片', async ({ page }) => {
    const statsCards = await page.locator('.stat-card').count()
    console.log('Stat cards found:', statsCards)
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

  test('显示Token消耗趋势图表', async ({ page }) => {
    await page.waitForSelector('.chart-card', { timeout: 10000 })
    const chartTitle = page.locator('.chart-card .el-card__header span').first()
    await expect(chartTitle).toContainText('Token')
  })

  test('显示API调用统计图表', async ({ page }) => {
    await page.waitForSelector('.chart-card', { timeout: 10000 })
    const chartTitles = await page.locator('.chart-card .el-card__header span').all()
    expect(chartTitles.length).toBeGreaterThan(1)
    const secondTitle = await chartTitles[1].textContent()
    expect(secondTitle).toContain('API')
  })
})
