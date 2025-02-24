import importlib
import inspect
import os
import sys
import typing
from pathlib import Path
from types import ModuleType
from typing import Callable, Optional, Tuple, Union

from flytekit.configuration.feature_flags import FeatureFlags
from flytekit.exceptions import system as _system_exceptions
from flytekit.loggers import logger


def import_module_from_file(module_name, file):
    try:
        spec = importlib.util.spec_from_file_location(module_name, file)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        return module
    except AssertionError:
        # handle where we can't determine the module of functions within the module
        return importlib.import_module(module_name)
    except Exception as exc:
        raise ModuleNotFoundError(f"Module from file {file} cannot be loaded") from exc


class InstanceTrackingMeta(type):
    """
    Please see the original class :py:class`flytekit.common.mixins.registerable._InstanceTracker` also and also look
    at the tests in the ``tests/flytekit/unit/core/tracker/test_tracking/`` folder to see how it's used.

    Basically, this will make instances of classes that use this metaclass aware of the module (the .py file) that
    caused the instance to be created. This is useful because it means that we can then (at least try to) find the
    variable that the instance was assigned to.
    """

    @staticmethod
    def _get_module_from_main(globals) -> Optional[str]:
        curdir = Path.cwd()
        file = globals.get("__file__")
        if file is None:
            return None

        file = Path(file)
        try:
            file_relative = file.relative_to(curdir)
        except ValueError:
            return None

        module_components = [*file_relative.with_suffix("").parts]
        module_name = ".".join(module_components)
        if len(module_components) == 0:
            return None

        # make sure current directory is in the PYTHONPATH.
        sys.path.insert(0, str(curdir))
        try:
            return import_module_from_file(module_name, file)
        except ModuleNotFoundError:
            return None

    @staticmethod
    def _find_instance_module():
        frame = inspect.currentframe()
        while frame:
            if frame.f_code.co_name == "<module>" and "__name__" in frame.f_globals:
                if frame.f_globals["__name__"] != "__main__":
                    return frame.f_globals["__name__"], None

                # Try to find the module and filename in the case that we're in the __main__ module
                # This is useful in cases that use FlyteRemote to load tasks/workflows that are defined
                # in the same file as where FlyteRemote is being invoked to register and execute Flyte
                # entities. One such case is with the `eager` decorator in the flytekit.experimental module.
                mod = InstanceTrackingMeta._get_module_from_main(frame.f_globals)
                if mod is None:
                    return None, None

                # This is used in find_lhs to find the trackable instances when the module is loaded from the __file__.
                return mod.__name__, mod.__file__
            frame = frame.f_back
        return None, None

    def __call__(cls, *args, **kwargs):
        o = super(InstanceTrackingMeta, cls).__call__(*args, **kwargs)
        mod_name, mod_file = InstanceTrackingMeta._find_instance_module()
        o._instantiated_in = mod_name
        o._module_file = mod_file
        return o


