from __future__ import annotations

import collections
import copy
import dataclasses
import datetime as _datetime
import enum
import inspect
import json
import json as _json
import mimetypes
import textwrap
import typing
from abc import ABC, abstractmethod
from functools import lru_cache
from typing import Dict, List, NamedTuple, Optional, Type, cast

from dataclasses_json import DataClassJsonMixin, dataclass_json
from google.protobuf import json_format as _json_format
from google.protobuf import struct_pb2 as _struct
from google.protobuf.json_format import MessageToDict as _MessageToDict
from google.protobuf.json_format import ParseDict as _ParseDict
from google.protobuf.message import Message
from google.protobuf.struct_pb2 import Struct
from marshmallow_enum import EnumField, LoadDumpOptions
from mashumaro.mixins.json import DataClassJSONMixin
from typing_extensions import Annotated, get_args, get_origin

from flytekit.core.annotation import FlyteAnnotation
from flytekit.core.context_manager import FlyteContext
from flytekit.core.hash import HashMethod
from flytekit.core.type_helpers import load_type_from_tag
from flytekit.core.utils import timeit
from flytekit.exceptions import user as user_exceptions
from flytekit.interaction.string_literals import literal_map_string_repr
from flytekit.lazy_import.lazy_module import is_imported
from flytekit.loggers import logger
from flytekit.models import interface as _interface_models
from flytekit.models import types as _type_models
from flytekit.models.annotation import TypeAnnotation as TypeAnnotationModel
from flytekit.models.core import types as _core_types
from flytekit.models.literals import (
    Blob,
    BlobMetadata,
    Literal,
    LiteralCollection,
    LiteralMap,
    Primitive,
    Scalar,
    Schema,
    StructuredDatasetMetadata,
    Union,
    Void,
)
from flytekit.models.types import LiteralType, SimpleType, StructuredDatasetType, TypeStructure, UnionType

T = typing.TypeVar("T")
DEFINITIONS = "definitions"
TITLE = "title"


class BatchSize:
    """
    This is used to annotate a FlyteDirectory when we want to download/upload the contents of the directory in batches. For example,

    @task
    def t1(directory: Annotated[FlyteDirectory, BatchSize(10)]) -> Annotated[FlyteDirectory, BatchSize(100)]:
        ...
        return FlyteDirectory(...)

    In the above example flytekit will download all files from the input `directory` in chunks of 10, i.e. first it
    downloads 10 files, loads them to memory, then writes those 10 to local disk, then it loads the next 10, so on
    and so forth. Similarly, for outputs, in this case flytekit is going to upload the resulting directory in chunks of
    100.
    """

    def __init__(self, val: int):
        self._val = val

    @property
    def val(self) -> int:
        return self._val


def get_batch_size(t: Type) -> Optional[int]:
    if is_annotated(t):
        for annotation in get_args(t)[1:]:
            if isinstance(annotation, BatchSize):
                return annotation.val
    return None


def modify_literal_uris(lit: Literal):
    """
    Modifies the literal object recursively to replace the URIs with the native paths in case they are of
    type "flyte://"
    """
    from flytekit.remote.remote_fs import FlytePathResolver

    if lit.collection:
        for l in lit.collection.literals:
            modify_literal_uris(l)
    elif lit.map:
        for k, v in lit.map.literals.items():
            modify_literal_uris(v)
    elif lit.scalar:
        if lit.scalar.blob and lit.scalar.blob.uri and lit.scalar.blob.uri.startswith(FlytePathResolver.protocol):
            lit.scalar.blob._uri = FlytePathResolver.resolve_remote_path(lit.scalar.blob.uri)
        elif lit.scalar.union:
            modify_literal_uris(lit.scalar.union.value)
        elif (
            lit.scalar.structured_dataset
            and lit.scalar.structured_dataset.uri
            and lit.scalar.structured_dataset.uri.startswith(FlytePathResolver.protocol)
        ):
            lit.scalar.structured_dataset._uri = FlytePathResolver.resolve_remote_path(
                lit.scalar.structured_dataset.uri
            )


class TypeTransformerFailedError(TypeError, AssertionError, ValueError):
    ...


class TypeTransformer(typing.Generic[T]):
    """
    Base transformer type that should be implemented for every python native type that can be handled by flytekit
    """

    def __init__(self, name: str, t: Type[T], enable_type_assertions: bool = True):
        self._t = t
        self._name = name
        self._type_assertions_enabled = enable_type_assertions

    @property
    def name(self):
        return self._name

    @property
    def python_type(self) -> Type[T]:
        """
        This returns the python type
        """
        return self._t

    @property
    def type_assertions_enabled(self) -> bool:
        """
        Indicates if the transformer wants type assertions to be enabled at the core type engine layer
        """
        return self._type_assertions_enabled

    def assert_type(self, t: Type[T], v: T):
        if not hasattr(t, "__origin__") and not isinstance(v, t):
            raise TypeTransformerFailedError(f"Expected value of type {t} but got '{v}' of type {type(v)}")

    @abstractmethod
    def get_literal_type(self, t: Type[T]) -> LiteralType:
        """
        Converts the python type to a Flyte LiteralType
        """
        raise NotImplementedError("Conversion to LiteralType should be implemented")

    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:
        """
        Converts the Flyte LiteralType to a python object type.
        """
        raise ValueError("By default, transformers do not translate from Flyte types back to Python types")

    @abstractmethod
    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        """
        Converts a given python_val to a Flyte Literal, assuming the given python_val matches the declared python_type.
        Implementers should refrain from using type(python_val) instead rely on the passed in python_type. If these
        do not match (or are not allowed) the Transformer implementer should raise an AssertionError, clearly stating
        what was the mismatch
        :param ctx: A FlyteContext, useful in accessing the filesystem and other attributes
        :param python_val: The actual value to be transformed
        :param python_type: The assumed type of the value (this matches the declared type on the function)
        :param expected: Expected Literal Type
        """
        raise NotImplementedError(f"Conversion to Literal for python type {python_type} not implemented")

    @abstractmethod
    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> Optional[T]:
        """
        Converts the given Literal to a Python Type. If the conversion cannot be done an AssertionError should be raised
        :param ctx: FlyteContext
        :param lv: The received literal Value
        :param expected_python_type: Expected native python type that should be returned
        """
        raise NotImplementedError(
            f"Conversion to python value expected type {expected_python_type} from literal not implemented"
        )

    def to_html(self, ctx: FlyteContext, python_val: T, expected_python_type: Type[T]) -> str:
        """
        Converts any python val (dataframe, int, float) to a html string, and it will be wrapped in the HTML div
        """
        return str(python_val)

    def __repr__(self):
        return f"{self._name} Transforms ({self._t}) to Flyte native"

    def __str__(self):
        return str(self.__repr__())


class SimpleTransformer(TypeTransformer[T]):
    """
    A Simple implementation of a type transformer that uses simple lambdas to transform and reduces boilerplate
    """

    def __init__(
        self,
        name: str,
        t: Type[T],
        lt: LiteralType,
        to_literal_transformer: typing.Callable[[T], Literal],
        from_literal_transformer: typing.Callable[[Literal], T],
    ):
        super().__init__(name, t)
        self._type = t
        self._lt = lt
        self._to_literal_transformer = to_literal_transformer
        self._from_literal_transformer = from_literal_transformer

    def get_literal_type(self, t: Optional[Type[T]] = None) -> LiteralType:
        return LiteralType.from_flyte_idl(self._lt.to_flyte_idl())

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        if type(python_val) != self._type:
            raise TypeTransformerFailedError(
                f"Expected value of type {self._type} but got '{python_val}' of type {type(python_val)}"
            )
        return self._to_literal_transformer(python_val)

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        expected_python_type = get_underlying_type(expected_python_type)

        if expected_python_type != self._type:
            raise TypeTransformerFailedError(
                f"Cannot convert to type {expected_python_type}, only {self._type} is supported"
            )

        try:  # todo(maximsmol): this is quite ugly and each transformer should really check their Literal
            res = self._from_literal_transformer(lv)
            if type(res) != self._type:
                raise TypeTransformerFailedError(f"Cannot convert literal {lv} to {self._type}")
            return res
        except AttributeError:
            # Assume that this is because a property on `lv` was None
            raise TypeTransformerFailedError(f"Cannot convert literal {lv} to {self._type}")

    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:
        if literal_type.simple is not None and literal_type.simple == self._lt.simple:
            return self.python_type
        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


class RestrictedTypeError(Exception):
    pass


class RestrictedTypeTransformer(TypeTransformer[T], ABC):
    """
    Types registered with the RestrictedTypeTransformer are not allowed to be converted to and from literals. In other words,
    Restricted types are not allowed to be used as inputs or outputs of tasks and workflows.
    """

    def __init__(self, name: str, t: Type[T]):
        super().__init__(name, t)

    def get_literal_type(self, t: Optional[Type[T]] = None) -> LiteralType:
        raise RestrictedTypeError(f"Transformer for type {self.python_type} is restricted currently")

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        raise RestrictedTypeError(f"Transformer for type {self.python_type} is restricted currently")

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        raise RestrictedTypeError(f"Transformer for type {self.python_type} is restricted currently")


