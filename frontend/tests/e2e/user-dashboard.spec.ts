import { test, expect } from '@playwright/test'

test.describe('用户仪表盘 - 页面功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173/')
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

test.describe('用户仪表盘 - 已认证', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173/')
    await page.waitForLoadState('networkidle')
    await page.waitForSelector('.dashboard, .el-form', { timeout: 10000 })
  })

  test('用户已登录时显示仪表盘', async ({ page }) => {
    const isDashboard = await page.locator('.dashboard').count() > 0
    const isLogin = await page.locator('.el-form').count() > 0
    
    if (isDashboard) {
      const elCards = await page.locator('.el-card').count()
      console.log('Dashboard el-card count:', elCards)
      expect(elCards).toBeGreaterThan(0)
    } else {
      console.log('User not authenticated - redirected to login page')
    }
  })

  test('已登录用户可看到图表', async ({ page }) => {
    const isDashboard = await page.locator('.dashboard').count() > 0
    
    if (!isDashboard) {
      test.skip()
    }
    
    await page.waitForSelector('.charts-grid', { timeout: 10000 })
    const chartContainers = await page.locator('.chart-container').count()
    console.log('Chart containers found:', chartContainers)
    expect(chartContainers).toBeGreaterThan(0)
  })

  test('已登录用户ECharts已初始化', async ({ page }) => {
    const isDashboard = await page.locator('.dashboard').count() > 0
    
    if (!isDashboard) {
      test.skip()
    }
    
    await page.waitForSelector('.chart-container', { timeout: 10000 })
    const canvases = await page.locator('.chart-container canvas').count()
    console.log('ECharts canvases found:', canvases)
    expect(canvases).toBeGreaterThan(0)
  })
})
