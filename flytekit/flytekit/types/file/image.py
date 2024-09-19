import pathlib
import typing
from typing import Type

import PIL.Image

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType

T = typing.TypeVar("T")


class PILImageTransformer(TypeTransformer[T]):
    """
    TypeTransformer that supports PIL.Image as a native type.
    """

    FILE_FORMAT = "PIL.Image"

    def __init__(self):
        super().__init__(name="PIL.Image", t=PIL.Image.Image)

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        return LiteralType(
            blob=_core_types.BlobType(
                format=self.FILE_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )

    def to_literal(
        self, ctx: FlyteContext, python_val: PIL.Image.Image, python_type: Type[T], expected: LiteralType
    ) -> Literal:
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.FILE_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )

        local_path = ctx.file_access.get_random_local_path() + ".png"
        pathlib.Path(local_path).parent.mkdir(parents=True, exist_ok=True)
        python_val.save(local_path)

        remote_path = ctx.file_access.get_random_remote_path(local_path)
        ctx.file_access.put_data(local_path, remote_path, is_multipart=False)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> PIL.Image.Image:
        try:
            uri = lv.scalar.blob.uri
        except AttributeError:
            raise TypeTransformerFailedError(f"Cannot convert from {lv} to {expected_python_type}")

        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(uri, local_path, is_multipart=False)

        return PIL.Image.open(local_path)

    def guess_python_type(self, literal_type: LiteralType) -> Type[T]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.FILE_FORMAT
        ):
            return PIL.Image.Image

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")

    def to_html(self, ctx: FlyteContext, python_val: PIL.Image.Image, expected_python_type: Type[T]) -> str:
        import base64
        from io import BytesIO

        buffered = BytesIO()
        python_val.save(buffered, format="PNG")
        img_base64 = base64.b64encode(buffered.getvalue()).decode()
        return f'<img src="data:image/png;base64,{img_base64}" alt="Rendered Image" />'


TypeEngine.register(PILImageTransformer())