class DataclassTransformer(TypeTransformer[object]):
    """
    The Dataclass Transformer provides a type transformer for dataclasses_json dataclasses.

    The Dataclass is converted to and from json and is transported between tasks using the proto.Structpb representation
    Also the type declaration will try to extract the JSON Schema for the object if possible and pass it with the
    definition.

    For Json Schema, we use https://github.com/fuhrysteve/marshmallow-jsonschema library.

    Example

    .. code-block:: python

        @dataclass
        class Test(DataClassJsonMixin):
           a: int
           b: str

        from marshmallow_jsonschema import JSONSchema
        t = Test(a=10,b="e")
        JSONSchema().dump(t.schema())

    Output will look like

    .. code-block:: json

        {'$schema': 'http://json-schema.org/draft-07/schema#',
         'definitions': {'TestSchema': {'properties': {'a': {'title': 'a',
             'type': 'number',
             'format': 'integer'},
            'b': {'title': 'b', 'type': 'string'}},
           'type': 'object',
           'additionalProperties': False}},
         '$ref': '#/definitions/TestSchema'}

    .. note::

        The schema support is experimental and is useful for auto-completing in the UI/CLI

    """

    def __init__(self):
        super().__init__("Object-Dataclass-Transformer", object)
        self._serializable_classes = [DataClassJSONMixin, DataClassJsonMixin]
        try:
            from mashumaro.mixins.orjson import DataClassORJSONMixin

            self._serializable_classes.append(DataClassORJSONMixin)
        except ModuleNotFoundError:
            pass

    def assert_type(self, expected_type: Type[DataClassJsonMixin], v: T):
        # Skip iterating all attributes in the dataclass if the type of v already matches the expected_type
        if type(v) == expected_type:
            return

        # @dataclass
        # class Foo(DataClassJsonMixin):
        #     a: int = 0
        #
        # @task
        # def t1(a: Foo):
        #     ...
        #
        # In above example, the type of v may not equal to the expected_type in some cases
        # For example,
        # 1. The input of t1 is another dataclass (bar), then we should raise an error
        # 2. when using flyte remote to execute the above task, the expected_type is guess_python_type (FooSchema) by default.
        # However, FooSchema is created by flytekit and it's not equal to the user-defined dataclass (Foo).
        # Therefore, we should iterate all attributes in the dataclass and check the type of value in dataclass matches the expected_type.

        expected_fields_dict = {}
        for f in dataclasses.fields(expected_type):
            expected_fields_dict[f.name] = f.type

        if isinstance(v, dict):
            original_dict = v

            # Find the Optional keys in expected_fields_dict
            optional_keys = {k for k, t in expected_fields_dict.items() if UnionTransformer.is_optional_type(t)}

            # Remove the Optional keys from the keys of original_dict
            original_key = set(original_dict.keys()) - optional_keys
            expected_key = set(expected_fields_dict.keys()) - optional_keys

            # Check if original_key is missing any keys from expected_key
            missing_keys = expected_key - original_key
            if missing_keys:
                raise TypeTransformerFailedError(
                    f"The original fields are missing the following keys from the dataclass fields: {list(missing_keys)}"
                )

            # Check if original_key has any extra keys that are not in expected_key
            extra_keys = original_key - expected_key
            if extra_keys:
                raise TypeTransformerFailedError(
                    f"The original fields have the following extra keys that are not in dataclass fields: {list(extra_keys)}"
                )

            for k, v in original_dict.items():
                if k in expected_fields_dict:
                    if isinstance(v, dict):
                        self.assert_type(expected_fields_dict[k], v)
                    else:
                        expected_type = expected_fields_dict[k]
                        original_type = type(v)
                        if UnionTransformer.is_optional_type(expected_type):
                            expected_type = UnionTransformer.get_sub_type_in_optional(expected_type)
                        if original_type != expected_type:
                            raise TypeTransformerFailedError(
                                f"Type of Val '{original_type}' is not an instance of {expected_type}"
                            )

        else:
            for f in dataclasses.fields(type(v)):  # type: ignore
                original_type = f.type
                expected_type = expected_fields_dict[f.name]

                if UnionTransformer.is_optional_type(original_type):
                    original_type = UnionTransformer.get_sub_type_in_optional(original_type)
                if UnionTransformer.is_optional_type(expected_type):
                    expected_type = UnionTransformer.get_sub_type_in_optional(expected_type)

                val = v.__getattribute__(f.name)
                if dataclasses.is_dataclass(val):
                    self.assert_type(expected_type, val)
                elif original_type != expected_type:
                    raise TypeTransformerFailedError(
                        f"Type of Val '{original_type}' is not an instance of {expected_type}"
                    )

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        """
        Extracts the Literal type definition for a Dataclass and returns a type Struct.
        If possible also extracts the JSONSchema for the dataclass.
        """
        if is_annotated(t):
            raise ValueError(
                "Flytekit does not currently have support for FlyteAnnotations applied to Dataclass."
                f"Type {t} cannot be parsed."
            )

        if not self.is_serializable_class(t):
            raise AssertionError(
                f"Dataclass {t} should be decorated with @dataclass_json or mixin with DataClassJSONMixin to be "
                f"serialized correctly"
            )
        schema = None
        try:
            if issubclass(t, DataClassJsonMixin):
                s = cast(DataClassJsonMixin, self._get_origin_type_in_annotation(t)).schema()
                for _, v in s.fields.items():
                    # marshmallow-jsonschema only supports enums loaded by name.
                    # https://github.com/fuhrysteve/marshmallow-jsonschema/blob/81eada1a0c42ff67de216923968af0a6b54e5dcb/marshmallow_jsonschema/base.py#L228
                    if isinstance(v, EnumField):
                        v.load_by = LoadDumpOptions.name
                # check if DataClass mixin
                from marshmallow_jsonschema import JSONSchema

                schema = JSONSchema().dump(s)
            else:  # DataClassJSONMixin
                from mashumaro.jsonschema import build_json_schema

                schema = build_json_schema(cast(DataClassJSONMixin, self._get_origin_type_in_annotation(t))).to_dict()
        except Exception as e:
            # https://github.com/lovasoa/marshmallow_dataclass/issues/13
            logger.warning(
                f"Failed to extract schema for object {t}, (will run schemaless) error: {e}"
                f"If you have postponed annotations turned on (PEP 563) turn it off please. Postponed"
                f"evaluation doesn't work with json dataclasses"
            )

        # Recursively construct the dataclass_type which contains the literal type of each field
        literal_type = {}

        # Get the type of each field from dataclass
        for field in t.__dataclass_fields__.values():  # type: ignore
            try:
                literal_type[field.name] = TypeEngine.to_literal_type(field.type)
            except Exception as e:
                logger.warning(
                    "Field {} of type {} cannot be converted to a literal type. Error: {}".format(
                        field.name, field.type, e
                    )
                )

        ts = TypeStructure(tag="", dataclass_type=literal_type)

        return _type_models.LiteralType(simple=_type_models.SimpleType.STRUCT, metadata=schema, structure=ts)

    def is_serializable_class(self, class_: Type[T]) -> bool:
        return any(issubclass(class_, serializable_class) for serializable_class in self._serializable_classes)

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        if isinstance(python_val, dict):
            json_str = json.dumps(python_val)
            return Literal(scalar=Scalar(generic=_json_format.Parse(json_str, _struct.Struct())))

        if not dataclasses.is_dataclass(python_val):
            raise TypeTransformerFailedError(
                f"{type(python_val)} is not of type @dataclass, only Dataclasses are supported for "
                f"user defined datatypes in Flytekit"
            )
        if not self.is_serializable_class(type(python_val)):
            raise TypeTransformerFailedError(
                f"Dataclass {python_type} should be decorated with @dataclass_json or inherit DataClassJSONMixin to be "
                f"serialized correctly"
            )
        self._serialize_flyte_type(python_val, python_type)

        json_str = python_val.to_json()  # type: ignore

        return Literal(scalar=Scalar(generic=_json_format.Parse(json_str, _struct.Struct())))  # type: ignore

    def _get_origin_type_in_annotation(self, python_type: Type[T]) -> Type[T]:
        # dataclass will try to hash python type when calling dataclass.schema(), but some types in the annotation is
        # not hashable, such as Annotated[StructuredDataset, kwtypes(...)]. Therefore, we should just extract the origin
        # type from annotated.
        if get_origin(python_type) is list:
            return typing.List[self._get_origin_type_in_annotation(get_args(python_type)[0])]  # type: ignore
        elif get_origin(python_type) is dict:
            return typing.Dict[  # type: ignore
                self._get_origin_type_in_annotation(get_args(python_type)[0]),
                self._get_origin_type_in_annotation(get_args(python_type)[1]),
            ]
        elif is_annotated(python_type):
            return get_args(python_type)[0]
        elif dataclasses.is_dataclass(python_type):
            for field in dataclasses.fields(copy.deepcopy(python_type)):
                field.type = self._get_origin_type_in_annotation(field.type)
        return python_type

    def _fix_structured_dataset_type(self, python_type: Type[T], python_val: typing.Any) -> T:
        # In python 3.7, 3.8, DataclassJson will deserialize Annotated[StructuredDataset, kwtypes(..)] to a dict,
        # so here we convert it back to the Structured Dataset.
        from flytekit.types.structured import StructuredDataset

        if python_type == StructuredDataset and type(python_val) == dict:
            return StructuredDataset(**python_val)
        elif get_origin(python_type) is list:
            return [self._fix_structured_dataset_type(get_args(python_type)[0], v) for v in python_val]  # type: ignore
        elif get_origin(python_type) is dict:
            return {  # type: ignore
                self._fix_structured_dataset_type(get_args(python_type)[0], k): self._fix_structured_dataset_type(
                    get_args(python_type)[1], v
                )
                for k, v in python_val.items()
            }
        elif dataclasses.is_dataclass(python_type):
            for field in dataclasses.fields(python_type):
                val = python_val.__getattribute__(field.name)
                python_val.__setattr__(field.name, self._fix_structured_dataset_type(field.type, val))
        return python_val

    def _serialize_flyte_type(self, python_val: T, python_type: Type[T]) -> typing.Any:
        """
        If any field inside the dataclass is flyte type, we should use flyte type transformer for that field.
        """
        from flytekit.types.directory.types import FlyteDirectory
        from flytekit.types.file import FlyteFile
        from flytekit.types.schema.types import FlyteSchema
        from flytekit.types.structured.structured_dataset import StructuredDataset

        # Handle Optional
        if get_origin(python_type) is typing.Union and type(None) in get_args(python_type):
            if python_val is None:
                return None
            return self._serialize_flyte_type(python_val, get_args(python_type)[0])

        if hasattr(python_type, "__origin__") and get_origin(python_type) is list:
            return [self._serialize_flyte_type(v, get_args(python_type)[0]) for v in cast(list, python_val)]

        if hasattr(python_type, "__origin__") and get_origin(python_type) is dict:
            return {
                k: self._serialize_flyte_type(v, get_args(python_type)[1]) for k, v in cast(dict, python_val).items()
            }

        if not dataclasses.is_dataclass(python_type):
            return python_val

        if inspect.isclass(python_type) and (
            issubclass(python_type, FlyteSchema)
            or issubclass(python_type, FlyteFile)
            or issubclass(python_type, FlyteDirectory)
            or issubclass(python_type, StructuredDataset)
        ):
            lv = TypeEngine.to_literal(FlyteContext.current_context(), python_val, python_type, None)
            # dataclasses_json package will extract the "path" from FlyteFile, FlyteDirectory, and write it to a
            # JSON which will be stored in IDL. The path here should always be a remote path, but sometimes the
            # path in FlyteFile and FlyteDirectory could be a local path. Therefore, reset the python value here,
            # so that dataclasses_json can always get a remote path.
            # In other words, the file transformer has special code that handles the fact that if remote_source is
            # set, then the real uri in the literal should be the remote source, not the path (which may be an
            # auto-generated random local path). To be sure we're writing the right path to the json, use the uri
            # as determined by the transformer.
            if issubclass(python_type, FlyteFile) or issubclass(python_type, FlyteDirectory):
                return python_type(path=lv.scalar.blob.uri)
            elif issubclass(python_type, StructuredDataset):
                sd = python_type(uri=lv.scalar.structured_dataset.uri)
                sd.file_format = lv.scalar.structured_dataset.metadata.structured_dataset_type.format
                return sd
            else:
                return python_val
        else:
            for v in dataclasses.fields(python_type):
                val = python_val.__getattribute__(v.name)
                field_type = v.type
                python_val.__setattr__(v.name, self._serialize_flyte_type(val, field_type))
            return python_val

    def _deserialize_flyte_type(self, python_val: T, expected_python_type: Type) -> Optional[T]:
        from flytekit.types.directory.types import FlyteDirectory, FlyteDirToMultipartBlobTransformer
        from flytekit.types.file.file import FlyteFile, FlyteFilePathTransformer
        from flytekit.types.schema.types import FlyteSchema, FlyteSchemaTransformer
        from flytekit.types.structured.structured_dataset import StructuredDataset, StructuredDatasetTransformerEngine

        # Handle Optional
        if get_origin(expected_python_type) is typing.Union and type(None) in get_args(expected_python_type):
            if python_val is None:
                return None
            return self._deserialize_flyte_type(python_val, get_args(expected_python_type)[0])

        if hasattr(expected_python_type, "__origin__") and expected_python_type.__origin__ is list:
            return [self._deserialize_flyte_type(v, expected_python_type.__args__[0]) for v in python_val]  # type: ignore

        if hasattr(expected_python_type, "__origin__") and expected_python_type.__origin__ is dict:
            return {k: self._deserialize_flyte_type(v, expected_python_type.__args__[1]) for k, v in python_val.items()}  # type: ignore

        if not dataclasses.is_dataclass(expected_python_type):
            return python_val

        if issubclass(expected_python_type, FlyteSchema):
            t = FlyteSchemaTransformer()
            return t.to_python_value(
                FlyteContext.current_context(),
                Literal(
                    scalar=Scalar(
                        schema=Schema(
                            cast(FlyteSchema, python_val).remote_path, t._get_schema_type(expected_python_type)
                        )
                    )
                ),
                expected_python_type,
            )
        elif issubclass(expected_python_type, FlyteFile):
            return FlyteFilePathTransformer().to_python_value(
                FlyteContext.current_context(),
                Literal(
                    scalar=Scalar(
                        blob=Blob(
                            metadata=BlobMetadata(
                                type=_core_types.BlobType(
                                    format="", dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
                                )
                            ),
                            uri=cast(FlyteFile, python_val).path,
                        )
                    )
                ),
                expected_python_type,
            )
        elif issubclass(expected_python_type, FlyteDirectory):
            return FlyteDirToMultipartBlobTransformer().to_python_value(
                FlyteContext.current_context(),
                Literal(
                    scalar=Scalar(
                        blob=Blob(
                            metadata=BlobMetadata(
                                type=_core_types.BlobType(
                                    format="", dimensionality=_core_types.BlobType.BlobDimensionality.MULTIPART
                                )
                            ),
                            uri=cast(FlyteDirectory, python_val).path,
                        )
                    )
                ),
                expected_python_type,
            )
        elif issubclass(expected_python_type, StructuredDataset):
            return StructuredDatasetTransformerEngine().to_python_value(
                FlyteContext.current_context(),
                Literal(
                    scalar=Scalar(
                        structured_dataset=StructuredDataset(
                            metadata=StructuredDatasetMetadata(
                                structured_dataset_type=StructuredDatasetType(
                                    format=cast(StructuredDataset, python_val).file_format
                                )
                            ),
                            uri=cast(StructuredDataset, python_val).uri,
                        )
                    )
                ),
                expected_python_type,
            )
        else:
            for f in dataclasses.fields(expected_python_type):
                value = python_val.__getattribute__(f.name)
                if hasattr(f.type, "__origin__") and f.type.__origin__ is list:
                    value = [self._deserialize_flyte_type(v, f.type.__args__[0]) for v in value]
                elif hasattr(f.type, "__origin__") and f.type.__origin__ is dict:
                    value = {k: self._deserialize_flyte_type(v, f.type.__args__[1]) for k, v in value.items()}
                else:
                    value = self._deserialize_flyte_type(value, f.type)
                python_val.__setattr__(f.name, value)
            return python_val

    def _fix_val_int(self, t: typing.Type, val: typing.Any) -> typing.Any:
        if val is None:
            return val

        if get_origin(t) is typing.Union and type(None) in get_args(t):
            # Handle optional type. e.g. Optional[int], Optional[dataclass]
            # Marshmallow doesn't support union type, so the type here is always an optional type.
            # https://github.com/marshmallow-code/marshmallow/issues/1191#issuecomment-480831796
            # Note: Union[None, int] is also an optional type, but Marshmallow does not support it.
            t = get_args(t)[0]

        if t == int:
            return int(val)

        if isinstance(val, list):
            # Handle nested List. e.g. [[1, 2], [3, 4]]
            return list(map(lambda x: self._fix_val_int(ListTransformer.get_sub_type(t), x), val))

        if isinstance(val, dict):
            ktype, vtype = DictTransformer.get_dict_types(t)
            # Handle nested Dict. e.g. {1: {2: 3}, 4: {5: 6}})
            return {
                self._fix_val_int(cast(type, ktype), k): self._fix_val_int(cast(type, vtype), v) for k, v in val.items()
            }

        if dataclasses.is_dataclass(t):
            return self._fix_dataclass_int(t, val)  # type: ignore

        return val

    def _fix_dataclass_int(self, dc_type: Type[DataClassJsonMixin], dc: DataClassJsonMixin) -> DataClassJsonMixin:
        """
        This is a performance penalty to convert to the right types, but this is expected by the user and hence
        needs to be done
        """
        # NOTE: Protobuf Struct does not support explicit int types, int types are upconverted to a double value
        # https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Value
        # Thus we will have to walk the given dataclass and typecast values to int, where expected.
        for f in dataclasses.fields(dc_type):
            val = dc.__getattribute__(f.name)
            dc.__setattr__(f.name, self._fix_val_int(f.type, val))
        return dc

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        if not dataclasses.is_dataclass(expected_python_type):
            raise TypeTransformerFailedError(
                f"{expected_python_type} is not of type @dataclass, only Dataclasses are supported for "
                "user defined datatypes in Flytekit"
            )
        if not self.is_serializable_class(expected_python_type):
            raise TypeTransformerFailedError(
                f"Dataclass {expected_python_type} should be decorated with @dataclass_json or mixin with DataClassJSONMixin to be "
                f"serialized correctly"
            )
        json_str = _json_format.MessageToJson(lv.scalar.generic)
        dc = expected_python_type.from_json(json_str)  # type: ignore

        dc = self._fix_structured_dataset_type(expected_python_type, dc)
        return self._fix_dataclass_int(expected_python_type, self._deserialize_flyte_type(dc, expected_python_type))

    # This ensures that calls with the same literal type returns the same dataclass. For example, `pyflyte run``
    # command needs to call guess_python_type to get the TypeEngine-derived dataclass. Without caching here, separate
    # calls to guess_python_type would result in a logically equivalent (but new) dataclass, which
    # TypeEngine.assert_type would not be happy about.
    @lru_cache(typed=True)
    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:  # type: ignore
        if literal_type.simple == SimpleType.STRUCT:
            if literal_type.metadata is not None:
                if DEFINITIONS in literal_type.metadata:
                    schema_name = literal_type.metadata["$ref"].split("/")[-1]
                    return convert_marshmallow_json_schema_to_python_class(
                        literal_type.metadata[DEFINITIONS], schema_name
                    )
                elif TITLE in literal_type.metadata:
                    schema_name = literal_type.metadata[TITLE]
                    return convert_mashumaro_json_schema_to_python_class(literal_type.metadata, schema_name)
        raise ValueError(f"Dataclass transformer cannot reverse {literal_type}")


