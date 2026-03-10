import { expect } from '@playwright/test'
import { adminPage, expectTableHasRows, getFixture } from '../util'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'

adminPage('shows runs list', async ({ adminPage }) => {
  const runsListUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/runs`
  console.log('navigating to ', runsListUrl)
  await adminPage.goto(runsListUrl)
  await expect(adminPage.getByRole('heading', { name: /Runs/i })).toBeVisible()
  await adminPage.waitForTimeout(1000)
  await expectTableHasRows(adminPage)

  // ensure we are in grouped view
  const viewByGroupToggle = await adminPage.getByTestId('group-runs-toggle')
  const isGroupView = await viewByGroupToggle.getAttribute('aria-checked')
  if (isGroupView === 'false') {
    console.log('toggling view by groups to true')
    await viewByGroupToggle.click()
  }
  const runFixture = await getFixture('run_with_groups')
  if (!runFixture || !runFixture.display_name) {
    throw new Error('Could not get run display name from fixture')
  }
  await expect(
    await adminPage.getByText(`${runFixture.env}.${runFixture.display_name}`),
  ).toBeInViewport()

  const taskFixture = await getFixture('add_ten')
  if (!taskFixture || !taskFixture.display_name) {
    throw new Error('Could not get task display name from fixture')
  }
  await expect(
    await adminPage.getByText(`${taskFixture.env}.${taskFixture.display_name}`),
  ).toBeInViewport()

  const timeFilter = await adminPage.getByText('Runs in the last 7 days')
  await timeFilter.click()
  const dialog = await adminPage.getByRole('dialog')
  const btn = await dialog.getByText('Last 30 min')
  await btn.click()
  const applyBtn = await dialog.getByText('Apply')
  await applyBtn.click()
  await expect(dialog).not.toBeInViewport()
  await adminPage.waitForTimeout(1000)
  await expect(
    adminPage.getByText(`${runFixture.env}.${runFixture.display_name}`),
  ).toBeInViewport()
})
