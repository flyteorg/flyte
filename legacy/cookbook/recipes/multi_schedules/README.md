[Back to Cookbook Menu](../..)

# How do I create schedules?


If you haven't read yet the [Launch Plans](https://lyft.github.io/flyte/user/concepts/launchplans_schedules.html) part of the documentation, please do so as it explains the basics of launch plans and schedules.

Schedules are set on launch plans, and as you can see from the [IDL](https://github.com/lyft/flyteidl/blob/e9727afcedf8d4c30a1fc2eeac45593e426d9bb0/protos/flyteidl/admin/schedule.proto#L20) can be either set with a traditional cron expression, or an interval.

Flyte uses common cron expression syntax so something like this will run every fifteen minutes, every day, from the top of the hour.

```python
SCHEDULE_EXPR = "0/15 * * * ? *"
```

The `create_launch_plan` function can take a schedule expression

```python
from workflows import workflows
from flytekit.common import schedules
scale_rotate_cronscheduled_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    schedule=schedules.CronSchedule("0/30 * * * ? *")
)
```

A similar result can be achieved with a fixed rate schedule,

```python
import datetime
from workflows import workflows
from flytekit.common import schedules
scale_rotate_fixedRateScheduled_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    schedule=schedules.FixedRate(duration=datetime.timedelta(minutes=30))
)
```

Please see the [schedules.py](schedules.py) file for the full example.

Note that, for convenience, when Flyte Admin launches a scheduled execution, it will include the scheduled time as an input argument, if you wish.  To use this feature, simply supply the name of the input that you'd like to receive it under to the schedule.  To clarify this is the scheduled time, not the time it actually launches - there's always going to be a few seconds worth of delay.

For example, if your workflow takes a datetime as an input named `trigger_time` and you would like it to receive the scheduled time at launch, you can specify it as,

```python
from multi_schedules.scheduled_workflow import ScheduledWorkflow
from flytekit.common import schedules
scheduled_time_lp = ScheduledWorkflow.create_launch_plan(
    schedule=schedules.CronSchedule("0/30 * * * ? *", kickoff_time_input_arg='trigger_time')
)
```
