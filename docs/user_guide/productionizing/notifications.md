# Notifications

```{eval-rst}
.. tags:: Intermediate

```

When a workflow is completed, users can be notified by:

- Email
- [Pagerduty](https://support.pagerduty.com/docs/email-integration-guide#integrating-with-a-pagerduty-service)
- [Slack](https://slack.com/help/articles/206819278-Send-emails-to-Slack)

The content of these notifications is configurable at the platform level.

## Code example

When a workflow reaches a specified [terminal workflow execution phase](https://github.com/flyteorg/flytekit/blob/v0.16.0b7/flytekit/core/notification.py#L10,L15), the {py:class}`flytekit:flytekit.Email`, {py:class}`flytekit:flytekit.PagerDuty`, or {py:class}`flytekit:flytekit.Slack` objects can be used in the construction of a {py:class}`flytekit:flytekit.LaunchPlan`.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/lp_notifications.py
:caption: productionizing/lp_notifications.py
:lines: 1
```

Consider the following example workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/lp_notifications.py
:caption: productionizing/lp_notifications.py
:lines: 3-14
```

Here are three scenarios that can help deepen your understanding of how notifications work:

1. Launch Plan triggers email notifications when the workflow execution reaches the `SUCCEEDED` phase.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/lp_notifications.py
:caption: productionizing/lp_notifications.py
:lines: 20-30
```

2. Notifications shine when used for scheduled workflows to alert for failures.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/lp_notifications.py
:caption: productionizing/lp_notifications.py
:lines: 33-44
```

3. Notifications can be combined with different permutations of terminal phases and recipient targets.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/lp_notifications.py
:caption: productionizing/lp_notifications.py
:lines: 48-70
```

4. You can use pyflyte register to register the launch plan and launch it in the web console to get the notifications.

```
pyflyte register lp_notifications.py
```

Choose the launch plan with notifications config
:::{figure} https://i.ibb.co/cLT5tRX/lp.png
:alt: Notifications Launch Plan
:class: with-shadow
:::


### Future work

Work is ongoing to support a generic event egress system that can be used to publish events for tasks, workflows, and workflow nodes. When this is complete, generic event subscribers can asynchronously process these events for a rich and fully customizable experience.

## Platform configuration changes

The `notifications` top-level portion of the Flyteadmin config specifies how to handle notifications.

As in schedules, the handling of notifications is composed of two parts— one part handles enqueuing notifications asynchronously. The other part handles processing pending notifications and sends out emails and alerts.

This is only supported for Flyte instances running on AWS.

### Config
#### For Sandbox
To publish notifications, you'll need to register a Sendgrid api key from [Sendgrid](https://sendgrid.com/), it's free for 100 emails per day. You have to add notifications config in your sandbox config file.

```yaml
# config-sandbox.yaml
notifications:
  type: sandbox  # noqa: F821
  emailer:
    emailServerConfig:
      serviceName: sendgrid
      apiKeyEnvVar: SENDGRID_API_KEY
    subject: "Notice: Execution \"{{ workflow.name }}\" has {{ phase }} in \"{{ domain }}\"."
    sender:  "flyte-notifications@company.com"
    body: >
       Execution \"{{ workflow.name }} [{{ name }}]\" has {{ phase }} in \"{{ domain }}\". View details at
       <a href=\http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}>
       http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}</a>. {{ error }}
```

Note that you should set and export the `SENDGRID_API_KEY` environment variable in your shell.

#### For AWS
To publish notifications, you'll need to set up an [SNS topic](https://aws.amazon.com/sns/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc).

To process notifications, you'll need to set up an [AWS SQS](https://aws.amazon.com/sqs/) queue to consume notification events. This queue must be configured as a subscription to your SNS topic you created above.

To publish notifications, you'll need a [verified SES email address](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html) which will be used to send notification emails and alerts using email APIs.

The role you use to run Flyteadmin must have permissions to read and write to your SNS topic and SQS queue.

Let's look into the following config section and explain what each value represents:

```bash
notifications:
  type: "aws"      # noqa: F821
  region: "us-east-1"
  publisher:
    topicName: "arn:aws:sns:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR TOPIC }}"
  processor:
    queueName: "{{ YOUR QUEUE NAME }}"
    accountId: "{{ YOUR ACCOUNT ID }}"
  emailer:
    subject: "Notice: Execution \"{{ workflow.name }}\" has {{ phase }} in \"{{ domain }}\"."
    sender:  "flyte-notifications@company.com"
    body: >
       Execution \"{{ workflow.name }} [{{ name }}]\" has {{ phase }} in \"{{ domain }}\". View details at
       <a href=\http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}>
       http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}</a>. {{ error }}
```

- **type**: AWS is the only cloud back-end supported for executing scheduled workflows; hence `"aws"` is the only valid value. By default, the no-op executor is used.
- **region**: Specifies the region AWS clients should use when creating SNS and SQS clients.
- **publisher**: Handles pushing notification events to your SNS topic.
  : - **topicName**: This is the arn of your SNS topic.
- **processor**: Handles recording notification events and enqueueing them to be processed asynchronously.
  : - **queueName**: Name of the SQS queue which will capture pending notification events.
    - **accountId**: AWS [account id](https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId).
- **emailer**: Encloses config details for sending and formatting emails used as notifications.
  : - **subject**: Configurable subject line used in notification emails.
    - **sender**: Your verified SES email sender.
    - **body**: Configurable email body used in notifications.

The complete set of parameters that can be used for email templating are checked in [here](https://github.com/flyteorg/flyteadmin/blob/a84223dab00dfa52d8ba1ed2d057e77b6c6ab6a7/pkg/async/notifications/email.go#L18,L30).


[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
