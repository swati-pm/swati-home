import { test, expect } from '@playwright/test'

// Inject admin session into sessionStorage to simulate logged-in admin.
// The frontend checks user.email === VITE_ADMIN_EMAIL for admin status.
// In docker-compose.test.yml, VITE_ADMIN_EMAIL=test@example.com.
async function loginAsAdmin(page) {
  await page.addInitScript(() => {
    sessionStorage.setItem(
      'swati-home-auth',
      JSON.stringify({
        email: 'test@example.com',
        name: 'Test Admin',
        picture: '',
        token: 'fake-test-token',
      })
    )
  })
}

// Intercept API calls that require auth, so the edit flow works
// without real Google JWT tokens.
async function mockAdminAPIs(page) {
  // Mock PUT /api/experiences/:id for saving edits
  await page.route('**/api/experiences/*', async (route) => {
    if (route.request().method() === 'PUT') {
      const body = JSON.parse(route.request().postData())
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(body),
      })
    } else {
      await route.fallback()
    }
  })

  // Mock POST /api/experiences for adding new
  await page.route('**/api/experiences', async (route) => {
    if (route.request().method() === 'POST') {
      const body = JSON.parse(route.request().postData())
      body.id = 'new-test-id'
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(body),
      })
    } else {
      await route.fallback()
    }
  })

  // Mock POST /api/experiences/suggest for AI suggestions
  await page.route('**/api/experiences/suggest', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        bullets: ['Spearheaded product strategy', 'Drove 20% revenue growth'],
      }),
    })
  })
}

