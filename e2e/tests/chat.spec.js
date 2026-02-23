import { test, expect } from '@playwright/test'

test.describe('Chat Widget', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('shows chat trigger on home page', async ({ page }) => {
    await expect(page.locator('.home-chat-trigger')).toBeVisible()
    await expect(page.locator('text=Ask me anything about Swati')).toBeVisible()
  })

  test('opens chat overlay when trigger is clicked', async ({ page }) => {
    await page.click('.home-chat-trigger')
    await expect(page.locator('.home-chat-overlay')).toBeVisible()
    await expect(page.locator('text=Ask about Swati')).toBeVisible()
    await expect(page.locator('input[placeholder="Type a question..."]')).toBeVisible()
  })

  test('shows welcome message in overlay', async ({ page }) => {
    await page.click('.home-chat-trigger')
    await expect(page.locator('.home-chat-welcome')).toBeVisible()
  })

  test('close button is visible in chat overlay', async ({ page }) => {
    await page.click('.home-chat-trigger')
    await expect(page.locator('.home-chat-close')).toBeVisible()
  })

  test('send button is disabled when input is empty', async ({ page }) => {
    await page.click('.home-chat-trigger')
    await expect(page.locator('.home-chat-send')).toBeDisabled()
  })

  test('send button is enabled when input has text', async ({ page }) => {
    await page.click('.home-chat-trigger')
    await page.fill('input[placeholder="Type a question..."]', 'Hello')
    await expect(page.locator('.home-chat-send')).toBeEnabled()
  })
})
