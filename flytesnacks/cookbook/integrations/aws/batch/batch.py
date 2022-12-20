"""
AWS Batch
#########

This example shows how to use a Flyte AWS batch plugin to execute a tasks on batch service.
With AWS Batch, there is no need to install and manage batch computing software or server clusters
that you use to run your jobs, allowing you to focus on analyzing results and solving problems.
"""

from flytekit import task, workflow
from flytekitplugins.awsbatch import AWSBatchConfig

# %%
# Use this to configure SubmitJobInput for a AWS batch job. Task's marked with this will automatically execute
# natively onto AWS batch service.
# Refer to AWS SubmitJobInput for more detail: https://docs.aws.amazon.com/sdk-for-go/api/service/batch/#SubmitJobInput
#
config = AWSBatchConfig(
    parameters={"codec": "mp4"},
    platformCapabilities="EC2",
    tags={"name": "flyte-example"},
)


@task(task_config=config)
def t1(a: int) -> int:
    return a + 2


@task(task_config=config)
def t2(b: int) -> int:
    return b * 10


@workflow
def my_wf(a: int) -> int:
    b = t1(a=a)
    return t2(b=b)


if __name__ == "__main__":
    print(f"Running my_wf(a=3') {my_wf(a=3)}")
