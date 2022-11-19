"""
#############
Notifications
#############

.. tags:: Intermediate

"""

# %%
# When a workflow is completed, users can be notified by:
#
# *  Email
# * `Pagerduty <https://support.pagerduty.com/docs/email-integration-guide#integrating-with-a-pagerduty-service>`__
# * `Slack <https://slack.com/help/articles/206819278-Send-emails-to-Slack>`__
#
# The content of these notifications is configurable at the platform level.
#
# ************
# Code Example
# ************
#
# When a workflow reaches a specified `terminal workflow execution phase <https://github.com/flyteorg/flytekit/blob/v0.16.0b7/flytekit/core/notification.py#L10,L15>`__,
# the :py:class:`flytekit:flytekit.Email`, :py:class:`flytekit:flytekit.PagerDuty`, or :py:class:`flytekit:flytekit.Slack`
# objects can be used in the construction of a :py:class:`flytekit:flytekit.LaunchPlan`.

from datetime import timedelta

# %%
# Consider the following example workflow:
from flytekit import Email, FixedRate, LaunchPlan, PagerDuty, Slack, WorkflowExecutionPhase, task, workflow


@task
def double_int_and_print(a: int) -> str:
    return str(a * 2)


@workflow
def int_doubler_wf(a: int) -> str:
    doubled = double_int_and_print(a=a)
    return doubled


# %%
# Here are three scenarios that can help deepen your understanding of how notifications work:
#
# 1. Launch Plan triggers email notifications when the workflow execution reaches the ``SUCCEEDED`` phase.
int_doubler_wf_lp = LaunchPlan.get_or_create(
    name="int_doubler_wf",
    workflow=int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        Email(
            phases=[WorkflowExecutionPhase.SUCCEEDED],
            recipients_email=["admin@example.com"],
        )
    ],
)

# %%
# 2. Notifications shine when used for scheduled workflows to alert for failures.


int_doubler_wf_scheduled_lp = LaunchPlan.get_or_create(
    name="int_doubler_wf_scheduled",
    workflow=int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        PagerDuty(
            phases=[WorkflowExecutionPhase.FAILED, WorkflowExecutionPhase.TIMED_OUT],
            recipients_email=["abc@pagerduty.com"],
        )
    ],
    schedule=FixedRate(duration=timedelta(days=1)),
)


# %%
# 3. Notifications can be combined with different permutations of terminal phases and recipient targets.

wacky_int_doubler_lp = LaunchPlan.get_or_create(
    name="wacky_int_doubler",
    workflow=int_doubler_wf,
    default_inputs={"a": 4},
    notifications=[
        Email(
            phases=[WorkflowExecutionPhase.FAILED],
            recipients_email=["me@example.com", "you@example.com"],
        ),
        Email(
            phases=[WorkflowExecutionPhase.SUCCEEDED],
            recipients_email=["myboss@example.com"],
        ),
        Slack(
            phases=[
                WorkflowExecutionPhase.SUCCEEDED,
                WorkflowExecutionPhase.ABORTED,
                WorkflowExecutionPhase.TIMED_OUT,
            ],
            recipients_email=["myteam@slack.com"],
        ),
    ],
)

# %%
# Future work
# ===========
# Work is ongoing to support a generic event egress system that can be used to publish events for tasks, workflows, and
# workflow nodes. When this is complete, generic event subscribers can asynchronously process these events for a rich
# and fully customizable experience.
#
# ******************************
# Platform Configuration Changes
# ******************************
#
# The ``notifications`` top-level portion of the Flyteadmin config specifies how to handle notifications.
#
# As in schedules, the handling of notifications is composed of two parts— one part handles enqueuing notifications asynchronously. The other part handles processing pending notifications and sends out emails and alerts.
#
# This is only supported for Flyte instances running on AWS.
#
# Config
# ======
#
# To publish notifications, you'll need to set up an `SNS topic <https://aws.amazon.com/sns/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc>`_.
#
# To process notifications, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue to consume notification events. This queue must be configured as a subscription to your SNS topic you created above.
#
# To publish notifications, you'll need a `verified SES email address <https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html>`_ which will be used to send notification emails and alerts using email APIs.
#
# The role you use to run Flyteadmin must have permissions to read and write to your SNS topic and SQS queue.
#
# Let's look into the following config section and explain what each value represents:
#
# .. code-block:: bash
#
#    notifications:
#      type: "aws"      # noqa: F821
#      region: "us-east-1"
#      publisher:
#        topicName: "arn:aws:sns:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR TOPIC }}"
#      processor:
#        queueName: "{{ YOUR QUEUE NAME }}"
#        accountId: "{{ YOUR ACCOUNT ID }}"
#      emailer:
#        subject: "Notice: Execution \"{{ workflow.name }}\" has {{ phase }} in \"{{ domain }}\"."
#        sender:  "flyte-notifications@company.com"
#        body: >
#           Execution \"{{ workflow.name }} [{{ name }}]\" has {{ phase }} in \"{{ domain }}\". View details at
#           <a href=\http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}>
#           http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}</a>. {{ error }}
#
# * **type**: AWS is the only cloud back-end supported for executing scheduled workflows; hence ``"aws"`` is the only valid value. By default, the no-op executor is used.
# * **region**: Specifies the region AWS clients should use when creating SNS and SQS clients.
# * **publisher**: Handles pushing notification events to your SNS topic.
#     * **topicName**: This is the arn of your SNS topic.
# * **processor**: Handles recording notification events and enqueueing them to be processed asynchronously.
#     * **queueName**: Name of the SQS queue which will capture pending notification events.
#     * **accountId**: AWS `account id <https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId>`_.
# * **emailer**: Encloses config details for sending and formatting emails used as notifications.
#     * **subject**: Configurable subject line used in notification emails.
#     * **sender**: Your verified SES email sender.
#     * **body**: Configurable email body used in notifications.
#
# The complete set of parameters that can be used for email templating are checked in `here <https://github.com/flyteorg/flyteadmin/blob/a84223dab00dfa52d8ba1ed2d057e77b6c6ab6a7/pkg/async/notifications/email.go#L18,L30>`_.
