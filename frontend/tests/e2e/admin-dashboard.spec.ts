import { test, expect } from '@playwright/test'

const adminBase = 'http://localhost:5174/admin.html'
const apiBase = process.env.API_BASE_URL || 'http://localhost:8080'

async function adminLogin(page: any) {
  const loginData = await page.evaluate(async (api: string) => {
    const response = await fetch(`${api}/api/v1/admin/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' })
    })
    return response.json()
  }, apiBase)
  
  await page.evaluate((data: any) => {
    localStorage.setItem('admin_token', data.token)
    localStorage.setItem('admin_user', JSON.stringify({ username: data.username, role: data.role }))
  }, loginData.data)
  
  return loginData
}

test.describe.serial('管理后台 - 页面功能', () => {
  test('页面加载成功', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    const body = await page.locator('body')
    await expect(body).toBeVisible()
  })

  test('页面包含Vue应用', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    const html = await page.content()
    expect(html.length).toBeGreaterThan(500)
    const hasApp = await page.locator('#app').count() > 0
    expect(hasApp).toBe(true)
  })
})

test.describe.serial('管理后台登录 - E2E', () => {
  test('管理员登录并跳转到仪表盘', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)
    
    const loginData = await adminLogin(page)
    expect(loginData.success).toBe(true)
    
    await page.goto(`${adminBase}#/dashboard`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(5000)
    
    const sidebar = await page.locator('.admin-sidebar').count()
    const layout = await page.locator('.admin-layout').count()
    expect(sidebar).toBeGreaterThan(0)
    expect(layout).toBeGreaterThan(0)
  })

  test('仪表盘显示统计数据卡片', async ({ page }) => {
    // Login first (new page, no localStorage)
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/dashboard`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const statsCards = await page.locator('.stat-card').count()
    expect(statsCards).toBeGreaterThanOrEqual(4)
  })

  test('图表区域已渲染', async ({ page }) => {
    // Login first (new page, no localStorage)
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/dashboard`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(5000)
    
    const chartContainers = await page.locator('.chart-container').count()
    expect(chartContainers).toBeGreaterThan(0)
  })
})
