import typing
from collections import OrderedDict

import flytekit.configuration
from flytekit import ContainerTask, Resources
from flytekit.configuration import FastSerializationSettings, Image, ImageConfig
from flytekit.core.base_task import kwtypes
from flytekit.core.launch_plan import LaunchPlan, ReferenceLaunchPlan
from flytekit.core.reference_entity import ReferenceSpec, ReferenceTemplate
from flytekit.core.task import ReferenceTask, task
from flytekit.core.workflow import ReferenceWorkflow, workflow
from flytekit.models.core import identifier as identifier_models
from flytekit.models.task import Resources as resource_model
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


def test_references():
    rlp = ReferenceLaunchPlan("media", "stg", "some.name", "cafe", inputs=kwtypes(in1=str), outputs=kwtypes())
    lp_model = get_serializable(OrderedDict(), serialization_settings, rlp)
    assert isinstance(lp_model, ReferenceSpec)
    assert isinstance(lp_model.template, ReferenceTemplate)
    assert lp_model.template.id == rlp.reference.id
    assert lp_model.template.resource_type == identifier_models.ResourceType.LAUNCH_PLAN

    rt = ReferenceTask("media", "stg", "some.name", "cafe", inputs=kwtypes(in1=str), outputs=kwtypes())
    task_spec = get_serializable(OrderedDict(), serialization_settings, rt)
    assert isinstance(task_spec, ReferenceSpec)
    assert isinstance(task_spec.template, ReferenceTemplate)
    assert task_spec.template.id == rt.reference.id
    assert task_spec.template.resource_type == identifier_models.ResourceType.TASK

    rw = ReferenceWorkflow("media", "stg", "some.name", "cafe", inputs=kwtypes(in1=str), outputs=kwtypes())
    wf_spec = get_serializable(OrderedDict(), serialization_settings, rw)
    assert isinstance(wf_spec, ReferenceSpec)
    assert isinstance(wf_spec.template, ReferenceTemplate)
    assert wf_spec.template.id == rw.reference.id
    assert wf_spec.template.resource_type == identifier_models.ResourceType.WORKFLOW


def test_basics():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a=y, b=b)
        return x, d

    wf_spec = get_serializable(OrderedDict(), serialization_settings, my_wf)
    assert len(wf_spec.template.interface.inputs) == 2
    assert len(wf_spec.template.interface.outputs) == 2
    assert len(wf_spec.template.nodes) == 2
    assert wf_spec.template.id.resource_type == identifier_models.ResourceType.WORKFLOW

    # Gets cached the first time around so it's not actually fast.
    ssettings = (
        serialization_settings.new_builder()
        .with_fast_serialization_settings(FastSerializationSettings(enabled=True))
        .build()
    )
    task_spec = get_serializable(OrderedDict(), ssettings, t1)
    assert "pyflyte-execute" in task_spec.template.container.args

    lp = LaunchPlan.create(
        "testlp",
        my_wf,
    )
    lp_model = get_serializable(OrderedDict(), serialization_settings, lp)
    assert lp_model.id.name == "testlp"


def test_fast():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    ssettings = (
        serialization_settings.new_builder()
        .with_fast_serialization_settings(FastSerializationSettings(enabled=True))
        .build()
    )
    task_spec = get_serializable(OrderedDict(), ssettings, t1)
    assert "pyflyte-fast-execute" in task_spec.template.container.args


def test_container():
    @task
    def t1(a: int) -> (int, str):
        return a + 2, str(a) + "-HELLO"

    t2 = ContainerTask(
        "raw",
        image="alpine",
        inputs=kwtypes(a=int, b=str),
        input_data_dir="/tmp",
        output_data_dir="/tmp",
        command=["cat"],
        arguments=["/tmp/a"],
        requests=Resources(mem="400Mi", cpu="1", gpu="2"),
    )

    ssettings = (
        serialization_settings.new_builder()
        .with_fast_serialization_settings(FastSerializationSettings(enabled=True))
        .build()
    )
    task_spec = get_serializable(OrderedDict(), ssettings, t2)
    assert "pyflyte" not in task_spec.template.container.args
    assert t2.get_container(ssettings).resources.requests[0].name == resource_model.ResourceName.CPU
    assert t2.get_container(ssettings).resources.requests[0].value == "1"
    assert t2.get_container(ssettings).resources.requests[1].name == resource_model.ResourceName.GPU
    assert t2.get_container(ssettings).resources.requests[1].value == "2"
    assert t2.get_container(ssettings).resources.requests[2].name == resource_model.ResourceName.MEMORY
    assert t2.get_container(ssettings).resources.requests[2].value == "400Mi"


def test_launch_plan_with_fixed_input():
    @task
    def greet(day_of_week: str, number: int, am: bool) -> str:
        greeting = "Have a great " + day_of_week + " "
        greeting += "morning" if am else "evening"
        return greeting + "!" * number

    @workflow
    def go_greet(day_of_week: str, number: int, am: bool = False) -> str:
        return greet(day_of_week=day_of_week, number=number, am=am)

    morning_greeting = LaunchPlan.create(
        "morning_greeting",
        go_greet,
        fixed_inputs={"am": True},
        default_inputs={"number": 1},
    )

    @workflow
    def morning_greeter_caller(day_of_week: str) -> str:
        greeting = morning_greeting(day_of_week=day_of_week)
        return greeting

    settings = (
        serialization_settings.new_builder()
        .with_fast_serialization_settings(FastSerializationSettings(enabled=True))
        .build()
    )
    task_spec = get_serializable(OrderedDict(), settings, morning_greeter_caller)
    assert len(task_spec.template.interface.inputs) == 1
    assert len(task_spec.template.interface.outputs) == 1
    assert len(task_spec.template.nodes) == 1
    assert len(task_spec.template.nodes[0].inputs) == 2
