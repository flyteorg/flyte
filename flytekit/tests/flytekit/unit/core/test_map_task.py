import functools
import typing
from collections import OrderedDict

import pytest

import flytekit.configuration
from flytekit import LaunchPlan, Resources, map_task
from flytekit.configuration import Image, ImageConfig
from flytekit.core.map_task import MapPythonTask, MapTaskResolver
from flytekit.core.task import TaskMetadata, task
from flytekit.core.workflow import workflow
from flytekit.tools.translator import get_serializable


@pytest.fixture
def serialization_settings():
    default_img = Image(name="default", fqn="test", tag="tag")
    return flytekit.configuration.SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env=None,
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
    )


@task
def t1(a: int) -> typing.Optional[str]:
    b = a + 2
    return str(b)


@task(cache=True, cache_version="1")
def t2(a: int) -> str:
    b = a + 2
    return str(b)


@task(cache=True, cache_version="1")
def t3(a: int, b: str, c: float) -> str:
    pass


# This test is for documentation.
def test_map_docs():
    # test_map_task_start
    @task
    def my_mappable_task(a: int) -> typing.Optional[str]:
        return str(a)

    @workflow
    def my_wf(x: typing.List[int]) -> typing.List[typing.Optional[str]]:
        return map_task(
            my_mappable_task,
            metadata=TaskMetadata(retries=1),
            concurrency=10,
            min_success_ratio=0.75,
        )(a=x).with_overrides(requests=Resources(cpu="10M"))

    # test_map_task_end

    res = my_wf(x=[1, 2, 3])
    assert res == ["1", "2", "3"]


def test_map_task_types():
    strs = map_task(t1, metadata=TaskMetadata(retries=1))(a=[5, 6])
    assert strs == ["7", "8"]

    with pytest.raises(TypeError):
        _ = map_task(t1, metadata=TaskMetadata(retries=1))(a=1)

    with pytest.raises(TypeError):
        _ = map_task(t1, metadata=TaskMetadata(retries=1))(a=["invalid", "args"])


def test_serialization(serialization_settings):
    maptask = map_task(t1, metadata=TaskMetadata(retries=1))
    task_spec = get_serializable(OrderedDict(), serialization_settings, maptask)

    # By default all map_task tasks will have their custom fields set.
    assert task_spec.template.custom["minSuccessRatio"] == 1.0
    assert task_spec.template.type == "container_array"
    assert task_spec.template.task_type_version == 1
    assert task_spec.template.container.args == [
        "pyflyte-map-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--checkpoint-path",
        "{{.checkpointOutputPrefix}}",
        "--prev-checkpoint",
        "{{.prevCheckpointPrefix}}",
        "--resolver",
        "MapTaskResolver",
        "--",
        "vars",
        "",
        "resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "task-module",
        "tests.flytekit.unit.core.test_map_task",
        "task-name",
        "t1",
    ]


@pytest.mark.parametrize(
    "custom_fields_dict, expected_custom_fields",
    [
        ({}, {"minSuccessRatio": 1.0}),
        ({"concurrency": 99}, {"parallelism": "99", "minSuccessRatio": 1.0}),
        ({"min_success_ratio": 0.271828}, {"minSuccessRatio": 0.271828}),
        ({"concurrency": 42, "min_success_ratio": 0.31415}, {"parallelism": "42", "minSuccessRatio": 0.31415}),
    ],
)
def test_serialization_of_custom_fields(custom_fields_dict, expected_custom_fields, serialization_settings):
    maptask = map_task(t1, **custom_fields_dict)
    task_spec = get_serializable(OrderedDict(), serialization_settings, maptask)

    assert task_spec.template.custom == expected_custom_fields


