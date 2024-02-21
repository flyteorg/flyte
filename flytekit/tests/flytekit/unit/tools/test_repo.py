import os
import pathlib
import tempfile

import mock
import pytest

import flytekit.configuration
from flytekit.configuration import DefaultImages, ImageConfig
from flytekit.tools.repo import find_common_root, load_packages_and_modules

task_text = """
from flytekit import task
@task
def t1(a: int):
    ...
"""


# Mock out the entities so the load function doesn't try to load everything
@mock.patch("flytekit.core.context_manager.FlyteEntities")
@mock.patch("flytekit.core.base_task.FlyteEntities")
def test_module_loading(mock_entities, mock_entities_2):
    entities = []
    mock_entities.entities = entities
    mock_entities_2.entities = entities
    with tempfile.TemporaryDirectory() as tmp_dir:
        # Create directories
        top_level = os.path.join(tmp_dir, "top")
        middle_level = os.path.join(top_level, "middle")
        bottom_level = os.path.join(middle_level, "bottom")
        os.makedirs(bottom_level)

        top_level_2 = os.path.join(tmp_dir, "top2")
        middle_level_2 = os.path.join(top_level_2, "middle")
        os.makedirs(middle_level_2)

        # Create init files
        pathlib.Path(os.path.join(top_level, "__init__.py")).touch()
        pathlib.Path(os.path.join(top_level, "a.py")).touch()
        pathlib.Path(os.path.join(middle_level, "__init__.py")).touch()
        pathlib.Path(os.path.join(middle_level, "a.py")).touch()
        pathlib.Path(os.path.join(bottom_level, "__init__.py")).touch()
        pathlib.Path(os.path.join(bottom_level, "a.py")).touch()
        with open(os.path.join(bottom_level, "a.py"), "w") as fh:
            fh.write(task_text)
        pathlib.Path(os.path.join(middle_level_2, "__init__.py")).touch()

        # Because they have different roots
        with pytest.raises(ValueError):
            find_common_root([middle_level_2, bottom_level])

        # But now add one more init file
        pathlib.Path(os.path.join(top_level_2, "__init__.py")).touch()

        # Now it should pass
        root = find_common_root([middle_level_2, bottom_level])
        assert pathlib.Path(root).resolve() == pathlib.Path(tmp_dir).resolve()

        # Now load them
        serialization_settings = flytekit.configuration.SerializationSettings(
            project="project",
            domain="domain",
            version="version",
            env=None,
            image_config=ImageConfig.auto(img_name=DefaultImages.default_image()),
        )

        x = load_packages_and_modules(serialization_settings, pathlib.Path(root), [bottom_level])
        assert len(x) == 1
