from datetime import datetime, timedelta, timezone

import pytest

from flytekit.models import literals
from flytekit.models import types as _types
from tests.flytekit.common import parameterizers


def test_retry_strategy():
    obj = literals.RetryStrategy(3)
    assert obj.retries == 3
    assert literals.RetryStrategy.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj != literals.RetryStrategy(4)

    with pytest.raises(Exception):
        obj = literals.RetryStrategy(-1)
        obj.to_flyte_idl()


def test_void():
    obj = literals.Void()
    assert literals.Void.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert literals.Void() == obj


def test_integer_primitive():
    obj = literals.Primitive(integer=1)
    assert obj.integer == 1
    assert obj.boolean is None
    assert obj.datetime is None
    assert obj.duration is None
    assert obj.float_value is None
    assert obj.string_value is None
    assert obj.value == 1
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=datetime.now())
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=1.0)
    assert obj != literals.Primitive(string_value="abc")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer == 1
    assert obj2.boolean is None
    assert obj2.datetime is None
    assert obj2.duration is None
    assert obj2.float_value is None
    assert obj2.string_value is None
    assert obj2.value == 1
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=datetime.now())
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=1.0)
    assert obj2 != literals.Primitive(string_value="abc")

    obj3 = literals.Primitive(integer=0)
    assert obj3.value == 0

    with pytest.raises(Exception):
        literals.Primitive(integer=1.0).to_flyte_idl()


def test_boolean_primitive():
    obj = literals.Primitive(boolean=True)
    assert obj.integer is None
    assert obj.boolean is True
    assert obj.datetime is None
    assert obj.duration is None
    assert obj.float_value is None
    assert obj.string_value is None
    assert obj.value is True
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=datetime.now())
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=1.0)
    assert obj != literals.Primitive(string_value="abc")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer is None
    assert obj2.boolean is True
    assert obj2.datetime is None
    assert obj2.duration is None
    assert obj2.float_value is None
    assert obj2.string_value is None
    assert obj2.value is True
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=datetime.now())
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=1.0)
    assert obj2 != literals.Primitive(string_value="abc")

    obj3 = literals.Primitive(boolean=False)
    assert obj3.value is False

    with pytest.raises(Exception):
        literals.Primitive(boolean=datetime.now()).to_flyte_idl()


def test_datetime_primitive():
    dt = datetime.utcnow().replace(tzinfo=timezone.utc)
    obj = literals.Primitive(datetime=dt)
    assert obj.integer is None
    assert obj.boolean is None
    assert obj.datetime == dt
    assert obj.duration is None
    assert obj.float_value is None
    assert obj.string_value is None
    assert obj.value == dt
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=dt + timedelta(seconds=1))
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=1.0)
    assert obj != literals.Primitive(string_value="abc")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer is None
    assert obj2.boolean is None
    assert obj2.datetime == dt
    assert obj2.duration is None
    assert obj2.float_value is None
    assert obj2.string_value is None
    assert obj2.value == dt
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=dt + timedelta(seconds=1))
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=1.0)
    assert obj2 != literals.Primitive(string_value="abc")

    with pytest.raises(Exception):
        literals.Primitive(datetime=1.0).to_flyte_idl()


def test_duration_primitive():
    duration = timedelta(seconds=1)
    obj = literals.Primitive(duration=duration)
    assert obj.integer is None
    assert obj.boolean is None
    assert obj.datetime is None
    assert obj.duration == duration
    assert obj.float_value is None
    assert obj.string_value is None
    assert obj.value == duration
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=datetime.now())
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=1.0)
    assert obj != literals.Primitive(string_value="abc")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer is None
    assert obj2.boolean is None
    assert obj2.datetime is None
    assert obj2.duration == duration
    assert obj2.float_value is None
    assert obj2.string_value is None
    assert obj2.value == duration
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=datetime.now())
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=1.0)
    assert obj2 != literals.Primitive(string_value="abc")

    with pytest.raises(Exception):
        literals.Primitive(duration=1.0).to_flyte_idl()


def test_float_primitive():
    obj = literals.Primitive(float_value=1.0)
    assert obj.integer is None
    assert obj.boolean is None
    assert obj.datetime is None
    assert obj.duration is None
    assert obj.float_value == 1.0
    assert obj.string_value is None
    assert obj.value == 1.0
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=datetime.now())
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=0.0)
    assert obj != literals.Primitive(string_value="abc")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer is None
    assert obj2.boolean is None
    assert obj2.datetime is None
    assert obj2.duration is None
    assert obj2.float_value == 1.0
    assert obj2.string_value is None
    assert obj2.value == 1.0
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=datetime.now())
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=0.0)
    assert obj2 != literals.Primitive(string_value="abc")

    obj3 = literals.Primitive(float_value=0.0)
    assert obj3.value == 0.0

    with pytest.raises(Exception):
        literals.Primitive(float_value="abc").to_flyte_idl()


