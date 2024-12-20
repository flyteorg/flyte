"""
Logic for serializing a basemodel to a literalmap that can be passed between flyte tasks.

The serialization process is as follows:

1. Serialize the basemodel to json, replacing all flyte types with unique placeholder strings
2. Serialize the flyte types to separate literals and store them in the flyte object store (a singleton object)
3. Return a literal map with the json and the flyte object store represented as a literalmap {placeholder: flyte type}

"""
import uuid
from typing import Any, Dict, Union, cast

from google.protobuf import json_format, struct_pb2
from typing_extensions import Annotated

from flytekit import lazy_module
from flytekit.core import context_manager, type_engine
from flytekit.models import literals

from . import commons

pydantic = lazy_module("pydantic")

BASEMODEL_JSON_KEY = "BaseModel JSON"
OBJECTS_KEY = "Serialized Flyte Objects"

SerializedBaseModel = Annotated[str, "A pydantic BaseModel that has been serialized with placeholders for Flyte types."]

ObjectStoreID = Annotated[str, "Key for unique literalmap of a serialized basemodel."]
LiteralObjID = Annotated[str, "Key for unique object in literal map."]
LiteralStore = Annotated[Dict[LiteralObjID, literals.Literal], "uid to literals for a serialized BaseModel"]


class BaseModelFlyteObjectStore:
    """
    This class is an intermediate store for python objects that are being serialized/deserialized.

    On serialization of a basemodel, flyte objects are serialized and stored in this object store.
    """

    def __init__(self) -> None:
        self.literal_store: LiteralStore = dict()

    def register_python_object(self, python_object: object) -> LiteralObjID:
        """Serialize to literal and return a unique identifier."""
        serialized_item = serialize_to_flyte_literal(python_object)
        identifier = make_identifier_for_serializeable(python_object)
        assert identifier not in self.literal_store
        self.literal_store[identifier] = serialized_item
        return identifier

    def to_literal(self) -> literals.Literal:
        """Convert the object store to a literal map."""
        return literals.Literal(map=literals.LiteralMap(literals=self.literal_store))


def serialize_basemodel(basemodel: pydantic.BaseModel) -> literals.Literal:
    """
    Serializes a given pydantic BaseModel instance into a LiteralMap.
    The BaseModel is first serialized into a JSON format, where all Flyte types are replaced with unique placeholder strings.
    The Flyte Types are serialized into separate Flyte literals
    """
    store = BaseModelFlyteObjectStore()
    basemodel_literal = serialize_basemodel_to_literal(basemodel, store)
    basemodel_literalmap = literals.LiteralMap(
        {
            BASEMODEL_JSON_KEY: basemodel_literal,  # json with flyte types replaced with placeholders
            OBJECTS_KEY: store.to_literal(),  # flyte type-engine serialized types
        }
    )
    literal = literals.Literal(map=basemodel_literalmap)  # type: ignore
    return literal


def serialize_basemodel_to_literal(
    basemodel: pydantic.BaseModel,
    flyteobject_store: BaseModelFlyteObjectStore,
) -> literals.Literal:
    """
    Serialize a pydantic BaseModel to json and protobuf, separating out the Flyte types into a separate store.
    On deserialization, the store is used to reconstruct the Flyte types.
    """

    def encoder(obj: Any) -> Union[str, commons.LiteralObjID]:
        if isinstance(obj, commons.PYDANTIC_SUPPORTED_FLYTE_TYPES):
            return flyteobject_store.register_python_object(obj)
        return basemodel.__json_encoder__(obj)

    basemodel_json = basemodel.json(encoder=encoder)
    return make_literal_from_json(basemodel_json)


def serialize_to_flyte_literal(python_obj: object) -> literals.Literal:
    """
    Use the Flyte TypeEngine to serialize a python object to a Flyte Literal.
    """
    python_type = type(python_obj)
    ctx = context_manager.FlyteContextManager().current_context()
    literal_type = type_engine.TypeEngine.to_literal_type(python_type)
    literal_obj = type_engine.TypeEngine.to_literal(ctx, python_obj, python_type, literal_type)
    return literal_obj


def make_literal_from_json(json: str) -> literals.Literal:
    """
    Converts the json representation of a pydantic BaseModel to a Flyte Literal.
    """
    return literals.Literal(scalar=literals.Scalar(generic=json_format.Parse(json, struct_pb2.Struct())))  # type: ignore


def make_identifier_for_serializeable(python_type: object) -> LiteralObjID:
    """
    Create a unique identifier for a python object.
    """
    unique_id = f"{type(python_type).__name__}_{uuid.uuid4().hex}"
    return cast(LiteralObjID, unique_id)
