from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

from flytekit.sdk.types import Types
from flytekit.sdk.tasks import inputs, outputs

from flytekit.sdk.workflow import workflow_class, Input
from flytekit.contrib.notebook.tasks import python_notebook

# The path
interactive_python = python_notebook(notebook_path="./notebook-task-examples/python-notebook.ipynb",
                                          inputs=inputs(pi=Types.Float),
                                          outputs=outputs(out=Types.Float),
                                          cpu_request="1",
                                          memory_request="1G"
                                        )

@workflow_class
class FlyteNotebookWorkflow(object):
    out2 = interactive_python(pi=3.14)