def test_string_primitive():
    obj = literals.Primitive(string_value="abc")
    assert obj.integer is None
    assert obj.boolean is None
    assert obj.datetime is None
    assert obj.duration is None
    assert obj.float_value is None
    assert obj.string_value == "abc"
    assert obj.value == "abc"
    assert obj != literals.Primitive(integer=0)
    assert obj != literals.Primitive(boolean=False)
    assert obj != literals.Primitive(datetime=datetime.now())
    assert obj != literals.Primitive(duration=timedelta(minutes=1))
    assert obj != literals.Primitive(float_value=0.0)
    assert obj != literals.Primitive(string_value="cba")

    obj2 = literals.Primitive.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.integer is None
    assert obj2.boolean is None
    assert obj2.datetime is None
    assert obj2.duration is None
    assert obj2.float_value is None
    assert obj2.string_value == "abc"
    assert obj2.value == "abc"
    assert obj2 != literals.Primitive(integer=0)
    assert obj2 != literals.Primitive(boolean=False)
    assert obj2 != literals.Primitive(datetime=datetime.now())
    assert obj2 != literals.Primitive(duration=timedelta(minutes=1))
    assert obj2 != literals.Primitive(float_value=0.0)
    assert obj2 != literals.Primitive(string_value="bca")

    obj3 = literals.Primitive(string_value="")
    assert obj3.value == ""

    with pytest.raises(Exception):
        literals.Primitive(string_value=1.0).to_flyte_idl()


def test_scalar_primitive():
    obj = literals.Scalar(primitive=literals.Primitive(float_value=5.6))
    assert obj.value.value == 5.6
    assert obj.error is None
    assert obj.blob is None
    assert obj.binary is None
    assert obj.schema is None
    assert obj.none_type is None

    x = obj.to_flyte_idl()
    assert x.primitive.float_value == 5.6

    obj2 = literals.Scalar.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.error is None
    assert obj2.blob is None
    assert obj2.binary is None
    assert obj2.schema is None
    assert obj2.none_type is None


def test_scalar_void():
    obj = literals.Scalar(none_type=literals.Void())
    assert obj.primitive is None
    assert obj.error is None
    assert obj.blob is None
    assert obj.binary is None
    assert obj.schema is None
    assert obj.none_type is not None

    obj2 = literals.Scalar.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.primitive is None
    assert obj2.error is None
    assert obj2.blob is None
    assert obj2.binary is None
    assert obj2.schema is None
    assert obj2.none_type is not None


def test_scalar_error():
    # TODO: implement error obj and write test
    pass


def test_scalar_binary():
    obj = literals.Scalar(binary=literals.Binary(b"value", "taggy"))
    assert obj.primitive is None
    assert obj.error is None
    assert obj.blob is None
    assert obj.binary is not None
    assert obj.schema is None
    assert obj.none_type is None
    assert obj.binary.tag == "taggy"
    assert obj.binary.value == b"value"

    obj2 = literals.Scalar.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.primitive is None
    assert obj2.error is None
    assert obj2.blob is None
    assert obj2.binary is not None
    assert obj2.schema is None
    assert obj2.none_type is None
    assert obj2.binary.tag == "taggy"
    assert obj2.binary.value == b"value"


def test_scalar_schema():
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

    schema = literals.Schema(uri="asdf", type=schema_type)
    obj = literals.Scalar(schema=schema)
    assert obj.primitive is None
    assert obj.error is None
    assert obj.blob is None
    assert obj.binary is None
    assert obj.schema is not None
    assert obj.none_type is None
    assert obj.value.type.columns[0].name == "a"
    assert len(obj.value.type.columns) == 6

    obj2 = literals.Scalar.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.primitive is None
    assert obj2.error is None
    assert obj2.blob is None
    assert obj2.binary is None
    assert obj2.schema is not None
    assert obj2.none_type is None
    assert obj2.value.type.columns[0].name == "a"
    assert len(obj2.value.type.columns) == 6


def test_structured_dataset():
    my_cols = [
        _types.StructuredDatasetType.DatasetColumn("a", _types.LiteralType(simple=_types.SimpleType.INTEGER)),
        _types.StructuredDatasetType.DatasetColumn("b", _types.LiteralType(simple=_types.SimpleType.STRING)),
        _types.StructuredDatasetType.DatasetColumn(
            "c", _types.LiteralType(collection_type=_types.LiteralType(simple=_types.SimpleType.INTEGER))
        ),
        _types.StructuredDatasetType.DatasetColumn(
            "d", _types.LiteralType(map_value_type=_types.LiteralType(simple=_types.SimpleType.INTEGER))
        ),
    ]
    ds = literals.StructuredDataset(
        uri="s3://bucket",
        metadata=literals.StructuredDatasetMetadata(
            structured_dataset_type=_types.StructuredDatasetType(columns=my_cols, format="parquet")
        ),
    )
    obj = literals.Scalar(structured_dataset=ds)
    assert obj.error is None
    assert obj.blob is None
    assert obj.binary is None
    assert obj.schema is None
    assert obj.none_type is None
    assert obj.structured_dataset is not None
    assert obj.value.uri == "s3://bucket"
    assert len(obj.value.metadata.structured_dataset_type.columns) == 4
    obj2 = literals.Scalar.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.blob is None
    assert obj2.binary is None
    assert obj2.schema is None
    assert obj2.none_type is None
    assert obj2.structured_dataset is not None
    assert obj2.value.uri == "s3://bucket"
    assert len(obj2.value.metadata.structured_dataset_type.columns) == 4


