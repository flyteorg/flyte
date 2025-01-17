import sys
import typing
from collections import OrderedDict
from functools import partial

import pytest

import flytekit.configuration
from flytekit.configuration import Image, ImageConfig
from flytekit.core.array_node_map_task import ArrayNodeMapTaskResolver
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.map_task import MapTaskResolver, map_task
from flytekit.core.task import TaskMetadata, task
from flytekit.core.workflow import workflow
from flytekit.experimental import map_task as array_node_map_task
from flytekit.tools.translator import gather_dependent_entities, get_serializable

default_img = Image(name="default", fqn="test", tag="tag")
serialization_settings = flytekit.configuration.SerializationSettings(
    project="project",
    domain="domain",
    version="version",
    env=None,
    image_config=ImageConfig(default_image=default_img, images=[default_img]),
)


def test_basics_1():
    @task
    def t1(a: int, b: str, c: float) -> int:
        return a + len(b) + int(c)

    outside_p = partial(t1, b="hello", c=3.14)

    @workflow
    def my_wf_1(a: int) -> typing.Tuple[int, int]:
        inner_partial = partial(t1, b="world", c=2.7)
        out = outside_p(a=a)
        inside = inner_partial(a=a)
        return out, inside

    with pytest.raises(Exception):
        get_serializable(OrderedDict(), serialization_settings, outside_p)

    # check the od todo
    od = OrderedDict()
    wf_1_spec = get_serializable(od, serialization_settings, my_wf_1)
    tts, wspecs, lps = gather_dependent_entities(od)
    tts = [t for t in tts.values()]
    assert len(tts) == 1
    assert len(wf_1_spec.template.nodes) == 2
    assert wf_1_spec.template.nodes[0].task_node.reference_id.name == tts[0].id.name
    assert wf_1_spec.template.nodes[1].task_node.reference_id.name == tts[0].id.name
    assert wf_1_spec.template.nodes[0].inputs[0].binding.promise.var == "a"
    assert wf_1_spec.template.nodes[0].inputs[1].binding.scalar is not None
    assert wf_1_spec.template.nodes[0].inputs[2].binding.scalar is not None

    @task
    def get_str() -> str:
        return "got str"

    bind_c = partial(t1, c=2.7)

    @workflow
    def my_wf_2(a: int) -> int:
        s = get_str()
        inner_partial = partial(bind_c, b=s)
        inside = inner_partial(a=a)
        return inside

    wf_2_spec = get_serializable(OrderedDict(), serialization_settings, my_wf_2)
    assert len(wf_2_spec.template.nodes) == 2


@pytest.mark.parametrize(
    "map_task_fn",
    [
        map_task,
        array_node_map_task,
    ],
)
def test_map_task_types(map_task_fn):
    @task
    def t3(a: int, b: str, c: float) -> str:
        return str(a) + b + str(c)

    t3_bind_b1 = partial(t3, b="hello")
    t3_bind_b2 = partial(t3, b="world")
    t3_bind_c1 = partial(t3_bind_b1, c=3.14)
    t3_bind_c2 = partial(t3_bind_b2, c=2.78)

    mt1 = map_task_fn(t3_bind_c1, metadata=TaskMetadata(cache=True, cache_version="1"))
    mt2 = map_task_fn(t3_bind_c2, metadata=TaskMetadata(cache=True, cache_version="1"))

    @task
    def print_lists(i: typing.List[str], j: typing.List[str]):
        print(f"First: {i}")
        print(f"Second: {j}")

    @workflow
    def wf_out(a: typing.List[int]):
        i = mt1(a=a)
        j = mt2(a=[3, 4, 5])
        print_lists(i=i, j=j)

    wf_out(a=[1, 2])

    @workflow
    def wf_in(a: typing.List[int]):
        mt_in1 = map_task_fn(t3_bind_c1, metadata=TaskMetadata(cache=True, cache_version="1"))
        mt_in2 = map_task_fn(t3_bind_c2, metadata=TaskMetadata(cache=True, cache_version="1"))
        i = mt_in1(a=a)
        j = mt_in2(a=[3, 4, 5])
        print_lists(i=i, j=j)

    wf_in(a=[1, 2])

    od = OrderedDict()
    wf_spec = get_serializable(od, serialization_settings, wf_in)
    tts, _, _ = gather_dependent_entities(od)
    assert len(tts) == 2  # one map task + the print task
    if map_task_fn == array_node_map_task:
        assert (
            wf_spec.template.nodes[0].array_node.node.task_node.reference_id.name
            == wf_spec.template.nodes[1].array_node.node.task_node.reference_id.name
        )
    elif map_task_fn == map_task:
        assert (
            wf_spec.template.nodes[0].task_node.reference_id.name
            == wf_spec.template.nodes[1].task_node.reference_id.name
        )
    else:
        raise ValueError("Unexpected map task fn")
    assert wf_spec.template.nodes[0].inputs[0].binding.promise is not None  # comes from wf input
    assert wf_spec.template.nodes[1].inputs[0].binding.collection is not None  # bound to static list
    assert wf_spec.template.nodes[1].inputs[1].binding.scalar is not None  # these are bound
    assert wf_spec.template.nodes[1].inputs[2].binding.scalar is not None


