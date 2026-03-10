import { expect } from '@playwright/test'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'
import { expectTableHasRows, getFixture, viewerPage } from '../util'

viewerPage('Task details page', async ({ viewerPage }) => {
  const fixture = getFixture('add_ten')
  if (!fixture) throw new Error('Could not get task data from fixture')
  const taskDetailUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/tasks/${fixture.env}.${fixture.name}`
  console.log('navigating to ', taskDetailUrl)
  await viewerPage.goto(taskDetailUrl)
  await expect(
    viewerPage.getByRole('heading', { name: /add_ten/i }),
  ).toBeVisible()
  const metadataSection = viewerPage.getByTestId('details-metadata')
  await expect(await metadataSection.textContent()).toContain(
    fixture.name,
  )
  await expect(await metadataSection.textContent()).toContain('SourceGitHub')
  const userIcon = metadataSection.getByLabel(`User: playwright-bot`)
  expect(userIcon).toBeVisible()

  // runs tab
  const runsTab = await viewerPage.getByRole('tab').filter({ hasText: 'Runs' })
  expect(runsTab).toHaveAttribute('data-selected')
  await expectTableHasRows(viewerPage)

  // spec tab
  const specTab = await viewerPage.getByRole('tab').filter({ hasText: 'Task' })
  await specTab.click()
  const descList = await viewerPage.getByTestId('dl-wrapper').last()
  await expect(await descList.textContent()).toContain(
    `${fixture.env}.${fixture.name}`,
  )
  await expect(descList).toContainText('python')

  // triggers tab
  const triggersTab = await viewerPage
    .getByRole('tab')
    .filter({ hasText: 'Triggers' })
  await triggersTab.click()
  await expect(viewerPage.getByText('No triggers')).toBeVisible()

  // // code tab
  // const codeTab = await viewerPage.getByRole('tab').filter({ hasText: 'Code' })
  // await codeTab.click()
  // await expect(codeTab).toHaveAttribute('data-selected')
  // // Code tab should show "Code" heading or a message about code availability
  // const codeHeading = await viewerPage.getByRole('heading', { name: 'Code' })
  // await expect(codeHeading).toBeVisible()

  // launch task
  const launchBtn = await viewerPage.getByRole('button', {
    name: 'Run',
    exact: true,
  })
  await launchBtn.click()

  const launchform = await viewerPage.getByTestId('drawer-dialog')
  await expect(launchform).toBeVisible()
  await expect(await launchform.textContent()).toContain(
    `${fixture.env}.${fixture.name}`,
  )
  expect(
    await launchform.getByRole('tab').filter({ hasText: 'Inputs' }),
  ).toHaveAttribute('data-selected')
})
