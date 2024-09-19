import datetime
import json

import pytest
from google.protobuf.json_format import Parse
from google.protobuf.struct_pb2 import Struct

from flytekit import BlobType
from flytekit.interaction.string_literals import (
    literal_map_string_repr,
    literal_string_repr,
    primitive_to_string,
    scalar_to_string,
)
from flytekit.models.literals import (
    Binary,
    Blob,
    BlobMetadata,
    Literal,
    LiteralCollection,
    LiteralMap,
    Primitive,
    Scalar,
    StructuredDataset,
    Union,
    Void,
)
from flytekit.models.types import Error, LiteralType, SimpleType


def test_primitive_to_string():
    primitive = Primitive(integer=1)
    assert primitive_to_string(primitive) == 1

    primitive = Primitive(float_value=1.0)
    assert primitive_to_string(primitive) == 1.0

    primitive = Primitive(boolean=True)
    assert primitive_to_string(primitive) is True

    primitive = Primitive(string_value="hello")
    assert primitive_to_string(primitive) == "hello"

    primitive = Primitive(datetime=datetime.datetime(2021, 1, 1))
    assert primitive_to_string(primitive) == "2021-01-01T00:00:00+00:00"

    primitive = Primitive(duration=datetime.timedelta(seconds=1))
    assert primitive_to_string(primitive) == 1.0

    with pytest.raises(ValueError):
        primitive_to_string(Primitive())


def test_scalar_to_string():
    scalar = Scalar(primitive=Primitive(integer=1))
    assert scalar_to_string(scalar) == 1

    scalar = Scalar(none_type=Void())
    assert scalar_to_string(scalar) is None

    scalar = Scalar(error=Error("n1", "error"))
    assert scalar_to_string(scalar) == "error"

    scalar = Scalar(structured_dataset=StructuredDataset(uri="uri"))
    assert scalar_to_string(scalar) == "uri"

    scalar = Scalar(
        blob=Blob(
            metadata=BlobMetadata(BlobType(format="", dimensionality=BlobType.BlobDimensionality.SINGLE)), uri="uri"
        )
    )
    assert scalar_to_string(scalar) == "uri"

    scalar = Scalar(binary=Binary(bytes("hello", "utf-8"), "text/plain"))
    assert scalar_to_string(scalar) == b"aGVsbG8="

    scalar = Scalar(generic=Parse(json.dumps({"a": "b"}), Struct()))
    assert scalar_to_string(scalar) == {"a": "b"}

    scalar = Scalar(
        union=Union(Literal(scalar=Scalar(primitive=Primitive(integer=1))), LiteralType(simple=SimpleType.INTEGER))
    )
    assert scalar_to_string(scalar) == 1


def test_literal_string_repr():
    lit = Literal(scalar=Scalar(primitive=Primitive(integer=1)))
    assert literal_string_repr(lit) == 1

    lit = Literal(collection=LiteralCollection(literals=[Literal(scalar=Scalar(primitive=Primitive(integer=1)))]))
    assert literal_string_repr(lit) == [1]

    lit = Literal(map=LiteralMap(literals={"a": Literal(scalar=Scalar(primitive=Primitive(integer=1)))}))
    assert literal_string_repr(lit) == {"a": 1}


def test_literal_map_string_repr():
    lit = LiteralMap(literals={"a": Literal(scalar=Scalar(primitive=Primitive(integer=1)))})
    assert literal_map_string_repr(lit) == {"a": 1}

    lit = LiteralMap(
        literals={
            "a": Literal(scalar=Scalar(primitive=Primitive(integer=1))),
            "b": Literal(scalar=Scalar(primitive=Primitive(integer=2))),
        }
    )
    assert literal_map_string_repr(lit) == {"a": 1, "b": 2}

    assert literal_map_string_repr(
        {
            "a": Literal(scalar=Scalar(primitive=Primitive(integer=1))),
            "b": Literal(scalar=Scalar(primitive=Primitive(integer=2))),
        }
    ) == {"a": 1, "b": 2}
