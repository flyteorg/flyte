import typing
from collections import OrderedDict

import pytest

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.condition import conditional
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.remote import FlyteTask, FlyteWorkflow
from flytekit.tools.translator import gather_dependent_entities, get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


@pytest.mark.skip(reason="branch nodes don't work yet.")
def test_wf_cond():
    @task
    def t1(a: int) -> int:
        return a + 2

    @workflow
    def my_sub_wf(a: int) -> int:
        return t1(a=a)

    @workflow
    def my_wf(a: int) -> int:
        d = conditional("test1").if_(a > 3).then(t1(a=a)).else_().then(my_sub_wf(a=a))
        return d

    get_serializable(OrderedDict(), serialization_settings, my_wf)


def test_wf_promote_subwf_lps():
    @task
    def t1(a: int) -> int:
        a = a + 2
        return a

    @workflow
    def subwf(a: int) -> int:
        return t1(a=a)

    sub_lp = LaunchPlan.get_or_create(subwf)

    @workflow
    def wf(b: int) -> int:
        return subwf(a=b)

    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, wf)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)
    sub_wf_dict = {s.id: s for s in wf_spec.sub_workflows}

    fwf = FlyteWorkflow.promote_from_model(
        wf_spec.template,
        sub_workflows=sub_wf_dict,
        node_launch_plans=lp_specs,
        tasks={k: FlyteTask.promote_from_model(t) for k, t in task_templates.items()},
    )
    assert len(fwf.outputs) == 1
    assert list(fwf.interface.inputs.keys()) == ["b"]
    assert len(fwf.nodes) == 1
    assert len(fwf.flyte_nodes) == 1

    # Test another subwf that calls a launch plan instead of the sub_wf directly
    @workflow
    def wf2(b: int) -> int:
        return sub_lp(a=b)

    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, wf2)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)

    fwf = FlyteWorkflow.promote_from_model(
        wf_spec.template,
        sub_workflows={},
        node_launch_plans=lp_specs,
        tasks={k: FlyteTask.promote_from_model(t) for k, t in task_templates.items()},
    )
    assert len(fwf.outputs) == 1
    assert list(fwf.interface.inputs.keys()) == ["b"]
    assert len(fwf.nodes) == 1
    assert len(fwf.flyte_nodes) == 1
    # The resource type will be different, so just check the name
    assert fwf.nodes[0].workflow_node.launchplan_ref.name == list(lp_specs.values())[0].workflow_id.name


def test_upstream():
    @task
    def t1(a: int) -> typing.Dict[str, str]:
        return {"a": str(a)}

    @task
    def t2(a: typing.Dict[str, str]) -> str:
        return " ".join([v for k, v in a.items()])

    @task
    def t3() -> str:
        return "hello"

    @workflow
    def my_wf(a: int) -> str:
        return t2(a=t1(a=a))

    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, my_wf)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)

    fwf = FlyteWorkflow.promote_from_model(
        wf_spec.template,
        sub_workflows={},
        node_launch_plans={},
        tasks={k: FlyteTask.promote_from_model(t) for k, t in task_templates.items()},
    )

    assert len(fwf.flyte_nodes[0].upstream_nodes) == 0
    assert len(fwf.flyte_nodes[1].upstream_nodes) == 1

    @workflow
    def parent(a: int) -> (str, str):
        first = my_wf(a=a)
        second = t3()
        return first, second

    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, parent)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)
    sub_wf_dict = {s.id: s for s in wf_spec.sub_workflows}

    fwf = FlyteWorkflow.promote_from_model(
        wf_spec.template,
        sub_workflows=sub_wf_dict,
        node_launch_plans={},
        tasks={k: FlyteTask.promote_from_model(v) for k, v in task_templates.items()},
    )
    # Test upstream nodes don't get confused by subworkflows
    assert len(fwf.flyte_nodes[0].upstream_nodes) == 0
    assert len(fwf.flyte_nodes[1].upstream_nodes) == 0
