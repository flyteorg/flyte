"""

Scheduling workflow executions with launch plans
----------------------------------------------------

For background on launch plans, refer to :any:`launch_plans`.

For up-to-date documentation on schedules, see the `official docs <https://lyft.github.io/flyte/user/concepts/launchplans_schedules.html#schedules>`_
"""

# %%
# Let's consider the following example workflow:
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
# Cron Expression
# ===============
# Cron expression strings use the `AWS syntax <http://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html#CronExpressions>`_.
# These are validated at launch plan registration time.
from flytekit import CronSchedule, LaunchPlan

# creates a launch plan that runs at 10am UTC every day.
cron_lp = LaunchPlan.create(
    "my_cron_scheduled_lp",
    date_formatter_wf,
    schedule=CronSchedule(
        # Note that kickoff_time_input_arg matches the workflow input we defined above: kickoff_time
        cron_expression="0 10 * * ? *",
        kickoff_time_input_arg="kickoff_time",
    ),
)

# %%
# Fixed Rate
# ==========
# If you prefer to use an interval rather than the cron syntax to schedule your workflows, this is currently supported
# for Flyte deployments hosted on AWS.
# To run ``date_formatter_wf`` every 10 minutes read on below:

from datetime import timedelta

from flytekit import FixedRate, LaunchPlan


@task
def be_positive(name: str) -> str:
    return f"You're awesome, {name}"


@workflow
def positive_wf(name: str):
    reminder = be_positive(name=name)
    print(f"{reminder}")


fixed_rate_lp = LaunchPlan.create(
    "my_fixed_rate_lp",
    positive_wf,
    # Note that the workflow above doesn't accept any kickoff time arguments.
    # We just omit the ``kickoff_time_input_arg`` from the FixedRate schedule invocation
    schedule=FixedRate(duration=timedelta(minutes=10)),
    fixed_inputs={"name": "you"},
)

# %%
# Once you've initialized your launch plan, don't forget to set it to active so that the schedule is run.
# TBD (katrogan)
