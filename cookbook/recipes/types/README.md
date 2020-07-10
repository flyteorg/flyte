# Types in Flyte through Flytekit (python)
Flyte's comes with its own type system, which closely maps most programming languages. These types are what power Flyte's magic of
 - Data lineage
 - Memoization
 - Auto parallelization
 - Simplifying access to data
 - Auto generated CLI and Launch UI

## Mapping of Types between Python and Flyte
As of flytekit 0.10.0 (notice why I added 0.10.0, we are working hard to auto-magically making it work for you - but until then), Users need to annotated their tasks and workflows and map their inputs and outputs to Flyte's own types. The following tables help you create a mental model of how to map the types. Once we understand the mapping, later sections will point to examples of how to use these types.

```python
from flytekit.sdk.types import Types
```

### Primitive types
| Python          | Flyte         |
|-----------------|---------------|
| int             | Types.Integer |
| float           | Types.Float   |
| str             | Types.String  |
| bool            | Types.Boolean |
| bytes/bytearray | Types.Binary  |
| complex         | Not supported |

### Time types
| Python             | Flyte           |
|--------------------|-----------------|
| datetime.timedelta | Types.Timedelta |
| datetime.datetime  | Types.Datetime  |

### Container types
| Python            | Flyte          |
|-------------------|----------------|
| List              | Types.List     |

Flyte type engine supports Dictionary type, but today it is not exposed to users. This is because the key can only be a string.

### IO Types
| Python                              | Flyte                 |
|-------------------------------------|-----------------------|
| file/file-like                      | Types.Blob, Types.Csv |
| Directory like                      | Types.MultipartBlob   |
| Pandas Dataframe (stored as a file) | Types.Schema*         |

**Types.Schema** Can represent any structured columnar or row structure. For example it could be used for Spark dataframes, vaex dataframes, raw parquet files, RecordIO structure etc

### Other types
| Python          | Flyte         |
|-----------------|---------------|
| JSON, json-like | Types.Generic |
| Proto           | Types.Proto*  |

**Types.Proto** is essentially passed through a binary byte array structure.

## Use primitives & time types
*Coming soon*

## Use container types
*Coming soon*

## Use Blob

## Use CSV

## Use MultipartBlob

## Use Schema and pandas data frames

## Use / pass - Json / Json like / Custom Python structs

One of the types that Flyte supports is a `struct type <https://github.com/lyft/flyteidl/blob/f8181796dc5cafe019b1493af1b64384ae1358f5/protos/flyteidl/core/types.proto#L20>`__.  The addition of this type was debated internally for some time as it effectively removes type-checking from Flyte, so use this with caution. It is a `scalar literal <https://github.com/lyft/flyteidl/blob/f8181796dc5cafe019b1493af1b64384ae1358f5/protos/flyteidl/core/literals.proto#L63>`__, even though in most cases it won't represent a scalar.

Flytekit implements this `here <https://github.com/lyft/flytekit/blob/1926b1285591ae941d7fc9bd4c2e4391c5c1b21b/flytekit/common/types/primitives.py#L501>`__.  Please see the included task/workflow in this directory for an example of its usage.

The UI currently does not support passing structs as inputs to workflows, so if you need to rely on this, you'll have to use **flyte-cli** for now. ::

```bash
    flyte-cli -p flytesnacks -d development execute-launch-plan -u lp:flytesnacks:development:workflows.recipe_3.tasks.GenericDemoWorkflow:477b61e4d9be818bbe6514500760053f4bc890db -r demo -- a='{"a": "hello", "b": "how are you", "c": ["array"], "d": {"nested": "value"}}'
```

## Pass your own proto objects