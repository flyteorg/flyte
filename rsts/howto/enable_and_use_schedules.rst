.. _howto_scheduling:

#################################################
How do I use Flyte scheduling?
#################################################

*******
Usage
*******

Launch plans can be set to run automatically on a schedule if the Flyte platform is properly configured.
You can even use the scheduled kick-off time in your workflow as an input.

There are two types of schedules, cron schedules, and fixed rate intervals.

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

Complete cron example
---------------------

For example, take the following workflow:

.. code:: python

    from flytekit workflow

    @workflow
    def MyWorkflow(an_input: int, another_input: int=10):
        ....

The above can be run on a cron schedule every 5 minutes like so:

.. code:: python

    from flytekit import CronSchedule, LaunchPlan

    cron_lp = LaunchPlan.create(
        "my_cron_lp",
        MyWorkflow,
        schedule=CronSchedule(cron_expression="0 5 * * ? *"),
        fixed_inputs={"an_input": 5},
    )


Fixed Rate Intervals
====================

Fixed rate schedules will run at the specified interval.

.. code-block::

    from flytekit import FixedRate
    from datetime import timedelta

    schedule = FixedRate(duration=timedelta(minutes=10))


Complete fixed rate example
---------------------------

.. code:: python

    from flytekit workflow

    @workflow
    def MyOtherWorkflow(triggered_time: datetime, an_input: int, another_input: int=10):
        ....


To run ``MyOtherWorkflow`` every 5 minutes with a value set for ``an_input`` and the scheduled execution time
assigned to the ``triggered_time`` input you could define the following launch plan:

.. code:: python

    from datetime import timedelta
    from flytekit import FixedRate, LaunchPlan

    fixed_rate_lp = LaunchPlan.create(
        "my_fixed_rate_lp",
        MyOtherWorkflow,
        # Note that kickoff_time_input_arg matches the workflow input we defined above: triggered_time
        schedule=FixedRate(duration=timedelta(minutes=5), kickoff_time_input_arg="triggered_time"),
        fixed_inputs={"an_input": 3},
    )

Please see a more complete example in the :std:ref:`User Guide <cookbook:sphx_glr_auto_deployment_workflow_lp_schedules.py>`.

Activating a schedule
=====================

Once you've initialized your launch plan, don't forget to set it to active so that the schedule is run.

You can use pyflyte in container ::

  pyflyte lp -p {{ your project }} -d {{ your domain }} activate-all

Or with flyte-cli view and activate launch plans ::

  flyte-cli -i -h localhost:30081 -p flyteexamples -d development list-launch-plan-versions

Extract the URN returned for the launch plan you're interested in and make the call to activate it ::

  flyte-cli update-launch-plan -i -h localhost:30081 --state active -u {{ urn }}

Verify your active launch plans::

  flyte-cli -i -h localhost:30081 -p flyteexamples -d development list-active-launch-plans

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

.. NOTE::

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
