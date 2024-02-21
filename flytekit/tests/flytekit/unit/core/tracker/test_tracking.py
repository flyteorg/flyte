import typing

import pytest

from flytekit import task
from flytekit.configuration.feature_flags import FeatureFlags
from flytekit.core.tracker import extract_task_module
from tests.flytekit.unit.core.tracker import d
from tests.flytekit.unit.core.tracker.b import b_local_a, local_b
from tests.flytekit.unit.core.tracker.c import b_in_c, c_local_a


def test_tracking():
    # Test that instantiated in returns the module (.py file) where the instance is instantiated, not where the class
    # is defined.
    assert b_local_a.instantiated_in == "tests.flytekit.unit.core.tracker.b"
    assert b_local_a.lhs == "b_local_a"

    # Test that even if the actual declaration that constructs the object is in a different file, instantiated_in
    # still shows the module where the Python file where the instance is assigned to a variable
    assert c_local_a.instantiated_in == "tests.flytekit.unit.core.tracker.c"
    assert c_local_a.lhs == "c_local_a"

    assert local_b.instantiated_in == "tests.flytekit.unit.core.tracker.b"
    assert local_b.lhs == "local_b"

    assert b_in_c.instantiated_in == "tests.flytekit.unit.core.tracker.c"
    assert b_in_c.lhs == "b_in_c"


def convert_to_test(d: dict) -> typing.Tuple[typing.List[str], typing.List]:
    ids = []
    test_vals = []
    for k, v in d.items():
        ids.append(k)
        test_vals.append(v)
    return ids, test_vals


NAMES, TESTS = convert_to_test(
    {
        "local-convert_to_test": (
            convert_to_test,
            (
                "tests.flytekit.unit.core.tracker.test_tracking.convert_to_test",
                "tests.flytekit.unit.core.tracker.test_tracking",
                "convert_to_test",
            ),
        ),
        "core.task": (task, ("flytekit.core.task.task", "flytekit.core.task", "task")),
        "current-mod-tasks": (
            d.tasks,
            ("tests.flytekit.unit.core.tracker.d.tasks", "tests.flytekit.unit.core.tracker.d", "tasks"),
        ),
        "tasks-core-task": (d.task, ("flytekit.core.task.task", "flytekit.core.task", "task")),
        "tracked-local": (
            local_b,
            ("tests.flytekit.unit.core.tracker.b.local_b", "tests.flytekit.unit.core.tracker.b", "local_b"),
        ),
        "tracked-b-in-c": (
            b_in_c,
            ("tests.flytekit.unit.core.tracker.c.b_in_c", "tests.flytekit.unit.core.tracker.c", "b_in_c"),
        ),
    }
)


@pytest.mark.parametrize(
    "test_input,expected",
    argvalues=TESTS,
    ids=NAMES,
)
def test_extract_task_module(test_input, expected):
    old = FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT
    FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT = "auto"
    try:
        # The last element is the full path of a local file, which is not stable across users / runs.
        assert extract_task_module(test_input)[:-1] == expected
    except Exception:
        FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT = old
        raise


local_task = task(d.inner_function)


def test_local_task_wrap():
    assert local_task.instantiated_in == "tests.flytekit.unit.core.tracker.test_tracking"
