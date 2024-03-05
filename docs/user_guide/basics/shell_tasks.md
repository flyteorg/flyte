---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(shell_task)=

# Shell tasks

```{eval-rst}
.. tags:: Basic
```

To execute bash scripts within Flyte, you can utilize the {py:class}`~flytekit.extras.tasks.shell.ShellTask` class.
This example includes three shell tasks to execute bash commands.

First, import the necessary libraries.

```{code-cell}
from pathlib import Path
from typing import Tuple

import flytekit
from flytekit import kwtypes, task, workflow
from flytekit.extras.tasks.shell import OutputLocation, ShellTask
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile
```

+++ {"lines_to_next_cell": 0}

With the required imports in place, you can proceed to define a shell task.
To create a shell task, provide a name for it, specify the bash script to be executed,
and define inputs and outputs if needed.

```{code-cell}
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
    output_locs=[OutputLocation(var="i", var_type=FlyteFile, location="{inputs.x}")],
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
    output_locs=[OutputLocation(var="j", var_type=FlyteFile, location="{inputs.y}.tar.gz")],
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
```

+++ {"lines_to_next_cell": 0}

Here's a breakdown of the parameters of the `ShellTask`:

- The `inputs` parameter allows you to specify the types of inputs that the task will accept
- The `output_locs` parameter is used to define the output locations, which can be `FlyteFile` or `FlyteDirectory`
- The `script` parameter contains the actual bash script that will be executed
  (`{inputs.x}`, `{outputs.j}`, etc. will be replaced with the actual input and output values).
- The `debug` parameter is helpful for debugging purposes

We define a task to instantiate `FlyteFile` and `FlyteDirectory`.
A `.gitkeep` file is created in the FlyteDirectory as a placeholder to ensure the directory exists.

```{code-cell}
@task
def create_entities() -> Tuple[FlyteFile, FlyteDirectory]:
    working_dir = Path(flytekit.current_context().working_directory)
    flytefile = working_dir / "test.txt"
    flytefile.touch()

    flytedir = working_dir / "testdata"
    flytedir.mkdir(exist_ok=True)

    flytedir_file = flytedir / ".gitkeep"
    flytedir_file.touch()
    return flytefile, flytedir
```

+++ {"lines_to_next_cell": 0}

We create a workflow to define the dependencies between the tasks.

```{code-cell}
@workflow
def shell_task_wf() -> FlyteFile:
    x, y = create_entities()
    t1_out = t1(x=x)
    t2_out = t2(x=t1_out, y=y)
    t3_out = t3(x=x, y=y, z=t2_out)
    return t3_out
```

+++ {"lines_to_next_cell": 0}

You can run the workflow locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Running shell_task_wf() {shell_task_wf()}")
```
