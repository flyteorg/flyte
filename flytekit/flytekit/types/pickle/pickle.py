import os
import typing
from typing import Type

import cloudpickle

from flytekit.core.context_manager import FlyteContext, FlyteContextManager
from flytekit.core.type_engine import TypeEngine, TypeTransformer
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType

T = typing.TypeVar("T")


class BatchSize:
    """
    Flyte-specific object used to wrap the hash function for a specific type
    """

    def __init__(self, val: int):
        self._val = val

    @property
    def val(self) -> int:
        return self._val


class FlytePickle(typing.Generic[T]):
    """
    This type is only used by flytekit internally. User should not use this type.
    Any type that flyte can't recognize will become FlytePickle
    """

    @classmethod
    def python_type(cls) -> typing.Type:
        return type(None)

    @classmethod
    def __class_getitem__(cls, python_type: typing.Type) -> typing.Type:
        if python_type is None:
            return cls

        class _SpecificFormatClass(FlytePickle):
            # Get the type engine to see this as kind of a generic
            __origin__ = FlytePickle

            @classmethod
            def python_type(cls) -> typing.Type:
                return python_type

        return _SpecificFormatClass

    @classmethod
    def to_pickle(cls, python_val: typing.Any) -> str:
        ctx = FlyteContextManager.current_context()
        local_dir = ctx.file_access.get_random_local_directory()
        os.makedirs(local_dir, exist_ok=True)
        local_path = ctx.file_access.get_random_local_path()
        uri = os.path.join(local_dir, local_path)
        with open(uri, "w+b") as outfile:
            cloudpickle.dump(python_val, outfile)

        return ctx.file_access.put_raw_data(uri)

    @classmethod
    def from_pickle(cls, uri: str) -> typing.Any:
        ctx = FlyteContextManager.current_context()
        # Deserialize the pickle, and return data in the pickle,
        # and download pickle file to local first if file is not in the local file systems.
        if ctx.file_access.is_remote(uri):
            local_path = ctx.file_access.get_random_local_path()
            ctx.file_access.get_data(uri, local_path, False)
            uri = local_path
        with open(uri, "rb") as infile:
            data = cloudpickle.load(infile)
        return data


class FlytePickleTransformer(TypeTransformer[FlytePickle]):
    PYTHON_PICKLE_FORMAT = "PythonPickle"

    def __init__(self):
        super().__init__(name="FlytePickle", t=FlytePickle)

    def assert_type(self, t: Type[T], v: T):
        # Every type can serialize to pickle, so we don't need to check the type here.
        ...

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        uri = lv.scalar.blob.uri
        return FlytePickle.from_pickle(uri)

    def to_literal(self, ctx: FlyteContext, python_val: T, python_type: Type[T], expected: LiteralType) -> Literal:
        if python_val is None:
            raise AssertionError("Cannot pickle None Value.")
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.PYTHON_PICKLE_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )
        remote_path = FlytePickle.to_pickle(python_val)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def guess_python_type(self, literal_type: LiteralType) -> typing.Type[FlytePickle[typing.Any]]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == FlytePickleTransformer.PYTHON_PICKLE_FORMAT
        ):
            return FlytePickle

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        lt = LiteralType(
            blob=_core_types.BlobType(
                format=self.PYTHON_PICKLE_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )
        lt.metadata = {"python_class_name": str(t)}
        return lt


TypeEngine.register(FlytePickleTransformer())
