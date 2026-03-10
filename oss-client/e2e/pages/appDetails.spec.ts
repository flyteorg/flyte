import { expect } from '@playwright/test'
import { BASE_URL, E2E_PROJECT_ORG } from '../constants'
import { getFixture, viewerPage } from '../util'

viewerPage('App details page', async ({ viewerPage }) => {
  const fixture = getFixture('hello_app')
  if (!fixture) throw new Error('Could not get app data from fixture')
  const appDetailUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/apps/${fixture.id}`
  console.log('navigating to ', appDetailUrl)
  await viewerPage.goto(appDetailUrl)
  await expect(
    viewerPage.getByRole('heading', { name: /hello-app/i }),
  ).toBeVisible()

  const metadataSection = viewerPage.getByTestId('details-metadata')
  await expect(await metadataSection.textContent()).toContain('TypeFastAPI')
  const userIcon = metadataSection.getByLabel(`User: playwright-bot`)
  await expect(userIcon).toBeVisible()

  // test app status indicator
  const statusIndicator = viewerPage.getByTestId('app-status')
  await expect(statusIndicator).toBeVisible()
  const statusText = await statusIndicator.textContent()
  expect(statusText).toMatch(
    /(unspecified|inactive|assigned|stopped|failed|deploying|pending|active|scaling up|scaling down)/i,
  )

  // logs tab (default)
  const logsTab = viewerPage.getByRole('tab').filter({ hasText: 'Logs' })
  await expect(logsTab).toHaveAttribute('data-selected')

  // app tab
  const appTab = viewerPage.getByRole('tab').filter({ hasText: 'App' })
  await appTab.click()
  await expect(appTab).toHaveAttribute('data-selected')
  const descList = viewerPage.getByTestId('dl-wrapper').first()
  await expect(descList).toContainText('HTTP Endpoints')

  // code tab
  // const codeTab = viewerPage.getByRole('tab').filter({ hasText: 'Code' })
  // await expect(codeTab).toBeEnabled() // wait for tab to be interactive
  // await codeTab.click()
  // await expect(codeTab).toHaveAttribute('data-selected')
  // const codeViewer = viewerPage.getByTestId('code-viewer')
  // await expect(codeViewer).toBeVisible()

  // app actions
  const startStopBtn = viewerPage.getByRole('button', {
    name: /start app|stop app/i,
  })
  await expect(startStopBtn).toBeVisible()
  if ((await startStopBtn.textContent())?.match(/stop app/i)) {
    await startStopBtn.click()
    const dialog = viewerPage.getByTestId('simple-dialog')
    await expect(dialog).toBeVisible()
    await expect(dialog).toContainText(`Stop ${fixture.id}?`)
    await expect(dialog).toContainText(
      'Are you sure you want to stop this app? It will no longer be accessible from the HTTP endpoint.',
    )
    const cancelBtn = dialog.getByRole('button').filter({ hasText: 'Cancel' })
    await cancelBtn.click()
    await expect(dialog).toBeHidden()
  }
})
