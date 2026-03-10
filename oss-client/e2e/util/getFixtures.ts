import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

type Fixture = {
  env: string
  name: string
  logs?: string
  display_name?: string // name of function
  id?: string //
  url?: string
  // eslint-disable-next-line
  inputs?: any
  // eslint-disable-next-line
  outputs?: any
  // triggers
  interval?: number
}

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

export let fixtures: Record<string, Fixture> = {}

const fixturePath = path.resolve(__dirname, '../scripts/fixtures.json')

if (fs.existsSync(fixturePath)) {
  try {
    fixtures = JSON.parse(fs.readFileSync(fixturePath, 'utf-8'))
  } catch (e) {
    console.error('⚠️ Failed to parse fixtures:', e)
  }
} else {
  console.warn(
    `⚠️ No fixtures.json found at ${fixturePath}. Did the seed step run?`,
  )
}

export const getFixture = (expectedKey: string) => {
  if (fixtures[expectedKey]) {
    return fixtures[expectedKey]
  } else {
    console.warn('⚠️ No fixture found with key - ', expectedKey)
    return null
  }
}
