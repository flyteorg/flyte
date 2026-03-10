// eslint-ignore no-console
import { select } from '@inquirer/prompts'
import { exec, spawn } from 'child_process'
import fs from 'fs'
import yargs from 'yargs'
import { hideBin } from 'yargs/helpers'

function openBrowser(url) {
  const cmd =
    process.platform === 'darwin'
      ? `open "${url}"`
      : process.platform === 'win32'
        ? `start "${url}"`
        : `xdg-open "${url}"`
  exec(cmd, (err) => {
    if (err) console.error('Failed to open browser:', err.message)
  })
}

// Spawn a subprocess with stdout piped so we can watch for the Next.js ready
// signal and open the browser, while still forwarding all output to the
// terminal. Returns the subprocess.
function spawnAndOpenWhenReady(command, args, env, url) {
  const subprocess = spawn(command, args, {
    env,
    stdio: ['inherit', 'pipe', 'inherit'],
    shell: true,
  })

  let opened = false
  subprocess.stdout.on('data', (data) => {
    process.stdout.write(data)
    if (!opened && /ready/i.test(data.toString())) {
      opened = true
      openBrowser(url)
    }
  })

  subprocess.on('close', (code) => {
    console.log(`Process exited with code ${code}`)
    process.exit(code)
  })

  return subprocess
}

const hosts = [
  'localhost',
  'byok.us-west-2.union.ai',
  'demo.hosted.unionai.cloud',
  'dogfood.cloud-staging.union.ai',
  'dogfood-gcp.cloud-staging.union.ai',
  'hosted.cloud-staging.union.ai',
  'playground.canary.unionai.cloud',
  'playground-gcp.canary.unionai.cloud',
  'serverless.union.ai',
  'serverless.staging.union.ai',
  'serverless-gcp.cloud-staging.union.ai',
  'serverless-1.us-east-2.s.union.ai',
  'utt-mgdp-stg-us-west-2.us-west-2.union.ai',
  'serving-mvp.us-west-2.union.ai',
  'union-internal.hosted.unionai.cloud',
]

const argv = yargs(hideBin(process.argv))
  .option('host', {
    alias: 'h',
    description: 'Specify the host',
    type: 'string',
  })
  .parse()

async function selectHost() {
  if (argv.host) {
    return argv.host
  }

  return select({
    message: 'choose a host',
    choices: hosts,
  })
}

function validateHost(host) {
  const etcPath = '/etc/hosts'
  try {
    const hostsFileContent = fs.readFileSync(etcPath, 'utf8')
    return hostsFileContent.includes(host)
  } catch (err) {
    console.error(`Error reading ${etcPath}:`, err.message)
    return false
  }
}

const BASE_OPTIONS = {
  PORT: 8080,
  NODE_ENV: 'development',
}
const main = async () => {
  const host = await selectHost()
  if (host === 'localhost') {
    const port = BASE_OPTIONS.PORT
    spawnAndOpenWhenReady(
      'pnpm',
      ['run dev:local'],
      {
        ...process.env,
        ...BASE_OPTIONS,
        NEXT_PUBLIC_ADMIN_API_URL: 'http://localhost:30080',
      },
      `http://localhost:${port}`,
    )
  } else {
    if (!validateHost(host)) {
      console.error(`You must add ${host} host to /etc/hosts`)
      process.exit(0)
    }
    const port = BASE_OPTIONS.PORT
    const localHost = `localhost.${host}`
    console.log(`Running: pnpm`)
    spawnAndOpenWhenReady(
      'pnpm',
      ['gen:ssl && NODE_ENV=development pnpm run dev:next'],
      {
        ...process.env,
        ...BASE_OPTIONS,
        NEXT_PUBLIC_ADMIN_DOMAIN: localHost,
        NEXT_PUBLIC_ADMIN_API_URL: `https://${host}`,
      },
      `https://${localHost}:${port}/v2`,
    )
  }
}
main()
