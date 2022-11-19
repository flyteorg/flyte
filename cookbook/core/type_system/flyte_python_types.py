"""
.. _flytekit_to_flyte_type_mapping:

Flyte and Python Types
----------------------

.. tags:: Basic

FlyteKit automatically maps Python types to Flyte types. This section provides details of the mappings, but for the most
part you can skip this section, as almost all of Python types are mapped automatically.

The following table provides a quick overview of how types are converted from the type-hints (python native) to Flyte-understood, cross-language types.

.. list-table:: Supported Python types and mapping to underlying Flyte Type
   :widths: auto
   :header-rows: 1

   * - Python Type
     - Flyte Type
     - Conversion
     - Comment
   * - int
     - Integer
     - Automatic
     - just use python 3 type hints
   * - float
     - Float
     - Automatic
     - just use python 3 type hints
   * - str
     - String
     - Automatic
     - just use python 3 type hints
   * - bool
     - Boolean
     - Automatic
     - just use python 3 type hints
   * - bytes/bytearray
     - binary
     - Not Supported
     - Let us know if this is an interesting usecase that you can currently support using your own transformers.
   * - complex
     - NA
     - Not Supported
     - Let us know if this is an interesting usecase that you can currently support using your own transformers.
   * - datetime.timedelta
     - Duration
     - Automatic
     - just use python 3 type hints
   * - datetime.datetime
     - Datetime
     - Automatic
     - just use python 3 type hints
   * - Univariate List / typing.List
     - Collection [ type ]
     - Automatic
     - Use python 3 type hints e.g ``typing.List[T], where T can be one of the other supported types in the table``
   * - file / file-like / os.PathLike / flytekit.types.file.FlyteFile
     - Blob - Single
     - Automatic
     - Use python 3 type hints. if using ``file / os.PathLike`` objects then, Flyte will default to binary protocol for the file. If using FlyteFile["protocol"], it is assumed that the file is in the specified protocol. e.g. "jpg", "png", "hdf5" etc
   * - Directory / flytekit.types.directory.FlyteDirectory
     - Blob - MultiPart
     - Automatic
     - Use python 3 type hints. If using FlyteDirectory["protocol"] it is assumed that all the files are of the specified protocol
   * - Typed dictionary with str key - typing.Dict[str, V]
     - Map[str, V]
     - Automatic
     - Use python 3 type hints e.g ``typing.Dict[str, V], where V can be one of the other supported types in the table even another Dictionary (nested)``
   * - Untyped dictionary - dict
     - JSON (struct.pb)
     - Automatic
     - Use python 3 type hints e.g ``dict``, it will be assumed that we can convert the dict to json. This may not always be possible and will cause a RuntimeError
   * - Dataclasses ``@dataclass``
     - Struct
     - Automatic
     - Use python 3 type hints. The class should be a pure value class and should be annotated with ``@dataclass and @dataclass_json``.
   * - pandas.DataFrame
     - Schema
     - Automatic
     - Use python 3 type hints. Pandas column types are not preserved
   * - pyspark.DataFrame
     - Schema
     - Automatic
     - Use python 3 type hints. Column types are not preserved. Install ``flytekitplugins-spark`` plugin using pip
   * - torch.Tensor & torch.nn.Module
     - Blob - Single
     - Automatic
     - Use PyTorch type hints.
   * - User defined types
     - Any
     - Custom Transformers
     - Use python 3 type hints. We use ``FlytePickle transformer`` by default, but users still can provide custom transformers. Refer to :ref:`advanced_custom_types`.

"""