def test_binding_data_scalar():
    obj = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=5)))
    assert obj.value.value.value == 5
    assert obj.promise is None
    assert obj.collection is None
    assert obj.map is None

    obj2 = literals.BindingData.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.value.value.value == 5
    assert obj2.promise is None
    assert obj2.collection is None
    assert obj2.map is None


def test_binding_data_map():
    b1 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=5)))
    b2 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=57)))
    b3 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=2)))
    binding_map_sub = literals.BindingDataMap(bindings={"first": b1, "second": b2})
    binding_map = literals.BindingDataMap(
        bindings={"three": b3, "sample_map": literals.BindingData(map=binding_map_sub)}
    )
    obj = literals.BindingData(map=binding_map)
    assert obj.scalar is None
    assert obj.promise is None
    assert obj.collection is None
    assert obj.value.bindings["three"].value.value.value == 2
    assert obj.value.bindings["sample_map"].value.bindings["second"].value.value.value == 57

    obj2 = literals.BindingData.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.scalar is None
    assert obj2.promise is None
    assert obj2.collection is None
    assert obj2.value.bindings["three"].value.value.value == 2
    assert obj2.value.bindings["sample_map"].value.bindings["first"].value.value.value == 5


def test_binding_data_promise():
    obj = literals.BindingData(promise=_types.OutputReference("some_node", "myvar"))
    assert obj.scalar is None
    assert obj.promise is not None
    assert obj.collection is None
    assert obj.map is None
    assert obj.value.node_id == "some_node"
    assert obj.value.var == "myvar"

    obj2 = literals.BindingData.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.scalar is None
    assert obj2.promise is not None
    assert obj2.collection is None
    assert obj2.map is None


def test_binding_data_collection():
    b1 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=5)))
    b2 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=57)))

    coll = literals.BindingDataCollection(bindings=[b1, b2])
    obj = literals.BindingData(collection=coll)
    assert obj.scalar is None
    assert obj.promise is None
    assert obj.collection is not None
    assert obj.map is None
    assert obj.value.bindings[0].value.value.value == 5
    assert obj.value.bindings[1].value.value.value == 57

    obj2 = literals.BindingData.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.scalar is None
    assert obj2.promise is None
    assert obj2.collection is not None
    assert obj2.map is None
    assert obj2.value.bindings[0].value.value.value == 5
    assert obj2.value.bindings[1].value.value.value == 57


def test_binding_data_collection_nested():
    b1 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=5)))
    b2 = literals.BindingData(scalar=literals.Scalar(primitive=literals.Primitive(integer=57)))

    bc_inner = literals.BindingDataCollection(bindings=[b1, b2])

    bd = literals.BindingData(collection=bc_inner)
    bc_outer = literals.BindingDataCollection(bindings=[bd])
    obj = literals.BindingData(collection=bc_outer)
    assert obj.scalar is None
    assert obj.promise is None
    assert obj.collection is not None
    assert obj.map is None

    obj2 = literals.BindingData.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.scalar is None
    assert obj2.promise is None
    assert obj2.collection is not None
    assert obj2.map is None
    assert obj2.value.bindings[0].value.bindings[0].value.value.value == 5
    assert obj2.value.bindings[0].value.bindings[1].value.value.value == 57


@pytest.mark.parametrize("scalar_value_pair", parameterizers.LIST_OF_SCALARS_AND_PYTHON_VALUES)
def test_scalar_literals(scalar_value_pair):
    scalar, _ = scalar_value_pair
    obj = literals.Literal(scalar=scalar, metadata={"hello": "world"})
    assert obj.value == scalar
    assert obj.scalar == scalar
    assert obj.collection is None
    assert obj.map is None
    assert obj.metadata["hello"] == "world"

    obj2 = literals.Literal.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.value == scalar
    assert obj2.scalar == scalar
    assert obj2.collection is None
    assert obj2.map is None
    assert obj2.metadata["hello"] == "world"

    obj = literals.Literal(scalar=scalar)
    obj2 = literals.Literal.from_flyte_idl(obj.to_flyte_idl())
    assert obj2.metadata is None


@pytest.mark.parametrize("literal_value_pair", parameterizers.LIST_OF_SCALAR_LITERALS_AND_PYTHON_VALUE)
def test_literal_collection(literal_value_pair):
    lit, _ = literal_value_pair
    obj = literals.LiteralCollection([lit, lit, lit])
    assert all(ll == lit for ll in obj.literals)
    assert len(obj.literals) == 3

    obj2 = literals.LiteralCollection.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert all(ll == lit for ll in obj.literals)
    assert len(obj.literals) == 3
