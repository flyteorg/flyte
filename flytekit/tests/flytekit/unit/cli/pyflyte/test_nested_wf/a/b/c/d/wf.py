from flytekit import task, workflow


@task
def t(m: str) -> str:
    return m


@workflow
def wf_id(m: str) -> str:
    return t(m=m)
