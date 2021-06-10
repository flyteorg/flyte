"""
.. _raw_container:

Using Raw Containers
---------------------

This example shows how it is possible to use arbitrary containers and pass data between them using Flyte.
Flyte mounts an input data volume where all the data needed by the container is available and an output data volume
for the container to write all the data which will be stored away.

The data is written as separate files, one per input variable. The format of the file is serialized strings.
Refer to the raw protocol to understand how to leverage this
"""
from flytekit import ContainerTask, kwtypes, workflow

# %%
# A :py:class:`flytekit.ContainerTask` denotes an arbitrary container. In the following example, the name of the task
# is square. This name has to be unique in the entire project. Users can specify some
#
# - ``input_data_dir`` -> where inputs will be written to
# - ``output_data_dir`` -> where Flyte will expect the outputs to exist.
#
# inputs and outputs specify the interface for the task, thus it should be an ordered dictionary of typed input and
# output variables
square = ContainerTask(
    name="square",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(val=int),
    outputs=kwtypes(out=int),
    image="alpine",
    command=[
        "sh",
        "-c",
        "echo $(( {{.Inputs.val}} * {{.Inputs.val}} )) | tee /var/outputs/out",
    ],
)


sum = ContainerTask(
    name="sum",
    input_data_dir="/var/flyte/inputs",
    output_data_dir="/var/flyte/outputs",
    inputs=kwtypes(x=int, y=int),
    outputs=kwtypes(out=int),
    image="alpine",
    command=[
        "sh",
        "-c",
        "echo $(( {{.Inputs.x}} + {{.Inputs.y}} )) | tee /var/flyte/outputs/out",
    ],
)


# %%
# As can be seen in this example, ContainerTasks can be interacted with like normal python functions, whose inputs
# correspond to the declared input variables
@workflow
def raw_container_wf(val1: int, val2: int) -> int:
    """
    These tasks can be invoked like simple python methods. But running them locally performs no execution, unless
    the execution is mocked.
    """
    return sum(x=square(val=val1), y=square(val=val2))


# %%
# ContainerTasks cannot really be executed locally as Flytekit is incapable of executing Containers currently.
# but it is possible to mock the execution.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running raw_container_wf(val1=5, val2=5) {raw_container_wf(val1=5, val2=5)}"
    )