class TrackedInstance(metaclass=InstanceTrackingMeta):
    """
    Please see the notes for the metaclass above first.

    This functionality has two use-cases currently,
    * Keep track of naming for non-function ``PythonAutoContainerTasks``.  That is, things like the
      :py:class:`flytekit.extras.sqlite3.task.SQLite3Task` task.
    * Task resolvers, because task resolvers are instances of :py:class:`flytekit.core.python_auto_container.TaskResolverMixin`
      classes, not the classes themselves, which means we need to look on the left hand side of them to see how to
      find them at task execution time.
    """

    def __init__(self, *args, **kwargs):
        self._instantiated_in = None
        self._module_file = None
        self._lhs = None
        super().__init__(*args, **kwargs)

    @property
    def instantiated_in(self) -> str:
        return self._instantiated_in

    @property
    def location(self) -> str:
        n, _, _, _ = extract_task_module(self)
        return n

    @property
    def lhs(self):
        if self._lhs is not None:
            return self._lhs
        return self.find_lhs()

    def find_lhs(self) -> str:
        if self._lhs is not None:
            return self._lhs

        if self._instantiated_in is None or self._instantiated_in == "":
            raise _system_exceptions.FlyteSystemException(f"Object {self} does not have an _instantiated in")

        logger.debug(f"Looking for LHS for {self} from {self._instantiated_in}")
        m = importlib.import_module(self._instantiated_in)
        for k in dir(m):
            try:
                if getattr(m, k) is self:
                    logger.debug(f"Found LHS for {self}, {k}")
                    self._lhs = k
                    return k
            except ValueError as err:
                # Empty pandas dataframes behave weirdly here such that calling `m.df` raises:
                # ValueError: The truth value of a {type(self).__name__} is ambiguous. Use a.empty, a.bool(), a.item(),
                #   a.any() or a.all()
                # Since dataframes aren't registrable entities to begin with we swallow any errors they raise and
                # continue looping through m.
                logger.warning("Caught ValueError {} while attempting to auto-assign name".format(err))

        # Try to find object in module when the tracked instance is defined in the __main__ module.
        # This section tries to find the matching object in the module when the module is loaded from the __file__.
        if self._module_file is not None:
            # Since the module loaded from the file is different from the original module that defined self, we need
            # to match by variable name and type.
            module = import_module_from_file(self._instantiated_in, self._module_file)

            def _candidate_name_matches(candidate) -> bool:
                if not hasattr(candidate, "name") or not hasattr(self, "name"):
                    return False
                return candidate.name == self.name

            for k in dir(module):
                try:
                    candidate = getattr(module, k)
                    # consider the variable equivalent to self if it's of the same type and name
                    if (
                        type(candidate) == type(self)
                        and _candidate_name_matches(candidate)
                        and candidate.instantiated_in == self.instantiated_in
                    ):
                        self._lhs = k
                        return k
                except ValueError as err:
                    logger.warning(f"Caught ValueError {err} while attempting to auto-assign name")

        logger.error(f"Could not find LHS for {self} in {self._instantiated_in}")
        raise _system_exceptions.FlyteSystemException(f"Error looking for LHS in {self._instantiated_in}")


def isnested(func: Callable) -> bool:
    """
    Returns true if a function is local to another function and is not accessible through a module

    This would essentially be any function with a `.<local>.` (defined within a function) e.g.

    .. code:: python

        def foo():
            def foo_inner():
                pass
            pass

    In the above example `foo_inner` is the local function or a nested function.
    """

    return hasattr(func, "__code__") and (func.__code__.co_flags & inspect.CO_NESTED != 0)


def is_functools_wrapped_module_level(func: Callable) -> bool:
    """Returns true if the function is a functools.wraps-updated function that is defined in the module-level scope.

    .. code:: python

        import functools

        def decorator(fn):
            @functools.wraps(fn)
            def wrapper(*args, **kwargs):
                return fn(*arks, **kwargs)

            return wrapper

        @decorator
        def foo():
            ...

        def define_inner_wrapped_fn():

            @decorator
            def foo_inner(*args, **kwargs):
                return fn(*arks, **kwargs)

            return foo_inner

        bar = define_inner_wrapped_fn()

        is_functools_wrapped_module_level(foo)  # True
        is_functools_wrapped_module_level(bar)  # False

    In this case, applying this function to ``foo`` returns true because ``foo`` was defined in the module-level scope.
    Applying this function to ``bar`` returns false because it's being assigned to ``foo_inner``, which is a
    functools-wrapped function but is actually defined in the local scope of ``define_inner_wrapped_fn``.

    This works because functools.wraps updates the __name__ and __qualname__ attributes of the wrapper to match the
    wrapped function. Since ``define_inner_wrapped_fn`` doesn't update the __qualname__ of ``foo_inner``, the inner
    function's __qualname__ won't match its __name__.
    """
    return hasattr(func, "__wrapped__") and func.__name__ == func.__qualname__


