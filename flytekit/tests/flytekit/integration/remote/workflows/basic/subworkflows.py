import typing

from flytekit import task, workflow


@task
def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
    return a + 2, "world"


@workflow
def my_subwf(a: int = 42) -> (str, str):
    x, y = t1(a=a)
    u, v = t1(a=x)
    return y, v


@workflow
def parent_wf(a: int) -> (int, str, str):
    x, y = t1(a=a)
    u, v = my_subwf(a=x)
    return x, u, v


if __name__ == "__main__":
    print(f"Running my_wf(a=3) {parent_wf(a=3)}")
