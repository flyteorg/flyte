import typing
from collections import OrderedDict

import pytest

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.base_task import TaskResolverMixin
from flytekit.core.class_based_resolver import ClassStorageTaskResolver
from flytekit.core.python_auto_container import default_task_resolver
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.tools.translator import get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


@workflow
def my_wf(a: int, b: str) -> typing.Tuple[int, str]:
    @task
    def t1(a: int) -> typing.Tuple[int, str]:
        return a + 2, "world"

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    x, y = t1(a=a)
    d = t2(a=y, b=b)
    return x, d


def test_wf_resolving():
    x = my_wf(a=3, b="hello")
    assert x == (5, "helloworld")

    # Because the workflow is nested inside a test, calling location will fail as it tries to find the LHS that the
    # workflow was assigned to
    assert my_wf.location == "tests.flytekit.unit.core.test_resolver.my_wf"

    workflows_tasks = my_wf.get_all_tasks()
    assert len(workflows_tasks) == 2  # Two tasks were declared inside

    # The tasks should get the location the workflow was assigned to as the resolver.
    # The args are the index.
    srz_t0_spec = get_serializable(OrderedDict(), serialization_settings, workflows_tasks[0])
    assert srz_t0_spec.template.container.args[-4:] == [
        "--resolver",
        "tests.flytekit.unit.core.test_resolver.my_wf",
        "--",
        "0",
    ]

    srz_t1_spec = get_serializable(OrderedDict(), serialization_settings, workflows_tasks[1])
    assert srz_t1_spec.template.container.args[-4:] == [
        "--resolver",
        "tests.flytekit.unit.core.test_resolver.my_wf",
        "--",
        "1",
    ]


def test_class_resolver():
    c = ClassStorageTaskResolver()
    assert c.name() != ""

    with pytest.raises(RuntimeError):
        c.load_task([])

    @task
    def t1(a: str, b: str) -> str:
        return b + a

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    c.add(t2)
    assert c.loader_args(None, t2) == ["0"]

    with pytest.raises(Exception):
        c.loader_args(t1)


def test_mixin():
    """
    This test is only to make codecov happy. Actual logic is already tested above.
    """
    x = TaskResolverMixin()
    x.location
    x.name()
    x.loader_args(None, None)
    x.get_all_tasks()
    x.load_task([])


def test_error():
    with pytest.raises(Exception):
        default_task_resolver.get_all_tasks()
