(raw_container)=

# Raw containers

```{eval-rst}
.. tags:: Containerization, Advanced
```

This example demonstrates how to use arbitrary containers in 5 different languages, all orchestrated in flytekit seamlessly. Flyte mounts an input data volume where all the data needed by the container is available, and an output data volume for the container to write all the data which will be stored away.

The data is written as separate files, one per input variable. The format of the file is serialized strings.
Refer to the raw protocol to understand how to leverage this.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/customizing_dependencies/customizing_dependencies/raw_container.py
:caption: customizing_dependencies/raw_container.py
:lines: 1-5
```

## Container tasks

A {py:class}`flytekit.ContainerTask` denotes an arbitrary container. In the following example, the name of the task
is `calculate_ellipse_area_shell`. This name has to be unique in the entire project. Users can specify:

- `input_data_dir` -> where inputs will be written to.
- `output_data_dir` -> where Flyte will expect the outputs to exist.

`inputs` and `outputs` specify the interface for the task; thus it should be an ordered dictionary of typed input and
output variables.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/customizing_dependencies/customizing_dependencies/raw_container.py
:caption: customizing_dependencies/raw_container.py
:lines: 15-112
```

As can be seen in this example, `ContainerTask`s can be interacted with like normal Python functions, whose inputs
correspond to the declared input variables. All data returned by the tasks are consumed and logged by a Flyte task.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/customizing_dependencies/customizing_dependencies/raw_container.py
:caption: customizing_dependencies/raw_container.py
:pyobject: wf
```

One of the benefits of raw container tasks is that Flytekit does not need to be installed in the target container.

:::{note}
Raw containers can be run locally when flytekit version >= 1.11.0.
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

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/customizing_dependencies/
