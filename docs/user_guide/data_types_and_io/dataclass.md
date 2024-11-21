(dataclass)=

# Dataclass

```{eval-rst}
.. tags:: Basic
```

When you've multiple values that you want to send across Flyte entities, you can use a `dataclass`.

Flytekit uses the [Mashumaro library](https://github.com/Fatal1ty/mashumaro)
to serialize and deserialize dataclasses.

With the 1.14 release, `flytekit` adopted `MessagePack` as the 
serialization format for dataclasses, overcoming a major limitation of serialization into a JSON string within a Protobuf `struct` datatype, like the previous versions do:

to store `int` types, Protobuf's `struct` converts them to `float`, forcing users to write boilerplate code to work around this issue.


:::{important}
If you're serializing dataclasses using `flytekit` version >= v1.14.0, and you want to produce Protobuf `struct 
literal` instead, you can set environment variable `FLYTE_USE_OLD_DC_FORMAT` to `true`.
:::


:::{important}
If you're using Flytekit version < v1.11.1, you will need to add `from dataclasses_json import dataclass_json` to your imports and decorate your dataclass with `@dataclass_json`.
:::

:::{important}
Flytekit version < v1.14.0 will produce protobuf struct literal for dataclasses.

Flytekit version >= v1.14.0 will produce msgpack bytes literal for dataclasses.

If you're using Flytekit version >= v1.14.0 and you want to produce protobuf struct literal for dataclasses, you can 
set environment variable  `FLYTE_USE_OLD_DC_FORMAT` to `true`.

For more details, you can refer the MSGPACK IDL RFC: https://github.com/flyteorg/flyte/blob/master/rfc/system/5741-binary-idl-with-message-pack.md
:::

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary dependencies:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 1-9
```

Build your custom image with ImageSpec:
```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 16-19
```

## Python types
We define a `dataclass` with `int`, `str` and `dict` as the data types.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:pyobject: Datum
```

You can send a `dataclass` between different tasks written in various languages, and input it through the Flyte console as raw JSON.

:::{note}
All variables in a data class should be **annotated with their type**. Failure to do should will result in an error.
:::

Once declared, a dataclass can be returned as an output or accepted as an input.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 32-47
```

## Flyte types
We also define a data class that accepts {std:ref}`StructuredDataset <structured_dataset>`,
{std:ref}`FlyteFile <files>` and {std:ref}`FlyteDirectory <folder>`.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 51-88
```

A data class supports the usage of data associated with Python types, data classes,
flyte file, flyte directory and structured dataset.

We define a workflow that calls the tasks created above.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:pyobject: dataclass_wf
```

You can run the workflow locally as follows:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/dataclass.py
:caption: data_types_and_io/dataclass.py
:lines: 101-102
```

To trigger a task that accepts a dataclass as an input with `pyflyte run`, you can provide a JSON file as an input:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/cfb5ea3b0d0502ef7df1f2e14f4a0d9b78250b6a/examples/data_types_and_io/data_types_and_io/dataclass.py \
  add --x dataclass_input.json --y dataclass_input.json
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
