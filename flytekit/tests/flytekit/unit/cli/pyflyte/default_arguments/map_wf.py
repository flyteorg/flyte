from typing import Dict

from flytekit import task, workflow


@task
def t1(x: Dict[str, int]) -> int:
    return sum(x.values())


@workflow
def wf(x: Dict[str, int] = {"a": 1, "b": 2, "c": 3}) -> int:
    return t1(x=x)
