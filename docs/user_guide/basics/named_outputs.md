(named_outputs)=

# Named outputs

```{eval-rst}
.. tags:: Basic
```

By default, Flyte employs a standardized convention to assign names to the outputs of tasks or workflows.
Each output is sequentially labeled as `o1`, `o2`, `o3`, ... `on`, where `o` serves as the standard prefix,
and `1`, `2`, ... `n` indicates the positional index within the returned values.

However, Flyte allows the customization of output names for tasks or workflows.
This customization becomes beneficial when you're returning multiple outputs
and you wish to assign a distinct name to each of them.

The following example illustrates the process of assigning names to outputs for both a task and a workflow.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/named_outputs.py
:caption: basics/named_outputs.py
:lines: 1-3
```

Define a `NamedTuple` and assign it as an output to a task:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/named_outputs.py
:caption: basics/named_outputs.py
:lines: 6-14
```

Likewise, assign a `NamedTuple` to the output of `intercept` task:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/named_outputs.py
:caption: basics/named_outputs.py
:lines: 18-26
```

:::{note}
While it's possible to create `NamedTuple`s directly within the code,
it's often better to declare them explicitly. This helps prevent potential linting errors in tools like mypy.

```
def slope() -> NamedTuple("slope_value", slope=float):
    pass
```
:::

You can easily unpack the `NamedTuple` outputs directly within a workflow.
Additionally, you can also have the workflow return a `NamedTuple` as an output.

:::{note}
Remember that we are extracting individual task execution outputs by dereferencing them.
This is necessary because `NamedTuple`s function as tuples and require this dereferencing:
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/named_outputs.py
:caption: basics/named_outputs.py
:lines: 32-39
```

You can run the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/named_outputs.py
:caption: basics/named_outputs.py
:lines: 43-44
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
