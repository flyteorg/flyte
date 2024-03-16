import sys
import typing

import pytest
from typing_extensions import Annotated

from flytekit import kwtypes
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.type_engine import LiteralsResolver, TypeEngine
from flytekit.models import interface as interface_models
from flytekit.models.literals import Literal, LiteralCollection, LiteralMap, Primitive, Scalar
from flytekit.types.structured.structured_dataset import StructuredDataset


@pytest.mark.parametrize(
    "literal_value,python_type,expected_python_value",
    [
        (
            Literal(
                collection=LiteralCollection(
                    literals=[
                        Literal(scalar=Scalar(primitive=Primitive(integer=1))),
                        Literal(scalar=Scalar(primitive=Primitive(integer=2))),
                        Literal(scalar=Scalar(primitive=Primitive(integer=3))),
                    ]
                )
            ),
            typing.List[int],
            [1, 2, 3],
        ),
        (
            Literal(
                map=LiteralMap(
                    literals={
                        "k1": Literal(scalar=Scalar(primitive=Primitive(string_value="v1"))),
                        "k2": Literal(scalar=Scalar(primitive=Primitive(string_value="2"))),
                    },
                )
            ),
            typing.Dict[str, str],
            {"k1": "v1", "k2": "2"},
        ),
    ],
)
def test_literals_resolver(literal_value, python_type, expected_python_value):
    lit_dict = {"a": literal_value}

    lr = LiteralsResolver(lit_dict)
    out = lr.get("a", python_type)
    assert out == expected_python_value


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_interface():
    import pandas as pd

    ctx = FlyteContextManager.current_context()
    lt = TypeEngine.to_literal_type(pd.DataFrame)
    df = pd.DataFrame({"name": ["Tom", "Joseph"], "age": [20, 22]})

    annotated_sd_type = Annotated[StructuredDataset, kwtypes(name=str, age=int)]
    df_literal_type = TypeEngine.to_literal_type(annotated_sd_type)
    assert df_literal_type.structured_dataset_type is not None
    assert len(df_literal_type.structured_dataset_type.columns) == 2
    assert df_literal_type.structured_dataset_type.columns[0].name == "name"
    assert df_literal_type.structured_dataset_type.columns[0].literal_type.simple is not None
    assert df_literal_type.structured_dataset_type.columns[1].name == "age"
    assert df_literal_type.structured_dataset_type.columns[1].literal_type.simple is not None

    sd = annotated_sd_type(df)
    sd_literal = TypeEngine.to_literal(ctx, sd, python_type=annotated_sd_type, expected=lt)

    lm = {
        "my_map": Literal(
            map=LiteralMap(
                literals={
                    "k1": Literal(scalar=Scalar(primitive=Primitive(string_value="v1"))),
                    "k2": Literal(scalar=Scalar(primitive=Primitive(string_value="2"))),
                },
            )
        ),
        "my_list": Literal(
            collection=LiteralCollection(
                literals=[
                    Literal(scalar=Scalar(primitive=Primitive(integer=1))),
                    Literal(scalar=Scalar(primitive=Primitive(integer=2))),
                    Literal(scalar=Scalar(primitive=Primitive(integer=3))),
                ]
            )
        ),
        "val_a": Literal(scalar=Scalar(primitive=Primitive(integer=21828))),
        "my_df": sd_literal,
    }

    variable_map = {
        "my_map": interface_models.Variable(type=TypeEngine.to_literal_type(typing.Dict[str, str]), description=""),
        "my_list": interface_models.Variable(type=TypeEngine.to_literal_type(typing.List[int]), description=""),
        "val_a": interface_models.Variable(type=TypeEngine.to_literal_type(int), description=""),
        "my_df": interface_models.Variable(type=df_literal_type, description=""),
    }

    lr = LiteralsResolver(lm, variable_map=variable_map, ctx=ctx)
    assert lr._ctx is ctx

    with pytest.raises(ValueError):
        lr["not"]  # noqa

    with pytest.raises(ValueError):
        lr.get_literal("not")

    # Test that just using [] works, guessing from the Flyte type is invoked
    result = lr["my_list"]
    assert result == [1, 2, 3]

    # Test that using get works, guessing from the Flyte type is invoked
    result = lr.get("my_map")
    assert result == {
        "k1": "v1",
        "k2": "2",
    }

    # Getting the literal will return the Literal object itself
    assert lr.get_literal("my_df") is sd_literal

    guessed_df = lr["my_df"]
    # Based on guessing, so no column information
    assert len(guessed_df.metadata.structured_dataset_type.columns) == 0
    guessed_df_2 = lr["my_df"]
    assert guessed_df is guessed_df_2

    # Update type hints with the annotated type
    lr.update_type_hints({"my_df": annotated_sd_type})
    del lr._native_values["my_df"]
    guessed_df = lr.get("my_df")
    # Using the user specified type, so number of columns is correct.
    assert len(guessed_df.metadata.structured_dataset_type.columns) == 2
