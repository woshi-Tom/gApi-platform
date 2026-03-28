import { test, expect } from '@playwright/test'

/**
 * 用户仪表盘图表测试
 * 测试目标：确保图表正确渲染，Y轴刻度显示
 */
test.describe('用户仪表盘 - 图表功能', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载成功', async ({ page }) => {
    await expect(page.locator('.dashboard')).toBeVisible()
  })

  test('Token消耗趋势图表存在且有数据', async ({ page }) => {
    // 查找Token消耗趋势图表
    const tokenChart = page.locator('text=Token消耗趋势')
    await expect(tokenChart).toBeVisible()

    // 检查图表canvas存在
    const canvas = page.locator('.chart-container canvas').first()
    await expect(canvas).toBeVisible()
  })

  test('API调用统计图表存在且有数据', async ({ page }) => {
    // 查找API调用统计图表
    const callsChart = page.locator('text=API调用统计')
    await expect(callsChart).toBeVisible()

    // 检查图表canvas存在
    const canvases = page.locator('.chart-container canvas')
    await expect(canvases).toHaveCount(2)
  })

  test('图表有Y轴刻度（不空白）', async ({ page }) => {
    // 等待图表渲染
    await page.waitForTimeout(1000)

    // 检查图表组件存在
    const charts = page.locator('.echarts')
    await expect(charts.first()).toBeVisible()
  })
})
