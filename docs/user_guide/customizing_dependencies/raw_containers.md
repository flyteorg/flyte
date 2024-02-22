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

(raw_container)=

# Raw containers

```{eval-rst}
.. tags:: Containerization, Advanced
```

This example demonstrates how to use arbitrary containers in 5 different languages, all orchestrated in flytekit seamlessly.
Flyte mounts an input data volume where all the data needed by the container is available, and an output data volume
for the container to write all the data which will be stored away.

The data is written as separate files, one per input variable. The format of the file is serialized strings.
Refer to the raw protocol to understand how to leverage this.

```{code-cell}
import logging

from flytekit import ContainerTask, kwtypes, task, workflow

logger = logging.getLogger(__file__)
```

+++ {"lines_to_next_cell": 0}

## Container tasks

A {py:class}`flytekit.ContainerTask` denotes an arbitrary container. In the following example, the name of the task
is `calculate_ellipse_area_shell`. This name has to be unique in the entire project. Users can specify:

- `input_data_dir` -> where inputs will be written to.
- `output_data_dir` -> where Flyte will expect the outputs to exist.

`inputs` and `outputs` specify the interface for the task; thus it should be an ordered dictionary of typed input and
output variables.

```{code-cell}
calculate_ellipse_area_shell = ContainerTask(
    name="ellipse-area-metadata-shell",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(a=float, b=float),
    outputs=kwtypes(area=float, metadata=str),
    image="ghcr.io/flyteorg/rawcontainers-shell:v2",
    command=[
        "./calculate-ellipse-area.sh",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ],
)

calculate_ellipse_area_python = ContainerTask(
    name="ellipse-area-metadata-python",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(a=float, b=float),
    outputs=kwtypes(area=float, metadata=str),
    image="ghcr.io/flyteorg/rawcontainers-python:v2",
    command=[
        "python",
        "calculate-ellipse-area.py",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ],
)

calculate_ellipse_area_r = ContainerTask(
    name="ellipse-area-metadata-r",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(a=float, b=float),
    outputs=kwtypes(area=float, metadata=str),
    image="ghcr.io/flyteorg/rawcontainers-r:v2",
    command=[
        "Rscript",
        "--vanilla",
        "calculate-ellipse-area.R",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ],
)

calculate_ellipse_area_haskell = ContainerTask(
    name="ellipse-area-metadata-haskell",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(a=float, b=float),
    outputs=kwtypes(area=float, metadata=str),
    image="ghcr.io/flyteorg/rawcontainers-haskell:v2",
    command=[
        "./calculate-ellipse-area",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ],
)

calculate_ellipse_area_julia = ContainerTask(
    name="ellipse-area-metadata-julia",
    input_data_dir="/var/inputs",
    output_data_dir="/var/outputs",
    inputs=kwtypes(a=float, b=float),
    outputs=kwtypes(area=float, metadata=str),
    image="ghcr.io/flyteorg/rawcontainers-julia:v2",
    command=[
        "julia",
        "calculate-ellipse-area.jl",
        "{{.inputs.a}}",
        "{{.inputs.b}}",
        "/var/outputs",
    ],
)


@task
def report_all_calculated_areas(
    area_shell: float,
    metadata_shell: str,
    area_python: float,
    metadata_python: str,
    area_r: float,
    metadata_r: str,
    area_haskell: float,
    metadata_haskell: str,
    area_julia: float,
    metadata_julia: str,
):
    logger.info(f"shell: area={area_shell}, metadata={metadata_shell}")
    logger.info(f"python: area={area_python}, metadata={metadata_python}")
    logger.info(f"r: area={area_r}, metadata={metadata_r}")
    logger.info(f"haskell: area={area_haskell}, metadata={metadata_haskell}")
    logger.info(f"julia: area={area_julia}, metadata={metadata_julia}")
```

+++ {"lines_to_next_cell": 0}

As can be seen in this example, `ContainerTask`s can be interacted with like normal Python functions, whose inputs
correspond to the declared input variables. All data returned by the tasks are consumed and logged by a Flyte task.

```{code-cell}
:lines_to_next_cell: 2

@workflow
def wf(a: float, b: float):
    # Calculate area in all languages
    area_shell, metadata_shell = calculate_ellipse_area_shell(a=a, b=b)
    area_python, metadata_python = calculate_ellipse_area_python(a=a, b=b)
    area_r, metadata_r = calculate_ellipse_area_r(a=a, b=b)
    area_haskell, metadata_haskell = calculate_ellipse_area_haskell(a=a, b=b)
    area_julia, metadata_julia = calculate_ellipse_area_julia(a=a, b=b)

    # Report on all results in a single task to simplify comparison
    report_all_calculated_areas(
        area_shell=area_shell,
        metadata_shell=metadata_shell,
        area_python=area_python,
        metadata_python=metadata_python,
        area_r=area_r,
        metadata_r=metadata_r,
        area_haskell=area_haskell,
        metadata_haskell=metadata_haskell,
        area_julia=area_julia,
        metadata_julia=metadata_julia,
    )
```

One of the benefits of raw container tasks is that Flytekit does not need to be installed in the target container.

:::{note}
Raw containers cannot be run locally at the moment.
:::

## Scripts

The contents of each script specified in the `ContainerTask` is as follows:

### calculate-ellipse-area.sh

```{literalinclude} raw-containers-supporting-files/per-language/shell/calculate-ellipse-area.sh
:language: shell
```

### calculate-ellipse-area.py

```{literalinclude} raw-containers-supporting-files/per-language/python/calculate-ellipse-area.py
:language: python
```

### calculate-ellipse-area.R

```{literalinclude} raw-containers-supporting-files/per-language/r/calculate-ellipse-area.R
:language: r
```

### calculate-ellipse-area.hs

```{literalinclude} raw-containers-supporting-files/per-language/haskell/calculate-ellipse-area.hs
:language: haskell
```

### calculate-ellipse-area.jl

```{literalinclude} raw-containers-supporting-files/per-language/julia/calculate-ellipse-area.jl
:language: julia
```
