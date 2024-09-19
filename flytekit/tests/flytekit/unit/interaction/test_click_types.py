import json
import tempfile
import typing
from datetime import datetime, timedelta
from enum import Enum

import click
import mock
import pytest
import yaml

from flytekit import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.interaction.click_types import (
    DateTimeType,
    DurationParamType,
    FileParamType,
    FlyteLiteralConverter,
    JsonParamType,
    key_value_callback,
)

dummy_param = click.Option(["--dummy"], type=click.STRING, default="dummy")


def test_file_param():
    m = mock.MagicMock()
    l = FileParamType().convert(__file__, m, m)
    assert l.path == __file__
    r = FileParamType().convert("https://tmp/file", m, m)
    assert r.path == "https://tmp/file"


class Color(Enum):
    RED = "red"
    GREEN = "green"
    BLUE = "blue"


@pytest.mark.parametrize(
    "python_type, python_value",
    [
        (typing.Union[typing.List[int], str, Color], "flyte"),
        (typing.Union[typing.List[int], str, Color], "red"),
        (typing.Union[typing.List[int], str, Color], [1, 2, 3]),
        (typing.List[int], [1, 2, 3]),
        (typing.Dict[str, int], {"flyte": 2}),
    ],
)
def test_literal_converter(python_type, python_value):
    ctx = FlyteContextManager.current_context()
    lt = TypeEngine.to_literal_type(python_type)

    lc = FlyteLiteralConverter(
        ctx,
        literal_type=lt,
        python_type=python_type,
        is_remote=True,
    )

    click_ctx = click.Context(click.Command("test_command"), obj={"remote": True})
    assert lc.convert(click_ctx, dummy_param, python_value) == TypeEngine.to_literal(ctx, python_value, python_type, lt)


@pytest.mark.parametrize(
    "python_type, python_str_value, python_value",
    [
        (Color, "red", Color.RED),
        (int, "1", 1),
        (float, "1.0", 1.0),
        (bool, "True", True),
        (bool, "False", False),
        (str, "flyte", "flyte"),
        (typing.List[int], "[1, 2, 3]", [1, 2, 3]),
        (typing.Dict[str, int], '{"a": 1}', {"a": 1}),
        (bool, "true", True),
        (bool, "false", False),
        (bool, "TRUE", True),
        (bool, "FALSE", False),
        (bool, "t", True),
        (bool, "f", False),
        (typing.Optional[int], "1", 1),
        (typing.Union[int, str, Color], "1", 1),
        (typing.Union[Color, str], "hello", "hello"),
        (typing.Union[Color, str], "red", Color.RED),
        (typing.Union[int, str, Color], "Hello", "Hello"),
        (typing.Union[int, str, Color], "red", Color.RED),
        (typing.Union[int, str, Color, float], "1.1", 1.1),
    ],
)
def test_enum_converter(python_type: typing.Type, python_str_value: str, python_value: typing.Any):
    p = dummy_param
    pt = python_type
    click_ctx = click.Context(click.Command("test_command"), obj={"remote": True})
    ctx = FlyteContextManager.current_context()
    lt = TypeEngine.to_literal_type(pt)
    lc = FlyteLiteralConverter(ctx, literal_type=lt, python_type=pt, is_remote=True)
    assert lc.convert(click_ctx, p, lc.click_type.convert(python_str_value, p, click_ctx)) == TypeEngine.to_literal(
        ctx, python_value, pt, lt
    )


def test_duration_type():
    t = DurationParamType()
    assert t.convert(value="1 day", param=None, ctx=None) == timedelta(days=1)

    with pytest.raises(click.BadParameter):
        t.convert(None, None, None)


def test_datetime_type():
    t = DateTimeType()

    assert t.convert("2020-01-01", None, None) == datetime(2020, 1, 1)

    now = datetime.now()
    v = t.convert("now", None, None)
    assert v.day == now.day
    assert v.month == now.month


def test_json_type():
    t = JsonParamType(typing.Dict[str, str])
    assert t.convert(value='{"a": "b"}', param=None, ctx=None) == {"a": "b"}

    with pytest.raises(click.BadParameter):
        t.convert(None, None, None)

    # test that it loads a json file
    with tempfile.NamedTemporaryFile("w", delete=False) as f:
        json.dump({"a": "b"}, f)
        f.flush()
        assert t.convert(value=f.name, param=None, ctx=None) == {"a": "b"}

    # test that if the file is not a valid json, it raises an error
    with tempfile.NamedTemporaryFile("w", delete=False) as f:
        f.write("asdf")
        f.flush()
        with pytest.raises(click.BadParameter):
            t.convert(value=f.name, param=dummy_param, ctx=None)

    # test if the file does not exist
    with pytest.raises(click.BadParameter):
        t.convert(value="asdf", param=None, ctx=None)

    # test if the file is yaml and ends with .yaml it works correctly
    with tempfile.NamedTemporaryFile("w", suffix=".yaml", delete=False) as f:
        yaml.dump({"a": "b"}, f)
        f.flush()
        assert t.convert(value=f.name, param=None, ctx=None) == {"a": "b"}


def test_key_value_callback():
    """Write a test that verifies that the callback works correctly."""
    ctx = click.Context(click.Command("test_command"), obj={"remote": True})
    assert key_value_callback(ctx, "a", None) is None
    assert key_value_callback(ctx, "a", ["a=b"]) == {"a": "b"}
    assert key_value_callback(ctx, "a", ["a=b", "c=d"]) == {"a": "b", "c": "d"}
    assert key_value_callback(ctx, "a", ["a=b", "c=d", "e=f"]) == {"a": "b", "c": "d", "e": "f"}
    with pytest.raises(click.BadParameter):
        key_value_callback(ctx, "a", ["a=b", "c"])
    with pytest.raises(click.BadParameter):
        key_value_callback(ctx, "a", ["a=b", "c=d", "e"])
    with pytest.raises(click.BadParameter):
        key_value_callback(ctx, "a", ["a=b", "c=d", "e=f", "g"])