class ProtobufTransformer(TypeTransformer[Message]):
    PB_FIELD_KEY = "pb_type"

    def __init__(self):
        super().__init__("Protobuf-Transformer", Message)

    @staticmethod
    def tag(expected_python_type: Type[T]) -> str:
        return f"{expected_python_type.__module__}.{expected_python_type.__name__}"

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        return LiteralType(simple=SimpleType.STRUCT, metadata={ProtobufTransformer.PB_FIELD_KEY: self.tag(t)})

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        struct = Struct()
        try:
            struct.update(_MessageToDict(cast(Message, python_val)))
        except Exception:
            raise TypeTransformerFailedError("Failed to convert to generic protobuf struct")
        return Literal(scalar=Scalar(generic=struct))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        if not (lv and lv.scalar and lv.scalar.generic is not None):
            raise TypeTransformerFailedError("Can only convert a generic literal to a Protobuf")

        pb_obj = expected_python_type()
        dictionary = _MessageToDict(lv.scalar.generic)
        pb_obj = _ParseDict(dictionary, pb_obj)  # type: ignore
        return pb_obj

    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:
        if (
            literal_type.simple == SimpleType.STRUCT
            and literal_type.metadata
            and literal_type.metadata.get(self.PB_FIELD_KEY, "")
        ):
            tag = literal_type.metadata[self.PB_FIELD_KEY]
            return load_type_from_tag(tag)
        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


