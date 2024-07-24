import typing
from collections import OrderedDict

import pytest

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core import context_manager
from flytekit.core.base_task import kwtypes
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.launch_plan import LaunchPlan, reference_launch_plan
from flytekit.core.promise import VoidPromise
from flytekit.core.reference import get_reference_entity
from flytekit.core.reference_entity import ReferenceEntity, ReferenceSpec, ReferenceTemplate, TaskReference
from flytekit.core.task import reference_task, task
from flytekit.core.testing import patch, task_mock
from flytekit.core.workflow import reference_workflow, workflow
from flytekit.models.core import identifier as _identifier_model
from flytekit.tools.translator import get_serializable


# This is used for docs
def test_ref_docs():
    # docs_ref_start
    ref_entity = get_reference_entity(
        _identifier_model.ResourceType.WORKFLOW,
        "project",
        "dev",
        "my.other.workflow",
        "abc123",
        inputs=kwtypes(a=str, b=int),
        outputs={},
    )
    # docs_ref_end

    with pytest.raises(Exception) as e:
        ref_entity()
    assert "You must mock this out" in f"{e}"


@reference_task(
    project="flytesnacks",
    domain="development",
    name="recipes.aaa.simple.join_strings",
    version="553018f39e519bdb2597b652639c30ce16b99c79",
)
def ref_t1(a: typing.List[str]) -> str:
    """
    The empty function acts as a convenient skeleton to make it intuitive to call/reference this task from workflows.
    The interface of the task must match that of the remote task. Otherwise, remote compilation of the workflow will
    fail.
    """
    ...


def test_ref():
    assert ref_t1.id.project == "flytesnacks"
    assert ref_t1.id.domain == "development"
    assert ref_t1.id.name == "recipes.aaa.simple.join_strings"
    assert ref_t1.id.version == "553018f39e519bdb2597b652639c30ce16b99c79"

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )
    spec = get_serializable(OrderedDict(), serialization_settings, ref_t1)
    assert isinstance(spec, ReferenceSpec)
    assert isinstance(spec.template, ReferenceTemplate)
    assert spec.template.id == ref_t1.id
    assert spec.template.resource_type == _identifier_model.ResourceType.TASK


def test_ref_task_more():
    @reference_task(
        project="flytesnacks",
        domain="development",
        name="recipes.aaa.simple.join_strings",
        version="553018f39e519bdb2597b652639c30ce16b99c79",
    )
    def ref_t1(a: typing.List[str]) -> str:
        ...

    @workflow
    def wf1(in1: typing.List[str]) -> str:
        return ref_t1(a=in1)

    with pytest.raises(Exception) as e:
        wf1(in1=["hello", "world"])
    assert "You must mock this out" in f"{e}"

    with task_mock(ref_t1) as mock:
        mock.return_value = "hello"
        assert wf1(in1=["hello", "world"]) == "hello"


def test_ref_task_more_2():
    # Test that >> will not break the mock.

    @reference_task(
        project="flytesnacks",
        domain="development",
        name="recipes.aaa.simple.join_strings",
        version="553018f39e519bdb2597b652639c30ce16b99c79",
    )
    def ref_t1(a: typing.List[str]) -> str:
        ...

    @reference_task(
        project="flytesnacks",
        domain="development",
        name="recipes.aaa.simple.join_string_second",
        version="553018f39e519bdb2597b652639c30ce16b99c79",
    )
    def ref_t2(a: typing.List[str]) -> str:
        ...

    @workflow
    def wf1(in1: typing.List[str]) -> str:
        x = ref_t1(a=in1)
        y = ref_t2(a=in1)
        y >> x
        return x

    with task_mock(ref_t1) as mock_x:
        with task_mock(ref_t2) as mock_y:
            mock_y.return_value = "ignored"
            mock_x.return_value = "hello"
            assert wf1(in1=["hello", "world"]) == "hello"


@reference_workflow(project="proj", domain="development", name="wf_name", version="abc")
def ref_wf1(a: int) -> typing.Tuple[str, str]:
    ...


def test_reference_workflow():
    @task
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def my_wf(a: int, b: str) -> (int, str, str):
        x, y = t1(a=a).with_overrides()
        u, v = ref_wf1(a=x)
        return x, u, v

    with pytest.raises(Exception):
        my_wf(a=3, b="foo")

    @patch(ref_wf1)
    def inner_test(ref_mock):
        ref_mock.return_value = ("hello", "alice")
        x, y, z = my_wf(a=3, b="foo")
        assert x == 5
        assert y == "hello"
        assert z == "alice"

    inner_test()

    # Ensure that the patching is only for the duration of that test
    with pytest.raises(Exception):
        my_wf(a=3, b="foo")


