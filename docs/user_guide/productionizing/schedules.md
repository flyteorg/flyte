---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(scheduling_launch_plan)=

# Schedules

```{eval-rst}
.. tags:: Basic
```

{ref}`flyte:divedeep-launchplans` can be set to run automatically on a schedule using the Flyte Native Scheduler.
For workflows that depend on knowing the kick-off time, Flyte supports passing in the scheduled time (not the actual time, which may be a few seconds off) as an argument to the workflow.

Check out a demo of how the Native Scheduler works:

```{eval-rst}
.. youtube:: sQoCp2qSQK4
```

:::{note}
Native scheduler doesn't support [AWS syntax](http://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions).
:::

+++ {"lines_to_next_cell": 0}

Consider the following example workflow:

```{code-cell}
from datetime import datetime

from flytekit import task, workflow


@task
def format_date(run_date: datetime) -> str:
    return run_date.strftime("%Y-%m-%d %H:%M")


@workflow
def date_formatter_wf(kickoff_time: datetime):
    formatted_kickoff_time = format_date(run_date=kickoff_time)
    print(formatted_kickoff_time)
```

+++ {"lines_to_next_cell": 0}

The `date_formatter_wf` workflow can be scheduled using either the `CronSchedule` or the `FixedRate` object.

(cron-schedules)=

## Cron schedules

[Cron](https://en.wikipedia.org/wiki/Cron) expression strings use this {ref}`syntax <concepts-schedules>`.
An incorrect cron schedule expression would lead to failure in triggering the schedule.

```{code-cell}
from flytekit import CronSchedule, LaunchPlan  # noqa: E402

# creates a launch plan that runs every minute.
cron_lp = LaunchPlan.get_or_create(
    name="my_cron_scheduled_lp",
    workflow=date_formatter_wf,
    schedule=CronSchedule(
        # Note that the ``kickoff_time_input_arg`` matches the workflow input we defined above: kickoff_time
        # But in case you are using the AWS scheme of schedules and not using the native scheduler then switch over the schedule parameter with cron_expression
        schedule="*/1 * * * *",  # Following schedule runs every min
        kickoff_time_input_arg="kickoff_time",
    ),
)
```

The `kickoff_time_input_arg` corresponds to the workflow input `kickoff_time`.
Specifying this argument means that Flyte will pass in the kick-off time of the
cron schedule into the `kickoff_time` argument of the `date_formatter_wf` workflow.

+++

## Fixed rate intervals

If you prefer to use an interval rather than a cron scheduler to schedule your workflows, you can use the fixed-rate scheduler.
A fixed-rate scheduler runs at the specified interval.

Here's an example:

```{code-cell}
from datetime import timedelta  # noqa: E402

from flytekit import FixedRate, LaunchPlan  # noqa: E402


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
```

This fixed-rate scheduler runs every ten minutes. Similar to a cron scheduler, a fixed-rate scheduler also accepts `kickoff_time_input_arg` (which is omitted in this example).

(activating-schedules)=

## Activating a schedule

After initializing your launch plan, [activate the specific version of the launch plan](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html) so that the schedule runs.

```bash
flytectl update launchplan -p flyteexamples -d development {{ name_of_lp }} --version <foo> --activate
```

+++

Verify if your launch plan was activated:

```bash
flytectl get launchplan -p flytesnacks -d development
```

+++

## Deactivating a schedule

You can [archive/deactivate the launch plan](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html) to deschedule any scheduled job associated with it.

```bash
flytectl update launchplan -p flyteexamples -d development {{ name_of_lp }} --version <foo> --archive
```

+++

## Platform configuration changes for AWS scheduler

The Scheduling feature can be run using the Flyte native scheduler which comes with Flyte. If you intend to use the AWS scheduler then it requires additional infrastructure to run, so these will have to be created and configured. The following sections are only required if you use the AWS scheme for the scheduler. You can still run the Flyte native scheduler on AWS.

### Setting up scheduled workflows

To run workflow executions based on user-specified schedules, you'll need to fill out the top-level `scheduler` portion of the flyteadmin application configuration.

In particular, you'll need to configure the two components responsible for scheduling workflows and processing schedule event triggers.

:::{note}
This functionality is currently only supported for AWS installs.
:::

#### Event scheduler

To schedule workflow executions, you'll need to set up an [AWS SQS](https://aws.amazon.com/sqs/) queue. A standard-type queue should suffice. The flyteadmin event scheduler creates [AWS CloudWatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/Create-CloudWatch-Events-Scheduled-Rule.html) event rules that invoke your SQS queue as a target.

With that in mind, let's take a look at an example `eventScheduler` config section and dive into what each value represents:

```bash
scheduler:
  eventScheduler:
    scheme: "aws"
    region: "us-east-1"
    scheduleRole: "arn:aws:iam::{{ YOUR ACCOUNT ID }}:role/{{ ROLE }}"
    targetName: "arn:aws:sqs:us-east-1:{{ YOUR ACCOUNT ID }}:{{ YOUR QUEUE NAME }}"
    scheduleNamePrefix: "flyte"
```

+++

- **scheme**: in this case because AWS is the only cloud back-end supported for scheduling workflows, only `"aws"` is a valid value. By default, the no-op scheduler is used.
- **region**: this specifies which region initialized AWS clients should use when creating CloudWatch rules.
- **scheduleRole** This is the IAM role ARN with permissions set to `Allow`
  : - `events:PutRule`
    - `events:PutTargets`
    - `events:DeleteRule`
    - `events:RemoveTargets`
- **targetName** this is the ARN for the SQS Queue you've allocated to scheduling workflows.
- **scheduleNamePrefix** this is an entirely optional prefix used when creating schedule rules. Because of AWS naming length restrictions, scheduled rules are a random hash and having a shared prefix makes these names more readable and indicates who generated the rules.

#### Workflow executor

Scheduled events which trigger need to be handled by the workflow executor, which subscribes to triggered events from the SQS queue configured above.

:::{NOTE}
Failure to configure a workflow executor will result in all your scheduled events piling up silently without ever kicking off workflow executions.
:::

Again, let's break down a sample config:

```bash
scheduler:
  eventScheduler:
    ...
  workflowExecutor:
    scheme: "aws"
    region: "us-east-1"
    scheduleQueueName: "{{ YOUR QUEUE NAME }}"
    accountId: "{{ YOUR ACCOUNT ID }}"
```

+++

- **scheme**: in this case because AWS is the only cloud back-end supported for executing scheduled workflows, only `"aws"` is a valid value. By default, the no-op executor is used and in case of sandbox we use `"local"` scheme which uses the Flyte native scheduler.
- **region**: this specifies which region AWS clients should use when creating an SQS subscriber client.
- **scheduleQueueName**: this is the name of the SQS Queue you've allocated to scheduling workflows.
- **accountId**: Your AWS [account id](https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#FindingYourAWSId).