class EnumTransformer(TypeTransformer[enum.Enum]):
    """
    Enables converting a python type enum.Enum to LiteralType.EnumType
    """

    def __init__(self):
        super().__init__(name="DefaultEnumTransformer", t=enum.Enum)

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        if is_annotated(t):
            raise ValueError(
                f"Flytekit does not currently have support \
                    for FlyteAnnotations applied to enums. {t} cannot be \
                    parsed."
            )

        values = [v.value for v in t]  # type: ignore
        if not isinstance(values[0], str):
            raise TypeTransformerFailedError("Only EnumTypes with value of string are supported")
        return LiteralType(enum_type=_core_types.EnumType(values=values))

    def to_literal(
        self, ctx: FlyteContext, python_val: enum.Enum, python_type: Type[T], expected: LiteralType
    ) -> Literal:
        if type(python_val).__class__ != enum.EnumMeta:
            raise TypeTransformerFailedError("Expected an enum")
        if type(python_val.value) != str:
            raise TypeTransformerFailedError("Only string-valued enums are supportedd")

        return Literal(scalar=Scalar(primitive=Primitive(string_value=python_val.value)))  # type: ignore

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        return expected_python_type(lv.scalar.primitive.string_value)  # type: ignore

    def guess_python_type(self, literal_type: LiteralType) -> Type[enum.Enum]:
        if literal_type.enum_type:
            return enum.Enum("DynamicEnum", {f"{i}": i for i in literal_type.enum_type.values})  # type: ignore
        raise ValueError(f"Enum transformer cannot reverse {literal_type}")


def generate_attribute_list_from_dataclass_json_mixin(schema: dict, schema_name: typing.Any):
    attribute_list = []
    for property_key, property_val in schema["properties"].items():
        if property_val.get("anyOf"):
            property_type = property_val["anyOf"][0]["type"]
        elif property_val.get("enum"):
            property_type = "enum"
        else:
            property_type = property_val["type"]
        # Handle list
        if property_type == "array":
            attribute_list.append((property_key, typing.List[_get_element_type(property_val["items"])]))  # type: ignore
        # Handle dataclass and dict
        elif property_type == "object":
            if property_val.get("anyOf"):
                sub_schemea = property_val["anyOf"][0]
                sub_schemea_name = sub_schemea["title"]
                attribute_list.append(
                    (property_key, convert_mashumaro_json_schema_to_python_class(sub_schemea, sub_schemea_name))
                )
            elif property_val.get("additionalProperties"):
                attribute_list.append(
                    (property_key, typing.Dict[str, _get_element_type(property_val["additionalProperties"])])  # type: ignore
                )
            else:
                sub_schemea_name = property_val["title"]
                attribute_list.append(
                    (property_key, convert_mashumaro_json_schema_to_python_class(property_val, sub_schemea_name))
                )
        elif property_type == "enum":
            attribute_list.append([property_key, str])  # type: ignore
        # Handle int, float, bool or str
        else:
            attribute_list.append([property_key, _get_element_type(property_val)])  # type: ignore
    return attribute_list


