from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Tuple, Type, Union

from dataclasses_json import DataClassJsonMixin
from typing_extensions import Annotated, get_args, get_origin

from flytekit import FlyteContext, lazy_module
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core.types import BlobType
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from flytekit.types.file import ONNXFile

np = lazy_module("numpy")
tf = lazy_module("tensorflow")
tf2onnx = lazy_module("tf2onnx")


@dataclass
class TensorFlow2ONNXConfig(DataClassJsonMixin):
    """
    TensorFlow2ONNXConfig is the config used during the tensorflow to ONNX conversion.

    Args:
      input_signature: The shape/dtype of the inputs to the model.
      custom_ops: If a model contains ops not recognized by ONNX runtime, you can tag these ops with a custom op domain so that the runtime can still open the model.
      target: List of workarounds applied to help certain platforms.
      custom_op_handlers: A dictionary of custom op handlers.
      custom_rewriter: A list of custom graph rewriters.
      opset: The ONNX opset number.
      extra_opset: The extra ONNX opset numbers to be used by, say, custom ops.
      shape_override: Dict with inputs that override the shapes given by tensorflow.
      inputs_as_nchw: Transpose inputs in list from nhwc to nchw.
      large_model: Whether to use the ONNX external tensor storage format.
    """

    input_signature: Union[tf.TensorSpec, np.ndarray]
    custom_ops: Optional[Dict[str, Any]] = None
    target: Optional[List[Any]] = None
    custom_op_handlers: Optional[Dict[Any, Tuple]] = None
    custom_rewriter: Optional[List[Any]] = None
    opset: Optional[int] = None
    extra_opset: Optional[List[int]] = None
    shape_override: Optional[Dict[str, List[Any]]] = None
    inputs_as_nchw: Optional[List[str]] = None
    large_model: bool = False


@dataclass
class TensorFlow2ONNX(DataClassJsonMixin):
    model: tf.keras.Model = field(default=None)


def extract_config(t: Type[TensorFlow2ONNX]) -> Tuple[Type[TensorFlow2ONNX], TensorFlow2ONNXConfig]:
    config = None
    if get_origin(t) is Annotated:
        base_type, config = get_args(t)
        if isinstance(config, TensorFlow2ONNXConfig):
            return base_type, config
        else:
            raise TypeTransformerFailedError(f"{t}'s config isn't of type TensorFlow2ONNX")
    return t, config


def to_onnx(ctx, model, config):
    local_path = ctx.file_access.get_random_local_path()

    tf2onnx.convert.from_keras(model, **config, output_path=local_path)

    return local_path


class TensorFlow2ONNXTransformer(TypeTransformer[TensorFlow2ONNX]):
    ONNX_FORMAT = "onnx"

    def __init__(self):
        super().__init__(name="TensorFlow ONNX", t=TensorFlow2ONNX)

    def get_literal_type(self, t: Type[TensorFlow2ONNX]) -> LiteralType:
        return LiteralType(blob=BlobType(format=self.ONNX_FORMAT, dimensionality=BlobType.BlobDimensionality.SINGLE))

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: TensorFlow2ONNX,
        python_type: Type[TensorFlow2ONNX],
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

    def guess_python_type(self, literal_type: LiteralType) -> Type[TensorFlow2ONNX]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.ONNX_FORMAT
        ):
            return TensorFlow2ONNX

        raise TypeTransformerFailedError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(TensorFlow2ONNXTransformer())
