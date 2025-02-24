import importlib.util
import sys
import types

LAZY_MODULES = []


class LazyModule(types.ModuleType):
    def __init__(self, module_name: str):
        super().__init__(module_name)
        self._module_name = module_name

    def __getattribute__(self, attr):
        raise ImportError(f"Module {object.__getattribute__(self, '_module_name')} is not yet installed.")


def is_imported(module_name):
    """
    This function is used to check if a module has been imported by the regular import.
    """
    return module_name in sys.modules and module_name not in LAZY_MODULES


def lazy_module(fullname):
    """
    This function is used to lazily import modules.  It is used in the following way:
    .. code-block:: python
        from flytekit.lazy_import import lazy_module
        sklearn = lazy_module("sklearn")
        sklearn.svm.SVC()
    :param Text fullname: The full name of the module to import
    """
    if fullname in sys.modules:
        return sys.modules[fullname]
    # https://docs.python.org/3/library/importlib.html#implementing-lazy-imports
    spec = importlib.util.find_spec(fullname)
    if spec is None:
        # Return a lazy module if the module is not found in the python environment,
        # so that we can raise a proper error when the user tries to access an attribute in the module.
        return LazyModule(fullname)
    loader = importlib.util.LazyLoader(spec.loader)
    spec.loader = loader
    module = importlib.util.module_from_spec(spec)
    sys.modules[fullname] = module
    LAZY_MODULES.append(module)
    loader.exec_module(module)
    return module
