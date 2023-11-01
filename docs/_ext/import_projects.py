import os
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from tempfile import TemporaryDirectory
from typing import Optional, List

from git import Repo
from sphinx.application import Sphinx
from sphinx.config import Config

__version__ = "0.0.0"


@dataclass
class ImportProjectsConfig:
    clone_dir: str


@dataclass
class Project:
    source: str
    dest: str
    local: bool = False
    gen_cmd: Optional[List[str]] = None
    docs_path: Optional[str] = None


def update_sys_path_for_flytekit():
    # add flytekit to python path
    flytekit_dir = os.path.abspath("./_src/flytekit")
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
    projects = [Project(**p) for p in config.import_projects]
    import_projects_config = ImportProjectsConfig(**config.import_projects_config)

    for project in projects:
        if project.local:
            local_dir = Path(app.srcdir) / project.source
        else:
            local_dir = Path(app.srcdir) / import_projects_config.clone_dir / project.dest
            if local_dir.exists():
                shutil.rmtree(local_dir)
            Repo.clone_from(project.source, local_dir)

        local_docs_path = local_dir / project.docs_path
        dest_docs_dir = Path(app.srcdir) / project.dest

        if project.gen_cmd:
            if isinstance(project.gen_cmd[0], list):
                for cmd in project.gen_cmd:
                    subprocess.run(cmd)
            else:
                subprocess.run(project.gen_cmd)

        shutil.copytree(local_docs_path, dest_docs_dir, dirs_exist_ok=True)
    
    # remove clone directories
    shutil.rmtree(import_projects_config.clone_dir)
    
    # add flytekit and plugins to path
    update_sys_path_for_flytekit()

    # TODO:
    # - need to handle intersphinx links somehow


def setup(app: Sphinx) -> dict:
    app.add_config_value("import_projects_config", None, False)
    app.add_config_value("import_projects", None, False)
    app.connect("config-inited", import_projects, priority=499)
    return {
        "version": __version__,
        "parallel_read_safe": True,
        "parallel_write_safe": True,
    }