class TypeEngine(typing.Generic[T]):
    """
    Core Extensible TypeEngine of Flytekit. This should be used to extend the capabilities of FlyteKits type system.
    Users can implement their own TypeTransformers and register them with the TypeEngine. This will allow special handling
    of user objects
    """

    _REGISTRY: typing.Dict[type, TypeTransformer[T]] = {}
    _RESTRICTED_TYPES: typing.List[type] = []
    _DATACLASS_TRANSFORMER: TypeTransformer = DataclassTransformer()  # type: ignore
    _ENUM_TRANSFORMER: TypeTransformer = EnumTransformer()  # type: ignore
    has_lazy_import = False

    @classmethod
    def register(
        cls,
        transformer: TypeTransformer,
        additional_types: Optional[typing.List[Type]] = None,
    ):
        """
        This should be used for all types that respond with the right type annotation when you use type(...) function
        """
        types = [transformer.python_type, *(additional_types or [])]
        for t in types:
            if t in cls._REGISTRY:
                existing = cls._REGISTRY[t]
                raise ValueError(
                    f"Transformer {existing.name} for type {t} is already registered."
                    f" Cannot override with {transformer.name}"
                )
            cls._REGISTRY[t] = transformer

    @classmethod
    def register_restricted_type(
        cls,
        name: str,
        type: Type[T],
    ):
        cls._RESTRICTED_TYPES.append(type)
        cls.register(RestrictedTypeTransformer(name, type))  # type: ignore

    @classmethod
    def register_additional_type(cls, transformer: TypeTransformer, additional_type: Type, override=False):
        if additional_type not in cls._REGISTRY or override:
            cls._REGISTRY[additional_type] = transformer

    @classmethod
    def get_transformer(cls, python_type: Type) -> TypeTransformer[T]:
        """
        The TypeEngine hierarchy for flyteKit. This method looksup and selects the type transformer. The algorithm is
        as follows

          d = dictionary of registered transformers, where is a python `type`
          v = lookup type
        Step 1:
            If the type is annotated with a TypeTransformer instance, use that.

        Step 2:
            find a transformer that matches v exactly

        Step 3:
            find a transformer that matches the generic type of v. e.g List[int], Dict[str, int] etc

        Step 4:
            Walk the inheritance hierarchy of v and find a transformer that matches the first base class.
            This is potentially non-deterministic - will depend on the registration pattern.

            Special case:
                If v inherits from Enum, use the Enum transformer even if Enum is not the first base class.

            TODO lets make this deterministic by using an ordered dict

        Step 5:
            if v is of type data class, use the dataclass transformer
        """
        cls.lazy_import_transformers()
        # Step 1
        if is_annotated(python_type):
            args = get_args(python_type)
            for annotation in args:
                if isinstance(annotation, TypeTransformer):
                    return annotation

            python_type = args[0]

        # Step 2
        # this makes sure that if it's a list/dict of annotated types, we hit the unwrapping code in step 2
        # see test_list_of_annotated in test_structured_dataset.py
        if (
            (not hasattr(python_type, "__origin__"))
            or (
                hasattr(python_type, "__origin__")
                and (python_type.__origin__ is not list and python_type.__origin__ is not dict)
            )
        ) and python_type in cls._REGISTRY:
            return cls._REGISTRY[python_type]

        # Step 3
        if hasattr(python_type, "__origin__"):
            # Handling of annotated generics, eg:
            # Annotated[typing.List[int], 'foo']
            if is_annotated(python_type):
                return cls.get_transformer(get_args(python_type)[0])

            if python_type.__origin__ in cls._REGISTRY:
                return cls._REGISTRY[python_type.__origin__]

            raise ValueError(f"Generic Type {python_type.__origin__} not supported currently in Flytekit.")

        # Step 4
        # To facilitate cases where users may specify one transformer for multiple types that all inherit from one
        # parent.
        if inspect.isclass(python_type) and issubclass(python_type, enum.Enum):
            # Special case: prevent that for a type `FooEnum(str, Enum)`, the str transformer is used.
            return cls._ENUM_TRANSFORMER

        for base_type in cls._REGISTRY.keys():
            if base_type is None:
                continue  # None is actually one of the keys, but isinstance/issubclass doesn't work on it
            try:
                if isinstance(python_type, base_type) or (
                    inspect.isclass(python_type) and issubclass(python_type, base_type)
                ):
                    return cls._REGISTRY[base_type]
            except TypeError:
                # As of python 3.9, calls to isinstance raise a TypeError if the base type is not a valid type, which
                # is the case for one of the restricted types, namely NamedTuple.
                logger.debug(f"Invalid base type {base_type} in call to isinstance", exc_info=True)

        # Step 5
        if dataclasses.is_dataclass(python_type):
            return cls._DATACLASS_TRANSFORMER

        # Step 6
        display_pickle_warning(str(python_type))
        from flytekit.types.pickle.pickle import FlytePickleTransformer

        return FlytePickleTransformer()

    @classmethod
    def lazy_import_transformers(cls):
        """
        Only load the transformers if needed.
        """
        if cls.has_lazy_import:
            return
        cls.has_lazy_import = True
        from flytekit.types.structured import (
            register_arrow_handlers,
            register_bigquery_handlers,
            register_pandas_handlers,
        )

        if is_imported("tensorflow"):
            from flytekit.extras import tensorflow  # noqa: F401
        if is_imported("torch"):
            from flytekit.extras import pytorch  # noqa: F401
        if is_imported("sklearn"):
            from flytekit.extras import sklearn  # noqa: F401
        if is_imported("pandas"):
            try:
                from flytekit.types.schema.types_pandas import PandasSchemaReader, PandasSchemaWriter  # noqa: F401
            except ValueError:
                logger.debug("Transformer for pandas is already registered.")
            register_pandas_handlers()
        if is_imported("pyarrow"):
            register_arrow_handlers()
        if is_imported("google.cloud.bigquery"):
            register_bigquery_handlers()
        if is_imported("numpy"):
            from flytekit.types import numpy  # noqa: F401
        if is_imported("PIL"):
            from flytekit.types.file import image  # noqa: F401

    @classmethod
    def to_literal_type(cls, python_type: Type) -> LiteralType:
        """
        Converts a python type into a flyte specific ``LiteralType``
        """
        transformer = cls.get_transformer(python_type)
        res = transformer.get_literal_type(python_type)
        data = None
        if is_annotated(python_type):
            for x in get_args(python_type)[1:]:
                if not isinstance(x, FlyteAnnotation):
                    continue
                if data is not None:
                    raise ValueError(
                        f"More than one FlyteAnnotation used within {python_type} typehint. Flytekit requires a max of one."
                    )
                data = x.data
        if data is not None:
            idl_type_annotation = TypeAnnotationModel(annotations=data)
            res = LiteralType.from_flyte_idl(res.to_flyte_idl())
            res._annotation = idl_type_annotation
        return res

    @classmethod
    def to_literal(cls, ctx: FlyteContext, python_val: typing.Any, python_type: Type, expected: LiteralType) -> Literal:
        """
        Converts a python value of a given type and expected ``LiteralType`` into a resolved ``Literal`` value.
        """
        from flytekit.core.promise import Promise, VoidPromise

        if isinstance(python_val, Promise):
            # In the example above, this handles the "in2=a" type of argument
            return python_val.val
        if isinstance(python_val, VoidPromise):
            raise AssertionError(
                f"Outputs of a non-output producing task {python_val.task_name} cannot be passed to another task."
            )
        if isinstance(python_val, tuple):
            raise AssertionError(
                "Tuples are not a supported type for individual values in Flyte - got a tuple -"
                f" {python_val}. If using named tuple in an inner task, please, de-reference the"
                "actual attribute that you want to use. For example, in NamedTuple('OP', x=int) then"
                "return v.x, instead of v, even if this has a single element"
            )
        if python_val is None and expected and expected.union_type is None:
            raise TypeTransformerFailedError(f"Python value cannot be None, expected {python_type}/{expected}")
        transformer = cls.get_transformer(python_type)
        if transformer.type_assertions_enabled:
            transformer.assert_type(python_type, python_val)

        # In case the value is an annotated type we inspect the annotations and look for hash-related annotations.
        hash = None
        if is_annotated(python_type):
            # We are now dealing with one of two cases:
            # 1. The annotated type is a `HashMethod`, which indicates that we should produce the hash using
            #    the method indicated in the annotation.
            # 2. The annotated type is being used for a different purpose other than calculating hash values, in which case
            #    we should just continue.
            for annotation in get_args(python_type)[1:]:
                if not isinstance(annotation, HashMethod):
                    continue
                hash = annotation.calculate(python_val)
                break

        lv = transformer.to_literal(ctx, python_val, python_type, expected)
        modify_literal_uris(lv)
        if hash is not None:
            lv.hash = hash
        return lv

    @classmethod
    def to_python_value(cls, ctx: FlyteContext, lv: Literal, expected_python_type: Type) -> typing.Any:
        """
        Converts a Literal value with an expected python type into a python value.
        """
        transformer = cls.get_transformer(expected_python_type)
        return transformer.to_python_value(ctx, lv, expected_python_type)

    @classmethod
    def to_html(cls, ctx: FlyteContext, python_val: typing.Any, expected_python_type: Type[typing.Any]) -> str:
        transformer = cls.get_transformer(expected_python_type)
        if is_annotated(expected_python_type):
            expected_python_type, *annotate_args = get_args(expected_python_type)
            from flytekit.deck.renderer import Renderable

            for arg in annotate_args:
                if isinstance(arg, Renderable):
                    return arg.to_html(python_val)
        return transformer.to_html(ctx, python_val, expected_python_type)

    @classmethod
    def named_tuple_to_variable_map(cls, t: typing.NamedTuple) -> _interface_models.VariableMap:
        """
        Converts a python-native ``NamedTuple`` to a flyte-specific VariableMap of named literals.
        """
        variables = {}
        for idx, (var_name, var_type) in enumerate(t.__annotations__.items()):
            literal_type = cls.to_literal_type(var_type)
            variables[var_name] = _interface_models.Variable(type=literal_type, description=f"{idx}")
        return _interface_models.VariableMap(variables=variables)

    @classmethod
    @timeit("Translate literal to python value")
    def literal_map_to_kwargs(
        cls, ctx: FlyteContext, lm: LiteralMap, python_types: typing.Dict[str, type]
    ) -> typing.Dict[str, typing.Any]:
        """
        Given a ``LiteralMap`` (usually an input into a task - intermediate), convert to kwargs for the task
        """
        if len(lm.literals) > len(python_types):
            raise ValueError(
                f"Received more input values {len(lm.literals)}" f" than allowed by the input spec {len(python_types)}"
            )
        kwargs = {}
        for i, k in enumerate(lm.literals):
            try:
                kwargs[k] = TypeEngine.to_python_value(ctx, lm.literals[k], python_types[k])
            except TypeTransformerFailedError as exc:
                raise TypeTransformerFailedError(f"Error converting input '{k}' at position {i}:\n  {exc}") from exc
        return kwargs

    @classmethod
    def dict_to_literal_map(
        cls,
        ctx: FlyteContext,
        d: typing.Dict[str, typing.Any],
        type_hints: Optional[typing.Dict[str, type]] = None,
    ) -> LiteralMap:
        """
        Given a dictionary mapping string keys to python values and a dictionary containing guessed types for such string keys,
        convert to a LiteralMap.
        """
        type_hints = type_hints or {}
        literal_map = {}
        for k, v in d.items():
            # The guessed type takes precedence over the type returned by the python runtime. This is needed
            # to account for the type erasure that happens in the case of built-in collection containers, such as
            # `list` and `dict`.
            python_type = type_hints.get(k, type(v))
            try:
                literal_map[k] = TypeEngine.to_literal(
                    ctx=ctx,
                    python_val=v,
                    python_type=python_type,
                    expected=TypeEngine.to_literal_type(python_type),
                )
            except TypeError:
                raise user_exceptions.FlyteTypeException(type(v), python_type, received_value=v)
        return LiteralMap(literal_map)

    @classmethod
    def get_available_transformers(cls) -> typing.KeysView[Type]:
        """
        Returns all python types for which transformers are available
        """
        return cls._REGISTRY.keys()

    @classmethod
    def guess_python_types(
        cls, flyte_variable_dict: typing.Dict[str, _interface_models.Variable]
    ) -> typing.Dict[str, type]:
        """
        Transforms a dictionary of flyte-specific ``Variable`` objects to a dictionary of regular python values.
        """
        python_types = {}
        for k, v in flyte_variable_dict.items():
            python_types[k] = cls.guess_python_type(v.type)
        return python_types

    @classmethod
    def guess_python_type(cls, flyte_type: LiteralType) -> type:
        """
        Transforms a flyte-specific ``LiteralType`` to a regular python value.
        """
        for _, transformer in cls._REGISTRY.items():
            try:
                return transformer.guess_python_type(flyte_type)
            except ValueError:
                logger.debug(f"Skipping transformer {transformer.name} for {flyte_type}")

        # Because the dataclass transformer is handled explicitly in the get_transformer code, we have to handle it
        # separately here too.
        try:
            return cls._DATACLASS_TRANSFORMER.guess_python_type(literal_type=flyte_type)
        except ValueError:
            logger.debug(f"Skipping transformer {cls._DATACLASS_TRANSFORMER.name} for {flyte_type}")
        raise ValueError(f"No transformers could reverse Flyte literal type {flyte_type}")


