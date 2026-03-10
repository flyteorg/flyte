const SLACK_APP_CHANNEL = 'team-ux-e2e'

/**
 * Post a message to Slack from within a Playwright test.
 * Soft-fails with a console warning if SLACK_APP_TOKEN is not set.
 *
 * Typical usage — skip a test suite and notify:
 *
 *   const fixture = getFixture('hello_app')
 *   if (!fixture) {
 *     await postToSlack('Skipping app tests because the deployment failed')
 *     test.skip()
 *     return
 *   }
 */
export async function postToSlack(message: string, channel = SLACK_APP_CHANNEL): Promise<void> {
  const token = process.env.SLACK_APP_TOKEN
  if (!token) {
    console.warn('SLACK_APP_TOKEN not set, skipping Slack notification')
    return
  }

  try {
    const res = await fetch('https://slack.com/api/chat.postMessage', {
      method: 'POST',
      body: JSON.stringify({ channel, text: message }),
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json; charset=utf-8',
        Accept: 'application/json',
      },
    })
    const json = (await res.json()) as { ok: boolean }
    if (!json.ok) {
      throw new Error(JSON.stringify(json))
    }
  } catch (e) {
    console.error('Failed to post to Slack', e)
  }
}
