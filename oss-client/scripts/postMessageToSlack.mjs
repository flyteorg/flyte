import yargs from 'yargs'
import { hideBin } from 'yargs/helpers'

// SLACK_APP_TOKEN should be stored in the AWS secret pulled by e2e-tests.sh / e2e-seed.sh.
// SLACK_APP_CHANNEL defaults to 'team-ux-e2e' but can be overridden with --channel.

const SLACK_APP_CHANNEL = 'team-ux-e2e'

const postMessageToSlack = async () => {
  if (!process.env.SLACK_APP_TOKEN) {
    console.warn('SLACK_APP_TOKEN not set, skipping Slack notification')
    return
  }

  const argv = yargs(hideBin(process.argv))
    .option('message', {
      alias: 'm',
      description: 'Plaintext message to post',
      type: 'string',
    })
    .option('channel', {
      alias: 'c',
      description: 'Channel to post',
      type: 'string',
    })
    .parse()

  if (!argv.message) {
    throw new Error(
      'Cannot post to Slack without a message, use --message "This is a message"',
    )
  }

  const channel = argv.channel || SLACK_APP_CHANNEL

  try {
    const data = JSON.stringify({
      channel,
      text: argv.message,
    })
    const res = await fetch('https://slack.com/api/chat.postMessage', {
      body: data,
      headers: {
        Authorization: `Bearer ${process.env.SLACK_APP_TOKEN}`,
        'Content-Type': 'application/json; charset=utf-8',
        Accept: 'application/json',
      },
      method: 'POST',
    })
    const json = await res.json()
    if (json.ok !== true) {
      throw new Error(JSON.stringify(json))
    }
  } catch (e) {
    console.error('Failed to post to Slack', e)
  }
}

postMessageToSlack()
