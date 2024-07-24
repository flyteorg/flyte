from typing import cast

from flyteidl.core.types_pb2 import SimpleType

from flytekit import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.types.error.error import ErrorTransformer, FlyteError


def test_error():
    err = FlyteError(message="err", failed_node_id="fn0")
    assert err.message == "err"
    assert err.failed_node_id == "fn0"


def test_error_transformer():
    err = FlyteError(message="err", failed_node_id="fn0")
    transformer = cast(ErrorTransformer, TypeEngine.get_transformer(FlyteError))
    lt_type = transformer.get_literal_type(FlyteError)
    assert lt_type.simple == SimpleType.ERROR

    pt = transformer.guess_python_type(lt_type)
    assert pt is FlyteError

    with FlyteContextManager.with_context(FlyteContextManager.current_context().new_builder()) as ctx:
        lt = transformer.to_literal(ctx, err, FlyteError, lt_type)
        assert lt.scalar.error.message == "err"
        assert lt.scalar.error.failed_node_id == "fn0"

        pv = transformer.to_python_value(ctx, lt, FlyteError)
        assert pv == err
