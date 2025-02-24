import contextlib
import importlib
import os
import pkgutil
import sys
from typing import Any, Iterator, List, Union


@contextlib.contextmanager
def add_sys_path(path: Union[str, os.PathLike]) -> Iterator[None]:
    """Temporarily add given path to `sys.path`."""
    path = os.fspath(path)
    try:
        sys.path.insert(0, path)
        yield
    finally:
        sys.path.remove(path)


def just_load_modules(pkgs: List[str]):
    """
    This one differs from the above in that we don't yield anything, just load all the modules.
    """
    for package_name in pkgs:
        package = importlib.import_module(package_name)

        # If it doesn't have a __path__ field, that means it's not a package, just a module
        if not hasattr(package, "__path__"):
            continue

        # Note that walk_packages takes an onerror arg and swallows import errors silently otherwise
        for _, name, _ in pkgutil.walk_packages(package.__path__, prefix=f"{package_name}."):
            importlib.import_module(name)


def load_object_from_module(object_location: str) -> Any:
    """
    # TODO: Handle corner cases, like where the first part is [] maybe
    """
    class_obj = object_location.split(".")
    class_obj_mod = class_obj[:-1]  # e.g. ['flytekit', 'core', 'python_auto_container']
    class_obj_key = class_obj[-1]  # e.g. 'default_task_class_obj'
    class_obj_mod = importlib.import_module(".".join(class_obj_mod))
    return getattr(class_obj_mod, class_obj_key)
