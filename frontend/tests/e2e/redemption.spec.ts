import { test, expect } from '@playwright/test'

const frontendBase = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:5173'
const apiBase = process.env.API_BASE_URL || 'http://localhost:8080'

async function userLogin(page: any, email: string, password: string) {
  const loginData = await page.evaluate(async ([api, em, pwd]) => {
    const response = await fetch(`${api}/api/v1/user/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: em, password: pwd })
    })
    return response.json()
  }, [apiBase, email, password])
  
  await page.evaluate((data: any) => {
    if (data.success && data.data) {
      localStorage.setItem('token', data.data.token)
      localStorage.setItem('user', JSON.stringify(data.data.user))
    }
  }, loginData)
  
  return loginData
}

test.describe.serial('用户兑换码功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${frontendBase}/`)
    await page.waitForLoadState('networkidle')
  })

  test('兑换码页面加载', async ({ page }) => {
    await page.goto(`${frontendBase}/redeem`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)
    
    const pageContent = await page.content()
    expect(pageContent.length).toBeGreaterThan(500)
    
    const hasRedeemForm = await page.locator('input[placeholder*="兑换码"]').count() > 0
    expect(hasRedeemForm).toBe(true)
  })

  test('兑换码页面显示兑换历史区域', async ({ page }) => {
    await page.goto(`${frontendBase}/redeem`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)
    
    const historySection = await page.locator('text=兑换历史').count()
    expect(historySection).toBeGreaterThan(0)
  })

  test('无效兑换码提示错误', async ({ page }) => {
    await userLogin(page, 'testuser123@test.com', 'password123')
    
    await page.goto(`${frontendBase}/redeem`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)
    
    const input = page.locator('input[placeholder*="兑换码"]')
    await input.fill('INVALID_CODE_12345')
    
    const redeemBtn = page.locator('button:has-text("立即兑换")')
    await redeemBtn.click()
    
    await page.waitForTimeout(3000)
    
    const errorMsg = await page.locator('text=/不存在|无效|失败/).count()
    expect(errorMsg).toBeGreaterThan(0)
  })
})

test.describe.serial('管理员兑换码管理', () => {
  const adminBase = 'http://localhost:5174/admin.html'

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
      if (data.success && data.data) {
        localStorage.setItem('admin_token', data.data.token)
        localStorage.setItem('admin_user', JSON.stringify({ username: data.username, role: data.role }))
      }
    }, loginData)
    
    return loginData
  }

  test('兑换码管理页面加载', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/redemption`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const pageContent = await page.content()
    expect(pageContent.length).toBeGreaterThan(1000)
    
    const hasRedemptionTitle = await page.locator('text=兑换码管理').count() > 0
    expect(hasRedemptionTitle).toBe(true)
  })

  test('兑换码列表显示', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/redemption`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const tableExists = await page.locator('table').count() > 0
    expect(tableExists).toBe(true)
  })

  test('生成兑换码按钮存在', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/redemption`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const generateBtn = await page.locator('button:has-text("生成兑换码")').count() > 0
    expect(generateBtn).toBe(true)
  })

  test('打开生成兑换码对话框', async ({ page }) => {
    await page.goto(adminBase)
    await page.waitForLoadState('networkidle')
    await adminLogin(page)
    
    await page.goto(`${adminBase}#/redemption`)
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(3000)
    
    const generateBtn = page.locator('button:has-text("生成兑换码")')
    await generateBtn.click()
    await page.waitForTimeout(1000)
    
    const dialogVisible = await page.locator('.el-dialog, [class*="dialog"]').count() > 0
    expect(dialogVisible).toBe(true)
  })
})
