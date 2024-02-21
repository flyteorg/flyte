import pytest
from flyteidl.core import types_pb2

from flytekit.models import types as _types
from flytekit.models.annotation import TypeAnnotation
from tests.flytekit.common import parameterizers


def test_simple_type():
    assert _types.SimpleType.NONE == types_pb2.NONE
    assert _types.SimpleType.INTEGER == types_pb2.INTEGER
    assert _types.SimpleType.FLOAT == types_pb2.FLOAT
    assert _types.SimpleType.STRING == types_pb2.STRING
    assert _types.SimpleType.BOOLEAN == types_pb2.BOOLEAN
    assert _types.SimpleType.DATETIME == types_pb2.DATETIME
    assert _types.SimpleType.DURATION == types_pb2.DURATION
    assert _types.SimpleType.BINARY == types_pb2.BINARY
    assert _types.SimpleType.ERROR == types_pb2.ERROR


def test_schema_column():
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.INTEGER == types_pb2.SchemaType.SchemaColumn.INTEGER
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.FLOAT == types_pb2.SchemaType.SchemaColumn.FLOAT
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.STRING == types_pb2.SchemaType.SchemaColumn.STRING
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.DATETIME == types_pb2.SchemaType.SchemaColumn.DATETIME
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.DURATION == types_pb2.SchemaType.SchemaColumn.DURATION
    assert _types.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN == types_pb2.SchemaType.SchemaColumn.BOOLEAN


def test_schema_type():
    obj = _types.SchemaType(
        [
            _types.SchemaType.SchemaColumn("a", _types.SchemaType.SchemaColumn.SchemaColumnType.INTEGER),
            _types.SchemaType.SchemaColumn("b", _types.SchemaType.SchemaColumn.SchemaColumnType.FLOAT),
            _types.SchemaType.SchemaColumn("c", _types.SchemaType.SchemaColumn.SchemaColumnType.STRING),
            _types.SchemaType.SchemaColumn("d", _types.SchemaType.SchemaColumn.SchemaColumnType.DATETIME),
            _types.SchemaType.SchemaColumn("e", _types.SchemaType.SchemaColumn.SchemaColumnType.DURATION),
            _types.SchemaType.SchemaColumn("f", _types.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN),
        ]
    )

    assert obj.columns[0].name == "a"
    assert obj.columns[1].name == "b"
    assert obj.columns[2].name == "c"
    assert obj.columns[3].name == "d"
    assert obj.columns[4].name == "e"
    assert obj.columns[5].name == "f"

    assert obj.columns[0].type == _types.SchemaType.SchemaColumn.SchemaColumnType.INTEGER
    assert obj.columns[1].type == _types.SchemaType.SchemaColumn.SchemaColumnType.FLOAT
    assert obj.columns[2].type == _types.SchemaType.SchemaColumn.SchemaColumnType.STRING
    assert obj.columns[3].type == _types.SchemaType.SchemaColumn.SchemaColumnType.DATETIME
    assert obj.columns[4].type == _types.SchemaType.SchemaColumn.SchemaColumnType.DURATION
    assert obj.columns[5].type == _types.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN

    assert obj == _types.SchemaType.from_flyte_idl(obj.to_flyte_idl())


def test_literal_types():
    obj = _types.LiteralType(simple=_types.SimpleType.INTEGER)
    assert obj.simple == _types.SimpleType.INTEGER
    assert obj.schema is None
    assert obj.collection_type is None
    assert obj.map_value_type is None
    assert obj == _types.LiteralType.from_flyte_idl(obj.to_flyte_idl())

    schema_type = _types.SchemaType(
        [
            _types.SchemaType.SchemaColumn("a", _types.SchemaType.SchemaColumn.SchemaColumnType.INTEGER),
            _types.SchemaType.SchemaColumn("b", _types.SchemaType.SchemaColumn.SchemaColumnType.FLOAT),
            _types.SchemaType.SchemaColumn("c", _types.SchemaType.SchemaColumn.SchemaColumnType.STRING),
            _types.SchemaType.SchemaColumn("d", _types.SchemaType.SchemaColumn.SchemaColumnType.DATETIME),
            _types.SchemaType.SchemaColumn("e", _types.SchemaType.SchemaColumn.SchemaColumnType.DURATION),
            _types.SchemaType.SchemaColumn("f", _types.SchemaType.SchemaColumn.SchemaColumnType.BOOLEAN),
        ]
    )
    obj = _types.LiteralType(schema=schema_type)
    assert obj.simple is None
    assert obj.schema == schema_type
    assert obj.collection_type is None
    assert obj.map_value_type is None
    assert obj == _types.LiteralType.from_flyte_idl(obj.to_flyte_idl())


def test_annotated_literal_types():
    obj = _types.LiteralType(simple=_types.SimpleType.INTEGER, annotation=TypeAnnotation(annotations={"foo": "bar"}))
    assert obj.simple == _types.SimpleType.INTEGER
    assert obj.schema is None
    assert obj.collection_type is None
    assert obj.map_value_type is None
    assert obj.annotation.annotations == {"foo": "bar"}
    assert obj == _types.LiteralType.from_flyte_idl(obj.to_flyte_idl())


@pytest.mark.parametrize("literal_type", parameterizers.LIST_OF_ALL_LITERAL_TYPES)
def test_literal_collections(literal_type):
    obj = _types.LiteralType(collection_type=literal_type)
    assert obj.collection_type == literal_type
    assert obj.simple is None
    assert obj.schema is None
    assert obj.map_value_type is None
    assert obj == _types.LiteralType.from_flyte_idl(obj.to_flyte_idl())


def test_output_reference():
    obj = _types.OutputReference(node_id="node1", var="var1")
    assert obj.node_id == "node1"
    assert obj.var == "var1"
    obj2 = _types.OutputReference.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_output_reference_with_attr_path():
    obj = _types.OutputReference(node_id="node1", var="var1", attr_path=["a", 0])
    assert obj.node_id == "node1"
    assert obj.var == "var1"
    assert obj.attr_path[0] == "a"
    assert obj.attr_path[1] == 0

    obj2 = _types.OutputReference.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
