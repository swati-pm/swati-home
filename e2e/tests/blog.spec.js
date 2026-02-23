import { test, expect } from '@playwright/test'

test.describe('Blog Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/#/blog')
  })

  test('displays blog section', async ({ page }) => {
    await expect(page.locator('.blog-title')).toBeVisible()
  })

  test('displays blog cards when posts exist', async ({ page }) => {
    // Wait for content to load
    await page.waitForTimeout(1000)

    const cards = page.locator('.blog-card')
    const count = await cards.count()

    if (count > 0) {
      const firstCard = cards.first()
      await expect(firstCard.locator('.blog-card-title')).toBeVisible()
    } else {
      await expect(page.locator('text=No posts yet.')).toBeVisible()
    }
  })

  test('does not show add button for non-admin', async ({ page }) => {
    await expect(page.locator('button:has-text("+ Add Post")')).toHaveCount(0)
  })
})
