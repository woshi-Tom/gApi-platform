import { test, expect } from '@playwright/test'

const frontendBase = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173'
const apiBase = process.env.API_BASE_URL || 'http://localhost:8080'

async function userLogin(page: any) {
  const loginData = await page.evaluate(async (api: string) => {
    const response = await fetch(`${api}/api/v1/user/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'testuser123@test.com', password: 'password123' })
    })
    return response.json()
  }, apiBase)
  
  await page.evaluate((data: any) => {
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
  }, loginData.data)
  
  return loginData
}

test.describe.serial('用户仪表盘 - 页面功能', () => {
  test('页面加载成功', async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    const body = await page.locator('body')
    await expect(body).toBeVisible()
  })

  test('页面包含Vue应用', async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    const html = await page.content()
    expect(html.length).toBeGreaterThan(1000)
    const hasApp = await page.locator('#app').count() > 0
    expect(hasApp).toBe(true)
  })
})

test.describe.serial('用户登录 - E2E', () => {
  test('用户登录并跳转到仪表盘', async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)

    await userLogin(page)
    
    // Navigate to trigger router
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    // Check if we're on dashboard
    const dashboardExists = await page.locator('.dashboard, [class*="dashboard"]').count() > 0
    expect(dashboardExists).toBe(true)
    
    const statsCards = await page.locator('.stat-card').count()
    expect(statsCards).toBeGreaterThan(0)
  })

  test('仪表盘显示统计数据卡片', async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const dashboardExists = await page.locator('.dashboard, [class*="dashboard"]').count() > 0
    if (!dashboardExists) {
      await userLogin(page)
      await page.goto(`${frontendBase}/`)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(3000)
    }
    
    const statsCards = await page.locator('.stat-card').count()
    console.log('Stat cards found:', statsCards)
    expect(statsCards).toBeGreaterThanOrEqual(4)
  })

  test('图表区域已渲染', async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const dashboardExists = await page.locator('.dashboard, [class*="dashboard"]').count() > 0
    if (!dashboardExists) {
      await userLogin(page)
      await page.goto(`${frontendBase}/`)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(3000)
    }
    
    await page.waitForTimeout(2000)
    const chartContainers = await page.locator('.chart-container').count()
    console.log('Chart containers found:', chartContainers)
    expect(chartContainers).toBeGreaterThan(0)
  })
})
