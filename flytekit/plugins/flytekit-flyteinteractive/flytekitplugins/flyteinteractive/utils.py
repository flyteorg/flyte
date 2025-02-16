import importlib
import os
import sys

from flyteidl.core import literals_pb2 as _literals_pb2

from flytekit.core import utils
from flytekit.core.context_manager import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.models import literals as _literal_models


def load_module_from_path(module_name, path):
    """
    Imports a Python module from a specified file path.

    Args:
        module_name (str): The name you want to assign to the imported module.
        path (str): The file system path to the Python file (.py) that contains the module you want to import.

    Returns:
        module: The imported module.

    Raises:
        ImportError: If the module cannot be loaded from the provided path, an ImportError is raised.
    """
    spec = importlib.util.spec_from_file_location(module_name, path)
    if spec is not None:
        module = importlib.util.module_from_spec(spec)
        sys.modules[module_name] = module
        spec.loader.exec_module(module)
        return module
    else:
        raise ImportError(f"Module at {path} could not be loaded")


def get_task_inputs(task_module_name, task_name, context_working_dir):
    """
    Read task input data from inputs.pb for a specific task function and convert it into Python types and structures.

    Args:
        task_module_name (str): The name of the Python module containing the task function.
        task_name (str): The name of the task function within the module.
        context_working_dir (str): The directory path where the input file and module file are located.

    Returns:
        dict: A dictionary containing the task inputs, converted into Python types and structures.
    """
    local_inputs_file = os.path.join(context_working_dir, "inputs.pb")
    input_proto = utils.load_proto_from_file(_literals_pb2.LiteralMap, local_inputs_file)
    idl_input_literals = _literal_models.LiteralMap.from_flyte_idl(input_proto)
    task_module = load_module_from_path(task_module_name, os.path.join(context_working_dir, f"{task_module_name}.py"))
    task_def = getattr(task_module, task_name)
    native_inputs = TypeEngine.literal_map_to_kwargs(
        FlyteContextManager().current_context(), idl_input_literals, task_def.python_interface.inputs
    )
    return native_inputs
