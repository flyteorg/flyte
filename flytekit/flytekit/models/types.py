import json as _json
import typing
from typing import Dict

from flyteidl.core import types_pb2 as _types_pb2
from google.protobuf import json_format as _json_format
from google.protobuf import struct_pb2 as _struct

from flytekit.models import common as _common
from flytekit.models.annotation import TypeAnnotation as TypeAnnotationModel
from flytekit.models.core import types as _core_types


class SimpleType(object):
    NONE = _types_pb2.NONE
    INTEGER = _types_pb2.INTEGER
    FLOAT = _types_pb2.FLOAT
    STRING = _types_pb2.STRING
    BOOLEAN = _types_pb2.BOOLEAN
    DATETIME = _types_pb2.DATETIME
    DURATION = _types_pb2.DURATION
    BINARY = _types_pb2.BINARY
    ERROR = _types_pb2.ERROR
    STRUCT = _types_pb2.STRUCT


class SchemaType(_common.FlyteIdlEntity):
    class SchemaColumn(_common.FlyteIdlEntity):
        class SchemaColumnType(object):
            INTEGER = _types_pb2.SchemaType.SchemaColumn.INTEGER
            FLOAT = _types_pb2.SchemaType.SchemaColumn.FLOAT
            STRING = _types_pb2.SchemaType.SchemaColumn.STRING
            DATETIME = _types_pb2.SchemaType.SchemaColumn.DATETIME
            DURATION = _types_pb2.SchemaType.SchemaColumn.DURATION
            BOOLEAN = _types_pb2.SchemaType.SchemaColumn.BOOLEAN

        def __init__(self, name, type):
            """
            :param Text name: Name for the column
            :param int type: Enum type from SchemaType.SchemaColumn.SchemaColumnType representing the type of the column
            """
            self._name = name
            self._type = type

        @property
        def name(self):
            """
            Name for the column
            :rtype: Text
            """
            return self._name

        @property
        def type(self):
            """
            Enum type from SchemaType.SchemaColumn.SchemaColumnType representing the type of the column
            :rtype: int
            """
            return self._type

        def to_flyte_idl(self):
            """
            :rtype: flyteidl.core.types_pb2.SchemaType.SchemaColumn
            """
            return _types_pb2.SchemaType.SchemaColumn(name=self.name, type=self.type)

        @classmethod
        def from_flyte_idl(cls, proto):
            """
            :param flyteidl.core.types_pb2.SchemaType.SchemaColumn proto:
            :rtype: SchemaType.SchemaColumn
            """
            return cls(name=proto.name, type=proto.type)

    def __init__(self, columns):
        """
        :param list[SchemaType.SchemaColumn] columns: A list of columns defining the underlying data frame.
        """
        self._columns = columns

    @property
    def columns(self):
        """
        A list of columns defining the underlying data frame.
        :rtype: list[SchemaType.SchemaColumn]
        """
        return self._columns

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.types_pb2.SchemaType
        """
        return _types_pb2.SchemaType(columns=[c.to_flyte_idl() for c in self.columns])

    @classmethod
    def from_flyte_idl(cls, proto):
        """
        :param flyteidl.core.types_pb2.SchemaType proto:
        :rtype: SchemaType
        """
        return cls(columns=[SchemaType.SchemaColumn.from_flyte_idl(c) for c in proto.columns])


class UnionType(_common.FlyteIdlEntity):
    """
    Models _types_pb2.UnionType
    """

    def __init__(self, variants: typing.List["LiteralType"]):
        self._variants = variants

    @property
    def variants(self) -> typing.List["LiteralType"]:
        return self._variants

    def to_flyte_idl(self) -> _types_pb2.UnionType:
        return _types_pb2.UnionType(
            variants=[val.to_flyte_idl() if val else None for val in self._variants],
        )

    @classmethod
    def from_flyte_idl(cls, proto: _types_pb2.UnionType):
        return cls(variants=[LiteralType.from_flyte_idl(v) for v in proto.variants])


class TypeStructure(_common.FlyteIdlEntity):
    """
    Models _types_pb2.TypeStructure
    """

    def __init__(self, tag: str, dataclass_type: Dict[str, "LiteralType"] = None):
        self._tag = tag
        self._dataclass_type = dataclass_type

    @property
    def tag(self) -> str:
        return self._tag

    @property
    def dataclass_type(self) -> Dict[str, "LiteralType"]:
        return self._dataclass_type

    def to_flyte_idl(self) -> _types_pb2.TypeStructure:
        return _types_pb2.TypeStructure(
            tag=self._tag,
            dataclass_type={k: v.to_flyte_idl() for k, v in self._dataclass_type.items()}
            if self._dataclass_type is not None
            else None,
        )

    @classmethod
    def from_flyte_idl(cls, proto: _types_pb2.TypeStructure):
        return cls(
            tag=proto.tag,
            dataclass_type={k: LiteralType.from_flyte_idl(v) for k, v in proto.dataclass_type.items()}
            if proto.dataclass_type is not None
            else None,
        )


class StructuredDatasetType(_common.FlyteIdlEntity):
    class DatasetColumn(_common.FlyteIdlEntity):
        def __init__(self, name: str, literal_type: "LiteralType"):
            self._name = name
            self._literal_type = literal_type

        @property
        def name(self) -> str:
            """
            Name for the column
            """
            return self._name

        @property
        def literal_type(self) -> "LiteralType":
            """
            A LiteralType that defines the type of this column
            """
            return self._literal_type

        def to_flyte_idl(self) -> _types_pb2.StructuredDatasetType.DatasetColumn:
            return _types_pb2.StructuredDatasetType.DatasetColumn(
                name=self.name, literal_type=self.literal_type.to_flyte_idl()
            )

        @classmethod
        def from_flyte_idl(
            cls, proto: _types_pb2.StructuredDatasetType.DatasetColumn
        ) -> _types_pb2.StructuredDatasetType.DatasetColumn:
            return cls(name=proto.name, literal_type=LiteralType.from_flyte_idl(proto.literal_type))

    def __init__(
        self,
        columns: typing.List[DatasetColumn] = None,
        format: str = "",
        external_schema_type: str = None,
        external_schema_bytes: bytes = None,
    ):
        self._columns = columns
        self._format = format
        self._external_schema_type = external_schema_type
        self._external_schema_bytes = external_schema_bytes

    @property
    def columns(self) -> typing.List[DatasetColumn]:
        return self._columns

    @columns.setter
    def columns(self, value):
        self._columns = value

    @property
    def format(self) -> str:
        return self._format

    @format.setter
    def format(self, format: str):
        self._format = format

    @property
    def external_schema_type(self) -> str:
        return self._external_schema_type

    @property
    def external_schema_bytes(self) -> bytes:
        return self._external_schema_bytes

    def to_flyte_idl(self) -> _types_pb2.StructuredDatasetType:
        return _types_pb2.StructuredDatasetType(
            columns=[c.to_flyte_idl() for c in self.columns] if self.columns else None,
            format=self.format,
            external_schema_type=self.external_schema_type if self.external_schema_type else None,
            external_schema_bytes=self.external_schema_bytes if self.external_schema_bytes else None,
        )

    @classmethod
    def from_flyte_idl(cls, proto: _types_pb2.StructuredDatasetType) -> _types_pb2.StructuredDatasetType:
        return cls(
            columns=[StructuredDatasetType.DatasetColumn.from_flyte_idl(c) for c in proto.columns],
            format=proto.format,
            external_schema_type=proto.external_schema_type,
            external_schema_bytes=proto.external_schema_bytes,
        )


class LiteralType(_common.FlyteIdlEntity):
    def __init__(
        self,
        simple=None,
        schema=None,
        collection_type=None,
        map_value_type=None,
        blob=None,
        enum_type=None,
        union_type=None,
        structured_dataset_type=None,
        metadata=None,
        structure=None,
        annotation=None,
    ):
        """
        This is a oneof message, only one of the kwargs may be set, representing one of the Flyte types.

        :param SimpleType simple: Enum type from SimpleType
        :param SchemaType schema: Type definition for a dataframe-like object.
        :param LiteralType collection_type: For list-like objects, this is the type of each entry in the list.
        :param LiteralType map_value_type: For map objects, this is the type of the value.  The key must always be a
            string.
        :param flytekit.models.core.types.BlobType blob: For blob objects, this describes the type.
        :param flytekit.models.core.types.EnumType enum_type: For enum objects, describes an enum
        :param flytekit.models.core.types.UnionType union_type: For union objects, describes an python union type.
        :param flytekit.models.core.types.TypeStructure structure: Type matching hints
        :param flytekit.models.core.types.StructuredDatasetType structured_dataset_type: structured dataset
        :param dict[Text, T] metadata: Additional data describing the type
        :param flytekit.models.annotation.TypeAnnotation annotation: Additional data
            describing the type intended to be saturated by the client
        """
        self._simple = simple
        self._schema = schema
        self._collection_type = collection_type
        self._map_value_type = map_value_type
        self._blob = blob
        self._enum_type = enum_type
        self._union_type = union_type
        self._structured_dataset_type = structured_dataset_type
        self._metadata = metadata
        self._structure = structure
        self._structured_dataset_type = structured_dataset_type
        self._metadata = metadata
        self._annotation = annotation

    @property
    def simple(self) -> SimpleType:
        return self._simple

    @property
    def schema(self) -> SchemaType:
        return self._schema

    @property
    def collection_type(self) -> "LiteralType":
        """
        The collection value type
        """
        return self._collection_type

    @property
    def map_value_type(self) -> "LiteralType":
        """
        The Value for a dictionary. Key is always string
        """
        return self._map_value_type

    @property
    def blob(self) -> _core_types.BlobType:
        return self._blob

    @property
    def enum_type(self) -> _core_types.EnumType:
        return self._enum_type

    @property
    def union_type(self) -> UnionType:
        return self._union_type

    @property
    def structure(self) -> TypeStructure:
        return self._structure

    @property
    def structured_dataset_type(self) -> StructuredDatasetType:
        return self._structured_dataset_type

    @property
    def metadata(self):
        """
        :rtype: dict[Text, T]
        """
        return self._metadata

    @property
    def annotation(self) -> TypeAnnotationModel:
        """
        :rtype: flytekit.models.annotation.TypeAnnotation
        """
        return self._annotation

    @metadata.setter
    def metadata(self, value):
        self._metadata = value

    @annotation.setter
    def annotation(self, value):
        self.annotation = value

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.types_pb2.LiteralType
        """

        if self.metadata is not None:
            metadata = _json_format.Parse(_json.dumps(self.metadata), _struct.Struct())
        else:
            metadata = None

        t = _types_pb2.LiteralType(
            simple=self.simple if self.simple is not None else None,
            schema=self.schema.to_flyte_idl() if self.schema is not None else None,
            collection_type=self.collection_type.to_flyte_idl() if self.collection_type is not None else None,
            map_value_type=self.map_value_type.to_flyte_idl() if self.map_value_type is not None else None,
            blob=self.blob.to_flyte_idl() if self.blob is not None else None,
            enum_type=self.enum_type.to_flyte_idl() if self.enum_type else None,
            union_type=self.union_type.to_flyte_idl() if self.union_type else None,
            structured_dataset_type=self.structured_dataset_type.to_flyte_idl()
            if self.structured_dataset_type
            else None,
            metadata=metadata,
            annotation=self.annotation.to_flyte_idl() if self.annotation else None,
            structure=self.structure.to_flyte_idl() if self.structure else None,
        )
        return t

    @classmethod
    def from_flyte_idl(cls, proto):
        """
        :param flyteidl.core.types_pb2.LiteralType proto:
        :rtype: LiteralType
        """
        collection_type = None
        map_value_type = None
        if proto.HasField("collection_type"):
            collection_type = LiteralType.from_flyte_idl(proto.collection_type)
        if proto.HasField("map_value_type"):
            map_value_type = LiteralType.from_flyte_idl(proto.map_value_type)
        return cls(
            simple=proto.simple if proto.HasField("simple") else None,
            schema=SchemaType.from_flyte_idl(proto.schema) if proto.HasField("schema") else None,
            collection_type=collection_type,
            map_value_type=map_value_type,
            blob=_core_types.BlobType.from_flyte_idl(proto.blob) if proto.HasField("blob") else None,
            enum_type=_core_types.EnumType.from_flyte_idl(proto.enum_type) if proto.HasField("enum_type") else None,
            union_type=UnionType.from_flyte_idl(proto.union_type) if proto.HasField("union_type") else None,
            structured_dataset_type=StructuredDatasetType.from_flyte_idl(proto.structured_dataset_type)
            if proto.HasField("structured_dataset_type")
            else None,
            metadata=_json_format.MessageToDict(proto.metadata) or None,
            structure=TypeStructure.from_flyte_idl(proto.structure) if proto.HasField("structure") else None,
            annotation=TypeAnnotationModel.from_flyte_idl(proto.annotation) if proto.HasField("annotation") else None,
        )


