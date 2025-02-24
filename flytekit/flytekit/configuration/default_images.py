import enum
import sys
import typing
from contextlib import suppress


class PythonVersion(enum.Enum):
    PYTHON_3_8 = (3, 8)
    PYTHON_3_9 = (3, 9)
    PYTHON_3_10 = (3, 10)
    PYTHON_3_11 = (3, 11)
    PYTHON_3_12 = (3, 12)


class DefaultImages(object):
    """
    We may want to load the default images from remote - maybe s3 location etc?
    """

    _DEFAULT_IMAGE_PREFIXES = {
        PythonVersion.PYTHON_3_8: "cr.flyte.org/flyteorg/flytekit:py3.8-",
        PythonVersion.PYTHON_3_9: "cr.flyte.org/flyteorg/flytekit:py3.9-",
        PythonVersion.PYTHON_3_10: "cr.flyte.org/flyteorg/flytekit:py3.10-",
        PythonVersion.PYTHON_3_11: "cr.flyte.org/flyteorg/flytekit:py3.11-",
        PythonVersion.PYTHON_3_12: "cr.flyte.org/flyteorg/flytekit:py3.12-",
    }

    @classmethod
    def default_image(cls) -> str:
        from flytekit.configuration.plugin import get_plugin

        with suppress(AttributeError):
            default_image = get_plugin().get_default_image()
            if default_image is not None:
                return default_image

        return cls.find_image_for()

    @classmethod
    def find_image_for(
        cls, python_version: typing.Optional[PythonVersion] = None, flytekit_version: typing.Optional[str] = None
    ) -> str:
        if python_version is None:
            python_version = PythonVersion((sys.version_info.major, sys.version_info.minor))

        return cls._DEFAULT_IMAGE_PREFIXES[python_version] + (
            flytekit_version.replace("v", "") if flytekit_version else cls.get_version_suffix()
        )

    @classmethod
    def get_version_suffix(cls) -> str:
        from flytekit import __version__

        if not __version__ or "dev" in __version__:
            version_suffix = "latest"
        else:
            version_suffix = __version__
        return version_suffix
