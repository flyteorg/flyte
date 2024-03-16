import json
import typing

from flytekit import task, workflow


@task
def convert_to_string(d: typing.Dict[str, str]) -> str:
    return json.dumps(d)


@workflow
def my_wf(d: typing.Dict[str, str]) -> str:
    return convert_to_string(d=d)


if __name__ == "__main__":
    print(f"Running my_wf(d={{'a': 'rwx', 'b': '42'}})={my_wf(d={'a': 'rwx', 'b': '42'})}")