class ListTransformer(TypeTransformer[T]):
    """
    Transformer that handles a univariate typing.List[T]
    """

    def __init__(self):
        super().__init__("Typed List", list)

    @staticmethod
    def get_sub_type(t: Type[T]) -> Type[T]:
        """
        Return the generic Type T of the List
        """
        if (sub_type := ListTransformer.get_sub_type_or_none(t)) is not None:
            return sub_type

        raise ValueError("Only generic univariate typing.List[T] type is supported.")

    @staticmethod
    def get_sub_type_or_none(t: Type[T]) -> Optional[Type[T]]:
        """
        Return the generic Type T of the List, or None if the generic type cannot be inferred
        """
        if hasattr(t, "__origin__"):
            # Handle annotation on list generic, eg:
            # Annotated[typing.List[int], 'foo']
            if is_annotated(t):
                return ListTransformer.get_sub_type(get_args(t)[0])

            if getattr(t, "__origin__") is list and hasattr(t, "__args__"):
                return getattr(t, "__args__")[0]

        return None

    def get_literal_type(self, t: Type[T]) -> Optional[LiteralType]:
        """
        Only univariate Lists are supported in Flyte
        """
        try:
            sub_type = TypeEngine.to_literal_type(self.get_sub_type(t))
            return _type_models.LiteralType(collection_type=sub_type)
        except Exception as e:
            raise ValueError(f"Type of Generic List type is not supported, {e}")

    @staticmethod
    def is_batchable(t: Type):
        """
        This function evaluates whether the provided type is batchable or not.
        It returns True only if the type is either List or Annotated(List) and the List subtype is FlytePickle.
        """
        from flytekit.types.pickle import FlytePickle

        if is_annotated(t):
            return ListTransformer.is_batchable(get_args(t)[0])
        if get_origin(t) is list:
            subtype = get_args(t)[0]
            if subtype == FlytePickle or (hasattr(subtype, "__origin__") and subtype.__origin__ == FlytePickle):
                return True
        return False

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        if type(python_val) != list:
            raise TypeTransformerFailedError("Expected a list")

        if ListTransformer.is_batchable(python_type):
            from flytekit.types.pickle.pickle import BatchSize, FlytePickle

            batch_size = len(python_val)  # default batch size
            # parse annotated to get the number of items saved in a pickle file.
            if is_annotated(python_type):
                for annotation in get_args(python_type)[1:]:
                    if isinstance(annotation, BatchSize):
                        batch_size = annotation.val
                        break
            if batch_size > 0:
                lit_list = [
                    TypeEngine.to_literal(ctx, python_val[i : i + batch_size], FlytePickle, expected.collection_type)
                    for i in range(0, len(python_val), batch_size)
                ]  # type: ignore
            else:
                lit_list = []
        else:
            t = self.get_sub_type(python_type)
            lit_list = [TypeEngine.to_literal(ctx, x, t, expected.collection_type) for x in python_val]  # type: ignore
        return Literal(collection=LiteralCollection(literals=lit_list))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> typing.List[typing.Any]:  # type: ignore
        try:
            lits = lv.collection.literals
        except AttributeError:
            raise TypeTransformerFailedError(
                (
                    f"The expected python type is '{expected_python_type}' but the received Flyte literal value "
                    f"is not a collection (Flyte's representation of Python lists)."
                )
            )
        if self.is_batchable(expected_python_type):
            from flytekit.types.pickle import FlytePickle

            batch_list = [TypeEngine.to_python_value(ctx, batch, FlytePickle) for batch in lits]
            if len(batch_list) > 0 and type(batch_list[0]) is list:
                # Make it have backward compatibility. The upstream task may use old version of Flytekit that
                # won't merge the elements in the list. Therefore, we should check if the batch_list[0] is the list first.
                return [item for batch in batch_list for item in batch]
            return batch_list
        else:
            st = self.get_sub_type(expected_python_type)
            return [TypeEngine.to_python_value(ctx, x, st) for x in lits]

    def guess_python_type(self, literal_type: LiteralType) -> list:  # type: ignore
        if literal_type.collection_type:
            ct: Type = TypeEngine.guess_python_type(literal_type.collection_type)
            return typing.List[ct]  # type: ignore
        raise ValueError(f"List transformer cannot reverse {literal_type}")


@lru_cache
def display_pickle_warning(python_type: str):
    # This is a warning that is only displayed once per python type
    logger.warning(
        f"Unsupported Type {python_type} found, Flyte will default to use PickleFile as the transport. "
        f"Pickle can only be used to send objects between the exact same version of Python, "
        f"and we strongly recommend to use python type that flyte support."
    )


def _add_tag_to_type(x: LiteralType, tag: str) -> LiteralType:
    x._structure = TypeStructure(tag=tag)
    return x


def _type_essence(x: LiteralType) -> LiteralType:
    if x.metadata is not None or x.structure is not None or x.annotation is not None:
        x = LiteralType.from_flyte_idl(x.to_flyte_idl())
        x._metadata = None
        x._structure = None
        x._annotation = None

    return x


def _are_types_castable(upstream: LiteralType, downstream: LiteralType) -> bool:
    if upstream.collection_type is not None:
        if downstream.collection_type is None:
            return False

        return _are_types_castable(upstream.collection_type, downstream.collection_type)

    if upstream.map_value_type is not None:
        if downstream.map_value_type is None:
            return False

        return _are_types_castable(upstream.map_value_type, downstream.map_value_type)

    # TODO: Structured dataset type matching requires that downstream structured datasets
    # are a strict sub-set of the upstream structured dataset.
    if upstream.structured_dataset_type is not None:
        if downstream.structured_dataset_type is None:
            return False

        usdt = upstream.structured_dataset_type
        dsdt = downstream.structured_dataset_type

        if usdt.format != dsdt.format:
            return False

        if usdt.external_schema_type != dsdt.external_schema_type:
            return False

        if usdt.external_schema_bytes != dsdt.external_schema_bytes:
            return False

        ucols = usdt.columns
        dcols = dsdt.columns

        if len(ucols) != len(dcols):
            return False

        for u, d in zip(ucols, dcols):
            if u.name != d.name:
                return False

            if not _are_types_castable(u.literal_type, d.literal_type):
                return False

        return True

    if upstream.union_type is not None:
        # for each upstream variant, there must be a compatible type downstream
        for v in upstream.union_type.variants:
            if not _are_types_castable(v, downstream):
                return False
        return True

    if downstream.union_type is not None:
        # there must be a compatible downstream type
        for v in downstream.union_type.variants:
            if _are_types_castable(upstream, v):
                return True

    if upstream.enum_type is not None:
        # enums are castable to string
        if downstream.simple == SimpleType.STRING:
            return True

    if _type_essence(upstream) == _type_essence(downstream):
        return True

    return False


class UnionTransformer(TypeTransformer[T]):
    """
    Transformer that handles a typing.Union[T1, T2, ...]
    """

    def __init__(self):
        super().__init__("Typed Union", typing.Union)

    @staticmethod
    def is_optional_type(t: Type[T]) -> bool:
        return get_origin(t) is typing.Union and type(None) in get_args(t)

    @staticmethod
    def get_sub_type_in_optional(t: Type[T]) -> Type[T]:
        """
        Return the generic Type T of the Optional type
        """
        return get_args(t)[0]

    def get_literal_type(self, t: Type[T]) -> Optional[LiteralType]:
        t = get_underlying_type(t)

        try:
            trans: typing.List[typing.Tuple[TypeTransformer, typing.Any]] = [
                (TypeEngine.get_transformer(x), x) for x in get_args(t)
            ]
            # must go through TypeEngine.to_literal_type instead of trans.get_literal_type
            # to handle Annotated
            variants = [_add_tag_to_type(TypeEngine.to_literal_type(x), t.name) for (t, x) in trans]
            return _type_models.LiteralType(union_type=UnionType(variants))
        except Exception as e:
            raise ValueError(f"Type of Generic Union type is not supported, {e}")

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        python_type = get_underlying_type(python_type)

        found_res = False
        is_ambiguous = False
        res = None
        res_type = None
        for i in range(len(get_args(python_type))):
            try:
                t = get_args(python_type)[i]
                trans: TypeTransformer[T] = TypeEngine.get_transformer(t)
                res = trans.to_literal(ctx, python_val, t, expected.union_type.variants[i])
                res_type = _add_tag_to_type(trans.get_literal_type(t), trans.name)
                if found_res:
                    is_ambiguous = True
                found_res = True
            except Exception as e:
                logger.debug(f"Failed to convert from {python_val} to {t}", e)
                continue

        if is_ambiguous:
            raise TypeError("Ambiguous choice of variant for union type")

        if found_res:
            return Literal(scalar=Scalar(union=Union(value=res, stored_type=res_type)))

        raise TypeTransformerFailedError(f"Cannot convert from {python_val} to {python_type}")

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> Optional[typing.Any]:
        expected_python_type = get_underlying_type(expected_python_type)

        union_tag = None
        union_type = None
        if lv.scalar is not None and lv.scalar.union is not None:
            union_type = lv.scalar.union.stored_type
            if union_type.structure is not None:
                union_tag = union_type.structure.tag

        found_res = False
        is_ambiguous = False
        cur_transformer = ""
        res = None
        res_tag = None
        for v in get_args(expected_python_type):
            try:
                trans: TypeTransformer[T] = TypeEngine.get_transformer(v)
                if union_tag is not None:
                    if trans.name != union_tag:
                        continue

                    expected_literal_type = TypeEngine.to_literal_type(v)
                    if not _are_types_castable(union_type, expected_literal_type):
                        continue

                    assert lv.scalar is not None  # type checker
                    assert lv.scalar.union is not None  # type checker

                    res = trans.to_python_value(ctx, lv.scalar.union.value, v)
                    if found_res:
                        is_ambiguous = True
                        cur_transformer = trans.name
                        break
                else:
                    res = trans.to_python_value(ctx, lv, v)
                    if found_res:
                        is_ambiguous = True
                        cur_transformer = trans.name
                        break
                res_tag = trans.name
                found_res = True
            except Exception as e:
                logger.debug(f"Failed to convert from {lv} to {v}", e)

        if is_ambiguous:
            raise TypeError(
                "Ambiguous choice of variant for union type. "
                + f"Both {res_tag} and {cur_transformer} transformers match"
            )

        if found_res:
            return res

        raise TypeError(f"Cannot convert from {lv} to {expected_python_type} (using tag {union_tag})")

    def guess_python_type(self, literal_type: LiteralType) -> type:
        if literal_type.union_type is not None:
            return typing.Union[tuple(TypeEngine.guess_python_type(v) for v in literal_type.union_type.variants)]  # type: ignore

        raise ValueError(f"Union transformer cannot reverse {literal_type}")


