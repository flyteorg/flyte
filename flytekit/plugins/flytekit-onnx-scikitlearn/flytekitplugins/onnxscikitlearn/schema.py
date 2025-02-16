from __future__ import annotations

import inspect
from dataclasses import dataclass, field
from typing import Any, Callable, Dict, List, Optional, Set, Tuple, Type, Union

from dataclasses_json import DataClassJsonMixin
from typing_extensions import Annotated, get_args, get_origin

from flytekit import FlyteContext, lazy_module
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core.types import BlobType
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from flytekit.types.file import ONNXFile

sklearn = lazy_module("sklearn")
skl2onnx = lazy_module("skl2onnx")
skl2onnx_data_types = lazy_module("skl2onnx.common.data_types")


@dataclass
class ScikitLearn2ONNXConfig(DataClassJsonMixin):
    """
    ScikitLearn2ONNXConfig is the config used during the scikitlearn to ONNX conversion.

    Args:
      initial_types: The types of the inputs to the model.
      name: The name of the graph in the produced ONNX model.
      doc_string: A string attached onto the produced ONNX model.
      target_opset: The ONNX opset number.
      custom_conversion_functions: A dictionary for specifying the user customized conversion function.
      custom_shape_calculators: A dictionary for specifying the user customized shape calculator.
      custom_parsers: Parsers determine which outputs are expected for which particular task.
      options: Specific options given to converters.
      intermediate: If True, the function returns the converted model and the instance of Topology used, else, it returns the converted model.
      naming: Change the way intermediates are named.
      white_op: White list of ONNX nodes allowed while converting a pipeline.
      black_op: Black list of ONNX nodes disallowed while converting a pipeline.
      verbose: Display progress while converting a model.
      final_types: Used to overwrite the type (if type is not None) and the name of every output.
    """

    initial_types: List[Tuple[str, Type]]
    name: Optional[str] = None
    doc_string: str = ""
    target_opset: Optional[int] = None
    custom_conversion_functions: Dict[Callable[..., Any], Callable[..., None]] = field(default_factory=dict)
    custom_shape_calculators: Dict[Callable[..., Any], Callable[..., None]] = field(default_factory=dict)
    custom_parsers: Dict[Callable[..., Any], Callable[..., None]] = field(default_factory=dict)
    options: Dict[Any, Any] = field(default_factory=dict)
    intermediate: bool = False
    naming: Optional[Union[str, Callable[..., Any]]] = None
    white_op: Optional[Set[str]] = None
    black_op: Optional[Set[str]] = None
    verbose: int = 0
    final_types: Optional[List[Tuple[str, Type]]] = None

    def __post_init__(self):
        validate_initial_types = [
            True for item in self.initial_types if item in inspect.getmembers(skl2onnx_data_types)
        ]
        if not all(validate_initial_types):
            raise ValueError("All types in initial_types must be in skl2onnx.common.data_types")

        if self.final_types:
            validate_final_types = [
                True for item in self.final_types if item in inspect.getmembers(skl2onnx_data_types)
            ]
            if not all(validate_final_types):
                raise ValueError("All types in final_types must be in skl2onnx.common.data_types")


@dataclass
class ScikitLearn2ONNX(DataClassJsonMixin):
    model: sklearn.base.BaseEstimator = field(default=None)


def extract_config(t: Type[ScikitLearn2ONNX]) -> Tuple[Type[ScikitLearn2ONNX], ScikitLearn2ONNXConfig]:
    config = None

    if get_origin(t) is Annotated:
        base_type, config = get_args(t)
        if isinstance(config, ScikitLearn2ONNXConfig):
            return base_type, config
        else:
            raise TypeTransformerFailedError(f"{t}'s config isn't of type ScikitLearn2ONNXConfig")
    return t, config


def to_onnx(ctx, model, config):
    local_path = ctx.file_access.get_random_local_path()

    onx = skl2onnx.convert_sklearn(model, **config)

    with open(local_path, "wb") as f:
        f.write(onx.SerializeToString())

    return local_path


class ScikitLearn2ONNXTransformer(TypeTransformer[ScikitLearn2ONNX]):
    ONNX_FORMAT = "onnx"

    def __init__(self):
        super().__init__(name="ScikitLearn ONNX", t=ScikitLearn2ONNX)

    def get_literal_type(self, t: Type[ScikitLearn2ONNX]) -> LiteralType:
        return LiteralType(blob=BlobType(format=self.ONNX_FORMAT, dimensionality=BlobType.BlobDimensionality.SINGLE))

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: ScikitLearn2ONNX,
        python_type: Type[ScikitLearn2ONNX],
        expected: LiteralType,
    ) -> Literal:
        python_type, config = extract_config(python_type)

        if config:
            local_path = to_onnx(ctx, python_val.model, config.__dict__.copy())
            remote_path = ctx.file_access.put_raw_data(local_path)
        else:
            raise TypeTransformerFailedError(f"{python_type}'s config is None")

        return Literal(
            scalar=Scalar(
                blob=Blob(
                    uri=remote_path,
                    metadata=BlobMetadata(
                        type=BlobType(format=self.ONNX_FORMAT, dimensionality=BlobType.BlobDimensionality.SINGLE)
                    ),
                )
            )
        )

    def to_python_value(
        self,
        ctx: FlyteContext,
        lv: Literal,
        expected_python_type: Type[ONNXFile],
    ) -> ONNXFile:
        if not lv.scalar.blob.uri:
            raise TypeTransformerFailedError(f"ONNX format isn't of the expected type {expected_python_type}")

        return ONNXFile(path=lv.scalar.blob.uri)

    def guess_python_type(self, literal_type: LiteralType) -> Type[ScikitLearn2ONNX]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.ONNX_FORMAT
        ):
            return ScikitLearn2ONNX

        raise TypeTransformerFailedError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(ScikitLearn2ONNXTransformer())
