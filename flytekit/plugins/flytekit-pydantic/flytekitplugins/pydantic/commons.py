import builtins
import datetime
import typing
from typing import Set

from typing_extensions import Annotated

from flytekit import lazy_module
from flytekit.core import type_engine

numpy = lazy_module("numpy")
pyarrow = lazy_module("pyarrow")

MODULES_TO_EXCLUDE_FROM_FLYTE_TYPES: Set[str] = {m.__name__ for m in [builtins, typing, datetime, pyarrow, numpy]}


def include_in_flyte_types(t: type) -> bool:
    if t is None:
        return False
    object_module = t.__module__
    if any(object_module.startswith(module) for module in MODULES_TO_EXCLUDE_FROM_FLYTE_TYPES):
        return False
    return True


type_engine.TypeEngine.lazy_import_transformers()  # loads all transformers
PYDANTIC_SUPPORTED_FLYTE_TYPES = tuple(
    filter(include_in_flyte_types, type_engine.TypeEngine.get_available_transformers())
)

# this is the UUID placeholder that is set in the serialized basemodel JSON, connecting that field to
# the literal map that holds the actual object that needs to be deserialized (w/ protobuf)
LiteralObjID = Annotated[str, "Key for unique object in literal map."]
