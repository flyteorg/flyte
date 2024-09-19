import logging
import typing
from datetime import datetime, timedelta

from croniter import croniter

from flytekit import LaunchPlan
from flytekit.core.workflow import ImperativeWorkflow, WorkflowBase, WorkflowFailurePolicy
from flytekit.remote.entities import FlyteLaunchPlan


def create_backfill_workflow(
    start_date: datetime,
    end_date: datetime,
    for_lp: typing.Union[LaunchPlan, FlyteLaunchPlan],
    parallel: bool = False,
    per_node_timeout: timedelta = None,
    per_node_retries: int = 0,
    failure_policy: typing.Optional[WorkflowFailurePolicy] = None,
) -> typing.Tuple[WorkflowBase, datetime, datetime]:
    """
    Generates a new imperative workflow for the launchplan that can be used to backfill the given launchplan.
    This can only be used to generate  backfilling workflow only for schedulable launchplans

    the Backfill plan is generated as (start_date - exclusive, end_date inclusive)

    .. code-block:: python
       :caption: Correct usage for dates example

        lp = Launchplan.get_or_create(...)
        start_date = datetime.datetime(2023, 1, 1)
        end_date =  start_date + datetime.timedelta(days=10)
        wf = create_backfill_workflow(start_date, end_date, for_lp=lp)


    .. code-block:: python
       :caption: Incorrect date example

        wf = create_backfill_workflow(end_date, start_date, for_lp=lp) # end_date is before start_date
        # OR
        wf = create_backfill_workflow(start_date, start_date, for_lp=lp) # start and end date are same


    :param start_date: datetime generate a backfill starting at this datetime (exclusive)
    :param end_date: datetime generate a backfill ending at this datetime (inclusive)
    :param for_lp: typing.Union[LaunchPlan, FlyteLaunchPlan] the backfill is generated for this launchplan
    :param parallel: if the backfill should be run in parallel. False (default) will run each bacfill sequentially
    :param per_node_timeout: timedelta Timeout to use per node
    :param per_node_retries: int Retries to user per node
    :param failure_policy: WorkflowFailurePolicy Failure policy to use for the backfill workflow
    :return: WorkflowBase, datetime datetime -> New generated workflow, datetime for first instance of backfill, datetime for last instance of backfill
    """
    if not for_lp:
        raise ValueError("Launch plan is required!")

    if start_date >= end_date:
        raise ValueError(
            f"for a backfill start date should be earlier than end date. Received {start_date} -> {end_date}"
        )

    schedule = for_lp.entity_metadata.schedule if isinstance(for_lp, FlyteLaunchPlan) else for_lp.schedule

    if schedule is None:
        raise ValueError("Backfill can only be created for scheduled launch plans")

    if schedule.cron_schedule is not None:
        cron_schedule = schedule.cron_schedule
    else:
        raise NotImplementedError("Currently backfilling only supports cron schedules.")

    logging.info(
        f"Generating backfill from {start_date} -> {end_date}. "
        f"Parallel?[{parallel}] FailurePolicy[{str(failure_policy)}]"
    )
    wf = ImperativeWorkflow(name=f"backfill-{for_lp.name}", failure_policy=failure_policy)

    input_name = schedule.kickoff_time_input_arg
    date_iter = croniter(cron_schedule.schedule, start_time=start_date, ret_type=datetime)
    prev_node = None
    actual_start = None
    actual_end = None
    while True:
        next_start_date = date_iter.get_next()
        if not actual_start:
            actual_start = next_start_date
        if next_start_date >= end_date:
            break
        actual_end = next_start_date
        inputs = {}
        if input_name:
            inputs[input_name] = next_start_date
        next_node = wf.add_launch_plan(for_lp, **inputs)
        next_node = next_node.with_overrides(
            name=f"b-{next_start_date}", retries=per_node_retries, timeout=per_node_timeout
        )
        if not parallel:
            if prev_node:
                prev_node.runs_before(next_node)
        prev_node = next_node

    if actual_end is None:
        raise StopIteration(
            f"The time window is too small for any backfill instances, first instance after start"
            f" date is {actual_start}"
        )

    return wf, actual_start, actual_end
