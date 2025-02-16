import typing
from datetime import datetime
from random import seed

import flytekit.configuration
from flytekit import dynamic, task, workflow
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.core.condition import conditional
from flytekit.core.context_manager import ExecutionState

# seed random number generator
seed(datetime.now().microsecond)


def test_dynamic_conditional():
    @task
    def split(in1: typing.List[int]) -> (typing.List[int], typing.List[int], int):
        return in1[0 : int(len(in1) / 2)], in1[int(len(in1) / 2) + 1 :], len(in1) / 2

    # One sample implementation for merging. In a more real world example, this might merge file streams and only load
    # chunks into the memory.
    @task
    def merge(x: typing.List[int], y: typing.List[int]) -> typing.List[int]:
        n1 = len(x)
        n2 = len(y)
        result = typing.List[int]()
        i = 0
        j = 0

        # Traverse both array
        while i < n1 and j < n2:
            # Check if current element of first array is smaller than current element of second array. If yes,
            # store first array element and increment first array index. Otherwise do same with second array
            if x[i] < y[j]:
                result.append(x[i])
                i = i + 1
            else:
                result.append(y[j])
                j = j + 1

        # Store remaining elements of first array
        while i < n1:
            result.append(x[i])
            i = i + 1

        # Store remaining elements of second array
        while j < n2:
            result.append(y[j])
            j = j + 1

        return result

    # This runs the sorting completely locally. It's faster and more efficient to do so if the entire list fits in memory.
    @task
    def merge_sort_locally(in1: typing.List[int]) -> typing.List[int]:
        return sorted(in1)

    @task
    def also_merge_sort_locally(in1: typing.List[int]) -> typing.List[int]:
        return sorted(in1)

    @dynamic
    def merge_sort_remotely(in1: typing.List[int]) -> typing.List[int]:
        x, y, new_count = split(in1=in1)
        sorted_x = merge_sort(in1=x, count=new_count)
        sorted_y = merge_sort(in1=y, count=new_count)
        return merge(x=sorted_x, y=sorted_y)

    @workflow
    def merge_sort(in1: typing.List[int], count: int) -> typing.List[int]:
        return (
            conditional("terminal_case")
            .if_(count < 500)
            .then(merge_sort_locally(in1=in1))
            .elif_(count < 1000)
            .then(also_merge_sort_locally(in1=in1))
            .else_()
            .then(merge_sort_remotely(in1=in1))
        )

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            flytekit.configuration.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        with context_manager.FlyteContextManager.with_context(
            ctx.with_execution_state(ctx.execution_state.with_params(mode=ExecutionState.Mode.TASK_EXECUTION))
        ) as ctx:
            dynamic_job_spec = merge_sort_remotely.compile_into_workflow(
                ctx, merge_sort_remotely._task_function, in1=[2, 3, 4, 5]
            )
            assert len(dynamic_job_spec.tasks) == 5
            assert len(dynamic_job_spec.subworkflows) == 1
