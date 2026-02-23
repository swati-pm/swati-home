import { test, expect } from '@playwright/test'

test.describe('Admin Login Page', () => {
  test('shows admin sign in page', async ({ page }) => {
    await page.goto('/#/admin')
    await expect(page.locator('.admin-title')).toHaveText('Admin Sign In')
    await expect(page.locator('.admin-btn')).toBeVisible()
  })

  test('does not show admin links in header for non-admin', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('a.header-link:has-text("Contact Requests")')).toHaveCount(0)
    await expect(page.locator('.header-signout')).toHaveCount(0)
  })
})
