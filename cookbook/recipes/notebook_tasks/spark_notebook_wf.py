from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

from flytekit.sdk.types import Types
from flytekit.sdk.tasks import inputs, outputs

from flytekit.sdk.workflow import workflow_class, Input
from flytekit.contrib.notebook.tasks import spark_notebook
from flytekit.models.task import  Container, Resources


interactive_spark = spark_notebook(notebook_path="./notebook-task-examples/spark-notebook.ipynb",
                                          inputs=inputs(partitions=Types.Integer),
                                          outputs=outputs(pi=Types.Float),
                                        )

@workflow_class
class FlyteNotebookSparkWorkflow(object):
    partitions = Input(Types.Integer, default=100)
    out1 = interactive_spark(partitions=partitions)
