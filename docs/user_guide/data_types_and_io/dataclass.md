(dataclass)=

# Dataclass

```{eval-rst}
.. tags:: Basic
```

When you've multiple values that you want to send across Flyte entities, you can use a `dataclass`.

Flytekit uses the [Mashumaro library](https://github.com/Fatal1ty/mashumaro)
to serialize and deserialize dataclasses.

:::{important}
If you're using Flytekit version below v1.10, you'll need to decorate with `@dataclass_json` using
`from dataclasses_json import dataclass_json` instead of inheriting from Mashumaro's `DataClassJSONMixin`.

If you're using Flytekit version >= v1.11.1, you don't need to decorate with `@dataclass_json` or
inherit from Mashumaro's `DataClassJSONMixin`.
:::

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 1-10
```

## Python types
We define a `dataclass` with `int`, `str` and `dict` as the data types.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:pyobject: Datum
```

You can send a `dataclass` between different tasks written in various languages, and input it through the Flyte console as raw JSON.

:::{note}
All variables in a data class should be **annotated with their type**. Failure to do should will result in an error.
:::

Once declared, a dataclass can be returned as an output or accepted as an input.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 28-43
```

## Flyte types
We also define a data class that accepts {std:ref}`StructuredDataset <structured_dataset>`,
{std:ref}`FlyteFile <files>` and {std:ref}`FlyteDirectory <folder>`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 47-84
```

A data class supports the usage of data associated with Python types, data classes,
flyte file, flyte directory and structured dataset.

We define a workflow that calls the tasks created above.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:pyobject: dataclass_wf
```

You can run the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 97-98
```

To trigger a task that accepts a dataclass as an input with `pyflyte run`, you can provide a JSON file as an input:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/dataclass.py \
  add --x dataclass_input.json --y dataclass_input.json
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
