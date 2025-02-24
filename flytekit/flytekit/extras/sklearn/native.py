import pathlib
from typing import Type, TypeVar

import joblib
import sklearn

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType

T = TypeVar("T")


class SklearnTypeTransformer(TypeTransformer[T]):
    def get_literal_type(self, t: Type[T]) -> LiteralType:
        return LiteralType(
            blob=_core_types.BlobType(
                format=self.SKLEARN_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
            )
        )

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: T,
        python_type: Type[T],
        expected: LiteralType,
    ) -> Literal:
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.SKLEARN_FORMAT,
                dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE,
            )
        )

        local_path = ctx.file_access.get_random_local_path() + ".joblib"
        pathlib.Path(local_path).parent.mkdir(parents=True, exist_ok=True)

        # save sklearn estimator to a file
        joblib.dump(python_val, local_path)

        remote_path = ctx.file_access.put_raw_data(local_path)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        try:
            uri = lv.scalar.blob.uri
        except AttributeError:
            TypeTransformerFailedError(f"Cannot convert from {lv} to {expected_python_type}")

        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(uri, local_path, is_multipart=False)

        # load sklearn estimator from a file
        return joblib.load(local_path)

    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.SKLEARN_FORMAT
        ):
            return T

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


class SklearnEstimatorTransformer(SklearnTypeTransformer[sklearn.base.BaseEstimator]):
    SKLEARN_FORMAT = "SklearnEstimator"

    def __init__(self):
        super().__init__(name="Sklearn Estimator", t=sklearn.base.BaseEstimator)

    def guess_python_type(self, literal_type: LiteralType) -> Type[sklearn.base.BaseEstimator]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.SKLEARN_FORMAT
        ):
            return sklearn.base.BaseEstimator

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(SklearnEstimatorTransformer())