def test_serialization_workflow_def(serialization_settings):
    @task
    def complex_task(a: int) -> str:
        b = a + 2
        return str(b)

    maptask = map_task(complex_task, metadata=TaskMetadata(retries=1))

    @workflow
    def w1(a: typing.List[int]) -> typing.List[str]:
        return maptask(a=a)

    @workflow
    def w2(a: typing.List[int]) -> typing.List[str]:
        return map_task(complex_task, metadata=TaskMetadata(retries=2))(a=a)

    serialized_control_plane_entities = OrderedDict()
    wf1_spec = get_serializable(serialized_control_plane_entities, serialization_settings, w1)
    assert wf1_spec.template is not None
    assert len(wf1_spec.template.nodes) == 1

    wf2_spec = get_serializable(serialized_control_plane_entities, serialization_settings, w2)
    assert wf2_spec.template is not None
    assert len(wf2_spec.template.nodes) == 1

    flyte_entities = list(serialized_control_plane_entities.keys())

    tasks_seen = []
    for entity in flyte_entities:
        if isinstance(entity, MapPythonTask) and "complex" in entity.name:
            tasks_seen.append(entity)

    assert len(tasks_seen) == 2
    print(tasks_seen[0])


def test_map_tasks_only():
    @workflow
    def wf1(a: int):
        print(f"{a}")

    with pytest.raises(ValueError):

        @workflow
        def wf2(a: typing.List[int]):
            return map_task(wf1)(a=a)

        wf2()

    lp = LaunchPlan.create("test", wf1)

    with pytest.raises(ValueError):

        @workflow
        def wf3(a: typing.List[int]):
            return map_task(lp)(a=a)

        wf3()


def test_inputs_outputs_length():
    @task
    def many_inputs(a: int, b: str, c: float) -> str:
        return f"{a} - {b} - {c}"

    m = map_task(many_inputs)
    assert m.python_interface.inputs == {"a": typing.List[int], "b": typing.List[str], "c": typing.List[float]}
    assert m.name == "tests.flytekit.unit.core.test_map_task.map_many_inputs_d41d8cd98f00b204e9800998ecf8427e"
    r_m = MapPythonTask(many_inputs)
    assert str(r_m.python_interface) == str(m.python_interface)

    p1 = functools.partial(many_inputs, c=1.0)
    m = map_task(p1)
    assert m.python_interface.inputs == {"a": typing.List[int], "b": typing.List[str], "c": float}
    assert m.name == "tests.flytekit.unit.core.test_map_task.map_many_inputs_4a8a08f09d37b73795649038408b5f33"
    r_m = MapPythonTask(many_inputs, bound_inputs=set("c"))
    assert str(r_m.python_interface) == str(m.python_interface)

    p2 = functools.partial(p1, b="hello")
    m = map_task(p2)
    assert m.python_interface.inputs == {"a": typing.List[int], "b": str, "c": float}
    assert m.name == "tests.flytekit.unit.core.test_map_task.map_many_inputs_74aefa13d6ab8e4bfbd241583749dfe8"
    r_m = MapPythonTask(many_inputs, bound_inputs={"c", "b"})
    assert str(r_m.python_interface) == str(m.python_interface)

    p3 = functools.partial(p2, a=1)
    m = map_task(p3)
    assert m.python_interface.inputs == {"a": int, "b": str, "c": float}
    assert m.name == "tests.flytekit.unit.core.test_map_task.map_many_inputs_a44c56c8177e32d3613988f4dba7962e"
    r_m = MapPythonTask(many_inputs, bound_inputs={"a", "c", "b"})
    assert str(r_m.python_interface) == str(m.python_interface)

    p3_1 = functools.partial(p2, a=1)
    m_1 = map_task(p3_1)
    assert m_1.python_interface.inputs == {"a": int, "b": str, "c": float}
    assert m_1.name == m.name

    with pytest.raises(TypeError):
        m(a=[1, 2, 3])

    @task
    def many_outputs(a: int) -> (int, str):
        return a, f"{a}"

    with pytest.raises(ValueError):
        _ = map_task(many_outputs)


def test_map_task_metadata():
    map_meta = TaskMetadata(retries=1)
    mapped_1 = map_task(t2, metadata=map_meta)
    assert mapped_1.metadata is map_meta
    mapped_2 = map_task(t2)
    assert mapped_2.metadata is t2.metadata


