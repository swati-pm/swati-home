import { test, expect } from '@playwright/test'

test.describe('Responsive Design', () => {
  test('mobile viewport renders header properly', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 })
    await page.goto('/')

    await expect(page.locator('.header-brand')).toBeVisible()
    await expect(page.locator('a.header-link:has-text("Home")')).toBeVisible()
    await expect(page.locator('a.header-link:has-text("Experience")')).toBeVisible()
    await expect(page.locator('a.header-link:has-text("Blog")')).toBeVisible()
    await expect(page.locator('a.header-link:has-text("Contact")')).toBeVisible()

    // Header should be stacked on mobile
    const header = page.locator('.header-content')
    await expect(header).toBeVisible()
  })

  test('tablet viewport renders properly', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/')

    await expect(page.locator('.header-brand')).toBeVisible()
    await expect(page.locator('.home-chat-trigger')).toBeVisible()
  })

  test('desktop viewport renders properly', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/')

    await expect(page.locator('.header-brand')).toBeVisible()
    await expect(page.locator('.home-chat-trigger')).toBeVisible()
  })

  test('home page content is visible on all viewports', async ({ page }) => {
    const viewports = [
      { width: 375, height: 812 },
      { width: 768, height: 1024 },
      { width: 1440, height: 900 },
    ]

    for (const viewport of viewports) {
      await page.setViewportSize(viewport)
      await page.goto('/')
      await expect(page.locator('.home-greeting')).toBeVisible()
      await expect(page.locator('.home-summary')).toBeVisible()
    }
  })
})
