import datetime
import sys
import typing
from dataclasses import dataclass
from typing import Dict, List

import pytest
from dataclasses_json import DataClassJsonMixin
from pytest import fixture
from typing_extensions import Annotated

import flytekit
from flytekit.core.base_sql_task import SQLTask
from flytekit.core.base_task import kwtypes
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.hash import HashMethod
from flytekit.core.local_cache import LocalTaskCache, _calculate_cache_key, _recursive_hash_placement
from flytekit.core.task import TaskMetadata, task
from flytekit.core.testing import task_mock
from flytekit.core.type_engine import TypeEngine
from flytekit.core.workflow import workflow
from flytekit.models.literals import Literal, LiteralCollection, LiteralMap, Primitive, Scalar
from flytekit.models.types import LiteralType, SimpleType
from flytekit.types.schema import FlyteSchema

# Global counter used to validate number of calls to cache
n_cached_task_calls = 0


@fixture(scope="function", autouse=True)
def setup():
    global n_cached_task_calls
    n_cached_task_calls = 0

    LocalTaskCache.initialize()
    LocalTaskCache.clear()


def test_to_confirm_that_cache_keys_include_function_name():
    """
    This test confirms that the function name is part of the cache key. It does so by defining 2 tasks with
    identical parameters and metadata (i.e. cache=True and cache version).
    """

    @task(cache=True, cache_version="v1")
    def f1(n: int) -> int:
        global n_cached_task_calls
        n_cached_task_calls += 1

        return n

    @task(cache=True, cache_version="v1")
    def f2(n: int) -> int:
        global n_cached_task_calls
        n_cached_task_calls += 1

        return n + 1

    @workflow
    def wf(n: int) -> typing.Tuple[int, int]:
        n_f1 = f1(n=n)
        n_f2 = f2(n=n)
        return n_f1, n_f2

    # This is demonstrating that calls to f1 and f2 are cached by input parameters.
    assert wf(n=1) == (1, 2)


def test_single_task_workflow():
    @task(cache=True, cache_version="v1")
    def is_even(n: int) -> bool:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return n % 2 == 0

    @task(cache=False)
    def uncached_task(a: int, b: int) -> int:
        return a + b

    @workflow
    def check_evenness(n: int) -> bool:
        uncached_task(a=n, b=n)
        return is_even(n=n)

    assert n_cached_task_calls == 0
    assert check_evenness(n=1) is False
    # Confirm task is called
    assert n_cached_task_calls == 1
    assert check_evenness(n=1) is False
    # Subsequent calls of the workflow with the same parameter do not bump the counter
    assert n_cached_task_calls == 1
    assert check_evenness(n=1) is False
    assert n_cached_task_calls == 1

    # Run workflow with a different parameter and confirm counter is bumped
    assert check_evenness(n=8) is True
    assert n_cached_task_calls == 2
    # Run workflow again with the same parameter and confirm the counter is not bumped
    assert check_evenness(n=8) is True
    assert n_cached_task_calls == 2


def test_cache_can_be_disabled(monkeypatch):
    monkeypatch.setenv("FLYTE_LOCAL_CACHE_ENABLED", "false")

    @task(cache=True, cache_version="v1")
    def is_even(n: int) -> bool:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return n % 2 == 0

    assert n_cached_task_calls == 0
    # Run once and check that the counter is increased
    assert is_even(n=1) is False
    assert n_cached_task_calls == 1

    # Run again and check that the counter is increased again i.e. no caching
    assert is_even(n=1) is False
    assert n_cached_task_calls == 2


def test_cache_can_be_overwrited(monkeypatch):
    @task(cache=True, cache_version="v1")
    def is_even(n: int) -> bool:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return n % 2 == 0

    assert n_cached_task_calls == 0

    # Run once and check that the counter is increased
    assert is_even(n=1) is False
    assert n_cached_task_calls == 1

    # Run again and check that the counter is not increased, i.e. cache hit
    assert is_even(n=1) is False
    assert n_cached_task_calls == 1

    # function changed
    @task(cache=True, cache_version="v1")
    def is_even(n: int) -> bool:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return n % 2 == 1

    monkeypatch.setenv("FLYTE_LOCAL_CACHE_OVERWRITE", "true")

    # Run again and check that the counter is increased again, i.e. cache overwrite
    assert is_even(n=1) is True
    assert n_cached_task_calls == 2

    monkeypatch.setenv("FLYTE_LOCAL_CACHE_OVERWRITE", "false")

    # Run again and check that the counter is not increased, i.e. cache hit
    assert is_even(n=1) is True
    assert n_cached_task_calls == 2