def test_map_task_resolver(serialization_settings):
    list_outputs = {"o0": typing.List[str]}
    mt = map_task(t3)
    assert mt.python_interface.inputs == {"a": typing.List[int], "b": typing.List[str], "c": typing.List[float]}
    assert mt.python_interface.outputs == list_outputs
    mtr = MapTaskResolver()
    assert mtr.name() == "MapTaskResolver"
    args = mtr.loader_args(serialization_settings, mt)
    t = mtr.load_task(loader_args=args)
    assert t.python_interface.inputs == mt.python_interface.inputs
    assert t.python_interface.outputs == mt.python_interface.outputs

    mt = map_task(functools.partial(t3, b="hello", c=1.0))
    assert mt.python_interface.inputs == {"a": typing.List[int], "b": str, "c": float}
    assert mt.python_interface.outputs == list_outputs
    mtr = MapTaskResolver()
    args = mtr.loader_args(serialization_settings, mt)
    t = mtr.load_task(loader_args=args)
    assert t.python_interface.inputs == mt.python_interface.inputs
    assert t.python_interface.outputs == mt.python_interface.outputs

    mt = map_task(functools.partial(t3, b="hello"))
    assert mt.python_interface.inputs == {"a": typing.List[int], "b": str, "c": typing.List[float]}
    assert mt.python_interface.outputs == list_outputs
    mtr = MapTaskResolver()
    args = mtr.loader_args(serialization_settings, mt)
    t = mtr.load_task(loader_args=args)
    assert t.python_interface.inputs == mt.python_interface.inputs
    assert t.python_interface.outputs == mt.python_interface.outputs


@pytest.mark.parametrize(
    "min_success_ratio, type_t",
    [
        (None, int),
        (1, int),
        (0.5, typing.Optional[int]),
    ],
)
def test_map_task_min_success_ratio(min_success_ratio, type_t):
    @task
    def some_task1(inputs: int) -> int:
        return inputs

    @workflow
    def my_wf1() -> typing.List[type_t]:
        return map_task(some_task1, min_success_ratio=min_success_ratio)(inputs=[1, 2, 3, 4])

    my_wf1()


def test_map_task_parameter_order():
    @task()
    def task1(a: int, b: float, c: str) -> str:
        return f"{a} - {b} - {c}"

    @task()
    def task2(b: float, c: str, a: int) -> str:
        return f"{a} - {b} - {c}"

    @task()
    def task3(c: str, a: int, b: float) -> str:
        return f"{a} - {b} - {c}"

    param_a = [1, 2, 3]
    param_b = [0.1, 0.2, 0.3]
    param_c = "c"

    m1 = map_task(functools.partial(task1, c=param_c))(a=param_a, b=param_b)
    m2 = map_task(functools.partial(task2, c=param_c))(a=param_a, b=param_b)
    m3 = map_task(functools.partial(task3, c=param_c))(a=param_a, b=param_b)

    assert m1 == m2 == m3 == ["1 - 0.1 - c", "2 - 0.2 - c", "3 - 0.3 - c"]


@pytest.mark.parametrize(
    "min_success_ratio, should_raise_error",
    [
        (None, True),
        (1, True),
        (0.75, False),
        (0.5, False),
    ],
)
def test_raw_execute_with_min_success_ratio(min_success_ratio, should_raise_error):
    @task
    def some_task1(inputs: int) -> int:
        if inputs == 2:
            raise ValueError("Unexpected inputs: 2")
        return inputs

    @workflow
    def my_wf1() -> typing.List[typing.Optional[int]]:
        return map_task(some_task1, min_success_ratio=min_success_ratio)(inputs=[1, 2, 3, 4])

    if should_raise_error:
        with pytest.raises(ValueError):
            my_wf1()
    else:
        assert my_wf1() == [1, None, 3, 4]


def test_map_task_override(serialization_settings):
    @task
    def my_mappable_task(a: int) -> typing.Optional[str]:
        return str(a)

    @workflow
    def wf(x: typing.List[int]):
        map_task(my_mappable_task)(a=x).with_overrides(container_image="random:image")

    assert wf.nodes[0].flyte_entity.run_task.container_image == "random:image"


def test_bounded_inputs_vars_order(serialization_settings):
    mt = map_task(functools.partial(t3, c=1.0, b="hello", a=1))
    mtr = MapTaskResolver()
    args = mtr.loader_args(serialization_settings, mt)

    assert args[1] == "a,b,c"
