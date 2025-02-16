from typing import Dict, List, NamedTuple, Optional, Union

import pytest

from flytekit.core import launch_plan
from flytekit.core.task import task
from flytekit.core.workflow import workflow
from flytekit.models import literals as _literal_models


def test_wf1_with_subwf():
    @task
    def t1(a: int) -> NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @task
    def t2(a: str, b: str) -> str:
        return b + a

    @workflow
    def my_subwf(a: int) -> (str, str):
        x, y = t1(a=a)
        u, v = t1(a=x)
        return y, v

    @workflow
    def my_wf(a: int, b: str) -> (int, str, str):
        x, y = t1(a=a).with_overrides()
        u, v = my_subwf(a=x)
        return x, u, v

    res = my_wf(a=5, b="hello ")
    assert res == (7, "world-9", "world-11")


def test_single_named_output_subwf():
    nt = NamedTuple("SubWfOutput", [("sub_int", int)])

    @task
    def t1(a: int) -> nt:
        a = a + 2
        return nt(a)

    @task
    def t2(a: int, b: int) -> nt:
        return nt(a + b)  # returns the named tuple

    @workflow
    def my_subwf(a: int) -> nt:
        x = t1(a=a)
        y = t2(a=a, b=x.sub_int)
        return nt(y.sub_int)

    res = my_subwf(a=3)
    # WF outputs don't currently return the namedtuple they're supposed to.
    assert res[0] == 8

    @workflow
    def my_wf(a: int) -> int:
        x = t1(a=a)
        u = my_subwf(a=x.sub_int)
        return u.sub_int

    res = my_wf(a=5)
    assert res == 16


def test_lp_default_handling():
    @task
    def t1(a: int) -> NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def my_wf(a: int, b: int) -> (str, str, int, int):
        x, y = t1(a=a)
        u, v = t1(a=b)
        return y, v, x, u

    lp = launch_plan.LaunchPlan.create("test1", my_wf)
    assert len(lp.parameters.parameters) == 2
    assert lp.parameters.parameters["a"].required
    assert lp.parameters.parameters["a"].default is None
    assert lp.parameters.parameters["b"].required
    assert lp.parameters.parameters["b"].default is None
    assert len(lp.fixed_inputs.literals) == 0

    lp_with_defaults = launch_plan.LaunchPlan.create("test2", my_wf, default_inputs={"a": 3})
    assert len(lp_with_defaults.parameters.parameters) == 2
    assert not lp_with_defaults.parameters.parameters["a"].required
    assert lp_with_defaults.parameters.parameters["a"].default == _literal_models.Literal(
        scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=3))
    )
    assert len(lp_with_defaults.fixed_inputs.literals) == 0

    lp_with_fixed = launch_plan.LaunchPlan.create("test3", my_wf, fixed_inputs={"a": 3})
    assert len(lp_with_fixed.parameters.parameters) == 1
    assert len(lp_with_fixed.fixed_inputs.literals) == 1
    assert lp_with_fixed.fixed_inputs.literals["a"] == _literal_models.Literal(
        scalar=_literal_models.Scalar(primitive=_literal_models.Primitive(integer=3))
    )

    @workflow
    def my_wf2(a: int, b: int = 42) -> (str, str, int, int):
        x, y = t1(a=a)
        u, v = t1(a=b)
        return y, v, x, u

    lp = launch_plan.LaunchPlan.create("test4", my_wf2)
    assert len(lp.parameters.parameters) == 2
    assert len(lp.fixed_inputs.literals) == 0

    lp_with_defaults = launch_plan.LaunchPlan.create("test5", my_wf2, default_inputs={"a": 3})
    assert len(lp_with_defaults.parameters.parameters) == 2
    assert len(lp_with_defaults.fixed_inputs.literals) == 0
    # Launch plan defaults override wf defaults
    assert lp_with_defaults(b=3) == ("world-5", "world-5", 5, 5)

    lp_with_fixed = launch_plan.LaunchPlan.create("test6", my_wf2, fixed_inputs={"a": 3})
    assert len(lp_with_fixed.parameters.parameters) == 1
    assert len(lp_with_fixed.fixed_inputs.literals) == 1
    # Launch plan defaults override wf defaults
    assert lp_with_fixed(b=3) == ("world-5", "world-5", 5, 5)

    lp_with_fixed = launch_plan.LaunchPlan.create("test7", my_wf2, fixed_inputs={"b": 3})
    assert len(lp_with_fixed.parameters.parameters) == 1
    assert len(lp_with_fixed.fixed_inputs.literals) == 1


def test_wf1_with_lp_node():
    @task
    def t1(a: int) -> NamedTuple("OutputsBC", t1_int_output=int, c=str):
        a = a + 2
        return a, "world-" + str(a)

    @workflow
    def my_subwf(a: int) -> (str, str):
        x, y = t1(a=a)
        u, v = t1(a=x)
        return y, v

    lp = launch_plan.LaunchPlan.create("lp_nodetest1", my_subwf)
    lp_with_defaults = launch_plan.LaunchPlan.create("lp_nodetest2", my_subwf, default_inputs={"a": 3})

    @workflow
    def my_wf(a: int = 42) -> (int, str, str):
        x, y = t1(a=a).with_overrides()
        u, v = lp(a=x)
        return x, u, v

    x = my_wf(a=5)
    assert x == (7, "world-9", "world-11")

    assert my_wf() == (44, "world-46", "world-48")

    @workflow
    def my_wf2(a: int = 42) -> (int, str, str, str):
        x, y = t1(a=a).with_overrides()
        u, v = lp_with_defaults()
        return x, y, u, v

    assert my_wf2() == (44, "world-44", "world-5", "world-7")

    @workflow
    def my_wf3(a: int = 42) -> (int, str, str, str):
        x, y = t1(a=a).with_overrides()
        u, v = lp_with_defaults(a=x)
        return x, y, u, v

    assert my_wf2() == (44, "world-44", "world-5", "world-7")


def test_optional_input():
    @task()
    def t1(a: Optional[int] = None, b: Optional[List[int]] = None, c: Optional[Dict[str, int]] = None) -> Optional[int]:
        ...

    @task()
    def t2(a: Union[int, Optional[List[int]], None] = None) -> Union[int, Optional[List[int]], None]:
        ...

    @workflow
    def wf(a: Optional[int] = 1) -> Optional[int]:
        t1()
        return t2(a=a)

    assert wf() is None

    with pytest.raises(ValueError, match="The default value for the optional type must be None, but got 3"):

        @task()
        def t3(c: Optional[int] = 3) -> Optional[int]:
            ...

        @workflow
        def wf():
            return t3()

        wf()
