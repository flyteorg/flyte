"""Serializes & deserializes the pydantic basemodels """

from typing import Dict, Type

from google.protobuf import json_format
from typing_extensions import Annotated

from flytekit import FlyteContext, lazy_module
from flytekit.core import type_engine
from flytekit.models import literals, types

from . import deserialization, serialization

pydantic = lazy_module("pydantic")

BaseModelLiterals = Annotated[
    Dict[str, literals.Literal],
    """
    BaseModel serialized to a LiteralMap consisting of:
        1) the basemodel json with placeholders for flyte types
        2) mapping from placeholders to serialized flyte type values in the object store
    """,
]


class BaseModelTransformer(type_engine.TypeTransformer[pydantic.BaseModel]):
    _TYPE_INFO = types.LiteralType(simple=types.SimpleType.STRUCT)

    def __init__(self):
        """Construct pydantic.BaseModelTransformer."""
        super().__init__(name="basemodel-transform", t=pydantic.BaseModel)

    def get_literal_type(self, t: Type[pydantic.BaseModel]) -> types.LiteralType:
        return types.LiteralType(simple=types.SimpleType.STRUCT)

    def to_literal(
        self,
        ctx: FlyteContext,
        python_val: pydantic.BaseModel,
        python_type: Type[pydantic.BaseModel],
        expected: types.LiteralType,
    ) -> literals.Literal:
        """Convert a given ``pydantic.BaseModel`` to the Literal representation."""
        return serialization.serialize_basemodel(python_val)

    def to_python_value(
        self,
        ctx: FlyteContext,
        lv: literals.Literal,
        expected_python_type: Type[pydantic.BaseModel],
    ) -> pydantic.BaseModel:
        """Re-hydrate the pydantic BaseModel object from Flyte Literal value."""
        basemodel_literals: BaseModelLiterals = lv.map.literals
        basemodel_json_w_placeholders = read_basemodel_json_from_literalmap(basemodel_literals)
        with deserialization.PydanticDeserializationLiteralStore.attach(
            basemodel_literals[serialization.OBJECTS_KEY].map
        ):
            return expected_python_type.parse_raw(basemodel_json_w_placeholders)


def read_basemodel_json_from_literalmap(lv: BaseModelLiterals) -> serialization.SerializedBaseModel:
    basemodel_literal: literals.Literal = lv[serialization.BASEMODEL_JSON_KEY]
    basemodel_json_w_placeholders = json_format.MessageToJson(basemodel_literal.scalar.generic)
    assert isinstance(basemodel_json_w_placeholders, str)
    return basemodel_json_w_placeholders


type_engine.TypeEngine.register(BaseModelTransformer())
