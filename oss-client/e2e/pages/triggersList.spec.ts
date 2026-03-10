import { expect } from '@playwright/test'
import { adminPage, expectTableHasRows, getFixture } from '../util'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'

adminPage('shows triggers list', async ({ adminPage }) => {
  const triggersListUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/triggers`
  console.log('navigating to ', triggersListUrl)
  await adminPage.goto(triggersListUrl)
  await expect(
    adminPage.getByRole('heading', { name: /Triggers/i }),
  ).toBeVisible()
  await expectTableHasRows(adminPage)
  const fixture = getFixture('list_trigger')
  if (!fixture) throw new Error('Could not get trigger data from fixture')
  const tableBody = await adminPage.locator('tbody')
  const row = await tableBody.locator('tr').filter({ hasText: fixture.name })
  const triggerName = `${fixture.env}.${fixture.name}`
  await expect(row).toContainText(triggerName)
  const userIcon = row.getByLabel(`User: playwright-bot`)
  await expect(userIcon).toBeVisible()
})
