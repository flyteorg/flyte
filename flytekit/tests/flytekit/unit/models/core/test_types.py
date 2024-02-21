from flyteidl.core import types_pb2 as _types_pb2

from flytekit.models.core import types as _types


def test_blob_dimensionality():
    assert _types.BlobType.BlobDimensionality.SINGLE == _types_pb2.BlobType.SINGLE
    assert _types.BlobType.BlobDimensionality.MULTIPART == _types_pb2.BlobType.MULTIPART


def test_blob_type():
    o = _types.BlobType(
        format="csv",
        dimensionality=_types.BlobType.BlobDimensionality.SINGLE,
    )
    assert o.format == "csv"
    assert o.dimensionality == _types.BlobType.BlobDimensionality.SINGLE

    o2 = _types.BlobType.from_flyte_idl(o.to_flyte_idl())
    assert o == o2
    assert o2.format == "csv"
    assert o2.dimensionality == _types.BlobType.BlobDimensionality.SINGLE


def test_enum_type():
    o = _types.EnumType(values=["x", "y"])
    assert o.values == ["x", "y"]
    v = o.to_flyte_idl()
    assert v
    assert v.values == ["x", "y"]

    o = _types.EnumType.from_flyte_idl(_types_pb2.EnumType(values=["a", "b"]))
    assert o.values == ["a", "b"]

    o = _types.EnumType(values=None)
    assert not o.values
    v = o.to_flyte_idl()
    assert v
    assert not v.values
