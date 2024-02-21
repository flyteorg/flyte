import typing
from collections import OrderedDict
from datetime import timedelta
from io import StringIO

from mock import patch

from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.core.condition import conditional
from flytekit.core.context_manager import ExecutionState, FlyteContextManager
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.gate import approve, sleep, wait_for_input
from flytekit.core.task import task
from flytekit.core.type_engine import TypeEngine
from flytekit.core.workflow import workflow
from flytekit.remote.entities import FlyteWorkflow
from flytekit.tools.translator import gather_dependent_entities, get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


def test_basic_sleep():
    @task
    def t1(a: int) -> int:
        return a + 5

    @workflow
    def wf_sleep() -> int:
        x = sleep(timedelta(seconds=10))
        b = t1(a=5)
        x >> b
        return b

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wf_sleep)
    assert len(wf_spec.template.nodes) == 2
    wf_spec.template.nodes[0].gate_node is not None
    wf_spec.template.nodes[0].gate_node.sleep.duration == timedelta(seconds=10)
    wf_spec.template.nodes[1].upstream_node_ids == ["n0"]


def test_basic_signal():
    @task
    def t1(a: int) -> int:
        return a + 5

    @task
    def t2(a: int) -> int:
        return a + 6

    @workflow
    def wf(a: int) -> typing.Tuple[int, int, int]:
        x = t1(a=a)
        s1 = wait_for_input("my-signal-name", timeout=timedelta(hours=1), expected_type=bool)
        s2 = wait_for_input("my-signal-name-2", timeout=timedelta(hours=2), expected_type=int)
        z = t1(a=5)
        y = t2(a=s2)
        q = t2(a=approve(y, "approvalfory", timeout=timedelta(hours=2)))
        x >> s1
        s1 >> z

        return y, z, q

    with patch("sys.stdin", StringIO("y\n3\ny\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        res = wf(a=5)
        assert res == (9, 10, 15)
        assert stdin.read() == ""  # all input consumed

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wf)
    assert len(wf_spec.template.nodes) == 7
    # The first t1 call
    assert wf_spec.template.nodes[0].task_node is not None

    # The first signal s1, dependent on the first t1 call
    assert wf_spec.template.nodes[1].upstream_node_ids == ["n0"]
    assert wf_spec.template.nodes[1].gate_node is not None
    assert wf_spec.template.nodes[1].gate_node.signal.signal_id == "my-signal-name"
    assert wf_spec.template.nodes[1].gate_node.signal.type.simple == 4
    assert wf_spec.template.nodes[1].gate_node.signal.output_variable_name == "o0"

    # The second signal
    assert wf_spec.template.nodes[2].upstream_node_ids == []
    assert wf_spec.template.nodes[2].gate_node is not None
    assert wf_spec.template.nodes[2].gate_node.signal.signal_id == "my-signal-name-2"
    assert wf_spec.template.nodes[2].gate_node.signal.type.simple == 1
    assert wf_spec.template.nodes[2].gate_node.signal.output_variable_name == "o0"

    # The second call to t1, dependent on the first signal
    assert wf_spec.template.nodes[3].upstream_node_ids == ["n1"]
    assert wf_spec.template.nodes[3].task_node is not None

    # The call to t2, dependent on the second signal
    assert wf_spec.template.nodes[4].upstream_node_ids == ["n2"]
    assert wf_spec.template.nodes[4].task_node is not None

    # Approval node
    assert wf_spec.template.nodes[5].gate_node is not None
    assert wf_spec.template.nodes[5].gate_node.approve is not None
    assert wf_spec.template.nodes[5].upstream_node_ids == ["n4"]
    assert len(wf_spec.template.nodes[5].inputs) == 1
    assert wf_spec.template.nodes[5].inputs[0].binding.promise.node_id == "n4"
    assert wf_spec.template.nodes[5].inputs[0].binding.promise.var == "o0"
    assert wf_spec.template.nodes[6].inputs[0].binding.promise.node_id == "n5"
    assert wf_spec.template.nodes[6].inputs[0].binding.promise.var == "o0"

    assert wf_spec.template.outputs[0].binding.promise.node_id == "n4"
    assert wf_spec.template.outputs[1].binding.promise.node_id == "n3"
    assert wf_spec.template.outputs[2].binding.promise.node_id == "n6"


def test_dyn_signal():
    @task
    def t1(a: int) -> int:
        return a + 5

    @task
    def t2(a: int) -> int:
        return a + 6

    @dynamic
    def dyn(a: int) -> typing.Tuple[int, int, int]:
        x = t1(a=a)
        s1 = wait_for_input("my-signal-name", timeout=timedelta(hours=1), expected_type=bool)
        s2 = wait_for_input("my-signal-name-2", timeout=timedelta(hours=2), expected_type=int)
        z = t1(a=5)
        y = t2(a=s2)
        q = t2(a=approve(y, "approvalfory", timeout=timedelta(hours=2)))
        x >> s1
        s1 >> z

        return y, z, q

    @workflow
    def wf_dyn(a: int) -> typing.Tuple[int, int, int]:
        y, z, q = dyn(a=a)
        return y, z, q

    with patch("sys.stdin", StringIO("y\n3\ny\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        res = wf_dyn(a=5)
        assert res == (9, 10, 15)
        assert stdin.read() == ""  # all input consumed

    wf_spec = get_serializable(OrderedDict(), serialization_settings, wf_dyn)
    assert len(wf_spec.template.nodes) == 1
    # The first t1 call
    assert wf_spec.template.nodes[0].task_node is not None

    with FlyteContextManager.with_context(
        FlyteContextManager.current_context().with_serialization_settings(serialization_settings)
    ) as ctx:
        with FlyteContextManager.with_context(
            ctx.with_execution_state(
                ctx.execution_state.with_params(
                    mode=ExecutionState.Mode.TASK_EXECUTION,
                )
            )
        ) as ctx:
            input_literal_map = TypeEngine.dict_to_literal_map(ctx, {"a": 50})
            dynamic_job_spec = dyn.dispatch_execute(ctx, input_literal_map)
            print(dynamic_job_spec)

        assert dynamic_job_spec.nodes[1].upstream_node_ids == ["dn0"]
        assert dynamic_job_spec.nodes[1].gate_node is not None
        assert dynamic_job_spec.nodes[1].gate_node.signal.signal_id == "my-signal-name"
        assert dynamic_job_spec.nodes[1].gate_node.signal.type.simple == 4
        assert dynamic_job_spec.nodes[1].gate_node.signal.output_variable_name == "o0"

        assert dynamic_job_spec.nodes[2].upstream_node_ids == []
        assert dynamic_job_spec.nodes[2].gate_node is not None
        assert dynamic_job_spec.nodes[2].gate_node.signal.signal_id == "my-signal-name-2"
        assert dynamic_job_spec.nodes[2].gate_node.signal.type.simple == 1
        assert dynamic_job_spec.nodes[2].gate_node.signal.output_variable_name == "o0"

        assert dynamic_job_spec.nodes[5].gate_node is not None
        assert dynamic_job_spec.nodes[5].gate_node.approve is not None
        assert dynamic_job_spec.nodes[5].upstream_node_ids == ["dn4"]
        assert len(dynamic_job_spec.nodes[5].inputs) == 1
        assert dynamic_job_spec.nodes[5].inputs[0].binding.promise.node_id == "dn4"
        assert dynamic_job_spec.nodes[5].inputs[0].binding.promise.var == "o0"
        assert dynamic_job_spec.nodes[6].inputs[0].binding.promise.node_id == "dn5"
        assert dynamic_job_spec.nodes[6].inputs[0].binding.promise.var == "o0"


def test_dyn_signal_no_approve():
    @task
    def t1(a: int) -> int:
        return a + 5

    @task
    def t2(a: int) -> int:
        return a + 6

    @dynamic
    def dyn(a: int) -> typing.Tuple[int, int]:
        x = t1(a=a)
        s1 = wait_for_input("my-signal-name", timeout=timedelta(hours=1), expected_type=bool)
        s2 = wait_for_input("my-signal-name-2", timeout=timedelta(hours=2), expected_type=int)
        z = t1(a=5)
        y = t2(a=s2)
        x >> s1
        s1 >> z

        return y, z

    @workflow
    def wf_dyn(a: int) -> typing.Tuple[int, int]:
        y, z = dyn(a=a)
        return y, z

    with patch("sys.stdin", StringIO("y\n3\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        wf_dyn(a=5)
        assert stdin.read() == ""  # all input consumed


def test_subwf():
    nt = typing.NamedTuple("Multi", [("named1", int), ("named2", int)])

    @task
    def nt1(a: int) -> nt:
        a = a + 2
        return nt(a, a)

    @workflow
    def subwf(a: int) -> nt:
        return nt1(a=a)

    @workflow
    def parent_wf(b: int) -> nt:
        out = subwf(a=b)
        return nt1(a=approve(out.named1, "subwf approve", timeout=timedelta(hours=2)))

    with patch("sys.stdin", StringIO("y\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        x = parent_wf(b=3)
        assert stdin.read() == ""  # all input consumed
        assert x == (7, 7)


def test_cond():
    @task
    def five() -> int:
        return 5

    @task
    def square(n: float) -> float:
        return n * n

    @task
    def double(n: float) -> float:
        return 2 * n

    @workflow
    def cond_wf() -> float:
        f = five()
        # Because approve itself produces a node, call approve outside of the conditional.
        app = approve(f, "jfdkl", timeout=timedelta(hours=2))
        return conditional("fractions").if_(app == 5).then(double(n=f)).else_().then(square(n=f))

    with patch("sys.stdin", StringIO("y\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        x = cond_wf()
        assert x == 10.0
        assert stdin.read() == ""


def test_cond_wait():
    @task
    def square(n: float) -> float:
        return n * n

    @task
    def double(n: float) -> float:
        return 2 * n

    @workflow
    def cond_wf(a: int) -> float:
        # Because approve itself produces a node, call approve outside of the conditional.
        input_1 = wait_for_input("top-input", timeout=timedelta(hours=1), expected_type=int)
        return conditional("fractions").if_(input_1 >= 5).then(double(n=a)).else_().then(square(n=a))

    with patch("sys.stdin", StringIO("3\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        x = cond_wf(a=3)
        assert x == 9
        assert stdin.read() == ""

    with patch("sys.stdin", StringIO("8\n")) as stdin, patch("sys.stdout", new_callable=StringIO):
        x = cond_wf(a=3)
        assert x == 6
        assert stdin.read() == ""


def test_promote():
    @task
    def t1(a: int) -> int:
        return a + 5

    @task
    def t2(a: int) -> int:
        return a + 6

    @workflow
    def wf(a: int) -> typing.Tuple[int, int, int]:
        zzz = sleep(timedelta(seconds=10))
        x = t1(a=a)
        s1 = wait_for_input("my-signal-name", timeout=timedelta(hours=1), expected_type=bool)
        s2 = wait_for_input("my-signal-name-2", timeout=timedelta(hours=2), expected_type=int)
        z = t1(a=5)
        y = t2(a=s2)
        q = t2(a=approve(y, "approvalfory", timeout=timedelta(hours=2)))
        zzz >> x
        x >> s1
        s1 >> z

        return y, z, q

    entries = OrderedDict()
    wf_spec = get_serializable(entries, serialization_settings, wf)
    tts, wf_specs, lp_specs = gather_dependent_entities(entries)

    fwf = FlyteWorkflow.promote_from_model(wf_spec.template, tasks=tts)
    assert fwf.template.nodes[2].gate_node is not None
