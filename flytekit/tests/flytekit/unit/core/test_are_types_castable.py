from flytekit.core.type_engine import _are_types_castable
from flytekit.models.annotation import TypeAnnotation
from flytekit.models.core.types import EnumType
from flytekit.models.types import LiteralType, SimpleType, StructuredDatasetType, TypeStructure, UnionType

str_type = LiteralType(simple=SimpleType.STRING)
int_type = LiteralType(simple=SimpleType.INTEGER)
none_type = LiteralType(simple=SimpleType.NONE)
bool_type = LiteralType(simple=SimpleType.BOOLEAN)

str_or_int = LiteralType(union_type=UnionType([str_type, int_type]))
int_or_str = LiteralType(union_type=UnionType([int_type, str_type]))
str_or_int_or_bool = LiteralType(union_type=UnionType([str_type, int_type, bool_type]))
optional_str = LiteralType(union_type=UnionType([str_type, none_type]))


def test_simple():
    assert _are_types_castable(str_type, str_type)
    assert not _are_types_castable(str_type, int_type)
    assert not _are_types_castable(int_type, str_type)


def test_metadata():
    a = LiteralType(simple=SimpleType.STRING, metadata={"test": 456})
    assert _are_types_castable(
        a,
        LiteralType(simple=SimpleType.STRING, metadata={"test": 123}),
    )
    # must not clobber metadata
    assert a.metadata == {"test": 456}


def test_annotation():
    a = LiteralType(simple=SimpleType.STRING, annotation=TypeAnnotation(annotations={"test": 456}))
    assert _are_types_castable(
        a,
        LiteralType(simple=SimpleType.STRING, annotation=TypeAnnotation(annotations={"test": 123})),
    )
    # must not clobber annotation
    assert a.annotation.annotations == {"test": 456}


def test_structure():
    a = LiteralType(simple=SimpleType.STRING, structure=TypeStructure(tag="a"))
    assert _are_types_castable(
        a,
        LiteralType(simple=SimpleType.STRING, structure=TypeStructure(tag="b")),
    )
    # must not clobber annotation
    assert a.structure.tag == "a"


def test_non_nullable():
    assert not _are_types_castable(none_type, str_type)
    assert not _are_types_castable(str_type, none_type)
    assert _are_types_castable(none_type, none_type)


def test_collection():
    assert _are_types_castable(LiteralType(collection_type=str_type), LiteralType(collection_type=str_type))
    assert not _are_types_castable(LiteralType(collection_type=str_type), LiteralType(collection_type=int_type))


def test_map():
    assert _are_types_castable(LiteralType(map_value_type=str_type), LiteralType(map_value_type=str_type))
    assert not _are_types_castable(LiteralType(map_value_type=str_type), LiteralType(map_value_type=int_type))


def test_structured_dataset():
    x = LiteralType(
        structured_dataset_type=StructuredDatasetType(
            columns=[
                StructuredDatasetType.DatasetColumn("a", str_type),
                StructuredDatasetType.DatasetColumn("b", int_type),
            ],
            format="abc",
            external_schema_type="zzz",
            external_schema_bytes=b"zzz",
        )
    )
    assert _are_types_castable(x, x)


def test_enum():
    e = LiteralType(enum_type=EnumType(["a", "b"]))
    # enum is a str
    assert _are_types_castable(e, str_type)
    # a str is not necessarily an enum
    assert not _are_types_castable(str_type, e)


def test_union():
    # str can be expanded to str | int
    assert _are_types_castable(str_type, str_or_int)
    # str can be expanded to int | str
    assert _are_types_castable(str_type, int_or_str)
    # str | int cannot be narrowed to str
    assert not _are_types_castable(str_or_int, str_type)
    # str | int == str | int
    assert _are_types_castable(str_or_int, str_or_int)
    # int | str == str | int
    assert _are_types_castable(int_or_str, str_or_int)
    # str is Optional[str]
    assert _are_types_castable(str_type, optional_str)
    # None is Optional[str]
    assert _are_types_castable(none_type, optional_str)
    # bool is not Optional[str]
    assert not _are_types_castable(bool_type, optional_str)


def test_collection_union():
    # a list of str is a list of (str | int)
    assert _are_types_castable(LiteralType(collection_type=str_type), LiteralType(collection_type=str_or_int))
    # a list of int is a list of (str | int)
    assert _are_types_castable(LiteralType(collection_type=int_type), LiteralType(collection_type=str_or_int))
    # a list of str or a list of int is a list of (str | int)
    assert _are_types_castable(
        LiteralType(
            union_type=UnionType([LiteralType(collection_type=int_type), LiteralType(collection_type=str_type)])
        ),
        LiteralType(collection_type=str_or_int),
    )
    assert _are_types_castable(
        LiteralType(
            union_type=UnionType([LiteralType(collection_type=int_type), LiteralType(collection_type=str_type)])
        ),
        LiteralType(collection_type=str_or_int_or_bool),
    )
    # a list of str or a list of bool is not a list of (str | int)
    assert not _are_types_castable(
        LiteralType(
            union_type=UnionType([LiteralType(collection_type=int_type), LiteralType(collection_type=bool_type)])
        ),
        LiteralType(collection_type=str_or_int),
    )
    # not the other way around
    assert not _are_types_castable(LiteralType(collection_type=str_or_int), LiteralType(collection_type=str_type))
    assert not _are_types_castable(LiteralType(collection_type=str_or_int), LiteralType(collection_type=int_type))
