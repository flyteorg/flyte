"""
Notifications
-------------

When a workflow completes, users can be notified by:

* email
* `pagerduty <https://www.pagerduty.com/>`__
* `slack <https://slack.com/>`__

The content of these notifications is configurable at the platform level.

*****
Usage
*****

When a workflow reaches a specified `terminal workflow execution phase <https://github.com/flyteorg/flytekit/blob/v0.16.0b7/flytekit/core/notification.py#L10,L15>`__
the :py:class:`flytekit:flytekit.Email`, :py:class:`flytekit:flytekit.PagerDuty`, or :py:class:`flytekit:flytekit.Slack`
objects can be used in the construction of a :py:class:`flytekit:flytekit.LaunchPlan`.

For example

.. code:: python

    from flytekit import Email, LaunchPlan
    from flytekit.models.core.execution import WorkflowExecutionPhase

    # This launch plan triggers email notifications when the workflow execution it triggered reaches the phase `SUCCEEDED`.
    my_notifiying_lp = LaunchPlan.create(
        "my_notifiying_lp",
        my_workflow_definition,
        default_inputs={"a": 4},
        notifications=[
            Email(
                phases=[WorkflowExecutionPhase.SUCCEEDED],
                recipients_email=["admin@example.com"],
            )
        ],
    )


See detailed usage examples in the :std:ref:`User Guide <cookbook:sphx_glr_auto_deployment_workflow_lp_notifications.py>`

Notifications can be combined with schedules to automatically alert you when a scheduled job succeeds or fails.

Future work
===========

Work is ongoing to support a generic event egress system that can be used to publish events for tasks, workflows and
workflow nodes. When this is complete, generic event subscribers can asynchronously process these vents for a rich
and fully customizable experience.


******************************
Platform Configuration Changes
******************************

Setting Up Workflow Notifications
=================================

The ``notifications`` top-level portion of the FlyteAdmin config specifies how to handle notifications.

As with schedules, the notifications handling is composed of two parts. One handles enqueuing notifications asynchronously and the second part handles processing pending notifications and actually firing off emails and alerts.

This is only supported for Flyte instances running on AWS.

Config
=======

To publish notifications, you'll need to set up an `SNS topic <https://aws.amazon.com/sns/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc>`_.

In order to process notifications, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue to consume notification events. This queue must be configured as a subscription to your SNS topic you created above.

In order to actually publish notifications, you'll need a `verified SES email address <https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html>`_ which will be used to send notification emails and alerts using email APIs.

The role you use to run FlyteAdmin must have permissions to read and write to your SNS topic and SQS queue.

Let's look at the following config section and explain what each value represents: ::

  notifications:
    type: "aws"
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

* **type**: Because AWS is the only cloud back-end supported for executing scheduled workflows in this case, only ``"aws"`` is a valid value. By default, the no-op executor is used.
* **region**: This specifies which region AWS clients will use when creating SNS and SQS clients.
* **publisher**: This handles pushing notification events to your SNS topic.
    * **topicName**: This is the arn of your SNS topic.
* **processor**: This handles the recording notification events and enqueueing them to be processed asynchronously.
    * **queueName**: This is the name of the SQS queue which will capture pending notification events.
    * **accountId**: Your AWS `account id <https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId>`_
* **emailer**: This section encloses config details for sending and formatting emails used as notifications.
    * **subject**: Configurable subject line used in notification emails.
    * **sender**: Your verified SES email sender.
    * **body**: Configurable email body used in notifications.

The full set of parameters which can be used for email templating are checked into `code <https://github.com/flyteorg/flyteadmin/blob/a84223dab00dfa52d8ba1ed2d057e77b6c6ab6a7/pkg/async/notifications/email.go#L18,L30>`_.

.. _admin-config-example:

Example config
==============

.. rli:: https://raw.githubusercontent.com/flyteorg/flyteadmin/master/flyteadmin_config.yaml
   :lines: 66-80

"""
