import inspect
import os
import sys
from contextlib import suppress


def _get_fast_register_modules(fast_register_path):
    debug_logs = os.environ.get("DEBUG_LOGS", "false").lower() == "true"
    # Get modules that are defined in fast_register_path
    fast_register_modules = []

    module_names = list(sys.modules)
    if debug_logs:
        print(f"Module names: {module_names}")
    for module_name in module_names:
        # Do not unload this module
        if module_name == "unionai_actor_executor._module_mgmt":
            continue

        try:
            module_file_path = inspect.getfile(sys.modules[module_name])
        except (TypeError, KeyError):
            continue

        absolute_file_path = os.path.abspath(module_file_path)

        if not os.path.commonpath([fast_register_path, absolute_file_path]) == fast_register_path:
            if debug_logs:
                print(f"Not unloading module {module_name} as it is not defined in under {fast_register_path}")
            continue

        fast_register_modules.append(module_name)

    return fast_register_modules


def reset_env(fast_register_path: str | None):

    # if env_keys:
    #     for key in env_keys:
    #         if key in os.environ:
    #             del os.environ[key]
    #             print(f"Removed environment variable: {key}")

    if fast_register_path:
        sys.path.remove(fast_register_path)
        # Unload modules that are user defined and resets Launchplan cache
        user_modules = _get_fast_register_modules(fast_register_path)
        for name in user_modules:
            with suppress(KeyError):
                del sys.modules[name]

    try:
        from flytekit import LaunchPlan
        if hasattr(LaunchPlan, "CACHE"):
            LaunchPlan.CACHE = {}
    except ImportError:
        pass

    # if cwd:
    #     os.chdir(cwd)
    #     print(f"Changed working directory to: {cwd}")
