import os
import typing
from typing import Dict, List

import pytest
from typing_extensions import Annotated  # type: ignore

from flytekit import map_task, task
from flytekit.core import context_manager
from flytekit.core.docstring import Docstring
from flytekit.core.interface import (
    extract_return_annotation,
    transform_function_to_interface,
    transform_inputs_to_parameters,
    transform_interface_to_list_interface,
    transform_interface_to_typed_interface,
    transform_variable_map,
)
from flytekit.models.core import types as _core_types
from flytekit.models.literals import Void
from flytekit.types.file import FlyteFile


def test_extract_only():
    def x() -> typing.NamedTuple("NT1", x_str=str, y_int=int):
        ...

    return_types = extract_return_annotation(typing.get_type_hints(x).get("return", None))
    assert len(return_types) == 2
    assert return_types["x_str"] == str
    assert return_types["y_int"] == int

    def t() -> List[int]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 1
    assert return_type["o0"]._name == "List"
    assert return_type["o0"].__origin__ == list

    def t() -> Dict[str, int]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 1
    assert return_type["o0"]._name == "Dict"
    assert return_type["o0"].__origin__ == dict

    def t(a: int, b: str) -> typing.Tuple[int, str]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 2
    assert return_type["o0"] == int
    assert return_type["o1"] == str

    def t(a: int, b: str) -> (int, str):
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 2
    assert return_type["o0"] == int
    assert return_type["o1"] == str

    def t(a: int, b: str) -> str:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 1
    assert return_type["o0"] == str

    def t(a: int, b: str):
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 0

    def t(a: int, b: str) -> None:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 0

    def t(a: int, b: str) -> List[int]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 1
    assert return_type["o0"] == List[int]

    def t(a: int, b: str) -> Dict[str, int]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t).get("return", None))
    assert len(return_type) == 1
    assert return_type["o0"] == Dict[str, int]


def test_named_tuples():
    nt1 = typing.NamedTuple("NT1", x_str=str, y_int=int)

    def x(a: int, b: str) -> typing.NamedTuple("NT1", x_str=str, y_int=int):
        return ("hello world", 5)

    def y(a: int, b: str) -> nt1:
        return nt1("hello world", 5)  # type: ignore

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(x).get("return", None)))
    assert result["x_str"].type.simple == 3
    assert result["y_int"].type.simple == 1

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(y).get("return", None)))
    assert result["x_str"].type.simple == 3
    assert result["y_int"].type.simple == 1


def test_unnamed_typing_tuple():
    def z(a: int, b: str) -> typing.Tuple[int, str]:
        return 5, "hello world"

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(z).get("return", None)))
    assert result["o0"].type.simple == 1
    assert result["o1"].type.simple == 3


def test_regular_tuple():
    def q(a: int, b: str) -> (int, str):
        return 5, "hello world"

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(q).get("return", None)))
    assert result["o0"].type.simple == 1
    assert result["o1"].type.simple == 3


def test_single_output_new_decorator():
    def q(a: int, b: str) -> int:
        return a + len(b)

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(q).get("return", None)))
    assert result["o0"].type.simple == 1


def test_sig_files():
    def q() -> os.PathLike:
        ...

    result = transform_variable_map(extract_return_annotation(typing.get_type_hints(q).get("return", None)))
    assert isinstance(result["o0"].type.blob, _core_types.BlobType)


def test_file_types():
    def t1() -> FlyteFile[typing.TypeVar("svg")]:
        ...

    return_type = extract_return_annotation(typing.get_type_hints(t1).get("return", None))
    assert return_type["o0"].extension() == FlyteFile[typing.TypeVar("svg")].extension()


def test_parameters_and_defaults():
    ctx = context_manager.FlyteContext.current_context()

    def z(a: int, b: str) -> typing.Tuple[int, str]:
        ...

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert params.parameters["a"].required
    assert params.parameters["a"].default is None
    assert params.parameters["b"].required
    assert params.parameters["b"].default is None

    def z(a: int, b: str = "hello") -> typing.Tuple[int, str]:
        ...

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert params.parameters["a"].required
    assert params.parameters["a"].default is None
    assert not params.parameters["b"].required
    assert params.parameters["b"].default.scalar.primitive.string_value == "hello"

    def z(a: int = 7, b: str = "eleven") -> typing.Tuple[int, str]:
        ...

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert not params.parameters["a"].required
    assert params.parameters["a"].default.scalar.primitive.integer == 7
    assert not params.parameters["b"].required
    assert params.parameters["b"].default.scalar.primitive.string_value == "eleven"

    def z(a: Annotated[int, "some annotation"]) -> Annotated[int, "some annotation"]:
        return a

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert params.parameters["a"].required
    assert params.parameters["a"].default is None
    assert our_interface.inputs == {"a": Annotated[int, "some annotation"]}
    assert our_interface.outputs == {"o0": Annotated[int, "some annotation"]}

    def z(
        a: typing.Optional[int] = None, b: typing.Optional[str] = None, c: typing.Union[typing.List[int], None] = None
    ) -> typing.Tuple[int, str]:
        ...

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert not params.parameters["a"].required
    assert params.parameters["a"].default.scalar.none_type == Void()
    assert not params.parameters["b"].required
    assert params.parameters["b"].default.scalar.none_type == Void()
    assert not params.parameters["c"].required
    assert params.parameters["c"].default.scalar.none_type == Void()


