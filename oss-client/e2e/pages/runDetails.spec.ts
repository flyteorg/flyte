import { expect } from '@playwright/test'
import { BASE_URL, E2E_PROJECT_ORG, E2E_PROJECT, E2E_ORG } from '../constants'
import { getFixture, viewerPage } from '../util/'

viewerPage('basic run details', async ({ viewerPage }) => {
  const fixture = getFixture('run_with_groups')
  if (!fixture?.id || !fixture.display_name)
    throw new Error('Could not get run name from fixture')
  const runDetailsUrl = `${BASE_URL}/${E2E_PROJECT_ORG}/runs/${fixture.id}`
  console.log('navigating to ', runDetailsUrl)
  await viewerPage.goto(runDetailsUrl)
  await expect(
    viewerPage.locator('h2', { hasText: fixture.display_name }),
  ).toBeVisible()
  // verify that root action has the same name as defined in task
  const firstAction = await viewerPage
    .getByTestId('sidebar')
    .getByTestId('timeline-action-label')
    .first()
  const text = await firstAction.textContent()
  expect(text).toBe(fixture.display_name)

  // verify number of sidebar items after collapse/uncollapse
  const allActions = await viewerPage
    .getByTestId('sidebar')
    .getByTestId('timeline-action-label')
    .count()
  // root + 10 children + 2 group folders
  expect(allActions).toBe(13)

  await viewerPage
    .getByTestId('sidebar')
    .getByTestId('timeline-group-chevron')
    .first()
    .click()

  await viewerPage
    .getByTestId('sidebar')
    .getByTestId('timeline-group-chevron')
    .last()
    .click()

  const updatedActions = await viewerPage
    .getByTestId('sidebar')
    .getByTestId('timeline-action-label')
    .count()
  // root + 2 group folders
  expect(updatedActions).toBe(3)

  const firstTab = await viewerPage.getByRole('tab').first()
  const firstTabText = await firstTab.textContent()
  expect(firstTabText).toBe('Summary')
  expect(firstTab).toHaveAttribute('data-selected')

  // subactions
  const subactionSection = await viewerPage.getByTestId(
    `tabsection-Sub actions`,
  )
  expect(subactionSection).toBeVisible()
  // check subaction table
  const subaction = await viewerPage
    .getByTestId('subaction-cell')
    .nth(4)
    .textContent()
  expect(subaction).toBe('Completed10(100.0%)')

  const outputs = await viewerPage.getByTestId(`tabsection-Output`)
  await outputs.scrollIntoViewIfNeeded()
  await viewerPage.waitForTimeout(1000)
  // Check that outputs contain the expected array values
  const outputsText = await outputs.textContent()
  expect(outputsText).toMatch(/o0.*\[/)
  expect(outputsText).toContain('0')
  expect(outputsText).toContain('4')
  expect(outputsText).toContain('8')
  expect(outputsText).toContain('12')
  expect(outputsText).toContain('18')
  
  // Check that inputs contain the expected key-value pair
  const inputSection = await viewerPage.getByTestId('tabsection-Input')
  await inputSection.scrollIntoViewIfNeeded()
  const inputsText = await inputSection.textContent()
  expect(inputsText).toMatch(/x\s*:\s*10/)

  // logs
  const logsTab = await viewerPage.getByRole('tab').nth(1)
  const logsTabText = await logsTab.textContent()
  expect(logsTabText).toBe('Logs')
  await logsTab.scrollIntoViewIfNeeded()
  await logsTab.click()
  await expect(logsTab).toHaveAttribute('data-selected')
  // check for expected log to be present
  const logviewer = await viewerPage.getByTestId('logviewer')
  await expect(logviewer).toBeVisible()
  if (fixture.logs) {
    // give logs plenty of time to show up to prevent flakiness
    await expect(logviewer).toContainText(fixture.logs, { timeout: 30000 })
  }

  // metrics
  const metricsTab = await viewerPage.getByRole('tab').nth(2)
  const metricsTabText = await metricsTab.textContent()
  expect(metricsTabText).toBe('Metrics')
  await metricsTab.click()
  await expect(metricsTab).toHaveAttribute('data-selected')
  const chartWrapper = viewerPage.getByTestId('run-metrics-chart').first()
  await expect(chartWrapper).toBeVisible()

  const svg = chartWrapper.locator('svg')
  await expect(svg).toBeVisible({ timeout: 1000 })

  const paths = svg.locator('path')
  await expect(await paths.count()).toBeGreaterThan(0)

  // spec
  const specTab = await viewerPage.getByRole('tab').nth(4)
  const specTabText = await specTab.textContent()
  expect(specTabText).toBe('Task')
  await specTab.click()
  const descList = viewerPage.getByTestId('dl-wrapper')
  await expect(descList).toBeVisible()
  await expect(descList).toContainText('Name')
  await expect(descList).toContainText(`${fixture.env}.${fixture.display_name}`)
  await expect(descList).toContainText('Project')
  await expect(descList).toContainText(E2E_PROJECT)
  await expect(descList).toContainText('Organization')
  await expect(descList).toContainText(E2E_ORG)
})
