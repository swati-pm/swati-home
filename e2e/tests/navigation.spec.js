import { test, expect } from '@playwright/test'

test.describe('Navigation', () => {
  test('loads the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.header-brand')).toBeVisible()
    await expect(page.locator('.home-greeting')).toBeVisible()
  })

  test('navigates to Experience page', async ({ page }) => {
    await page.goto('/')
    await page.click('a.header-link:has-text("Experience")')
    await expect(page).toHaveURL(/#\/experience/)
    await expect(page.locator('.experience-title')).toBeVisible()
  })

  test('navigates to Blog page', async ({ page }) => {
    await page.goto('/')
    await page.click('a.header-link:has-text("Blog")')
    await expect(page).toHaveURL(/#\/blog/)
  })

  test('navigates to Contact page', async ({ page }) => {
    await page.goto('/')
    await page.click('a.header-link:has-text("Contact")')
    await expect(page).toHaveURL(/#\/contact/)
    await expect(page.locator('.contact-page-title')).toBeVisible()
  })

  test('navigates back to Home from header brand', async ({ page }) => {
    await page.goto('/#/contact')
    await page.click('.header-brand')
    await expect(page).toHaveURL(/#\/$/)
  })

  test('header highlights active page', async ({ page }) => {
    await page.goto('/#/experience')
    const link = page.locator('a.header-link.active')
    await expect(link).toHaveText('Experience')
  })

  test('footer is visible on all pages', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('footer')).toBeVisible()
  })
})
