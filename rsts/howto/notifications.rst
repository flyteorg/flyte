.. _howto-notifications:

####################################################
How do I configure Flyte for scheduling Workflows?
####################################################

.. todo:: combine docs

Setting up scheduled workflows
==============================

In order to run workflow executions based on user-specified schedules you'll need to fill out the top-level ``scheduler`` portion of the flyteadmin application configuration.

In particular you'll need to configure the two components responsible for scheduling workflows and processing schedule event triggers.

Note this functionality is currently only supported for AWS installs.

Set-up the Event Scheduler
--------------------------

In order to schedule workflow executions, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue. A standard type queue should suffice. The flyteadmin event scheduler creates `AWS CloudWatch <https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/Create-CloudWatch-Events-Scheduled-Rule.html>`_ event rules that invokes your SQS queue as a target.

With that in mind, let's take a look at an example ``eventScheduler`` config section and dive into what each value represents: ::

    scheduler:
      eventScheduler:
        scheme: "aws"
        region: "us-east-1"
        scheduleRole: "arn:aws:iam::{{ YOUR ACCOUNT ID }}:role/{{ ROLE }}"
        targetName: "arn:aws:sqs:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR QUEUE NAME }}"
        scheduleNamePrefix: "flyte"

* **scheme**: in this case because AWS is the only cloud back-end supported for scheduling workflows, only ``"aws"`` is a valid value. By default, the no-op scheduler is used.
* **region**: this specifies which region initialized AWS clients should will use when creating CloudWatch rules
* **scheduleRole** This is the IAM role ARN with permissions set to ``Allow``
    * 'events:PutRule'
    * 'events:PutTargets'
    * 'events:DeleteRule'
    * 'events:RemoveTargets'
* **targetName** this is the ARN for the SQS Queue you've allocated to scheduling workflows
* **scheduleNamePrefix** this is an entirely optional prefix used when creating schedule rules. Because of AWS naming length restrictions, scheduled rules are a random hash and having a shared prefix makes these names more readable and indicates who generated the rules

Set-up the Workflow Executor
----------------------------
Scheduled events which trigger need to be handled by the workflow executor, which subscribes to triggered events from the SQS queue you've configured above.

Again, let's break down a sample config: ::

    scheduler:
      eventScheduler:
        ...
      workflowExecutor:
        scheme: "aws"
        region: "us-east-1"
        scheduleQueueName: "{{ YOUR QUEUE NAME }}"
        accountId: "{{ YOUR ACCOUNT ID }}"

* **scheme**: in this case because AWS is the only cloud back-end supported for executing scheduled workflows, only ``"aws"`` is a valid value. By default, the no-op executor is used.
* **region**: this specifies which region AWS clients should will use when creating an SQS subscriber client
* **scheduleQueueName**: this is the name of the SQS Queue you've allocated to scheduling workflows
* **accountId**: Your AWS `account id <https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId>`_

.. CAUTION::
   Failure to configure a workflow executor will result in all your scheduled events piling up silently without ever kicking off workflow executions.

Setting up workflow notifications
=================================

The ``notifications`` top-level portion of the flyteadmin config specifies how to handle notifications.

As like in schedules, the notifications handling is composed of two parts. One handles enqueuing notifications asynchronously and the second part handles processing pending notifications and actually firing off emails and alerts.

This is only supported for Flyte instances running on AWS.

Config
------

To publish notifications, you'll need to set up an `SNS topic <https://aws.amazon.com/sns/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc>`_.

In order to process notifications, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue to consume notification events. This queue must be configured as a subscription to your SNS topic you created above.

In order to actually publish notifications, you'll need a `verified SES email address <https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-addresses-and-domains.html>`_ which will be used to send notification emails and alerts using email APIs.

The role you use to run flyteadmin must have permissions to read and write to your SNS topic and SQS queue.

Let's look at the following config section and go into what each value represents: ::

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

* **type**: in this case because AWS is the only cloud back-end supported for executing scheduled workflows, only ``"aws"`` is a valid value. By default, the no-op executor is used.
* **region**: this specifies which region AWS clients should will use when creating SNS and SQS clients
* **publisher**: This handles pushing notification events to your SNS topic
    * **topicName**: This is the arn of your SNS topic
* **processor**: This handles the recording notification events and enqueueing them to be processed asynchronously
    * **queueName**: This is the name of the SQS queue which will capture pending notification events
    * **accountId**: Your AWS `account id <https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId>`_
* **emailer**: This section encloses config details for sending and formatting emails used as notifications
    * **subject**: Configurable subject line used in notification emails
    * **sender**: Your verified SES email sender
    * **body**: Configurable email body used in notifications

The full set of parameters which can be used for email templating are checked into `code <https://github.com/lyft/flyteadmin/blob/a84223dab00dfa52d8ba1ed2d057e77b6c6ab6a7/pkg/async/notifications/email.go#L18,L30>`_.

.. _admin-config-example:

Example config
==============

.. rli:: https://raw.githubusercontent.com/flyteorg/flyteadmin/master/flyteadmin_config.yaml


FlyteAdmin Remote Cluster Access
================================

Some deployments of Flyte may choose to run the control plane separate from the data plane. Flyte Admin is designed to create kubernetes resources in one or more Flyte data plane clusters. For Admin to access remote clusters, it needs credentials to each cluster. In kubernetes, scoped service credentials are created by configuring a “Role” resource in a Kubernetes cluster. When you attach that role to a “ServiceAccount”, Kubernetes generates a bearer token that permits access. We create a flyteadmin `ServiceAccount <https://github.com/lyft/flyte/blob/c0339e7cc4550a9b7eb78d6fb4fc3884d65ea945/artifacts/base/adminserviceaccount/adminserviceaccount.yaml>`_ in each data plane cluster to generate these tokens.

When you first create the Flyte Admin ServiceAccount in a new cluster, a bearer token is generated, and will continue to allow access unless the ServiceAccount is deleted. Once we create the Flyte Admin ServiceAccount on a cluster, we should never delete it. In order to feed the credentials to Flyte Admin, you must retrieve them from your new data plane cluster, and upload them to Admin somehow (within Lyft, we use Confidant for example). 

The credentials have two parts (ca cert, bearer token). Find the generated secret via ::

  kubectl get secrets -n flyte | grep flyteadmin-token

Once you have the name of the secret, you can copy the ca cert to your clipboard with ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.ca\.crt}' | base64 -D | pbcopy

You can copy the bearer token to your clipboard with ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.token}’ | base64 -D | pbcopy

