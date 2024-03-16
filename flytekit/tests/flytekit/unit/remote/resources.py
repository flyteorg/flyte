from flytekit import task, workflow


@task
def say_hello(a: int) -> str:
    return "hello world" + str(a)


@workflow
def hello_wf(a: int) -> str:
    res = say_hello(a=a)
    return res