@pytest.mark.parametrize(
    "map_task_fn",
    [
        map_task,
        array_node_map_task,
    ],
)
def test_lists_cannot_be_used_in_partials(map_task_fn):
    @task
    def t(a: int, b: typing.List[str]) -> str:
        return str(a) + str(b)

    with pytest.raises(ValueError):
        map_task_fn(partial(t, b=["hello", "world"]))(a=[1, 2, 3])

    @task
    def t_multilist(a: int, b: typing.List[float], c: typing.List[int]) -> str:
        return str(a) + str(b) + str(c)

    with pytest.raises(ValueError):
        map_task_fn(partial(t_multilist, b=[3.14, 12.34, 9876.5432], c=[42, 99]))(a=[1, 2, 3, 4])

    @task
    def t_list_of_lists(a: typing.List[typing.List[float]], b: int) -> str:
        return str(a) + str(b)

    with pytest.raises(ValueError):
        map_task_fn(partial(t_list_of_lists, a=[[3.14]]))(b=[1, 2, 3, 4])


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
@pytest.mark.parametrize(
    "map_task_fn",
    [
        map_task,
        array_node_map_task,
    ],
)
def test_everything(map_task_fn):
    import pandas as pd

    @task
    def get_static_list() -> typing.List[float]:
        return [3.14, 2.718]

    @task
    def get_list_of_pd(s: int) -> typing.List[pd.DataFrame]:
        df1 = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
        df2 = pd.DataFrame({"Name": ["Rachel", "Eve", "Mary"], "Age": [22, 23, 24]})
        if s == 2:
            return [df1, df2]
        else:
            return [df1, df2, df1]

    @task
    def t3(a: int, b: str, c: typing.List[float], d: typing.List[float], a2: pd.DataFrame) -> str:
        return str(a) + f"pdsize{len(a2)}" + b + str(c) + "&&" + str(d)

    t3_bind_b2 = partial(t3, b="world")
    # TODO: partial lists are not supported yet.
    # t3_bind_b1 = partial(t3, b="hello")
    # t3_bind_c1 = partial(t3_bind_b1, c=[6.674, 1.618, 6.626], d=[1.0])
    # mt1 = map_task_fn(t3_bind_c1)

    mt1 = map_task_fn(t3_bind_b2)

    if map_task_fn == array_node_map_task:
        mr = ArrayNodeMapTaskResolver()
    elif map_task_fn == map_task:
        mr = MapTaskResolver()
    else:
        raise ValueError("Unexpected map task fn")
    aa = mr.loader_args(serialization_settings, mt1)
    # Check bound vars
    aa = aa[1].split(",")
    aa.sort()
    assert aa == ["b"]

    @task
    def print_lists(i: typing.List[str], j: typing.List[str], k: typing.List[str]) -> str:
        print(f"First: {i}")
        print(f"Second: {j}")
        print(f"Third: {k}")
        return f"{i}-{j}-{k}"

    @dynamic
    def dt1(a: typing.List[int], a2: typing.List[pd.DataFrame], sl: typing.List[float]) -> str:
        i = mt1(a=a, a2=a2, c=[[1.1, 2.0, 3.0], [1.1, 2.0, 3.0]], d=[sl, sl])
        mt_in2 = map_task_fn(t3_bind_b2)
        dfs = get_list_of_pd(s=3)
        j = mt_in2(a=[3, 4, 5], a2=dfs, c=[[1.0], [2.0], [3.0]], d=[sl, sl, sl])

        # Test a2 bound to a fixed dataframe
        t3_bind_a2 = partial(t3_bind_b2, a2=a2[0])

        mt_in3 = map_task_fn(t3_bind_a2)

        aa = mr.loader_args(serialization_settings, mt_in3)
        # Check bound vars
        aa = aa[1].split(",")
        aa.sort()
        assert aa == ["a2", "b"]

        k = mt_in3(a=[3, 4, 5], c=[[1.0], [2.0], [3.0]], d=[sl, sl, sl])
        return print_lists(i=i, j=j, k=k)

    @workflow
    def wf_dt(a: typing.List[int]) -> str:
        sl = get_static_list()
        dfs = get_list_of_pd(s=2)
        return dt1(a=a, a2=dfs, sl=sl)

    print(wf_dt(a=[1, 2]))
    assert (
        wf_dt(a=[1, 2])
        == "['1pdsize2world[1.1, 2.0, 3.0]&&[3.14, 2.718]', '2pdsize3world[1.1, 2.0, 3.0]&&[3.14, 2.718]']-['3pdsize2world[1.0]&&[3.14, 2.718]', '4pdsize3world[2.0]&&[3.14, 2.718]', '5pdsize2world[3.0]&&[3.14, 2.718]']-['3pdsize2world[1.0]&&[3.14, 2.718]', '4pdsize2world[2.0]&&[3.14, 2.718]', '5pdsize2world[3.0]&&[3.14, 2.718]']"
    )