class OutputReference(_common.FlyteIdlEntity):
    def __init__(self, node_id, var, attr_path: typing.List[typing.Union[str, int]] = None):
        """
        A reference to an output produced by a node. The type can be retrieved -and validated- from
            the underlying interface of the node.

        :param Text node_id: Node id must exist at the graph layer.
        :param Text var: Variable name must refer to an output variable for the node.
        :param List[Union[str, int]] attr_path: The attribute path the promise will be resolved with.
        """
        self._node_id = node_id
        self._var = var
        self._attr_path = attr_path if attr_path is not None else []

    @property
    def node_id(self):
        """
        Node id must exist at the graph layer.
        :rtype: Text
        """
        return self._node_id

    @property
    def var(self):
        """
        Variable name must refer to an output variable for the node.
        :rtype: Text
        """
        return self._var

    @property
    def attr_path(self) -> typing.List[typing.Union[str, int]]:
        """
        The attribute path the promise will be resolved with.
        :rtype: list[union[str, int]]
        """
        return self._attr_path

    @var.setter
    def var(self, var_name):
        self._var = var_name

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.types.OutputReference
        """
        return _types_pb2.OutputReference(
            node_id=self.node_id,
            var=self.var,
            attr_path=[
                _types_pb2.PromiseAttribute(
                    string_value=p if type(p) == str else None,
                    int_value=p if type(p) == int else None,
                )
                for p in self._attr_path
            ],
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.types.OutputReference pb2_object:
        :rtype: OutputReference
        """
        return cls(
            node_id=pb2_object.node_id,
            var=pb2_object.var,
            attr_path=[p.string_value or p.int_value for p in pb2_object.attr_path],
        )


class Error(_common.FlyteIdlEntity):
    def __init__(self, failed_node_id: str, message: str):
        self._message = message
        self._failed_node_id = failed_node_id

    @property
    def message(self) -> str:
        return self._message

    @property
    def failed_node_id(self) -> str:
        return self._failed_node_id

    def to_flyte_idl(self) -> _types_pb2.Error:
        return _types_pb2.Error(
            message=self._message,
            failed_node_id=self._failed_node_id,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _types_pb2.Error) -> "Error":
        """
        :param flyteidl.core.types.Error pb2_object:
        :rtype: Error
        """
        return cls(failed_node_id=pb2_object.failed_node_id, message=pb2_object.message)
