.. _howto_scheduling:

#################################################
How do I use Flyte scheduling?
#################################################

*******
Usage
*******

Launch plans can be set to run automatically on a schedule if the Flyte platform is properly configured. There are two types of schedules, cron schedules, and fixed rate intervals.

Cron Schedules
==============

Cron expression strings use the `AWS syntax <http://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions>`_.
These are validated at launch plan registration time.

.. code-block::

    from flytekit import CronSchedule

    schedule = CronSchedule(
        cron_expression="0 10 * * ? *",
    )


This ``schedule`` object can then be used in the construction of a :py:class:`flytekit:flytekit.LaunchPlan`.


Fixed Rate Intervals
====================

Fixed rate schedules will run at the specified interval.

.. code-block::

    from flytekit import FixedRate
    from datetime import timedelta

    schedule = FixedRate(duration=timedelta(minutes=10))

Please see a more complete example in the :std:ref:`cookbook <cookbook:launch_plans>`.


******************************
Platform Configuration Changes
******************************

Scheduling features requires additional infrastructure to run so these will have to be created and configured.

Setting up scheduled workflows
==============================

In order to run workflow executions based on user-specified schedules you'll need to fill out the top-level ``scheduler`` portion of the flyteadmin application configuration.

In particular you'll need to configure the two components responsible for scheduling workflows and processing schedule event triggers.

Note this functionality is currently only supported for AWS installs.

Event Scheduler
---------------

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
    * ``events:PutRule``
    * ``events:PutTargets``
    * ``events:DeleteRule``
    * ``events:RemoveTargets``
* **targetName** this is the ARN for the SQS Queue you've allocated to scheduling workflows
* **scheduleNamePrefix** this is an entirely optional prefix used when creating schedule rules. Because of AWS naming length restrictions, scheduled rules are a random hash and having a shared prefix makes these names more readable and indicates who generated the rules

Workflow Executor
-----------------
Scheduled events which trigger need to be handled by the workflow executor, which subscribes to triggered events from the SQS queue you've configured above.

.. CAUTION::

   Failure to configure a workflow executor will result in all your scheduled events piling up silently without ever kicking off workflow executions.

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
