# Tests that a LaunchPlan with inputs can be used in a workflow for flytectl compile
from flytekit import LaunchPlan, task, workflow

@task
def my_task(num: int) -> int:
    return num + 1


@workflow
def inner_workflow(num: int) -> int:
    return my_task(num=num)


@workflow
def outer_workflow() -> int:
    return LaunchPlan.get_or_create(inner_workflow, "name_override", default_inputs={"num": 42})()