def test_parameters_with_docstring():
    ctx = context_manager.FlyteContext.current_context()

    def z(a: int, b: str) -> typing.Tuple[int, str]:
        """
        function z

        :param a: foo
        :param b: bar
        :return: ramen
        """
        ...

    our_interface = transform_function_to_interface(z, Docstring(callable_=z))
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert params.parameters["a"].var.description == "foo"
    assert params.parameters["b"].var.description == "bar"


def test_transform_interface_to_typed_interface_with_docstring():
    # sphinx style
    def z(a: int, b: str) -> typing.Tuple[int, str]:
        """
        function z

        :param a: foo
        :param b: bar
        :return: ramen
        """
        ...

    our_interface = transform_function_to_interface(z, Docstring(callable_=z))
    typed_interface = transform_interface_to_typed_interface(our_interface)
    assert typed_interface.inputs.get("a").description == "foo"
    assert typed_interface.inputs.get("b").description == "bar"
    assert typed_interface.outputs.get("o1").description == "ramen"

    # numpy style, multiple return values, shared descriptions
    def z(a: int, b: str) -> typing.Tuple[int, str]:
        """
        function z

        Parameters
        ----------
        a : int
            foo
        b : str
            bar

        Returns
        -------
        out1, out2 : tuple
            ramen
        """
        ...

    our_interface = transform_function_to_interface(z, Docstring(callable_=z))
    typed_interface = transform_interface_to_typed_interface(our_interface)
    assert typed_interface.inputs.get("a").description == "foo"
    assert typed_interface.inputs.get("b").description == "bar"
    assert typed_interface.outputs.get("o0").description == "ramen"
    assert typed_interface.outputs.get("o1").description == "ramen"

    # numpy style, multiple return values, named
    def z(a: int, b: str) -> typing.NamedTuple("NT", x_str=str, y_int=int):
        """
        function z

        Parameters
        ----------
        a : int
            foo
        b : str
            bar

        Returns
        -------
        x_str : str
            description for x_str
        y_int : int
            description for y_int
        """
        ...

    our_interface = transform_function_to_interface(z, Docstring(callable_=z))
    typed_interface = transform_interface_to_typed_interface(our_interface)
    assert typed_interface.inputs.get("a").description == "foo"
    assert typed_interface.inputs.get("b").description == "bar"
    assert typed_interface.outputs.get("x_str").description == "description for x_str"
    assert typed_interface.outputs.get("y_int").description == "description for y_int"


def test_parameter_change_to_pickle_type():
    ctx = context_manager.FlyteContext.current_context()

    class Foo:
        def __init__(self, name):
            self.name = name

    def z(a: Foo) -> Foo:
        ...

    our_interface = transform_function_to_interface(z)
    params = transform_inputs_to_parameters(ctx, our_interface)
    assert params.parameters["a"].required
    assert params.parameters["a"].default is None
    assert our_interface.outputs["o0"] == Foo
    assert our_interface.inputs["a"] == Foo


def test_doc_string():
    @task
    def t1(a: int) -> int:
        """Set the temperature value.

        The value of the temp parameter is stored as a value in
        the class variable temperature.
        """
        return a

    assert t1.docs.short_description == "Set the temperature value."
    assert (
        t1.docs.long_description.value
        == "The value of the temp parameter is stored as a value in\nthe class variable temperature."
    )


@pytest.mark.parametrize(
    "optional_outputs, expected_type",
    [
        (False, int),
        (True, typing.Optional[int]),
    ],
)
def test_transform_interface_to_list_interface(optional_outputs, expected_type):
    @task
    def t() -> int:
        ...

    list_interface = transform_interface_to_list_interface(t.python_interface, set(), optional_outputs=optional_outputs)
    assert list_interface.outputs["o0"] == typing.List[expected_type]


@pytest.mark.parametrize(
    "min_success_ratio, expected_type",
    [
        (None, str),
        (0, typing.Optional[str]),
        (1, str),
        (0.42, typing.Optional[str]),
        (0.5, typing.Optional[str]),
        (0.999999, typing.Optional[str]),
    ],
)
def test_map_task_interface(min_success_ratio, expected_type):
    @task
    def t() -> str:
        ...

    mt = map_task(t, min_success_ratio=min_success_ratio)

    assert mt.python_interface.outputs["o0"] == typing.List[expected_type]
