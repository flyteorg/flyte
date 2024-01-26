import inspect
import os
import re
import shutil
import subprocess
import sys
from dataclasses import dataclass, field
from functools import partial
from pathlib import Path
from typing import Optional, List, Union

import git
from git import Repo
from docutils import nodes
from docutils.statemachine import StringList, string2lines
from sphinx.application import Sphinx
from sphinx.config import Config
from sphinx.util.docutils import SphinxDirective

__version__ = "0.0.0"


@dataclass
class ImportProjectsConfig:
    clone_dir: str
    flytekit_api_dir: str
    source_regex_mapping: dict = field(default_factory=dict)
    list_table_toc: List[str] = field(default_factory=list)
    dev_build: bool = False


@dataclass
class Project:
    name: str
    source: str
    dest: str
    local: bool = False
    cmd: Optional[List[Union[str, List[str]]]] = None
    docs_path: Optional[str] = None
    refresh: bool = False


TOC_TEMPLATE = """
```{{toctree}}
:maxdepth: 1
:hidden:
{toc}
```
"""

TABLE_TEMPLATE = """
```{{list-table}}
{content}
```
"""

VERSION_PATTERN = r"^v[0-9]+.[0-9]+.[0-9]+$"


class ListTableToc(SphinxDirective):
    """Custom directive to convert list-table into both list-table and toctree."""

    has_content = True

    def run(self) -> list:
        return [self.parse()]

    def parse(self):
        """Parses the list-table and adds a toctree"""
        toc = ""

        # finds all doc references in the form <doc/path>
        for doc in re.findall(r"<(.+)>", self.block_text):
            toc += f"\n{doc}"

        container = nodes.container("")
        toc = inspect.cleandoc(TOC_TEMPLATE.format(toc=toc))
        table = inspect.cleandoc(TABLE_TEMPLATE.format(content=self.block_text))
        content = f"{toc}\n\n{table}"

        self.state.nested_parse(StringList(string2lines(content)), 0, container)
        return container


def update_sys_path_for_flytekit(import_project_config: ImportProjectsConfig):
    # create flytekit/_version.py file manually
    with open(f"{import_project_config.flytekit_api_dir}/flytekit/_version.py", "w") as f:
        f.write(f'__version__ = "dev"')


    # add flytekit to python path
    flytekit_dir = os.path.abspath(import_project_config.flytekit_api_dir)
    flytekit_src_dir = os.path.abspath(os.path.join(flytekit_dir, "flytekit"))
    plugins_dir = os.path.abspath(os.path.join(flytekit_dir, "plugins"))

    sys.path.insert(0, flytekit_src_dir)
    sys.path.insert(0, flytekit_dir)

    # add plugins to python path
    for possible_plugin_dir in os.listdir(plugins_dir):
        dir_path = os.path.abspath((os.path.join(plugins_dir, possible_plugin_dir)))
        plugin_path = os.path.abspath(os.path.join(dir_path, "flytekitplugins"))
        if os.path.isdir(dir_path) and os.path.exists(plugin_path):
            sys.path.insert(0, dir_path)


def update_html_context(project: Project, tag: str, commit: str, config: Config):
    tag_url = "#" if tag == "dev" else f"{project.source}/releases/tag/{tag}"
    commit_url = f"{project.source}/tree/{commit}"

    config.html_context[f"{project.name}_tag"] = tag
    config.html_context[f"{project.name}_tag_url"] = tag_url
    config.html_context[f"{project.name}_commit"] = commit
    config.html_context[f"{project.name}_commit_url"] = commit_url


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

    if not hasattr(config, "html_context"):
        config.html_context = {}

    show_repo_tags = False
    for project in projects:
        if project.local:
            local_dir = srcdir / project.source
            try:
                repo = Repo(local_dir)
                show_repo_tags = True
            except git.InvalidGitRepositoryError:
                repo = None
        else:
            local_dir = srcdir / import_projects_config.clone_dir / project.dest
            shutil.rmtree(local_dir, ignore_errors=True)
            repo = Repo.clone_from(project.source, local_dir)
            show_repo_tags = True

        local_docs_path = local_dir / project.docs_path
        dest_docs_dir = srcdir / project.dest

        if repo:
            tags = sorted(
                [t for t in repo.tags if re.match(VERSION_PATTERN, t.name)],
                key=lambda t: t.commit.committed_datetime
            )
            if not tags or import_projects_config.dev_build:
                # If dev_build is specified or the tags don't exist just use the
                # current commit. This occurs when the git repo is a shallow
                # clone.
                tag_str = "dev"
                commit = str(repo.head.commit)[:7]
            else:
                tag = tags[-1]
                tag_str = str(tag)
                commit = str(tag.commit)[:7]
                repo.git.checkout(tag_str)

            update_html_context(project, tag_str, commit, config)

        if project.refresh or not dest_docs_dir.exists():
            shutil.rmtree(dest_docs_dir, ignore_errors=True)
            shutil.copytree(local_docs_path, dest_docs_dir, dirs_exist_ok=True)

        if project.cmd:
            if isinstance(project.cmd[0], list):
                for c in project.cmd:
                    subprocess.run(c)
            else:
                subprocess.run(project.cmd)

    config.html_context["show_repo_tags"] = show_repo_tags

    # remove cloned directories
    shutil.rmtree(import_projects_config.clone_dir)

    # add flytekit and plugins to path so that API reference can build
    update_sys_path_for_flytekit(import_projects_config)

    # add functions to clean up source and docstring refs
    for i, (patt, repl) in enumerate(import_projects_config.source_regex_mapping.items()):
        app.connect(
            "source-read",
            partial(replace_refs_in_files, patt, repl),
            priority=i,
        )
        app.connect(
            "autodoc-process-docstring",
            partial(replace_refs_in_docstrings, patt, repl),
            priority=i,
        )


def replace_refs_in_files(patt: str, repl: str, app: Sphinx, docname: str, source: List[str]):
    text = source[0]

    if re.search(patt, text):
        text = re.sub(patt, repl, text)

    # replace source file
    source[0] = text


def replace_refs_in_docstrings(
    patt: str, repl: str, app: Sphinx, what: str, name: str, obj: str, options: dict, lines: List[str],
):
    replace = {}
    for i, text in enumerate(lines):
        if re.search(patt, text):
            text = re.sub(patt, repl, text)
            replace[i] = text

    for i, text in replace.items():
        lines[i] = text


def add_list_table_toc(app: Sphinx, docname: str, source: List[str]):
    """This replaces list-table directives in specific documents with list-table-toc.

    This is important for automatically generating a toctree and list-table from
    a list-table.
    """
    if docname in app.config.import_projects_config["list_table_toc"]:
        text = source[0]
        text = re.sub(r"{list-table}", r"{list-table-toc}", text)
        source[0] = text


def setup(app: Sphinx) -> dict:
    app.add_config_value("import_projects_config", None, False)
    app.add_config_value("import_projects", None, False)
    app.add_directive("list-table-toc", ListTableToc)
    app.connect("config-inited", import_projects, priority=0)
    app.connect("source-read", add_list_table_toc, priority=0)

    return {
        "version": __version__,
        "parallel_read_safe": True,
        "parallel_write_safe": True,
    }
