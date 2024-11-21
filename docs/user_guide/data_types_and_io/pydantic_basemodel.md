(pydantic_basemodel)=

# Pydantic BaseModel

```{eval-rst}
.. tags:: Basic
```

`flytekit` version >=1.14 supports natively the `JSON` format that Pydantic `BaseModel` produces,  enhancing the 
interoperability of Pydantic BaseModels with the Flyte type system.

:::{important}
Pydantic BaseModel V2 only works when you are using flytekit version >= v1.14.0.
:::

With the 1.14 release, `flytekit` adopted `MessagePack` as the serialization format for Pydantic `BaseModel`, 
overcoming a major limitation of serialization into a JSON string within a Protobuf `struct` datatype like the previous versions do:

to store `int` types, Protobuf's `struct` converts them to `float`, forcing users to write boilerplate code to work around this issue.

:::{important}
By default, `flytekit >= 1.14` will produce `msgpack` bytes literals when serializing, preserving the types defined in your `BaseModel` class.
If you're serializing `BaseModel` using `flytekit` version >= v1.14.0 and you want to produce Protobuf `struct` literal instead, you can set environment variable `FLYTE_USE_OLD_DC_FORMAT` to `true`.

For more details, you can refer the MESSAGEPACK IDL RFC: https://github.com/flyteorg/flyte/blob/master/rfc/system/5741-binary-idl-with-message-pack.md
:::

```{note}
You can put Dataclass and FlyteTypes (FlyteFile, FlyteDirectory, FlyteSchema, and StructuredDataset) in a pydantic BaseModel.
```

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
We define a `pydantic basemodel` with `int`, `str` and `dict` as the data types.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/pydantic_basemodel.py
:caption: data_types_and_io/pydantic_basemodel.py
:pyobject: Datum
```

You can send a `pydantic basemodel` between different tasks written in various languages, and input it through the Flyte console as raw JSON.

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