class DictTransformer(TypeTransformer[dict]):
    """
    Transformer that transforms a univariate dictionary Dict[str, T] to a Literal Map or
    transforms a untyped dictionary to a JSON (struct/Generic)
    """

    def __init__(self):
        super().__init__("Typed Dict", dict)

    @staticmethod
    def get_dict_types(t: Optional[Type[dict]]) -> typing.Tuple[Optional[type], Optional[type]]:
        """
        Return the generic Type T of the Dict
        """
        _origin = get_origin(t)
        _args = get_args(t)
        if _origin is not None:
            if _origin is Annotated:
                raise ValueError(
                    f"Flytekit does not currently have support \
                        for FlyteAnnotations applied to dicts. {t} cannot be \
                        parsed."
                )
            if _origin is dict and _args is not None:
                return _args  # type: ignore
        return None, None

    @staticmethod
    def dict_to_generic_literal(v: dict) -> Literal:
        """
        Creates a flyte-specific ``Literal`` value from a native python dictionary.
        """
        return Literal(scalar=Scalar(generic=_json_format.Parse(_json.dumps(v), _struct.Struct())))

    def get_literal_type(self, t: Type[dict]) -> LiteralType:
        """
        Transforms a native python dictionary to a flyte-specific ``LiteralType``
        """
        tp = self.get_dict_types(t)
        if tp:
            if tp[0] == str:
                try:
                    sub_type = TypeEngine.to_literal_type(cast(type, tp[1]))
                    return _type_models.LiteralType(map_value_type=sub_type)
                except Exception as e:
                    raise ValueError(f"Type of Generic List type is not supported, {e}")
        return _type_models.LiteralType(simple=_type_models.SimpleType.STRUCT)

    def to_literal(
        self, ctx: FlyteContext, python_val: typing.Any, python_type: Type[dict], expected: LiteralType
    ) -> Literal:
        if type(python_val) != dict:
            raise TypeTransformerFailedError("Expected a dict")

        if expected and expected.simple and expected.simple == SimpleType.STRUCT:
            return self.dict_to_generic_literal(python_val)

        lit_map = {}
        for k, v in python_val.items():
            if type(k) != str:
                raise ValueError("Flyte MapType expects all keys to be strings")
            # TODO: log a warning for Annotated objects that contain HashMethod
            k_type, v_type = self.get_dict_types(python_type)
            lit_map[k] = TypeEngine.to_literal(ctx, v, cast(type, v_type), expected.map_value_type)
        return Literal(map=LiteralMap(literals=lit_map))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[dict]) -> dict:
        if lv and lv.map and lv.map.literals is not None:
            tp = self.get_dict_types(expected_python_type)
            if tp is None or tp[0] is None:
                raise TypeError(
                    "TypeMismatch: Cannot convert to python dictionary from Flyte Literal Dictionary as the given "
                    "dictionary does not have sub-type hints or they do not match with the originating dictionary "
                    "source. Flytekit does not currently support implicit conversions"
                )
            if tp[0] != str:
                raise TypeError("TypeMismatch. Destination dictionary does not accept 'str' key")
            py_map = {}
            for k, v in lv.map.literals.items():
                py_map[k] = TypeEngine.to_python_value(ctx, v, cast(Type, tp[1]))
            return py_map

        # for empty generic we have to explicitly test for lv.scalar.generic is not None as empty dict
        # evaluates to false
        if lv and lv.scalar and lv.scalar.generic is not None:
            try:
                return _json.loads(_json_format.MessageToJson(lv.scalar.generic))
            except TypeError:
                raise TypeTransformerFailedError(f"Cannot convert from {lv} to {expected_python_type}")
        raise TypeTransformerFailedError(f"Cannot convert from {lv} to {expected_python_type}")

    def guess_python_type(self, literal_type: LiteralType) -> Union[Type[dict], typing.Dict[Type, Type]]:
        if literal_type.map_value_type:
            mt = TypeEngine.guess_python_type(literal_type.map_value_type)
            return typing.Dict[str, mt]  # type: ignore

        if literal_type.simple == SimpleType.STRUCT:
            if literal_type.metadata is None:
                return dict  # type: ignore

        raise ValueError(f"Dictionary transformer cannot reverse {literal_type}")


class TextIOTransformer(TypeTransformer[typing.TextIO]):
    """
    Handler for TextIO
    """

    def __init__(self):
        super().__init__(name="TextIO", t=typing.TextIO)

    def _blob_type(self) -> _core_types.BlobType:
        return _core_types.BlobType(
            format=mimetypes.types_map[".txt"],
            dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
        )

    def get_literal_type(self, t: typing.TextIO) -> LiteralType:  # type: ignore
        return _type_models.LiteralType(blob=self._blob_type())

    def to_literal(
        self, ctx: FlyteContext, python_val: typing.TextIO, python_type: Type[typing.TextIO], expected: LiteralType
    ) -> Literal:
        raise NotImplementedError("Implement handle for TextIO")

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[typing.TextIO]
    ) -> typing.TextIO:
        # TODO rename to get_auto_local_path()
        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(lv.scalar.blob.uri, local_path, is_multipart=False)
        # TODO it is probably the responsibility of the framework to close() this
        return open(local_path, "r")


class BinaryIOTransformer(TypeTransformer[typing.BinaryIO]):
    """
    Handler for BinaryIO
    """

    def __init__(self):
        super().__init__(name="BinaryIO", t=typing.BinaryIO)

    def _blob_type(self) -> _core_types.BlobType:
        return _core_types.BlobType(
            format=mimetypes.types_map[".bin"],
            dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
        )

    def get_literal_type(self, t: Type[typing.BinaryIO]) -> LiteralType:
        return _type_models.LiteralType(
            blob=self._blob_type(),
        )

    def to_literal(
        self, ctx: FlyteContext, python_val: typing.BinaryIO, python_type: Type[typing.BinaryIO], expected: LiteralType
    ) -> Literal:
        raise NotImplementedError("Implement handle for TextIO")

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[typing.BinaryIO]
    ) -> typing.BinaryIO:
        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(lv.scalar.blob.uri, local_path, is_multipart=False)
        # TODO it is probability the responsibility of the framework to close this
        return open(local_path, "rb")


def generate_attribute_list_from_dataclass_json(schema: dict, schema_name: typing.Any):
    attribute_list = []
    for property_key, property_val in schema[schema_name]["properties"].items():
        property_type = property_val["type"]
        # Handle list
        if property_val["type"] == "array":
            attribute_list.append((property_key, List[_get_element_type(property_val["items"])]))  # type: ignore[misc,index]
        # Handle dataclass and dict
        elif property_type == "object":
            if property_val.get("$ref"):
                name = property_val["$ref"].split("/")[-1]
                attribute_list.append((property_key, convert_marshmallow_json_schema_to_python_class(schema, name)))
            elif property_val.get("additionalProperties"):
                attribute_list.append(
                    (property_key, Dict[str, _get_element_type(property_val["additionalProperties"])])  # type: ignore[misc,index]
                )
            else:
                attribute_list.append((property_key, Dict[str, _get_element_type(property_val)]))  # type: ignore[misc,index]
        # Handle int, float, bool or str
        else:
            attribute_list.append([property_key, _get_element_type(property_val)])  # type: ignore
    return attribute_list


def convert_marshmallow_json_schema_to_python_class(
    schema: dict, schema_name: typing.Any
) -> Type[dataclasses.dataclass()]:  # type: ignore
    """
    Generate a model class based on the provided JSON Schema
    :param schema: dict representing valid JSON schema
    :param schema_name: dataclass name of return type
    """

    attribute_list = generate_attribute_list_from_dataclass_json(schema, schema_name)
    return dataclass_json(dataclasses.make_dataclass(schema_name, attribute_list))


