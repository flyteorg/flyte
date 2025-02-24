from dataclasses import dataclass
from typing import Type, TypeVar

from mashumaro.mixins.json import DataClassJSONMixin

from flytekit.core.context_manager import FlyteContext
from flytekit.core.type_engine import TypeEngine, TypeTransformer, TypeTransformerFailedError
from flytekit.models import types as _type_models
from flytekit.models.literals import Error, Literal, Scalar
from flytekit.models.types import LiteralType

T = TypeVar("T")


@dataclass
class FlyteError(DataClassJSONMixin):
    """
    Special Task type that will be used in the failure node. Propeller will pass this error to failure task, so users
    have to add an input with this type to the failure task.
    """

    message: str
    failed_node_id: str


class ErrorTransformer(TypeTransformer[FlyteError]):
    """
    Enables converting a python type FlyteError to LiteralType.Error
    """

    def __init__(self):
        super().__init__(name="FlyteError", t=FlyteError)

    def get_literal_type(self, t: Type[T]) -> LiteralType:
        return LiteralType(simple=_type_models.SimpleType.ERROR)

    def to_literal(
        self, ctx: FlyteContext, python_val: FlyteError, python_type: Type[T], expected: LiteralType
    ) -> Literal:
        if type(python_val) != FlyteError:
            raise TypeTransformerFailedError(
                f"Expected value of type {FlyteError} but got '{python_val}' of type {type(python_val)}"
            )
        return Literal(scalar=Scalar(error=Error(message=python_val.message, failed_node_id=python_val.failed_node_id)))

    def to_python_value(self, ctx: FlyteContext, lv: Literal, expected_python_type: Type[T]) -> T:
        if not (lv and lv.scalar and lv.scalar.error is not None):
            raise TypeTransformerFailedError("Can only convert a generic literal to FlyteError")
        return FlyteError(message=lv.scalar.error.message, failed_node_id=lv.scalar.error.failed_node_id)

    def guess_python_type(self, literal_type: LiteralType) -> Type[FlyteError]:
        if literal_type.simple and literal_type.simple == _type_models.SimpleType.ERROR:
            return FlyteError

        raise ValueError(f"Transformer {self} cannot reverse {literal_type}")


TypeEngine.register(ErrorTransformer())
