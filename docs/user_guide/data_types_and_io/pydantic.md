(pydantic)=

# Pydantic BaseModel

```{eval-rst}
.. tags:: Basic
```

When you've multiple values that you want to send across Flyte entities, and you want them to have, you can use a `pydantic.BaseModel`.
Note:
You can put Dataclass and FlyteTypes (FlyteFile, FlyteDirectory, FlyteSchema, and StructuredDataset) in a pydantic BaseModel.

:::{important}
Pydantic BaseModel V2 only works when you are using flytekit version >= v1.14.0.
:::

:::{important}
If you're using Flytekit version >= v1.14.0 and you want to produce protobuf struct literal for pydantic basemodels,
you can set environment variable  `FLYTE_USE_OLD_DC_FORMAT` to `true`.

For more details, you can refer the MSGPACK IDL RFC: https://github.com/flyteorg/flyte/blob/master/rfc/system/5741-binary-idl-with-message-pack.md
:::


```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary dependencies:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:lines: 1-9
```

Build your custom image with ImageSpec:
```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:lines: 11-14
```

## Python types
We define a `dataclass` with `int`, `str` and `dict` as the data types.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:pyobject: Datum
```

You can send a `dataclass` between different tasks written in various languages, and input it through the Flyte console as raw JSON.

:::{note}
All variables in a data class should be **annotated with their type**. Failure to do should will result in an error.
:::

Once declared, a dataclass can be returned as an output or accepted as an input.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:lines: 26-41
```

## Flyte types
We also define a data class that accepts {std:ref}`StructuredDataset <structured_dataset>`,
{std:ref}`FlyteFile <files>` and {std:ref}`FlyteDirectory <folder>`.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:lines: 45-86
```

A data class supports the usage of data associated with Python types, data classes,
flyte file, flyte directory and structured dataset.

We define a workflow that calls the tasks created above.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:pyobject: basemodel_wf
```

You can run the workflow locally as follows:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:lines: 99-100
```

To trigger a task that accepts a dataclass as an input with `pyflyte run`, you can provide a JSON file as an input:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/b71e01d45037cea883883f33d8d93f258b9a5023/examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py \
  basemodel_wf --x 1 --y 2
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
