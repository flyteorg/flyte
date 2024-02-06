(data_types_and_io)=

# Data Types and IO

Flyte being a data-aware orchestration platform, types play a vital role within it.
This section provides an introduction to the wide range of data types that Flyte supports.
These types serve a dual purpose by not only validating the data but also enabling seamless
transfer of data between local and cloud storage.
They enable:

- Data lineage
- Memoization
- Auto parallelization
- Simplifying access to data
- Auto generated CLI and launch UI

For a more comprehensive understanding of how Flyte manages data, refer to the
{std:ref}`Understand How Flyte Handles Data <flyte:divedeep-data-management>` guide.

(python_to_flyte_type_mapping)=

## Mapping Python to Flyte types

Flytekit automatically translates most Python types into Flyte types.
Here's a breakdown of these mappings:

```{eval-rst}
.. list-table::
    :widths: auto
    :header-rows: 1

    * - Python Type
      - Flyte Type
      - Conversion
      - Comment
    * - ``int``
      - ``Integer``
      - Automatic
      - Use Python 3 type hints.
    * - ``float``
      - ``Float``
      - Automatic
      - Use Python 3 type hints.
    * - ``str``
      - ``String``
      - Automatic
      - Use Python 3 type hints.
    * - ``bool``
      - ``Boolean``
      - Automatic
      - Use Python 3 type hints.
    * - ``bytes``/``bytearray``
      - ``Binary``
      - Not Supported
      - You have the option to employ your own custom type transformer.
    * - ``complex``
      - NA
      - Not Supported
      - You have the option to employ your own custom type transformer.
    * - ``datetime.timedelta``
      - ``Duration``
      - Automatic
      - Use Python 3 type hints.
    * - ``datetime.datetime``
      - ``Datetime``
      - Automatic
      - Use Python 3 type hints.
    * - ``datetime.date``
      - ``Datetime``
      - Automatic
      - Use Python 3 type hints.
    * - ``typing.List[T]`` / ``list[T]``
      - ``Collection [T]``
      - Automatic
      - Use ``typing.List[T]`` or ``list[T]``, where ``T`` can represent one of the other supported types listed in the table.
    * - ``typing.Iterator[T]``
      - ``Collection [T]``
      - Automatic
      - Use ``typing.Iterator[T]``, where ``T`` can represent one of the other supported types listed in the table.
    * - File / file-like / ``os.PathLike``
      - ``FlyteFile``
      - Automatic
      - If you're using ``file`` or ``os.PathLike`` objects, Flyte will default to the binary protocol for the file.
        When using ``FlyteFile["protocol"]``, it is assumed that the file is in the specified protocol, such as 'jpg', 'png', 'hdf5', etc.
    * - Directory
      - ``FlyteDirectory``
      - Automatic
      - When using ``FlyteDirectory["protocol"]``, it is assumed that all the files belong to the specified protocol.
    * - ``typing.Dict[str, V]`` / ``dict[str, V]``
      - ``Map[str, V]``
      - Automatic
      - Use ``typing.Dict[str, V]`` or ``dict[str, V]``, where ``V`` can be one of the other supported types in the table,
        including a nested dictionary.
    * - ``dict``
      - JSON (``struct.pb``)
      - Automatic
      - Use ``dict``. It's assumed that the untyped dictionary can be converted to JSON.
        However, this may not always be possible and could result in a ``RuntimeError``.
    * - ``@dataclass``
      - ``Struct``
      - Automatic
      - The class should be a pure value class that inherits from Mashumaro's DataClassJSONMixin,
        and be annotated with the ``@dataclass`` decorator.
    * - ``np.ndarray``
      - File
      - Automatic
      - Use ``np.ndarray`` as a type hint.
    * - ``pandas.DataFrame``
      - Structured Dataset
      - Automatic
      - Use ``pandas.DataFrame`` as a type hint. Pandas column types aren't preserved.
    * - ``pyspark.DataFrame``
      - Structured Dataset
      - To utilize the type, install the ``flytekitplugins-spark`` plugin.
      - Use ``pyspark.DataFrame`` as a type hint.
    * - ``pydantic.BaseModel``
      - ``Map``
      - To utilize the type, install the ``flytekitplugins-pydantic`` plugin.
      - Use ``pydantic.BaseModel`` as a type hint.
    * - ``torch.Tensor`` / ``torch.nn.Module``
      - File
      - To utilize the type, install the ``torch`` library.
      - Use ``torch.Tensor`` or ``torch.nn.Module`` as a type hint, and you can use their derived types.
    * - ``tf.keras.Model``
      - File
      - To utilize the type, install the ``tensorflow`` library.
      - Use ``tf.keras.Model`` and its derived types.
    * - ``sklearn.base.BaseEstimator``
      - File
      - To utilize the type, install the ``scikit-learn`` library.
      - Use ``sklearn.base.BaseEstimator`` and its derived types.
    * - User defined types
      - Any
      - Custom transformers
      - The ``FlytePickle`` transformer is the default option, but you can also define custom transformers.
        **For instructions on building custom type transformers, please refer to :ref:`this section <advanced_custom_types>`**.
```

```{auto-examples-toc}
file
folder
structured_dataset
dataclass
attribute_access
pytorch_type
enum_type
pickle_type
```
