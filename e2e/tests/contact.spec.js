import { test, expect } from '@playwright/test'

test.describe('Contact Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/#/contact')
  })

  test('displays the contact form', async ({ page }) => {
    await expect(page.locator('text=Contact Me')).toBeVisible()
    await expect(page.locator('input[name="name"]')).toBeVisible()
    await expect(page.locator('input[name="email"]')).toBeVisible()
    await expect(page.locator('input[name="subject"]')).toBeVisible()
    await expect(page.locator('textarea[name="message"]')).toBeVisible()
    await expect(page.locator('button:has-text("Send Message")')).toBeVisible()
  })

  test('submits contact form successfully', async ({ page }) => {
    await page.fill('input[name="name"]', 'E2E Test User')
    await page.fill('input[name="email"]', 'e2e@test.com')
    await page.fill('input[name="subject"]', 'E2E Test')
    await page.fill('textarea[name="message"]', 'This is an automated E2E test message.')

    await page.click('button:has-text("Send Message")')

    await expect(page.locator('text=Message Sent!')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('text=Back to Home')).toBeVisible()
  })

  test('shows phone field (optional)', async ({ page }) => {
    await expect(page.locator('input[name="phone"]')).toBeVisible()
  })
})
