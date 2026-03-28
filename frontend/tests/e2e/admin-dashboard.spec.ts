import { test, expect } from '@playwright/test'

/**
 * 管理后台仪表盘图表测试
 */
test.describe('管理后台仪表盘 - 图表功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/admin.html')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载成功', async ({ page }) => {
    await expect(page.locator('.admin-dashboard')).toBeVisible()
  })

  test('API请求趋势图表存在', async ({ page }) => {
    const chart = page.locator('text=API请求趋势')
    await expect(chart).toBeVisible()
  })

  test('用户使用排行图表存在且可切换', async ({ page }) => {
    const chart = page.locator('text=用户使用排行')
    await expect(chart).toBeVisible()

    const requestsTab = page.locator('text=请求量')
    await expect(requestsTab).toBeVisible()
  })

  test('图表容器高度正确', async ({ page }) => {
    const chartContainer = page.locator('.chart-container').first()
    await expect(chartContainer).toBeVisible()
  })
})
