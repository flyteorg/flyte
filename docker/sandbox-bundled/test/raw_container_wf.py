from flytekit import ContainerTask, kwtypes, workflow

raw_container_task = ContainerTask(
    name="raw-container-task",
    inputs=kwtypes(command=str),
    outputs=kwtypes(output=str),
    image="busybox:latest",
    command=[
        "sh",
        "-c",
        "cat /var/flyte/inputs/command | sh - > /var/flyte/outputs/output",
    ],
)


@workflow
def wf() -> str:
    x = raw_container_task(command="ls")
    return x
