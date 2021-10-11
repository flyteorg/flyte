"""
Scheduling Workflows
--------------------
For background on launch plans, refer to :any:`launch_plans`.
Launch plans can be set to run automatically on a schedule using the flyte native scheduler.
For workflows that depend on knowing the kick-off time, Flyte also supports passing in the scheduled time (not the actual time, which may be a few seconds off) as an argument to the workflow. 

.. note::

  Native scheduler doesn't support `AWS syntax <http://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions>`_.

Check out a demo of how Native Scheduler works below:

.. youtube:: sQoCp2qSQK4

"""

# %%
# Consider the following example workflow:
from datetime import datetime

from flytekit import task, workflow


@task
def format_date(run_date: datetime) -> str:
    return run_date.strftime("%Y-%m-%d %H:%M")


@workflow
def date_formatter_wf(kickoff_time: datetime):
    formatted_kickoff_time = format_date(run_date=kickoff_time)
    print(formatted_kickoff_time)


# %%
# The `date_formatter_wf` workflow can be scheduled using either the `CronSchedule` or the `FixedRate` object.
#
# Cron Schedules
# ##############
#
# Cron expression strings use the following `syntax <https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format>`_.
# An incorrect cron schedule expression would lead to failure in triggering the schedule
from flytekit import CronSchedule, LaunchPlan

# creates a launch plan that runs at 10am UTC every day.
cron_lp = LaunchPlan.get_or_create(
    name="my_cron_scheduled_lp",
    workflow=date_formatter_wf,
    schedule=CronSchedule(
        # Note that kickoff_time_input_arg matches the workflow input we defined above: kickoff_time
        # But in case you are using the AWS scheme of schedules and not using the native scheduler then switch over the schedule parameter with cron_expression
        schedule="*/1 * * * *", # Following schedule runs every min 
        kickoff_time_input_arg="kickoff_time",
    ),
)

# %%
# The ``kickoff_time_input_arg`` corresponds to the workflow input ``kickoff_time``. This means that the workflow gets triggered only after the specified kickoff time, and it thereby runs at 10 AM UTC every day.

# %%
# Fixed Rate Intervals
# ####################
#     
# If you prefer to use an interval rather than a cron scheduler to schedule your workflows, you can use the fixed-rate scheduler. 
# A fixed-rate scheduler runs at the specified interval.
#
# Here's an example:

from datetime import timedelta

from flytekit import FixedRate, LaunchPlan


@task
def be_positive(name: str) -> str:
    return f"You're awesome, {name}"


@workflow
def positive_wf(name: str):
    reminder = be_positive(name=name)
    print(f"{reminder}")


fixed_rate_lp = LaunchPlan.get_or_create(
    name="my_fixed_rate_lp",
    workflow=positive_wf,
    # Note that the workflow above doesn't accept any kickoff time arguments.
    # We just omit the ``kickoff_time_input_arg`` from the FixedRate schedule invocation
    schedule=FixedRate(duration=timedelta(minutes=10)),
    fixed_inputs={"name": "you"},
)

# %%
# This fixed-rate scheduler runs every ten minutes. Similar to a cron scheduler, a fixed-rate scheduler also accepts ``kickoff_time_input_arg`` (which is omitted in this example).
# 
# Activating a Schedule 
# #####################
#
# Once you've initialized your launch plan, don't forget to set it to active so that the schedule is run.
# You can use pyflyte in container:
#
# .. code-block:: bash
# 
#   pyflyte lp -p {{ your project }} -d {{ your domain }} activate-all

# %%
# (or)
#
# With flyte-cli:
#
# - View and activate launch plans: flyte-cli -i -h localhost:30081 -p flyteexamples -d development list-launch-plan-versions
# 
# .. code-block:: bash
#
#   flyte-cli -i -h localhost:30081 -p flyteexamples -d development list-launch-plan-versions
 
# %%
# - Extract the URN returned for the launch plan you're interested in and make the call to activate it: 
# 
# .. code-block:: bash
# 
#   flyte-cli update-launch-plan -i -h localhost:30081 --state active -u {{ urn }}
#
# .. tip::
#   The equivalent command in `flytectl <https://docs.flyte.org/projects/flytectl/en/latest/index.html>`__ is:
#
#   .. code-block:: bash
#
#       flytectl update launchplan -p flyteexamples -d development {{ name_of_lp }} --activate``
#   
#   Example: 
# 
#   .. code-block:: bash
#
#       flytectl update launchplan -p flyteexamples -d development core.basic.lp.go_greet --activate

