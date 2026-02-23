import { test, expect } from '@playwright/test'

test.describe('Experience Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/#/experience')
  })

  test('displays experience cards', async ({ page }) => {
    await expect(page.locator('.exp-card').first()).toBeVisible()
  })

  test('experience card shows company and role', async ({ page }) => {
    const firstCard = page.locator('.exp-card').first()
    await expect(firstCard.locator('.exp-card-company')).toBeVisible()
    await expect(firstCard.locator('.exp-card-role')).toBeVisible()
  })

  test('experience card shows bullet points', async ({ page }) => {
    const firstCard = page.locator('.exp-card').first()
    await expect(firstCard.locator('.exp-card-bullets li').first()).toBeVisible()
  })

  test('does not show edit buttons for non-admin', async ({ page }) => {
    await expect(page.locator('button:has-text("Edit")')).toHaveCount(0)
    await expect(page.locator('button:has-text("+ Add Experience")')).toHaveCount(0)
  })
})