test.describe('Edit Experience Flow', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
    await mockAdminAPIs(page)
    await page.goto('/#/experience')
    // Wait for experience cards to load
    await expect(page.locator('.exp-card').first()).toBeVisible()
  })

  test('admin sees Edit buttons on experience cards', async ({ page }) => {
    const editButtons = page.locator('.exp-card-edit')
    await expect(editButtons.first()).toBeVisible()
    await expect(editButtons.first()).toHaveText('Edit')
  })

  test('admin sees Add Experience button', async ({ page }) => {
    await expect(page.locator('.add-btn')).toBeVisible()
    await expect(page.locator('.add-btn')).toHaveText('+ Add Experience')
  })

  test('clicking Edit opens modal with pre-filled data', async ({ page }) => {
    const firstCard = page.locator('.exp-card').first()
    const company = await firstCard.locator('.exp-card-company').textContent()
    const role = await firstCard.locator('.exp-card-role').textContent()

    await firstCard.locator('.exp-card-edit').click()

    // Modal should appear
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal-title')).toHaveText('Edit Experience')

    // Fields should be pre-filled
    const companyInput = page.locator('input[name="company"]')
    await expect(companyInput).toHaveValue(company)
    const roleInput = page.locator('input[name="role"]')
    await expect(roleInput).toHaveValue(role)
  })

  test('editing fields and saving updates the experience', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // Clear and type new company name
    const companyInput = page.locator('input[name="company"]')
    await companyInput.clear()
    await companyInput.fill('Updated Corp')

    // Clear and type new role
    const roleInput = page.locator('input[name="role"]')
    await roleInput.clear()
    await roleInput.fill('Director of Engineering')

    // Save
    await page.locator('button:has-text("Save")').click()

    // Modal should close
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('cancel closes the modal without saving', async ({ page }) => {
    const firstCard = page.locator('.exp-card').first()
    const originalCompany = await firstCard.locator('.exp-card-company').textContent()

    await firstCard.locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // Edit company
    const companyInput = page.locator('input[name="company"]')
    await companyInput.clear()
    await companyInput.fill('Should Not Save')

    // Cancel
    await page.locator('button:has-text("Cancel")').click()

    // Modal should close
    await expect(page.locator('.modal')).not.toBeVisible()

    // Original company should remain
    await expect(firstCard.locator('.exp-card-company')).toHaveText(originalCompany)
  })

  test('clicking overlay background closes modal', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // Click the overlay (outside the modal)
    await page.locator('.modal-overlay').click({ position: { x: 10, y: 10 } })

    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('editing bullets textarea', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    const textarea = page.locator('textarea[name="bullets"]')
    await textarea.clear()
    await textarea.fill('First bullet point\nSecond bullet point\nThird bullet point')

    await expect(textarea).toHaveValue('First bullet point\nSecond bullet point\nThird bullet point')
  })

  test('Add Experience opens empty form', async ({ page }) => {
    await page.locator('.add-btn').click()

    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal-title')).toHaveText('Add Experience')

    // All fields should be empty
    await expect(page.locator('input[name="company"]')).toHaveValue('')
    await expect(page.locator('input[name="role"]')).toHaveValue('')
    await expect(page.locator('input[name="location"]')).toHaveValue('')
    await expect(page.locator('input[name="startDate"]')).toHaveValue('')
    await expect(page.locator('input[name="endDate"]')).toHaveValue('')
    await expect(page.locator('textarea[name="bullets"]')).toHaveValue('')
  })

  test('Suggest Improvements button visible when editing with bullets', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // The suggest button should be visible since we're editing an existing experience with bullets
    await expect(page.locator('.btn-suggest')).toBeVisible()
    await expect(page.locator('.btn-suggest')).toHaveText('Suggest Improvements')
  })

  test('Suggest Improvements button not visible for new experience', async ({ page }) => {
    await page.locator('.add-btn').click()
    await expect(page.locator('.modal')).toBeVisible()

    // No suggest button for new experience
    await expect(page.locator('.btn-suggest')).not.toBeVisible()
  })

  test('clicking Suggest Improvements shows suggestions panel', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    await page.locator('.btn-suggest').click()

    // Loading state
    await expect(page.locator('.btn-suggest')).toHaveText('Getting suggestions...')

    // Suggestions panel appears
    await expect(page.locator('.suggest-panel')).toBeVisible()
    await expect(page.locator('.suggest-panel-title')).toHaveText('Suggested Improvements')
    await expect(page.locator('.suggest-item').first()).toBeVisible()

    // Accept and Dismiss buttons
    await expect(page.locator('button:has-text("Accept")')).toBeVisible()
    await expect(page.locator('button:has-text("Dismiss")')).toBeVisible()
  })

  test('accepting suggestions replaces bullets', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    await page.locator('.btn-suggest').click()
    await expect(page.locator('.suggest-panel')).toBeVisible()

    await page.locator('button:has-text("Accept")').click()

    // Panel should close
    await expect(page.locator('.suggest-panel')).not.toBeVisible()

    // Bullets textarea should have the suggested content
    const textarea = page.locator('textarea[name="bullets"]')
    await expect(textarea).toHaveValue('Spearheaded product strategy\nDrove 20% revenue growth')
  })

  test('dismissing suggestions keeps original bullets', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    const textarea = page.locator('textarea[name="bullets"]')
    const originalBullets = await textarea.inputValue()

    await page.locator('.btn-suggest').click()
    await expect(page.locator('.suggest-panel')).toBeVisible()

    await page.locator('button:has-text("Dismiss")').click()

    // Panel should close
    await expect(page.locator('.suggest-panel')).not.toBeVisible()

    // Original bullets should remain
    await expect(textarea).toHaveValue(originalBullets)
  })

  test('role description toggle shows textarea', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // Click the role description toggle
    await page.locator('.suggest-role-toggle').click()

    // Role description textarea should appear
    const roleDesc = page.locator('textarea[placeholder*="Paste a job description"]')
    await expect(roleDesc).toBeVisible()

    // Type a role description
    await roleDesc.fill('Senior PM role requiring data analytics expertise')
    await expect(roleDesc).toHaveValue('Senior PM role requiring data analytics expertise')
  })

  test('suggest with role description sends it in request', async ({ page }) => {
    let capturedBody = null
    await page.route('**/api/experiences/suggest', async (route) => {
      capturedBody = JSON.parse(route.request().postData())
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ bullets: ['Tailored bullet'] }),
      })
    })

    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    await page.locator('.suggest-role-toggle').click()
    await page.locator('textarea[placeholder*="Paste a job description"]').fill('Data PM role')

    await page.locator('.btn-suggest').click()
    await expect(page.locator('.suggest-panel')).toBeVisible()

    expect(capturedBody).not.toBeNull()
    expect(capturedBody.role_description).toBe('Data PM role')
  })

  test('suggest error shows error banner', async ({ page }) => {
    // Override the suggest mock to return an error
    await page.route('**/api/experiences/suggest', async (route) => {
      await route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Service Unavailable' }),
      })
    })

    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    await page.locator('.btn-suggest').click()

    // Error banner should appear
    await expect(page.locator('.suggest-error')).toBeVisible()
    await expect(page.locator('.suggest-error')).toContainText('Failed to get suggestions')
  })

  test('suggest error can be dismissed', async ({ page }) => {
    await page.route('**/api/experiences/suggest', async (route) => {
      await route.fulfill({ status: 503, body: 'error' })
    })

    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await page.locator('.btn-suggest').click()
    await expect(page.locator('.suggest-error')).toBeVisible()

    await page.locator('.suggest-error-close').click()
    await expect(page.locator('.suggest-error')).not.toBeVisible()
  })

  test('full edit flow: edit, suggest, accept, save', async ({ page }) => {
    await page.locator('.exp-card').first().locator('.exp-card-edit').click()
    await expect(page.locator('.modal')).toBeVisible()

    // Edit company
    const companyInput = page.locator('input[name="company"]')
    await companyInput.clear()
    await companyInput.fill('Updated Corp')

    // Get suggestions
    await page.locator('.btn-suggest').click()
    await expect(page.locator('.suggest-panel')).toBeVisible()

    // Accept suggestions
    await page.locator('button:has-text("Accept")').click()
    await expect(page.locator('.suggest-panel')).not.toBeVisible()

    // Verify bullets were updated
    const textarea = page.locator('textarea[name="bullets"]')
    await expect(textarea).toHaveValue('Spearheaded product strategy\nDrove 20% revenue growth')

    // Save
    await page.locator('button:has-text("Save")').click()

    // Modal should close
    await expect(page.locator('.modal')).not.toBeVisible()
  })
})