# %%
# - Verify your active launch plans: 
# 
# .. code-block:: bash
# 
#   flyte-cli -i -h localhost:30081 -p flyteexamples -d development list-active-launch-plans
# 
# .. tip::
#   The equivalent command in `flytectl <https://docs.flyte.org/projects/flytectl/en/latest/index.html>`__ is:
#   
#   .. code-block:: bash 
#
#       flytectl get launchplan -p flytesnacks -d development``

# %%        
# Platform Configuration Changes For AWS Scheduler
# ################################################
# 
# Scheduling feature can be run using the flyte native scheduler which comes with flyte but if you intend to use the AWS scheduler then it require additional infrastructure to run, so these will have to be created and configured.The following sections are only required if you use AWS scheme for the scheduler. You can even run the flyte native scheduler on AWS though
# 
# Setting up Scheduled Workflows
# ==============================
# To run workflow executions based on user-specified schedules, you'll need to fill out the top-level ``scheduler`` portion of the flyteadmin application configuration.
#
# In particular, you'll need to configure the two components responsible for scheduling workflows and processing schedule event triggers.
#
# .. note::
#   This functionality is currently only supported for AWS installs.
#
# Event Scheduler
# ^^^^^^^^^^^^^^^
#
# To schedule workflow executions, you'll need to set up an `AWS SQS <https://aws.amazon.com/sqs/>`_ queue. A standard-type queue should suffice. The flyteadmin event scheduler creates `AWS CloudWatch <https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/Create-CloudWatch-Events-Scheduled-Rule.html>`_ event rules that invoke your SQS queue as a target.
#
# With that in mind, let's take a look at an example ``eventScheduler`` config section and dive into what each value represents: 
#
# .. code-block:: bash
# 
#    scheduler:
#      eventScheduler:
#        scheme: "aws"
#        region: "us-east-1"
#        scheduleRole: "arn:aws:iam::{{ YOUR ACCOUNT ID }}:role/{{ ROLE }}"
#        targetName: "arn:aws:sqs:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR QUEUE NAME }}"
#        scheduleNamePrefix: "flyte"

# %%            
# * **scheme**: in this case because AWS is the only cloud back-end supported for scheduling workflows, only ``"aws"`` is a valid value. By default, the no-op scheduler is used.
# * **region**: this specifies which region initialized AWS clients should will use when creating CloudWatch rules
# * **scheduleRole** This is the IAM role ARN with permissions set to ``Allow``
#     * ``events:PutRule``
#     * ``events:PutTargets``
#     * ``events:DeleteRule``
#     * ``events:RemoveTargets``
# * **targetName** this is the ARN for the SQS Queue you've allocated to scheduling workflows
# * **scheduleNamePrefix** this is an entirely optional prefix used when creating schedule rules. Because of AWS naming length restrictions, scheduled rules are a random hash and having a shared prefix makes these names more readable and indicates who generated the rules.
#
# Workflow Executor
# ^^^^^^^^^^^^^^^^^
# Scheduled events which trigger need to be handled by the workflow executor, which subscribes to triggered events from the SQS queue configured above.
#
# .. NOTE::
#
#    Failure to configure a workflow executor will result in all your scheduled events piling up silently without ever kicking off workflow executions.
#
# Again, let's break down a sample config:
#
# .. code-block:: bash
#
#    scheduler:
#      eventScheduler:
#        ...
#      workflowExecutor:
#        scheme: "aws"
#        region: "us-east-1"
#        scheduleQueueName: "{{ YOUR QUEUE NAME }}"
#        accountId: "{{ YOUR ACCOUNT ID }}"

# %%            
# * **scheme**: in this case because AWS is the only cloud back-end supported for executing scheduled workflows, only ``"aws"`` is a valid value. By default, the no-op executor is used and in case of sandbox we use ``"local"`` scheme which uses the flyte native scheduler.
# * **region**: this specifies which region AWS clients should will use when creating an SQS subscriber client
# * **scheduleQueueName**: this is the name of the SQS Queue you've allocated to scheduling workflows
# * **accountId**: Your AWS `account id <https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId>`_
