from datetime import datetime, timedelta

import pytest

from flytekit import CronSchedule, FixedRate, LaunchPlan, task, workflow
from flytekit.remote.backfill import create_backfill_workflow


@task
def tk(t: datetime, v: int):
    print(f"Invoked at {t} with v {v}")


@workflow
def example_wf(t: datetime, v: int):
    tk(t=t, v=v)


def test_create_backfiller_error():
    no_schedule = LaunchPlan.get_or_create(
        workflow=example_wf,
        name="nos",
        fixed_inputs={"v": 10},
    )
    rate_schedule = LaunchPlan.get_or_create(
        workflow=example_wf,
        name="rate",
        fixed_inputs={"v": 10},
        schedule=FixedRate(duration=timedelta(days=1)),
    )
    start_date = datetime(2022, 12, 1, 8)
    end_date = start_date + timedelta(days=10)

    with pytest.raises(ValueError):
        create_backfill_workflow(start_date, end_date, no_schedule)

    with pytest.raises(ValueError):
        create_backfill_workflow(end_date, start_date, no_schedule)

    with pytest.raises(ValueError):
        create_backfill_workflow(end_date, start_date, None)

    with pytest.raises(NotImplementedError):
        create_backfill_workflow(start_date, end_date, rate_schedule)


def test_create_backfiller():
    daily_lp = LaunchPlan.get_or_create(
        workflow=example_wf,
        name="daily",
        fixed_inputs={"v": 10},
        schedule=CronSchedule(schedule="0 8 * * *", kickoff_time_input_arg="t"),
    )

    start_date = datetime(2022, 12, 1, 8)
    end_date = start_date + timedelta(days=10)

    wf, start, end = create_backfill_workflow(start_date, end_date, daily_lp)
    assert isinstance(wf.nodes[0].flyte_entity, LaunchPlan)
    b0, b1 = wf.nodes[0].bindings[0], wf.nodes[0].bindings[1]
    assert b0.var == "t"
    assert b0.binding.scalar.primitive.datetime.day == 2
    assert b1.var == "v"
    assert b1.binding.scalar.primitive.integer == 10
    assert len(wf.nodes) == 9
    assert len(wf.nodes[0].upstream_nodes) == 0
    assert len(wf.nodes[1].upstream_nodes) == 1
    assert wf.nodes[1].upstream_nodes[0] == wf.nodes[0]
    assert start
    assert end


def test_create_backfiller_parallel():
    daily_lp = LaunchPlan.get_or_create(
        workflow=example_wf,
        name="daily",
        fixed_inputs={"v": 10},
        schedule=CronSchedule(schedule="0 8 * * *", kickoff_time_input_arg="t"),
    )

    start_date = datetime(2022, 12, 1, 8)
    end_date = start_date + timedelta(days=10)

    wf, start, end = create_backfill_workflow(start_date, end_date, daily_lp, parallel=True)
    assert isinstance(wf.nodes[0].flyte_entity, LaunchPlan)
    b0, b1 = wf.nodes[0].bindings[0], wf.nodes[0].bindings[1]
    assert b0.var == "t"
    assert b0.binding.scalar.primitive.datetime.day == 2
    assert b1.var == "v"
    assert b1.binding.scalar.primitive.integer == 10
    assert len(wf.nodes) == 9
    assert len(wf.nodes[0].upstream_nodes) == 0
    assert len(wf.nodes[1].upstream_nodes) == 0
    assert start
    assert end
