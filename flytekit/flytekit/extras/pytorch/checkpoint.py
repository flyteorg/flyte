import pathlib
import typing
from dataclasses import asdict, dataclass, fields, is_dataclass
from typing import Any, Callable, Dict, NamedTuple, Optional, Type, Union

import torch
from dataclasses_json import DataClassJsonMixin
from typing_extensions import Protocol

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType


class IsDataclass(Protocol):
    __dataclass_fields__: Dict
    __dataclass_params__: Dict
    __post_init__: Optional[Callable]


@dataclass
class PyTorchCheckpoint(DataClassJsonMixin):
    """
    This class is helpful to save a checkpoint.
    """

    module: Optional[torch.nn.Module] = None
    hyperparameters: Optional[Union[Dict[str, Any], NamedTuple, IsDataclass]] = None
    optimizer: Optional[torch.optim.Optimizer] = None

    def __post_init__(self):
        if not (
            isinstance(self.hyperparameters, dict)
            or (is_dataclass(self.hyperparameters) and not isinstance(self.hyperparameters, type))
            or (isinstance(self.hyperparameters, tuple) and hasattr(self.hyperparameters, "_fields"))
            or (self.hyperparameters is None)
        ):
            raise TypeTransformerFailedError(
                f"hyperparameters must be a dict, dataclass, or NamedTuple. Got {type(self.hyperparameters)}"
            )

        if not (self.module or self.hyperparameters or self.optimizer):
            raise TypeTransformerFailedError("Must have at least one of module, hyperparameters, or optimizer")


class PyTorchCheckpointTransformer(TypeTransformer[PyTorchCheckpoint]):
    """
    TypeTransformer that supports serializing and deserializing checkpoint.
    """

    PYTORCH_CHECKPOINT_FORMAT = "PyTorchCheckpoint"

    def __init__(self):
        super().__init__(name="PyTorch Checkpoint", t=PyTorchCheckpoint)

    def get_literal_type(self, t: Type[PyTorchCheckpoint]) -> LiteralType:
        return LiteralType(
            blob=_core_types.BlobType(
                format=self.PYTORCH_CHECKPOINT_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: PyTorchCheckpoint,
        python_type: Type[PyTorchCheckpoint],
        expected: LiteralType,
    ) -> Literal:
        meta = BlobMetadata(
            type=_core_types.BlobType(
                format=self.PYTORCH_CHECKPOINT_FORMAT, dimensionality=_core_types.BlobType.BlobDimensionality.SINGLE
            )
        )

        local_path = ctx.file_access.get_random_local_path() + ".pt"
        pathlib.Path(local_path).parent.mkdir(parents=True, exist_ok=True)

        to_save = {}
        for field in fields(python_val):
            value = getattr(python_val, field.name)

            if value and field.name in ["module", "optimizer"]:
                to_save[field.name + "_state_dict"] = getattr(value, "state_dict")()
            elif value and field.name == "hyperparameters":
                if isinstance(value, dict):
                    to_save.update(value)
                elif isinstance(value, tuple):
                    to_save.update(value._asdict())
                elif is_dataclass(value):
                    to_save.update(asdict(value))

        if not to_save:
            raise TypeTransformerFailedError(f"Cannot save empty {python_val}")

        # save checkpoint to a file
        torch.save(to_save, local_path)

        remote_path = ctx.file_access.put_raw_data(local_path)
        return Literal(scalar=Scalar(blob=Blob(metadata=meta, uri=remote_path)))

    def to_python_value(
        self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[PyTorchCheckpoint]
    ) -> PyTorchCheckpoint:
        try:
            uri = lv.scalar.blob.uri
        except AttributeError:
            TypeTransformerFailedError(f"Cannot convert from {lv} to {expected_python_type}")

        local_path = ctx.file_access.get_random_local_path()
        ctx.file_access.get_data(uri, local_path, is_multipart=False)

        # cpu <-> gpu conversion
        if torch.cuda.is_available():
            map_location = "cuda:0"
        else:
            map_location = torch.device("cpu")

        # load checkpoint from a file
        return typing.cast(PyTorchCheckpoint, torch.load(local_path, map_location=map_location))

    def guess_python_type(self, literal_type: LiteralType) -> Type[PyTorchCheckpoint]:
        if (
            literal_type.blob is not None
            and literal_type.blob.dimensionality == _core_types.BlobType.BlobDimensionality.SINGLE
            and literal_type.blob.format == self.PYTORCH_CHECKPOINT_FORMAT
        ):
            return PyTorchCheckpoint

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(PyTorchCheckpointTransformer())
