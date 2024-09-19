import typing
from collections import OrderedDict

import typing_extensions

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.annotation import FlyteAnnotation
from flytekit.core.task import task
from flytekit.models.annotation import TypeAnnotation
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)
entity_mapping: OrderedDict = OrderedDict()


@task
def x(a: typing_extensions.Annotated[int, FlyteAnnotation({"foo": {"bar": 1}})], b: str):
    ...


@task
def y0(a: typing.List[typing_extensions.Annotated[int, FlyteAnnotation({"foo": {"bar": 1}})]]):
    ...


@task
def y1(a: typing_extensions.Annotated[typing.List[int], FlyteAnnotation({"foo": {"bar": 1}})]):
    ...


def test_get_variable_descriptions():
    x_tsk = get_serializable(entity_mapping, serialization_settings, x)
    x_input_vars = x_tsk.template.interface.inputs

    a_ann = x_input_vars["a"].type.annotation
    assert isinstance(a_ann, TypeAnnotation)
    assert a_ann.annotations["foo"] == {"bar": 1}

    b_ann = x_input_vars["b"].type.annotation
    assert b_ann is None

    # Annotated simple type within list generic
    y0_tsk = get_serializable(entity_mapping, serialization_settings, y0)
    y0_input_vars = y0_tsk.template.interface.inputs
    y0_a_ann = y0_input_vars["a"].type.collection_type.annotation
    assert isinstance(y0_a_ann, TypeAnnotation)
    assert y0_a_ann.annotations["foo"] == {"bar": 1}

    # Annotated list generic
    y1_tsk = get_serializable(entity_mapping, serialization_settings, y1)
    y1_input_vars = y1_tsk.template.interface.inputs
    y1_a_ann = y1_input_vars["a"].type.annotation
    assert isinstance(y1_a_ann, TypeAnnotation)
    assert y1_a_ann.annotations["foo"] == {"bar": 1}
