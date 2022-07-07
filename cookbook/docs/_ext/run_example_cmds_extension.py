import inspect
from pathlib import Path
from typing import NamedTuple

from docutils import nodes
from docutils.statemachine import StringList, string2lines
from sphinx.util.docutils import SphinxDirective


TEMPLATE = """
.. dropdown:: :fa:`play` Run this example

   You can run this example script with the ``pyflyte run`` CLI or ``FlyteRemote``:

   .. admonition:: Prerequisites
      :class: important

      Make sure you have your :ref:`development environment <env_setup>` ready:

   .. prompt:: bash

      cd cookbook/{example_root_dir}

   .. tabbed:: pyflyte run

      In a terminal session, run:

      .. literalinclude:: /../{pyflyte_run_path}
         :language: bash

   .. tabbed:: FlyteRemote API

      Execute the following code in a python session or jupyter notebook:

      .. literalinclude:: /../{flyte_remote_path}
         :language: python

"""


TestScriptPaths = NamedTuple(
    "TestScriptPaths",
    [
        ("example_root_dir", str),
        ("pyflyte_run_path", str),
        ("flyte_remote_path", str),
    ]
)


class RunExampleCmds(SphinxDirective):
    has_content = True

    def run(self) -> list:
        return [self.parse_rst_template()]

    def get_test_script_paths(self) -> TestScriptPaths:
        full_source_path = self.get_source_info()[0]
        rel_path = Path(full_source_path.split("/auto/")[-1])
        run_dir = rel_path.parent / "_run"
        return TestScriptPaths(
            str(rel_path.parent),
            str(run_dir / f"run_{rel_path.stem}.sh"),
            str(run_dir / f"run_{rel_path.stem}.py"),
        )

    def parse_rst_template(self):
        """Parse rst template for run commands into docutils nodes.

        This was adapted from this example: https://www.bustawin.com/sphinx-extension-devel/
        """
        test_script_paths = self.get_test_script_paths()
        container = nodes.container("")
        template = inspect.cleandoc(
            TEMPLATE.format(
                example_root_dir=test_script_paths.example_root_dir,
                pyflyte_run_path=test_script_paths.pyflyte_run_path,
                flyte_remote_path=test_script_paths.flyte_remote_path,
            )
        )
        self.state.nested_parse(StringList(string2lines(template)), 0, container)
        return container


def setup(app: object) -> dict:
    app.add_directive("run-example-cmds", RunExampleCmds)
