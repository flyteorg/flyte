import { expect } from '@playwright/test'
import { adminPage, expectTableHasRows } from '../util/'
import { BASE_URL } from '../constants'

adminPage('shows project list', async ({ adminPage }) => {
  const projectsUrl = `${BASE_URL}/projects`
  console.log('navigating to ', projectsUrl)
  await adminPage.goto(projectsUrl)
  await expect(
    adminPage.getByRole('heading', { name: /Projects/i }),
  ).toBeVisible()
  await expectTableHasRows(adminPage)
})