def test_ref_plain_no_outputs():
    r1 = ReferenceEntity(
        TaskReference("proj", "domain", "some.name", "abc"),
        inputs=kwtypes(a=str, b=int),
        outputs={},
    )

    # Reference entities should always raise an exception when not mocked out.
    with pytest.raises(Exception) as e:
        r1(a="fdsa", b=3)
    assert "You must mock this out" in f"{e}"

    @workflow
    def wf1(a: str, b: int):
        r1(a=a, b=b)

    @patch(r1)
    def inner_test(ref_mock):
        ref_mock.return_value = None
        x = wf1(a="fdsa", b=3)
        assert x is None

    inner_test()

    nt1 = typing.NamedTuple("DummyNamedTuple", t1_int_output=int, c=str)

    @task
    def t1(a: int) -> nt1:
        a = a + 2
        return nt1(a, "world-" + str(a))  # type: ignore

    @workflow
    def wf2(a: int):
        t1_int, c = t1(a=a)
        r1(a=c, b=t1_int)

    @patch(r1)
    def inner_test2(ref_mock):
        ref_mock.return_value = None
        x = wf2(a=3)
        assert x is None
        ref_mock.assert_called_with(a="world-5", b=5)

    inner_test2()

    # Test nodes
    node_r1 = wf2._nodes[1]
    assert node_r1._upstream_nodes[0] is wf2._nodes[0]


def test_ref_plain_two_outputs():
    r1 = ReferenceEntity(
        TaskReference("proj", "domain", "some.name", "abc"),
        inputs=kwtypes(a=str, b=int),
        outputs=kwtypes(x=bool, y=int),
    )

    ctx = context_manager.FlyteContext.current_context()
    with context_manager.FlyteContextManager.with_context(ctx.with_new_compilation_state()):
        xx, yy = r1(a="five", b=6)
        # Note - misnomer, these are not SdkNodes, they are core.Nodes
        assert xx.ref.node is yy.ref.node
        assert xx.var == "x"
        assert yy.var == "y"
        assert xx.ref.node_id == "n0"
        assert len(xx.ref.node.bindings) == 2

    @task
    def t2(q: bool, r: int) -> str:
        return f"q: {q} r: {r}"

    @workflow
    def wf1(a: str, b: int) -> str:
        x_out, y_out = r1(a=a, b=b)
        return t2(q=x_out, r=y_out)

    @patch(r1)
    def inner_test(ref_mock):
        ref_mock.return_value = (False, 30)
        x = wf1(a="hello", b=10)
        assert x == "q: False r: 30"

    inner_test()


@pytest.mark.parametrize(
    "resource_type",
    [_identifier_model.ResourceType.LAUNCH_PLAN, _identifier_model.ResourceType.TASK],
)
def test_lps(resource_type):
    ref_entity = get_reference_entity(
        resource_type,
        "proj",
        "dom",
        "app.other.flyte_entity",
        "123",
        inputs=kwtypes(a=str, b=int),
        outputs={},
    )

    ctx = context_manager.FlyteContext.current_context()
    with pytest.raises(Exception) as e:
        ref_entity()
    assert "You must mock this out" in f"{e}"

    with context_manager.FlyteContextManager.with_context(ctx.with_new_compilation_state()) as ctx:
        with pytest.raises(Exception) as e:
            ref_entity()
        assert "was not specified for function" in f"{e}"

        output = ref_entity(a="hello", b=3)
        assert isinstance(output, VoidPromise)

    @workflow
    def wf1(a: str, b: int):
        ref_entity(a=a, b=b)

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )
    wf_spec = get_serializable(OrderedDict(), serialization_settings, wf1)
    assert len(wf_spec.template.interface.inputs) == 2
    assert len(wf_spec.template.interface.outputs) == 0
    assert len(wf_spec.template.nodes) == 1
    if resource_type == _identifier_model.ResourceType.LAUNCH_PLAN:
        assert wf_spec.template.nodes[0].workflow_node.launchplan_ref.project == "proj"
        assert wf_spec.template.nodes[0].workflow_node.launchplan_ref.name == "app.other.flyte_entity"
    else:
        assert wf_spec.template.nodes[0].task_node.reference_id.project == "proj"
        assert wf_spec.template.nodes[0].task_node.reference_id.name == "app.other.flyte_entity"


def test_ref_sub_wf():
    ref_entity = get_reference_entity(
        _identifier_model.ResourceType.WORKFLOW,
        "proj",
        "dom",
        "app.other.sub_wf",
        "123",
        inputs=kwtypes(a=str, b=int),
        outputs={},
    )

    ctx = context_manager.FlyteContext.current_context()
    with pytest.raises(Exception) as e:
        ref_entity()
    assert "You must mock this out" in f"{e}"

    with context_manager.FlyteContextManager.with_context(ctx.with_new_compilation_state()) as ctx:
        with pytest.raises(Exception) as e:
            ref_entity()
        assert "was not specified for function" in f"{e}"

        output = ref_entity(a="hello", b=3)
        assert isinstance(output, VoidPromise)

    @workflow
    def wf1(a: str, b: int):
        ref_entity(a=a, b=b)

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )
    with pytest.raises(Exception, match="currently unsupported"):
        # Subworkflow as references don't work (probably ever). The reason is because we'd need to make a network call
        # to admin to get the structure of the subworkflow and the whole point of reference entities is that there
        # is no network call.
        get_serializable(OrderedDict(), serialization_settings, wf1)