def convert_mashumaro_json_schema_to_python_class(
    schema: dict, schema_name: typing.Any
) -> Type[dataclasses.dataclass()]:  # type: ignore
    """
    Generate a model class based on the provided JSON Schema
    :param schema: dict representing valid JSON schema
    :param schema_name: dataclass name of return type
    """

    attribute_list = generate_attribute_list_from_dataclass_json_mixin(schema, schema_name)
    return dataclass_json(dataclasses.make_dataclass(schema_name, attribute_list))


def _get_element_type(element_property: typing.Dict[str, str]) -> Type:
    element_type = (
        [e_property["type"] for e_property in element_property["anyOf"]]  # type: ignore
        if element_property.get("anyOf")
        else element_property["type"]
    )
    element_format = element_property["format"] if "format" in element_property else None

    if type(element_type) == list:
        # Element type of Optional[int] is [integer, None]
        return typing.Optional[_get_element_type({"type": element_type[0]})]  # type: ignore

    if element_type == "string":
        return str
    elif element_type == "integer":
        return int
    elif element_type == "boolean":
        return bool
    elif element_type == "number":
        if element_format == "integer":
            return int
        else:
            return float
    return str


def dataclass_from_dict(cls: type, src: typing.Dict[str, typing.Any]) -> typing.Any:
    """
    Utility function to construct a dataclass object from dict
    """
    field_types_lookup = {field.name: field.type for field in dataclasses.fields(cls)}

    constructor_inputs = {}
    for field_name, value in src.items():
        if dataclasses.is_dataclass(field_types_lookup[field_name]):
            constructor_inputs[field_name] = dataclass_from_dict(field_types_lookup[field_name], value)
        else:
            constructor_inputs[field_name] = value

    return cls(**constructor_inputs)


def _check_and_covert_float(lv: Literal) -> float:
    if lv.scalar.primitive.float_value is not None:
        return lv.scalar.primitive.float_value
    elif lv.scalar.primitive.integer is not None:
        return float(lv.scalar.primitive.integer)
    raise TypeTransformerFailedError(f"Cannot convert literal {lv} to float")


def _check_and_convert_void(lv: Literal) -> None:
    if lv.scalar.none_type is None:
        raise TypeTransformerFailedError(f"Cannot convert literal {lv} to None")
    return None


def _register_default_type_transformers():
    TypeEngine.register(
        SimpleTransformer(
            "int",
            int,
            _type_models.LiteralType(simple=_type_models.SimpleType.INTEGER),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(integer=x))),
            lambda x: x.scalar.primitive.integer,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "float",
            float,
            _type_models.LiteralType(simple=_type_models.SimpleType.FLOAT),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(float_value=x))),
            _check_and_covert_float,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "bool",
            bool,
            _type_models.LiteralType(simple=_type_models.SimpleType.BOOLEAN),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(boolean=x))),
            lambda x: x.scalar.primitive.boolean,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "str",
            str,
            _type_models.LiteralType(simple=_type_models.SimpleType.STRING),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(string_value=x))),
            lambda x: x.scalar.primitive.string_value,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "datetime",
            _datetime.datetime,
            _type_models.LiteralType(simple=_type_models.SimpleType.DATETIME),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(datetime=x))),
            lambda x: x.scalar.primitive.datetime,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "timedelta",
            _datetime.timedelta,
            _type_models.LiteralType(simple=_type_models.SimpleType.DURATION),
            lambda x: Literal(scalar=Scalar(primitive=Primitive(duration=x))),
            lambda x: x.scalar.primitive.duration,
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "date",
            _datetime.date,
            _type_models.LiteralType(simple=_type_models.SimpleType.DATETIME),
            lambda x: Literal(
                scalar=Scalar(primitive=Primitive(datetime=_datetime.datetime.combine(x, _datetime.time.min)))
            ),  # convert datetime to date
            lambda x: x.scalar.primitive.datetime.date(),  # get date from datetime
        )
    )

    TypeEngine.register(
        SimpleTransformer(
            "none",
            type(None),
            _type_models.LiteralType(simple=_type_models.SimpleType.NONE),
            lambda x: Literal(scalar=Scalar(none_type=Void())),
            lambda x: _check_and_convert_void(x),
        ),
        [None],
    )
    TypeEngine.register(ListTransformer())
    TypeEngine.register(UnionTransformer())
    TypeEngine.register(DictTransformer())
    TypeEngine.register(TextIOTransformer())
    TypeEngine.register(BinaryIOTransformer())
    TypeEngine.register(EnumTransformer())
    TypeEngine.register(ProtobufTransformer())

    # inner type is. Also unsupported are typing's Tuples. Even though you can look inside them, Flyte's type system
    # doesn't support these currently.
    # Confusing note: typing.NamedTuple is in here even though task functions themselves can return them. We just mean
    # that the return signature of a task can be a NamedTuple that contains another NamedTuple inside it.
    # Also, it's not entirely true that Flyte IDL doesn't support tuples. We can always fake them as structs, but we'll
    # hold off on doing that for now, as we may amend the IDL formally to support tuples.
    TypeEngine.register_restricted_type("non typed tuple", tuple)
    TypeEngine.register_restricted_type("non typed tuple", typing.Tuple)
    TypeEngine.register_restricted_type("named tuple", NamedTuple)


class LiteralsResolver(collections.UserDict):
    """
    LiteralsResolver is a helper class meant primarily for use with the FlyteRemote experience or any other situation
    where you might be working with LiteralMaps. This object allows the caller to specify the Python type that should
    correspond to an element of the map.

    TODO: Consider inheriting from collections.UserDict instead of manually having the _native_values cache
    """

    def __init__(
        self,
        literals: typing.Dict[str, Literal],
        variable_map: Optional[Dict[str, _interface_models.Variable]] = None,
        ctx: Optional[FlyteContext] = None,
    ):
        """
        :param literals: A Python map of strings to Flyte Literal models.
        :param variable_map: This map should be basically one side (either input or output) of the Flyte
          TypedInterface model and is used to guess the Python type through the TypeEngine if a Python type is not
          specified by the user. TypeEngine guessing is flaky though, so calls to get() should specify the as_type
          parameter when possible.
        """
        super().__init__(literals)
        if literals is None:
            raise ValueError("Cannot instantiate LiteralsResolver without a map of Literals.")
        self._literals = literals
        self._variable_map = variable_map
        self._native_values: Dict[str, type] = {}
        self._type_hints: Dict[str, type] = {}
        self._ctx = ctx

    def __str__(self) -> str:
        if self.literals:
            if len(self.literals) == len(self.native_values):
                return str(self.native_values)
            if self.native_values:
                header = "Partially converted to native values, call get(key, <type_hint>) to convert rest...\n"
                strs = []
                for key, literal in self._literals.items():
                    if key in self._native_values:
                        strs.append(f"{key}: " + str(self._native_values[key]) + "\n")
                    else:
                        lit_txt = str(self._literals[key])
                        lit_txt = textwrap.indent(lit_txt, " " * (len(key) + 2))
                        strs.append(f"{key}: \n" + lit_txt)

                return header + "{\n" + textwrap.indent("".join(strs), " " * 2) + "\n}"
            else:
                return str(literal_map_string_repr(self.literals))
        return "{}"

    def __repr__(self):
        return self.__str__()

    @property
    def native_values(self) -> typing.Dict[str, typing.Any]:
        return self._native_values

    @property
    def variable_map(self) -> Optional[Dict[str, _interface_models.Variable]]:
        return self._variable_map

    @property
    def literals(self):
        return self._literals

    def update_type_hints(self, type_hints: typing.Dict[str, typing.Type]):
        self._type_hints.update(type_hints)

    def get_literal(self, key: str) -> Literal:
        if key not in self._literals:
            raise ValueError(f"Key {key} is not in the literal map")

        return self._literals[key]

    def __getitem__(self, key: str):
        # First check to see if it's even in the literal map.
        if key not in self._literals:
            raise ValueError(f"Key {key} is not in the literal map")

        # Return the cached value if it's cached
        if key in self._native_values:
            return self._native_values[key]

        return self.get(key)

    def get(self, attr: str, as_type: Optional[typing.Type] = None) -> typing.Any:  # type: ignore
        """
        This will get the ``attr`` value from the Literal map, and invoke the TypeEngine to convert it into a Python
        native value. A Python type can optionally be supplied. If successful, the native value will be cached and
        future calls will return the cached value instead.

        :param attr:
        :param as_type:
        :return: Python native value from the LiteralMap
        """
        if attr not in self._literals:
            raise AttributeError(f"Attribute {attr} not found")
        if attr in self.native_values:
            return self.native_values[attr]

        if as_type is None:
            if attr in self._type_hints:
                as_type = self._type_hints[attr]
            else:
                if self.variable_map and attr in self.variable_map:
                    try:
                        as_type = TypeEngine.guess_python_type(self.variable_map[attr].type)
                    except ValueError as e:
                        logger.error(f"Could not guess a type for Variable {self.variable_map[attr]}")
                        raise e
                else:
                    raise ValueError("as_type argument not supplied and Variable map not specified in LiteralsResolver")
        val = TypeEngine.to_python_value(
            self._ctx or FlyteContext.current_context(), self._literals[attr], cast(Type, as_type)
        )
        self._native_values[attr] = val
        return val


_register_default_type_transformers()


def is_annotated(t: Type) -> bool:
    return get_origin(t) is Annotated


def get_underlying_type(t: Type) -> Type:
    """Return the underlying type for annotated types or the type itself"""
    if is_annotated(t):
        return get_args(t)[0]
    return t
