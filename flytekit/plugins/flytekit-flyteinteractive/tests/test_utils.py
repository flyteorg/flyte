import os

from flytekitplugins.flyteinteractive import get_task_inputs
from flytekitplugins.flyteinteractive.utils import load_module_from_path


def test_load_module_from_path():
    module_name = "task"
    module_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), "testdata", "task.py")
    task_name = "t1"
    task_module = load_module_from_path(module_name, module_path)
    assert hasattr(task_module, task_name)
    task_def = getattr(task_module, task_name)
    assert task_def(a=6, b=3) == 2


def test_get_task_inputs():
    test_working_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)), "testdata")
    native_inputs = get_task_inputs("task", "t1", test_working_dir)
    assert native_inputs == {"a": 30, "b": 0}
