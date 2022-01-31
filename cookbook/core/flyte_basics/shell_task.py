"""
Run Bash Scripts Using ShellTask
--------------------------------

To run bash scripts from within Flyte, ShellTask can be used. In this example, let's define three ShellTasks to run simple bash commands.

.. note::
    The new input/output placeholder syntax of ``ShellTask`` is available starting Flytekit 0.30.0b8+.
"""
import os
from typing import Tuple

import flytekit
from flytekit import kwtypes, task, workflow
from flytekit.extras.tasks.shell import OutputLocation, ShellTask
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile


t1 = ShellTask(
    name="task_1",
    debug=True,
    script="""
    set -ex
    echo "Hey there! Let's run some bash scripts using Flyte's ShellTask."
    echo "Showcasing Flyte's Shell Task." >> {inputs.x}
    if grep "Flyte" {inputs.x}
    then
        echo "Found it!" >> {inputs.x}
    else
        echo "Not found!"
    fi
    """,
    inputs=kwtypes(x=FlyteFile),
    output_locs=[
        OutputLocation(var="i", var_type=FlyteFile, location="{inputs.x}")
    ],
)


t2 = ShellTask(
    name="task_2",
    debug=True,
    script="""
    set -ex
    cp {inputs.x} {inputs.y}
    tar -zcvf {outputs.j} {inputs.y}
    """,
    inputs=kwtypes(x=FlyteFile, y=FlyteDirectory),
    output_locs=[
        OutputLocation(var="j", var_type=FlyteFile, location="{inputs.y}.tar.gz")
    ],
)


t3 = ShellTask(
    name="task_3",
    debug=True,
    script="""
    set -ex
    tar -zxvf {inputs.z}
    cat {inputs.y}/$(basename {inputs.x}) | wc -m > {outputs.k}
    """,
    inputs=kwtypes(x=FlyteFile, y=FlyteDirectory, z=FlyteFile),
    output_locs=[OutputLocation(var="k", var_type=FlyteFile, location="output.txt")],
)


# %%
# * The ``inputs`` parameter is useful to specify the types of inputs that the task will accept
# * The ``output_locs`` parameter is helpful to specify the output locations, could be a ``FlyteFile`` or ``FlyteDirectory``
# * The ``script`` parameter is the actual bash script that will be executed (``{inputs.x}``, ``{outputs.j}``, etc. will be replaced with the actual input and output values)
# * The ``debug`` parameter is useful for debugging
#
# Next, we define a task to create FlyteFile and FlyteDirectory.
@task
def create_entities() -> Tuple[FlyteFile, FlyteDirectory]:
    working_dir = flytekit.current_context().working_directory
    flytefile = os.path.join(working_dir, "test.txt")
    os.open(flytefile, os.O_CREAT)
    flytedir = os.path.join(working_dir, "testdata")
    os.makedirs(flytedir, exist_ok=True)
    return flytefile, flytedir


# %%
# The data passage between tasks can be witnessed in the following workflow:
@workflow
def wf() -> FlyteFile:
    x, y = create_entities()
    t1_out = t1(x=x)
    t2_out = t2(x=t1_out, y=y)
    t3_out = t3(x=x, y=y, z=t2_out)
    return t3_out


if __name__ == "__main__":
    print(f"Running wf() {wf()}")
