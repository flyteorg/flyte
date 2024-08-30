.. _deployment-configuration-notifications:

#############
Notifications
#############

.. tags:: Infrastructure, Advanced

When a workflow completes, users can be notified by email,  `Pagerduty <https://support.pagerduty.com/docs/email-integration-guide#integrating-with-a-pagerduty-service>`__,
or `Slack <https://slack.com/help/articles/206819278-Send-emails-to-Slack>`__.

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


See detailed usage examples in the :std:doc:`/user_guide/productionizing/notifications`

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

This is only supported for Flyte instances running on AWS or GCP.

AWS Config
==========

To publish notifications, you'll need to set up an `SNS topic <https://aws.amazon.com/sns/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc>`_.

In order to process notifications, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue to consume notification events. This queue must be configured as a subscription to your SNS topic you created above.

In order to actually publish notifications, you'll need a `verified SES email address <https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html>`_ which will be used to send notification emails and alerts using email APIs.

The role you use to run FlyteAdmin must have permissions to read and write to your SNS topic and SQS queue.

Let's look at the following config section and explain what each value represents:

.. code-block:: yaml

  notifications:
    # By default, the no-op executor is used.
    type: "aws"

    # This specifies which region AWS clients will use when creating SNS and SQS clients.
    region: "us-east-1"

    # This handles pushing notification events to your SNS topic.
    publisher:

      # This is the arn of your SNS topic.
      topicName: "arn:aws:sns:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR TOPIC }}"

    # This handles the recording notification events and enqueueing them to be
    # processed asynchronously.
    processor:

      # This is the name of the SQS queue which will capture pending notification events.
      queueName: "{{ YOUR QUEUE NAME }}"

      # Your AWS `account id, see: https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId
      accountId: "{{ YOUR ACCOUNT ID }}"

    # This section encloses config details for sending and formatting emails
    # used as notifications.
    emailer:

      # Configurable subject line used in notification emails.
      subject: "Notice: Execution \"{{ workflow.name }}\" has {{ phase }} in \"{{ domain }}\"."

      # Your verified SES email sender.
      sender:  "flyte-notifications@company.com"

      # Configurable email body used in notifications.
      body: >
        Execution \"{{ workflow.name }} [{{ name }}]\" has {{ phase }} in \"{{ domain }}\". View details at
        <a href=\http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}>
        http://flyte.company.com/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}</a>. {{ error }}

The full set of parameters which can be used for email templating are checked
into `code <https://github.com/flyteorg/flyteadmin/blob/a84223dab00dfa52d8ba1ed2d057e77b6c6ab6a7/pkg/async/notifications/email.go#L18,L30>`_.

.. _admin-config-example:

Example config
--------------

You can find the full configuration file `here <https://github.com/flyteorg/flyteadmin/blob/master/flyteadmin_config.yaml>`__.

.. rli:: https://raw.githubusercontent.com/flyteorg/flyteadmin/master/flyteadmin_config.yaml
   :caption: flyteadmin/flyteadmin_config.yaml
   :lines: 91-105

GCP Config
==========

You'll need to set up a `Pub/Sub topic <https://cloud.google.com/pubsub/docs/create-topic>`__ to publish notifications to, 
and a `Pub/Sub subscriber <https://cloud.google.com/pubsub/docs/subscription-overview>`__ to consume from that topic 
and process notifications. The GCP service account used by FlyteAdmin must also have Pub/Sub publish and subscribe permissions.

Email service
-------------

In order to actually publish notifications, you'll need an account with an external email service which will be 
used to send notification emails and alerts using email APIs. 

Currently, `SendGrid <https://sendgrid.com/en-us>`__ is the only supported external email service, 
and you will need to have a verified SendGrid sender. Create a SendGrid API key with ``Mail Send`` permissions 
and save it to a file ``key``. 

Create a K8s secret in FlyteAdmin's cluster with that file:

.. prompt:: bash $
  
    kubectl create secret generic -n flyte --from-file key sendgrid-key

Mount the secret by adding the following to the ``flyte-core`` values YAML:

.. code-block:: yaml

    flyteadmin:
      additionalVolumes:
      - name: sendgrid-key
        secret:
          secretName: sendgrid-key
          items:
            - key: key
              path: key
      additionalVolumeMounts:
      - name: sendgrid-key
        mountPath: /sendgrid

Config
------

In the ``flyte-core`` values YAML, the top-level ``notifications`` config should be
placed under ``workflow_notifications``.

.. code-block:: yaml

    workflow_notifications:
      enabled: true
      config:
        notifications:
          type: gcp
          gcp:
            projectId: "{{ YOUR PROJECT ID }}"
          publisher:
            topicName: "{{ YOUR PUB/SUB TOPIC NAME }}"
          processor:
            queueName: "{{ YOUR PUB/SUB SUBSCRIBER NAME }}"
          emailer:
            emailServerConfig:
              serviceName: sendgrid
              apiKeyFilePath: /sendgrid/key
            subject: "Flyte execution \"{{ name }}\" has {{ phase }} in \"{{ project }}\"."
            sender: "{{ YOUR SENDGRID SENDER EMAIL }}"
            body: View details at <a href=https://{{ YOUR FLYTE HOST }}/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}>https://{{ YOUR FLYTE HOST }}/console/projects/{{ project }}/domains/{{ domain }}/executions/{{ name }}</a>