def test_shared_tasks_in_two_separate_workflows():
    @task(cache=True, cache_version="0.0.1")
    def is_odd(n: int) -> bool:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return n % 2 == 1

    @workflow
    def check_oddness_wf1(n: int) -> bool:
        return is_odd(n=n)

    @workflow
    def check_oddness_wf2(n: int) -> bool:
        return is_odd(n=n)

    assert n_cached_task_calls == 0
    assert check_oddness_wf1(n=42) is False
    assert check_oddness_wf1(n=99) is True
    assert n_cached_task_calls == 2

    # The next two executions of the *_wf2 workflow are going to
    # hit the cache for the calls to `is_odd`
    assert check_oddness_wf2(n=42) is False
    assert check_oddness_wf2(n=99) is True
    assert n_cached_task_calls == 2


# TODO add test with typing.List[str]


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_sql_task():
    import pandas as pd

    sql = SQLTask(
        "my-query",
        query_template="SELECT * FROM hive.city.fact_airport_sessions WHERE ds = '{{ .Inputs.ds }}' LIMIT 10",
        inputs=kwtypes(ds=datetime.datetime),
        outputs=kwtypes(results=FlyteSchema),
        metadata=TaskMetadata(retries=2, cache=True, cache_version="0.1"),
    )

    @task(cache=True, cache_version="0.1.2")
    def t1() -> datetime.datetime:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return datetime.datetime.now()

    @workflow
    def my_wf() -> FlyteSchema:
        dt = t1()
        return sql(ds=dt)

    with task_mock(sql) as mock:
        mock.return_value = pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})
        assert n_cached_task_calls == 0
        assert (my_wf().open().all() == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()
        assert n_cached_task_calls == 1
        # The second and third calls hit the cache
        assert (my_wf().open().all() == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()
        assert n_cached_task_calls == 1
        assert (my_wf().open().all() == pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()
        assert n_cached_task_calls == 1


def test_wf_custom_types():
    @dataclass
    class MyCustomType(DataClassJsonMixin):
        x: int
        y: str

    @task(cache=True, cache_version="a.b.c")
    def t1(a: int) -> MyCustomType:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return MyCustomType(x=a, y="t1")

    @task(cache=True, cache_version="v1")
    def t2(a: MyCustomType, b: str) -> (MyCustomType, int):
        global n_cached_task_calls
        n_cached_task_calls += 1
        return MyCustomType(x=a.x, y=f"{a.y} {b}"), 5

    @workflow
    def my_wf(a: int, b: str) -> (MyCustomType, int):
        return t2(a=t1(a=a), b=b)

    assert n_cached_task_calls == 0
    c, v = my_wf(a=10, b="hello")
    assert v == 5
    assert c.x == 10
    assert c.y == "t1 hello"
    assert n_cached_task_calls == 2
    c, v = my_wf(a=10, b="hello")
    assert v == 5
    assert c.x == 10
    assert c.y == "t1 hello"
    assert n_cached_task_calls == 2


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_wf_schema_to_df():
    import pandas as pd

    schema1 = FlyteSchema[kwtypes(x=int, y=str)]

    @task(cache=True, cache_version="v0")
    def t1() -> schema1:
        global n_cached_task_calls
        n_cached_task_calls += 1

        s = schema1()
        s.open().write(pd.DataFrame(data={"x": [1, 2], "y": ["3", "4"]}))
        return s

    @task(cache=True, cache_version="v1")
    def t2(df: pd.DataFrame) -> int:
        global n_cached_task_calls
        n_cached_task_calls += 1

        return len(df.columns.values)

    @workflow
    def wf() -> int:
        return t2(df=t1())

    assert n_cached_task_calls == 0
    x = wf()
    assert x == 2
    assert n_cached_task_calls == 2
    # Second call does not bump the counter
    x = wf()
    assert x == 2
    assert n_cached_task_calls == 2


def test_dict_wf_with_constants():
    @task(cache=True, cache_version="v99")
    def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
        global n_cached_task_calls
        n_cached_task_calls += 1

        return a + 2, "world"

    @task(cache=True, cache_version="v101")
    def t2(a: typing.Dict[str, str]) -> str:
        global n_cached_task_calls
        n_cached_task_calls += 1

        return " ".join([v for k, v in a.items()])

    @workflow
    def my_wf(a: int, b: str) -> (int, str):
        x, y = t1(a=a)
        d = t2(a={"key1": b, "key2": y})
        return x, d

    assert n_cached_task_calls == 0
    x = my_wf(a=5, b="hello")
    assert x == (7, "hello world")
    assert n_cached_task_calls == 2
    # Second call does not bump the counter
    x = my_wf(a=5, b="hello")
    assert x == (7, "hello world")
    assert n_cached_task_calls == 2


def test_set_integer_literal_hash_is_cached():
    """
    Test to confirm that the local cache is set in the case of integers, even if we
    return an annotated integer. In order to make this very explicit, we define a constant hash
    function, i.e. the same value is returned by it regardless of the input.
    """

    def constant_hash_function(a: int) -> str:
        return "hash"

    @task
    def t0(a: int) -> Annotated[int, HashMethod(function=constant_hash_function)]:
        return a

    @task(cache=True, cache_version="0.0.1")
    def t1(cached_a: int) -> int:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return cached_a

    @workflow
    def wf(a: int) -> int:
        annotated_a = t0(a=a)
        return t1(cached_a=annotated_a)

    assert n_cached_task_calls == 0
    assert wf(a=3) == 3
    assert n_cached_task_calls == 1
    # Confirm that the value is cached due to the fact the hash value is constant, regardless
    # of the value passed to the cacheable task.
    assert wf(a=2) == 3
    assert n_cached_task_calls == 1
    # Confirm that the cache is hit if we execute the workflow with the same value as previous run.
    assert wf(a=2) == 3
    assert n_cached_task_calls == 1


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_pass_annotated_to_downstream_tasks():
    @task
    def t0(a: int) -> Annotated[int, HashMethod(function=str)]:
        return a + 1

    @task(cache=True, cache_version="42")
    def downstream_t(a: int) -> int:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return a + 2

    @dynamic
    def t1(a: int) -> int:
        v = t0(a=a)

        # We should have a cache miss in the first call to downstream_t and have a cache hit
        # on the second call.
        downstream_t(a=v)
        v_2 = downstream_t(a=v)

        return v_2

    assert n_cached_task_calls == 0
    assert t1(a=3) == 6
    assert n_cached_task_calls == 1


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_pd_dataframe_hash():
    """
    Test that cache is hit in the case of pd dataframes where we annotated dataframes to hash
    the contents of the dataframes.
    """
    import pandas as pd

    def hash_pd_dataframe(df: pd.DataFrame) -> str:
        return str(pd.util.hash_pandas_object(df))

    @task
    def uncached_data_reading_task() -> Annotated[pd.DataFrame, HashMethod(hash_pd_dataframe)]:
        return pd.DataFrame({"column_1": [1, 2, 3]})

    @task(cache=True, cache_version="0.1")
    def cached_data_processing_task(data: pd.DataFrame) -> pd.DataFrame:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return data * 2

    @workflow
    def my_workflow():
        raw_data = uncached_data_reading_task()
        cached_data_processing_task(data=raw_data)

    assert n_cached_task_calls == 0
    my_workflow()
    assert n_cached_task_calls == 1

    # Confirm that we see a cache hit in the case of annotated dataframes.
    my_workflow()
    assert n_cached_task_calls == 1


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_list_of_pd_dataframe_hash():
    """
    Test that cache is hit in the case of a list of pd dataframes where we annotated dataframes to hash
    the contents of the dataframes.
    """
    import pandas as pd

    def hash_pd_dataframe(df: pd.DataFrame) -> str:
        return str(pd.util.hash_pandas_object(df))

    @task
    def uncached_data_reading_task() -> List[Annotated[pd.DataFrame, HashMethod(hash_pd_dataframe)]]:
        return [pd.DataFrame({"column_1": [1, 2, 3]}), pd.DataFrame({"column_1": [10, 20, 30]})]

    @task(cache=True, cache_version="0.1")
    def cached_data_processing_task(data: List[pd.DataFrame]) -> List[pd.DataFrame]:
        global n_cached_task_calls
        n_cached_task_calls += 1
        return [df * 2 for df in data]

    @workflow
    def my_workflow():
        raw_data = uncached_data_reading_task()
        cached_data_processing_task(data=raw_data)

    assert n_cached_task_calls == 0
    my_workflow()
    assert n_cached_task_calls == 1

    # Confirm that we see a cache hit in the case of annotated dataframes.
    my_workflow()
    assert n_cached_task_calls == 1


def test_cache_key_repetition():
    pt = Dict
    lt = TypeEngine.to_literal_type(pt)
    ctx = FlyteContextManager.current_context()
    kwargs = {
        "a": 0.41083513079747874,
        "b": 0.7773927872515183,
        "c": 17,
    }
    keys = set()
    for i in range(0, 100):
        lit = TypeEngine.to_literal(ctx, kwargs, Dict, lt)
        lm = LiteralMap(
            literals={
                "d": lit,
            }
        )
        key = _calculate_cache_key("t1", "007", lm)
        keys.add(key)

    assert len(keys) == 1


def test_stable_cache_key():
    """
    The intent of this test is to ensure cache keys are stable across releases and python versions.
    """
    pt = Dict
    lt = TypeEngine.to_literal_type(pt)
    ctx = FlyteContextManager.current_context()
    kwargs = {
        "a": 42,
        "b": "abcd",
        "c": 0.12349,
        "d": [1, 2, 3],
        "e": {
            "e_a": 11,
            "e_b": list(range(1000)),
            "e_c": {
                "e_c_a": 12.34,
                "e_c_b": "a string",
            },
        },
    }
    lit = TypeEngine.to_literal(ctx, kwargs, Dict, lt)
    lm = LiteralMap(
        literals={
            "lit_1": lit,
            "lit_2": TypeEngine.to_literal(ctx, 99, int, LiteralType(simple=SimpleType.INTEGER)),
            "lit_3": TypeEngine.to_literal(ctx, 3.14, float, LiteralType(simple=SimpleType.FLOAT)),
            "lit_4": TypeEngine.to_literal(ctx, True, bool, LiteralType(simple=SimpleType.BOOLEAN)),
        }
    )
    key = _calculate_cache_key("task_name_1", "31415", lm)
    assert key == "task_name_1-31415-404b45f8556276183621d4bf37f50049"


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def calculate_cache_key_multiple_times(x, n=1000):
    import pandas as pd

    series = pd.Series(
        [
            _calculate_cache_key(
                task_name="task_name",
                cache_version="cache_version",
                input_literal_map=LiteralMap(
                    literals={
                        "d": TypeEngine.to_literal(
                            ctx=FlyteContextManager.current_context(),
                            expected=TypeEngine.to_literal_type(Dict),
                            python_type=Dict,
                            python_val=x,
                        ),
                    }
                ),
            )
            for _ in range(n)
        ]
    ).value_counts()
    return series


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
@pytest.mark.parametrize(
    "d",
    [
        dict(a=1, b=2, c=3),
        dict(x=dict(a=1, b=2, c=3)),
        dict(xs=[dict(a=1, b=2, c=3), dict(y=dict(a=10, b=20, c=30))]),
    ],
)
def test_cache_key_consistency(d):
    assert len(calculate_cache_key_multiple_times(d)) == 1


def test_literal_hash_placement():
    """
    Test that hashes on literal collections and maps are preserved by the
    _recursive_hash_placement function used in cache key calculations.
    """
    lit = Literal(scalar=Scalar(primitive=Primitive(string_value="test")))

    litmap = Literal(map=LiteralMap(literals={"test": lit}), hash="0xffff")
    litcoll = Literal(collection=LiteralCollection(literals=[lit]), hash="0xffff")

    assert litmap.hash == _recursive_hash_placement(litmap).hash
    assert litcoll.hash == _recursive_hash_placement(litcoll).hash


@task(cache=True, cache_version="v0")
def t2(n: int) -> int:
    ctx = flytekit.current_context()
    cp = ctx.checkpoint
    cp.write(bytes(n + 1))
    return n + 1


def test_checkpoint_cached_task():
    assert t2(n=5) == 6
