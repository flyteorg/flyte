from .basemodel_transformer import BaseModelTransformer
from .deserialization import set_validators_on_supported_flyte_types as _set_validators_on_supported_flyte_types

_set_validators_on_supported_flyte_types()  # enables you to use flytekit.types in pydantic model
