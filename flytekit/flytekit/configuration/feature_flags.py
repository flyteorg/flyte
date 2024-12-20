import os


def _get(key: str, default_val: str) -> str:
    v = os.environ.get(key)
    return v if v else default_val


class FeatureFlags:
    FLYTE_PYTHON_PACKAGE_ROOT = _get("FLYTE_PYTHON_PACKAGE_ROOT", "auto")
    """
    Valid values,
    "auto" -> Automatically locate the fully qualified module name. This assumes that every python package directory is valid and has an _init__.py file
     "." -> Assuming current directory as the root
     or an actual path -> path to the root, this will be used to locate the root.
    """
