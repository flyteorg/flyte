import pytest

from flytekit.models import interface, types
from tests.flytekit.common.parameterizers import LIST_OF_ALL_LITERAL_TYPES


@pytest.mark.parametrize("literal_type", LIST_OF_ALL_LITERAL_TYPES)
def test_variable_type(literal_type):
    var = interface.Variable(type=literal_type, description="abc")
    assert var.type == literal_type
    assert var.description == "abc"
    assert var == interface.Variable.from_flyte_idl(var.to_flyte_idl())


@pytest.mark.parametrize("literal_type", LIST_OF_ALL_LITERAL_TYPES)
def test_typed_interface(literal_type):
    typed_interface = interface.TypedInterface(
        {"a": interface.Variable(literal_type, "description1")},
        {"b": interface.Variable(literal_type, "description2"), "c": interface.Variable(literal_type, "description3")},
    )

    assert typed_interface.inputs["a"].type == literal_type
    assert typed_interface.outputs["b"].type == literal_type
    assert typed_interface.outputs["c"].type == literal_type
    assert typed_interface.inputs["a"].description == "description1"
    assert typed_interface.outputs["b"].description == "description2"
    assert typed_interface.outputs["c"].description == "description3"
    assert len(typed_interface.inputs) == 1
    assert len(typed_interface.outputs) == 2

    pb = typed_interface.to_flyte_idl()
    deserialized_typed_interface = interface.TypedInterface.from_flyte_idl(pb)
    assert typed_interface == deserialized_typed_interface

    assert deserialized_typed_interface.inputs["a"].type == literal_type
    assert deserialized_typed_interface.outputs["b"].type == literal_type
    assert deserialized_typed_interface.outputs["c"].type == literal_type
    assert deserialized_typed_interface.inputs["a"].description == "description1"
    assert deserialized_typed_interface.outputs["b"].description == "description2"
    assert deserialized_typed_interface.outputs["c"].description == "description3"
    assert len(deserialized_typed_interface.inputs) == 1
    assert len(deserialized_typed_interface.outputs) == 2


def test_parameter():
    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    obj = interface.Parameter(var=v)
    assert obj.var == v

    obj2 = interface.Parameter.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.var == v


def test_parameter_map():
    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    p = interface.Parameter(var=v)

    obj = interface.ParameterMap({"ppp": p})
    obj2 = interface.ParameterMap.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_variable_map():
    v = interface.Variable(types.LiteralType(simple=types.SimpleType.BOOLEAN), "asdf asdf asdf")
    obj = interface.VariableMap({"vvv": v})

    obj2 = interface.VariableMap.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
