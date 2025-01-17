import typing
from collections import OrderedDict

import pytest

from flytekit import dynamic
from flytekit.configuration import FastSerializationSettings, Image, ImageConfig, SerializationSettings
from flytekit.core import context_manager
from flytekit.core.context_manager import ExecutionState
from flytekit.core.launch_plan import LaunchPlan
from flytekit.core.task import task
from flytekit.core.type_engine import TypeEngine
from flytekit.core.workflow import workflow
from flytekit.exceptions.user import FlyteAssertion
from flytekit.models.admin.workflow import WorkflowSpec
from flytekit.models.task import TaskSpec
from flytekit.remote import FlyteLaunchPlan, FlyteTask, FlyteWorkflow
from flytekit.remote.interface import TypedInterface
from flytekit.tools.translator import gather_dependent_entities, get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


@task
def t1(a: int) -> int:
    return a + 2


@task
def t2(a: int, b: str) -> str:
    return b + str(a)


@workflow
def sub_wf(a: int, b: str) -> (int, str):
    x = t1(a=a)
    d = t2(a=x, b=b)
    return x, d


t1_spec = get_serializable(OrderedDict(), serialization_settings, t1)
ft = FlyteTask.promote_from_model(t1_spec.template)


def test_fetched_task():
    @workflow
    def wf(a: int) -> int:
        return ft(a=a).with_overrides(node_name="foobar")

    # Should not work unless mocked out.
    with pytest.raises(Exception, match="cannot be run locally"):
        wf(a=3)

    # Should have one task template
    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, wf)
    vals = [v for v in serialized.values()]
    tts = [f for f in filter(lambda x: isinstance(x, TaskSpec), vals)]
    assert len(tts) == 1
    assert wf_spec.template.nodes[0].id == "foobar"
    assert wf_spec.template.outputs[0].binding.promise.node_id == "foobar"


def test_misnamed():
    with pytest.raises(FlyteAssertion):

        @workflow
        def wf(a: int) -> int:
            return ft(b=a)

        wf()


def test_calling_lp():
    sub_wf_lp = LaunchPlan.get_or_create(sub_wf)
    serialized = OrderedDict()
    lp_model = get_serializable(serialized, serialization_settings, sub_wf_lp)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)
    for wf_id, spec in wf_specs.items():
        break

    remote_lp = FlyteLaunchPlan.promote_from_model(lp_model.id, lp_model.spec)
    # To pretend that we've fetched this launch plan from Admin, also fill in the Flyte interface, which isn't
    # part of the IDL object but is something FlyteRemote does
    remote_lp._interface = TypedInterface.promote_from_model(spec.template.interface)
    serialized = OrderedDict()

    @workflow
    def wf2(a: int) -> typing.Tuple[int, str]:
        return remote_lp(a=a, b="hello")

    wf_spec = get_serializable(serialized, serialization_settings, wf2)
    print(wf_spec.template.nodes[0].workflow_node.launchplan_ref)
    assert wf_spec.template.nodes[0].workflow_node.launchplan_ref == lp_model.id


def test_dynamic():
    @dynamic
    def my_subwf(a: int) -> typing.List[int]:
        s = []
        for i in range(a):
            s.append(ft(a=i))
        return s

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            context_manager.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
                fast_serialization_settings=FastSerializationSettings(enabled=True),
            )
        )
    ) as ctx:
        with context_manager.FlyteContextManager.with_context(
            ctx.with_execution_state(
                ctx.execution_state.with_params(
                    mode=ExecutionState.Mode.TASK_EXECUTION,
                )
            )
        ) as ctx:
            input_literal_map = TypeEngine.dict_to_literal_map(ctx, {"a": 2})
            # Test that it works
            dynamic_job_spec = my_subwf.dispatch_execute(ctx, input_literal_map)
            assert len(dynamic_job_spec._nodes) == 2
            assert len(dynamic_job_spec.tasks) == 1
            assert dynamic_job_spec.tasks[0].id == ft.id

            # Test that the fast execute stuff does not get applied because the commands of tasks fetched from
            # Admin should never change.
            args = " ".join(dynamic_job_spec.tasks[0].container.args)
            assert not args.startswith("pyflyte-fast-execute")


def test_calling_wf():
    # No way to fetch from Admin in unit tests so we serialize and then promote back
    serialized = OrderedDict()
    wf_spec: WorkflowSpec = get_serializable(serialized, serialization_settings, sub_wf)
    task_templates, wf_specs, lp_specs = gather_dependent_entities(serialized)
    fwf = FlyteWorkflow.promote_from_model(
        wf_spec.template, tasks={k: FlyteTask.promote_from_model(t) for k, t in task_templates.items()}
    )

    @workflow
    def parent_1(a: int, b: str) -> typing.Tuple[int, str]:
        y = t1(a=a)
        return fwf(a=y, b=b)

    # No way to fetch from Admin in unit tests so we serialize and then promote back
    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, parent_1)
    # Get task_specs from the second one, merge with the first one. Admin normally would be the one to do this.
    task_templates_p1, wf_specs, lp_specs = gather_dependent_entities(serialized)
    for k, v in task_templates.items():
        task_templates_p1[k] = v

    # Pick out the subworkflow templates from the ordereddict. We can't use the output of the gather_dependent_entities
    # function because that only looks for WorkflowSpecs
    subwf_templates = {
        x.template.id: x.template for x in list(filter(lambda x: isinstance(x, WorkflowSpec), serialized.values()))
    }
    fwf_p1 = FlyteWorkflow.promote_from_model(
        wf_spec.template,
        sub_workflows=subwf_templates,
        tasks={k: FlyteTask.promote_from_model(t) for k, t in task_templates_p1.items()},
    )

    @workflow
    def parent_2(a: int, b: str) -> typing.Tuple[int, str]:
        x, y = fwf_p1(a=a, b=b)
        z = t1(a=x)
        return z, y

    serialized = OrderedDict()
    wf_spec = get_serializable(serialized, serialization_settings, parent_2)
    # Make sure both were picked up.
    assert len(wf_spec.sub_workflows) == 2
