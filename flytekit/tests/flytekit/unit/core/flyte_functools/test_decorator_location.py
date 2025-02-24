import importlib

from flytekit.core.tracker import extract_task_module


def test_dont_use_wrapper_location():
    m = importlib.import_module("tests.flytekit.unit.core.flyte_functools.decorator_usage")
    get_data_task = getattr(m, "get_data")
    assert "decorator_source" not in get_data_task.name
    assert "decorator_usage" in get_data_task.name

    a, b, c, _ = extract_task_module(get_data_task)
    assert (a, b, c) == (
        "tests.flytekit.unit.core.flyte_functools.decorator_usage.get_data",
        "tests.flytekit.unit.core.flyte_functools.decorator_usage",
        "get_data",
    )
