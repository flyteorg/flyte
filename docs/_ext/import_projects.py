import os
import re
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List, Union

from git import Repo
from sphinx.application import Sphinx
from sphinx.config import Config

__version__ = "0.0.0"


@dataclass
class ImportProjectsConfig:
    clone_dir: str
    flytekit_api_dir: str


@dataclass
class Project:
    source: str
    dest: str
    local: bool = False
    cmd: Optional[List[Union[str, List[str]]]] = None
    docs_path: Optional[str] = None
    refresh: bool = False


def update_sys_path_for_flytekit(import_project_config: ImportProjectsConfig):
    # add flytekit to python path
    flytekit_dir = os.path.abspath(import_project_config.flytekit_api_dir)
    flytekit_src_dir = os.path.abspath(os.path.join(flytekit_dir, "flytekit"))
    plugins_dir = os.path.abspath(os.path.join(flytekit_dir, "plugins"))

    sys.path.insert(0, flytekit_src_dir)
    sys.path.insert(0, flytekit_dir)

    for possible_plugin_dir in os.listdir(plugins_dir):
        dir_path = os.path.abspath((os.path.join(plugins_dir, possible_plugin_dir)))
        plugin_path = os.path.abspath(os.path.join(dir_path, "flytekitplugins"))
        if os.path.isdir(dir_path) and os.path.exists(plugin_path):
            sys.path.insert(0, dir_path)


def import_projects(app: Sphinx, config: Config):
    """Clone projects from git or copy from local directory."""
    projects = [Project(**p) for p in config.import_projects]
    import_projects_config = ImportProjectsConfig(**config.import_projects_config)

    srcdir = Path(app.srcdir)

    for _dir in (
        import_projects_config.clone_dir,
        import_projects_config.flytekit_api_dir,
    ):
        (srcdir / _dir).mkdir(parents=True, exist_ok=True)

    for project in projects:
        if project.local:
            local_dir = srcdir / project.source
        else:
            local_dir = srcdir / import_projects_config.clone_dir / project.dest
            shutil.rmtree(local_dir, ignore_errors=True)
            Repo.clone_from(project.source, local_dir)

        local_docs_path = local_dir / project.docs_path
        dest_docs_dir = srcdir / project.dest

        if project.refresh or not dest_docs_dir.exists():
            shutil.rmtree(dest_docs_dir, ignore_errors=True)
            shutil.copytree(local_docs_path, dest_docs_dir, dirs_exist_ok=True)

        if project.cmd:
            if isinstance(project.cmd[0], list):
                for c in project.cmd:
                    subprocess.run(c)
            else:
                subprocess.run(project.cmd)

    # remove cloned directories
    shutil.rmtree(import_projects_config.clone_dir)

    # add flytekit and plugins to path so that API reference can build
    update_sys_path_for_flytekit(import_projects_config)


# should handle cases like:
# - :ref:`cookbook:label` -> :ref:`label`
# - :ref:`Text <cookbook:label>` -> :ref:`Text <label>`
INTERSPHINX_REFS_PATTERN = r"([`<])(flyte:|flytekit:|flytectl:|flyteidl:|cookbook:|idl:)"

PROTO_REF_GOOGLE = "ref_google"


def remove_intersphinx_refs_in_files(app: Sphinx, docname: str, source: List[str]):
    """
    This removes intersphinx refs from source files since we're building the
    docs in one place.
    """
    text = source[0]

    if re.search(INTERSPHINX_REFS_PATTERN, text):
        text = re.sub(INTERSPHINX_REFS_PATTERN, r"\1", text)

    # replace source file
    source[0] = text


def remove_intersphinx_in_docstrings(
    app: Sphinx, what: str, name: str, obj: str, options: dict, lines: List[str],
):
    """
    This removes intersphinx refs from docstrings since we're building the
    docs in one place.
    """
    replace = {}
    for i, text in enumerate(lines):
        if re.search(INTERSPHINX_REFS_PATTERN, text):
            text = re.sub(INTERSPHINX_REFS_PATTERN, r"\1", text)
            replace[i] = text

    for i, text in replace.items():
        lines[i] = text


def setup(app: Sphinx) -> dict:
    app.add_config_value("import_projects_config", None, False)
    app.add_config_value("import_projects", None, False)
    app.connect("config-inited", import_projects, priority=1)
    app.connect("source-read", remove_intersphinx_refs_in_files)
    app.connect("autodoc-process-docstring", remove_intersphinx_in_docstrings)
    return {
        "version": __version__,
        "parallel_read_safe": True,
        "parallel_write_safe": True,
    }