def test_lp_with_output():
    ref_lp = get_reference_entity(
        _identifier_model.ResourceType.LAUNCH_PLAN,
        "proj",
        "dom",
        "app.other.flyte_entity",
        "123",
        inputs=kwtypes(a=str, b=int),
        outputs=kwtypes(x=bool, y=int),
    )

    @task
    def t1() -> (str, int):
        return "hello", 88

    @task
    def t2(q: bool, r: int) -> str:
        return f"q: {q} r: {r}"

    @workflow
    def wf1() -> str:
        t1_str, t1_int = t1()
        x_out, y_out = ref_lp(a=t1_str, b=t1_int)
        return t2(q=x_out, r=y_out)

    @patch(ref_lp)
    def inner_test(ref_mock):
        ref_mock.return_value = (False, 30)
        x = wf1()
        assert x == "q: False r: 30"
        ref_mock.assert_called_with(a="hello", b=88)

    inner_test()

    serialization_settings = flytekit.configuration.SerializationSettings(
        project="test_proj",
        domain="test_domain",
        version="abc",
        image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
        env={},
    )
    wf_spec = get_serializable(OrderedDict(), serialization_settings, wf1)
    assert wf_spec.template.nodes[1].workflow_node.launchplan_ref.project == "proj"
    assert wf_spec.template.nodes[1].workflow_node.launchplan_ref.name == "app.other.flyte_entity"


def test_lp_from_ref_wf():
    @reference_workflow(project="project", domain="domain", name="name", version="version")
    def ref_wf1(p1: str, p2: str) -> None:
        ...

    lp = LaunchPlan.create("reference-wf-12345", ref_wf1, fixed_inputs={"p1": "p1-value", "p2": "p2-value"})
    assert lp.name == "reference-wf-12345"
    assert lp.workflow == ref_wf1
    assert lp.workflow.id.name == "name"
    assert lp.workflow.id.project == "project"
    assert lp.workflow.id.domain == "domain"
    assert lp.workflow.id.version == "version"


def test_ref_lp_from_decorator():
    @reference_launch_plan(project="project", domain="domain", name="name", version="version")
    def ref_lp1(p1: str, p2: str) -> int:
        ...

    assert ref_lp1.id.name == "name"
    assert ref_lp1.id.project == "project"
    assert ref_lp1.id.domain == "domain"
    assert ref_lp1.id.version == "version"
    assert ref_lp1.python_interface.inputs == {"p1": str, "p2": str}
    assert ref_lp1.python_interface.outputs == {"o0": int}


def test_ref_lp_from_decorator_with_named_outputs():
    @reference_launch_plan(project="project", domain="domain", name="name", version="version")
    def ref_lp1(p1: str, p2: str) -> typing.NamedTuple("RefLPOutput", o1=int, o2=str):
        ...

    assert ref_lp1.python_interface.outputs == {"o1": int, "o2": str}


def test_ref_dynamic_task():
    @reference_task(
        project="flytesnacks",
        domain="development",
        name="sample.reference.task",
        version="553018f39e519bdb2597b652639c30ce16b99c79",
    )
    def ref_t1(a: int) -> str:
        ...

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @dynamic
    def my_subwf(a: int) -> typing.List[str]:
        s = []
        for i in range(a):
            s.append(ref_t1(a=i))
        return s

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            flytekit.configuration.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        new_exc_state = ctx.execution_state.with_params(mode=context_manager.ExecutionState.Mode.TASK_EXECUTION)
        with context_manager.FlyteContextManager.with_context(ctx.with_execution_state(new_exc_state)) as ctx:
            with pytest.raises(Exception, match="currently unsupported"):
                my_subwf.compile_into_workflow(ctx, my_subwf._task_function, a=5)


def test_ref_dynamic_lp():
    @dynamic
    def my_subwf(a: int) -> typing.List[int]:
        @reference_launch_plan(project="project", domain="domain", name="name", version="version")
        def ref_lp1(p1: str, p2: str) -> int:
            ...

        s = []
        for i in range(a):
            s.append(ref_lp1(p1="hello", p2=str(a)))
        return s

    with context_manager.FlyteContextManager.with_context(
        context_manager.FlyteContextManager.current_context().with_serialization_settings(
            flytekit.configuration.SerializationSettings(
                project="test_proj",
                domain="test_domain",
                version="abc",
                image_config=ImageConfig(Image(name="name", fqn="image", tag="name")),
                env={},
            )
        )
    ) as ctx:
        new_exc_state = ctx.execution_state.with_params(mode=context_manager.ExecutionState.Mode.TASK_EXECUTION)
        with context_manager.FlyteContextManager.with_context(ctx.with_execution_state(new_exc_state)) as ctx:
            djspec = my_subwf.compile_into_workflow(ctx, my_subwf._task_function, a=5)
            assert len(djspec.nodes) == 5
