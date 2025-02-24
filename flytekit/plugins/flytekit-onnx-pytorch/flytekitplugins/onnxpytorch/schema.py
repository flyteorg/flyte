from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple, Type, Union

from dataclasses_json import DataClassJsonMixin
from typing_extensions import Annotated, get_args, get_origin

from flytekit import FlyteContext, lazy_module
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core.types import BlobType
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from flytekit.types.file import ONNXFile

torch = lazy_module("torch")


@dataclass
class PyTorch2ONNXConfig(DataClassJsonMixin):
    """
    PyTorch2ONNXConfig is the config used during the pytorch to ONNX conversion.

    Args:
      args: The input to the model.
      export_params: Whether to export all the parameters.
      verbose: Whether to print description of the ONNX model.
      training: Whether to export the model in training mode or inference mode.
      opset_version: The ONNX version to export the model to.
      input_names: Names to assign to the input nodes of the graph.
      output_names: Names to assign to the output nodes of the graph.
      operator_export_type: How to export the ops.
      do_constant_folding: Whether to apply constant folding for optimization.
      dynamic_axes: Specify axes of tensors as dynamic.
      keep_initializers_as_inputs: Whether to add the initializers as inputs to the graph.
      custom_opsets: A dictionary of opset domain name and version.
      export_modules_as_functions: Whether to export modules as functions.
    """

    args: Union[Tuple, torch.Tensor]
    export_params: bool = True
    verbose: bool = False
    training: torch.onnx.TrainingMode = torch.onnx.TrainingMode.EVAL
    opset_version: int = 9
    input_names: List[str] = field(default_factory=list)
    output_names: List[str] = field(default_factory=list)
    operator_export_type: Optional[torch.onnx.OperatorExportTypes] = None
    do_constant_folding: bool = False
    dynamic_axes: Union[Dict[str, Dict[int, str]], Dict[str, List[int]]] = field(default_factory=dict)
    keep_initializers_as_inputs: Optional[bool] = None
    custom_opsets: Dict[str, int] = field(default_factory=dict)
    export_modules_as_functions: Union[bool, set[Type]] = False


@dataclass
class PyTorch2ONNX(DataClassJsonMixin):
    model: Union[torch.nn.Module, torch.jit.ScriptModule, torch.jit.ScriptFunction] = field(default=None)


def extract_config(t: Type[PyTorch2ONNX]) -> Tuple[Type[PyTorch2ONNX], PyTorch2ONNXConfig]:
    config = None
    if get_origin(t) is Annotated:
        base_type, config = get_args(t)
        if isinstance(config, PyTorch2ONNXConfig):
            return base_type, config
        else:
            raise TypeTransformerFailedError(f"{t}'s config isn't of type PyTorch2ONNXConfig")
    return t, config


def to_onnx(ctx, model, config):
    local_path = ctx.file_access.get_random_local_path()

    torch.onnx.export(
        model,
        **config,
        f=local_path,
    )

    return local_path


class PyTorch2ONNXTransformer(TypeTransformer[PyTorch2ONNX]):
    ONNX_FORMAT = "onnx"

    def __init__(self):
        super().__init__(name="PyTorch ONNX", t=PyTorch2ONNX)

    def get_literal_type(self, t: Type[PyTorch2ONNX]) -> LiteralType:
        return LiteralType(blob=BlobType(format=self.ONNX_FORMAT, dimensionality=BlobType.BlobDimensionality.SINGLE))

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: PyTorch2ONNX,
        python_type: Type[PyTorch2ONNX],
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
        if not (lv.scalar.blob.uri and lv.scalar.blob.metadata.format == self.ONNX_FORMAT):
            raise TypeTransformerFailedError(f"ONNX format isn't of the expected type {expected_python_type}")

        return ONNXFile(path=lv.scalar.blob.uri)

    def guess_python_type(self, literal_type: LiteralType) -> Type[PyTorch2ONNX]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.ONNX_FORMAT
        ):
            return PyTorch2ONNX

        raise TypeTransformerFailedError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(PyTorch2ONNXTransformer())
