import { Page, test } from '@playwright/test'
import { withRetry } from './misc'
import { HOST } from '../constants'

export async function signIn(page: Page, role: 'admin' | 'viewer' = 'admin') {
  await withRetry(
    async () => {
      const creds = getRoleCredentials(role)
      await page.goto(HOST)
      await enterCreds(page, creds.username, creds.password)
      await page.waitForTimeout(1000)
    },
    3,
    'signIn',
  )
}

function getRoleCredentials(role: 'admin' | 'viewer') {
  switch (role) {
    case 'admin':
      if (!process.env.ADMIN_USERNAME || !process.env.ADMIN_PASSWORD) {
        throw new Error('No admin username/password passed via env vars')
      }
      return {
        username: process.env.ADMIN_USERNAME,
        password: process.env.ADMIN_PASSWORD,
      }
    case 'viewer':
      if (!process.env.VIEWER_USERNAME || !process.env.VIEWER_PASSWORD) {
        throw new Error('No viewer username/password passed via env vars')
      }
      return {
        username: process.env.VIEWER_USERNAME,
        password: process.env.VIEWER_PASSWORD,
      }
    default:
      return {
        username: process.env.VIEWER_USERNAME!,
        password: process.env.VIEWER_PASSWORD!,
      }
  }
}

async function enterCreds(page: Page, username: string, password: string) {
  const usernameInput = page.locator('input[type="text"]')
  const passwordInput = page.locator('input[type="password"]')

  await usernameInput.waitFor({ state: 'visible' })
  await usernameInput.click()
  await usernameInput.fill(username)

  await passwordInput.waitFor({ state: 'visible' })
  await passwordInput.click()
  await passwordInput.fill(password)

  await page.click('input[type="submit"]')
}

export const adminPage = test.extend<{ adminPage: Page }>({
  adminPage: async ({ browser }, run) => {
    console.log('Signing in as admin...')
    const page = await browser.newPage()
    await signIn(page, 'admin')
    try {
      await page.waitForURL(/\/console(\/.*)?$/, { timeout: 10000 })
    } catch {
      // no-op
    }
    await run(page)
  },
})

export const viewerPage = test.extend<{ viewerPage: Page }>({
  viewerPage: async ({ page }, run) => {
    console.log('Signing in as viewer...')
    await signIn(page, 'viewer')
    try {
      await page.waitForURL(/\/console(\/.*)?$/, { timeout: 4000 })
    } catch {
      // no-op
    }
    await run(page)
  },
})
