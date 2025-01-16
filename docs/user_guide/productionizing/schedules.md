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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

Consider the following example workflow:

```{literalinclude} /examples/productionizing/productionizing/lp_schedules.py
:caption: productionizing/lp_schedules.py
:lines: 1-14
```

The `date_formatter_wf` workflow can be scheduled using either the `CronSchedule` or the `FixedRate` object.

(cron-schedules)=

## Cron schedules

[Cron](https://en.wikipedia.org/wiki/Cron) expression strings use this {ref}`syntax <concepts-schedules>`.
An incorrect cron schedule expression would lead to failure in triggering the schedule.

```{literalinclude} /examples/productionizing/productionizing/lp_schedules.py
:caption: productionizing/lp_schedules.py
:lines: 17-29
```

The `kickoff_time_input_arg` corresponds to the workflow input `kickoff_time`.
Specifying this argument means that Flyte will pass in the kick-off time of the
cron schedule into the `kickoff_time` argument of the `date_formatter_wf` workflow.

## Fixed rate intervals

If you prefer to use an interval rather than a cron scheduler to schedule your workflows, you can use the fixed-rate scheduler. A fixed-rate scheduler runs at the specified interval.

Here's an example:

```{literalinclude} /examples/productionizing/productionizing/lp_schedules.py
:caption: productionizing/lp_schedules.py
:lines: 34-57
```

This fixed-rate scheduler runs every ten minutes. Similar to a cron scheduler, a fixed-rate scheduler also accepts `kickoff_time_input_arg` (which is omitted in this example).

(activating-schedules)=

## Activating a schedule

After initializing your launch plan, [activate the specific version of the launch plan](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html) so that the schedule runs.

```bash
flytectl update launchplan -p flyteexamples -d development {{ name_of_lp }} --version <foo> --activate
```

Verify if your launch plan was activated:

```bash
flytectl get launchplan -p flytesnacks -d development
```

## Deactivating a schedule

You can [archive/deactivate the launch plan](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_update_launchplan.html) to deschedule any scheduled job associated with it.

```bash
flytectl update launchplan -p flyteexamples -d development {{ name_of_lp }} --version <foo> --archive
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
