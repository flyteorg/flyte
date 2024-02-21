import typing

from flytekit import task, workflow


@task
def t1(a: int) -> typing.Iterator[int]:
    for i in range(a):
        yield i


@task
def t2(ls: typing.Iterator[int]) -> typing.List[int]:
    return [x for x in ls]


@workflow
def wf(a: int) -> typing.List[int]:
    return t1(a=a)


def test_iterator():
    assert wf(a=4) == [0, 1, 2, 3]