def istestfunction(func) -> bool:
    """
    Return true if the function is defined in a test module.

    A test module has to have `test_` as the prefix or `_test` as the suffix.
    False in all other cases.
    """
    mod = inspect.getmodule(func)
    if mod:
        mod_name = mod.__name__
        if "." in mod_name:
            mod_name = mod_name.split(".")[-1]
        return mod_name.startswith("test_") or mod_name.endswith("_test")
    return False


class _ModuleSanitizer(object):
    """
    Sanitizes and finds the absolute module path irrespective of the import location.
    """

    def __init__(self):
        self._module_cache = {}

    def _resolve_abs_module_name(self, path: str, package_root: typing.Optional[str] = None) -> str:
        """
        Recursively finds the root python package under-which basename exists
        """
        # If we have already computed the module for this directory - return
        if path in self._module_cache:
            return self._module_cache[path]

        basename = os.path.basename(path)
        dirname = os.path.dirname(path)

        # Let us remove any extensions like .py
        basename = os.path.splitext(basename)[0]

        if dirname == package_root:
            return basename

        # If we have reached a directory with no __init__, ignore
        if "__init__.py" not in os.listdir(dirname):
            return basename

        # Now recurse down such that we can extract the absolute module path
        mod_name = self._resolve_abs_module_name(dirname, package_root)
        final_mod_name = f"{mod_name}.{basename}" if mod_name else basename
        self._module_cache[path] = final_mod_name
        return final_mod_name

    def get_absolute_module_name(self, path: str, package_root: typing.Optional[str] = None) -> str:
        """
        Returns the absolute module path for a given python file path. This assumes that every module correctly contains
        a __init__.py file. Absence of this file, indicates the root.
        """
        return self._resolve_abs_module_name(path, package_root)


_mod_sanitizer = _ModuleSanitizer()


def _task_module_from_callable(f: Callable):
    mod = inspect.getmodule(f)
    mod_name = getattr(mod, "__name__", f.__module__)
    name = f.__name__.split(".")[-1]
    return mod, mod_name, name


def extract_task_module(f: Union[Callable, TrackedInstance]) -> Tuple[str, str, str, str]:
    """
    Returns the task-name, absolute module and the string name of the callable.
    :param f: A task or any other callable
    :return: [name to use: str, module_name: str, function_name: str, full_path: str]
    """

    if isinstance(f, TrackedInstance):
        if hasattr(f, "task_function"):
            mod, mod_name, name = _task_module_from_callable(f.task_function)
        elif f.instantiated_in:
            mod = importlib.import_module(f.instantiated_in)
            mod_name = mod.__name__
            name = f.lhs
        else:
            raise AssertionError(f"Unable to determine module of {f}")
    else:
        mod, mod_name, name = _task_module_from_callable(f)

    if mod is None:
        raise AssertionError(f"Unable to determine module of {f}")

    if mod_name == "__main__":
        if hasattr(f, "task_function"):
            f = f.task_function
        inspect_file = inspect.getfile(f)  # type: ignore
        return name, "", name, os.path.abspath(inspect_file)

    mod_name = get_full_module_path(mod, mod_name)
    return f"{mod_name}.{name}", mod_name, name, os.path.abspath(inspect.getfile(mod))


def get_full_module_path(mod: ModuleType, mod_name: str) -> str:
    if FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT != ".":
        package_root = (
            FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT if FeatureFlags.FLYTE_PYTHON_PACKAGE_ROOT != "auto" else None
        )
        new_mod_name = _mod_sanitizer.get_absolute_module_name(inspect.getabsfile(mod), package_root)
        # We only replace the mod_name if it is more specific, else we already have a fully resolved path
        if len(new_mod_name) > len(mod_name):
            mod_name = new_mod_name
    return mod_name
