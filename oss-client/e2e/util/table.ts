import { expect, Page } from '@playwright/test'

export const expectTableHasRows = async (page: Page) => {
  await page.waitForSelector('table', { timeout: 10000 })
  const table = page.locator('table')
  await expect(table).toBeVisible()
  const rows = table.locator('tbody tr')
  // Check that at least one row exists
  const rowCount = await rows.count()
  expect(rowCount).toBeGreaterThan(0)
}
