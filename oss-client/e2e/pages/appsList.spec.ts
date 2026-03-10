import { expect } from '@playwright/test'
import { adminPage, expectTableHasRows } from '../util'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'

adminPage('shows apps list', async ({ adminPage }) => {
  const appsListUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/apps`
  console.log('navigating to ', appsListUrl)
  await adminPage.goto(appsListUrl)
  await expect(adminPage.getByRole('heading', { name: /Apps/i })).toBeVisible()
  await expectTableHasRows(adminPage)
})
