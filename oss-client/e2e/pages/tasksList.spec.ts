import { expect } from '@playwright/test'
import { adminPage, expectTableHasRows, getFixture } from '../util'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'

adminPage('shows tasks list', async ({ adminPage }) => {
  const tasksListUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/tasks`
  console.log('navigating to ', tasksListUrl)
  await adminPage.goto(tasksListUrl)
  await expect(adminPage.getByRole('heading', { name: /Tasks/i })).toBeVisible()
  await expectTableHasRows(adminPage)

  const fixture = getFixture('add_ten')
  if (!fixture) throw new Error('Could not get task data from fixture')

  const tableBody = await adminPage.locator('tbody')

  const row = await tableBody
    .locator('tr')
    .filter({ hasText: fixture?.display_name })

  expect(row).toBeVisible()
  expect(row).toContainText(fixture?.env)
  const userIcon = row.getByLabel(`User: playwright-bot`)
  expect(userIcon).toBeVisible()

  // last updated filter
  const updatedFilter = await adminPage
    .locator('button')
    .filter({ hasText: 'Last updated' })
  await updatedFilter.click()
  const menu = await adminPage.getByRole('dialog')
  expect(menu).toBeVisible()
  await expect(menu).toContainText('Last hour')
  await menu.locator('button').filter({ hasText: 'Last hour ' }).click()
  await adminPage.waitForTimeout(1000)
  const rows = await tableBody.locator('tr')
  // at least one task has been updated in the last hour
  await expect(await rows.count()).toBeGreaterThanOrEqual(1)
  const clearBtn = await adminPage.getByLabel('Clear')
  await clearBtn.click()
  const updatedRows = await adminPage.locator('tr')
  await expect(await updatedRows.count()).toBeGreaterThan(1)

  // deployed by filter
  const deploydFilter = await adminPage
    .locator('button')
    .filter({ hasText: 'Deployed by' })
  await deploydFilter.click()
  const dialog = await adminPage.getByRole('dialog')
  expect(dialog).toBeVisible()
  await dialog.locator('input').fill('playwright-bot')
  await adminPage.waitForTimeout(1000)
  await expect(await dialog.locator('button').count()).toEqual(1)
  await dialog.locator('button').click()
  await expect(dialog.locator('button')).toHaveAttribute('data-checked', 'true')
  await adminPage.waitForTimeout(2000)
  await expect(await rows.count()).toBeGreaterThanOrEqual(1)
  await adminPage.locator('button').filter({ hasText: 'Clear all' }).click()
  await adminPage.waitForTimeout(2000)
  await expect(await rows.count()).toBeGreaterThan(1)

  // search
  const searchInput = adminPage.locator('input[type="Search"]').last()
  if (!fixture.display_name) throw new Error('No task name added to fixture')
  await searchInput.fill(fixture.display_name)
  await adminPage.waitForTimeout(1500)
  await expect(rows).toHaveCount(1)
  searchInput.clear()
  await adminPage.waitForTimeout(1500)
  await expect(await rows.count()).toBeGreaterThan(1)
})
