---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(testing_agents_locally)=
# Testing agents in a local Python environment

You can test agents locally without running the backend server.

To test an agent locally, create a class for the agent task that inherits from `SyncAgentExecutorMixin` or `AsyncAgentExecutorMixin`.
These mixins can handle synchronous and asynchronous tasks, respectively, and allow flytekit to mimic FlytePropeller's behavior in calling the agent.

## Example Task

To test the example agent defined in {ref}`developing agents <developing_agents>`, copy the following code to a file called `task.py`, modifying as needed.

```python
# task.py
from collections.abc import Callable
from dataclasses import dataclass
from typing import Any, Dict, Optional

from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct

from flytekit import PythonFunctionTask
from flytekit.configuration import SerializationSettings
from flytekit.extend import TaskPlugins
from flytekit.extend.backend.base_agent import AsyncAgentExecutorMixin


@dataclass
class ExampleConfig(object):
    """
    ExampleConfig should be used to configure an ExampleTask.
    """

    environment: str


# Add `AsyncAgentExecutorMixin` or `SyncAgentExecutorMixin` to the class to tell flytekit to use the agent to run the task.
# This task extends PythonFunctionTask but you can extend different base tasks depending on your needs (ie. SQLTask)
class ExampleTask(AsyncAgentExecutorMixin, PythonFunctionTask[ExampleConfig]):
    # This must match the task type defined in the agent
    _TASK_TYPE = "example"

    def __init__(
        self,
        task_config: Optional[ExampleConfig],
        task_function: Callable,
        **kwargs,
    ) -> None:
        outputs = None
        super().__init__(
            task_config=task_config,
            task_function=task_function,
            task_type=self._TASK_TYPE,
            **kwargs,
        )

    def get_custom(self, settings: SerializationSettings) -> Dict[str, Any]:
        """
        Return plugin-specific data as a serializable dictionary. This is required for your plugin to access task_template.custom.
        """
        config = {
            "environment": self.task_config.environment,
        }
        s = Struct()
        s.update(config)
        return json_format.MessageToDict(s)


# Register the Example Task into the flytekit core plugin system
TaskPlugins.register_pythontask_plugin(ExampleConfig, ExampleTask)
```

```{note}

The ExampleTask implements `get_custom` which is originally defined in the base Task object. You will need to implement
`_get_custom` if you wish to pass plugin-specific data through the task_template's `custom` field.

```

Flytekit will automatically use the agent to run the task in the local execution.
```python
# example.py
from flytekit.configuration.default_images import DefaultImages
from flytekit import task

# Import agent to trigger agent registration
from .agent import ExampleAgent
# Import task to trigger task registration and pass plugin specific config
from .task import ExampleConfig


@task(task_config=ExampleConfig(environment="dev"), container_image=DefaultImages.default_image())
def say_hello(name: str) -> str:
    print(f"Hello, {name}!")
    return f"Hello, {name}!"

```

You can run locally and test the agent with the following command:

```bash
pyflyte run example.py example --name world
Running Execution on local.
create called task_template=<FlyteLiteral(TaskTemplate) id { resource_type: TASK name: "armada_test_agent.example.say_hello" } type: "example" metadata { runtime { type: FLYTE_SDK version: "1.12.0" flavor: "python" } retries { } } interface { inputs { variables { key: "name" value { type { simple: STRING } description: "name" } } } outputs { variables { key: "o0" value { type { simple: STRING } description: "o0" } } } } custom { fields { key: "environment" value { string_value: "dev" } } } container { image: "cr.flyte.org/flyteorg/flytekit:py3.11-1.12.0" args: "pyflyte-execute" args: "--inputs" args: "/tmp/flyte-od3iwm33/raw/0991ba2310db3416e9ad85aba218c0d9/inputs.pb" args: "--output-prefix" args: "/tmp/flyte-od3iwm33/raw/0991ba2310db3416e9ad85aba218c0d9" args: "--raw-output-data-prefix" args: "/tmp/flyte-od3iwm33/raw/0991ba2310db3416e9ad85aba218c0d9/raw_output" args: "--checkpoint-path" args: "/tmp/flyte-od3iwm33/raw/0991ba2310db3416e9ad85aba218c0d9/checkpoint_output" args: "--prev-checkpoint" args: "/tmp/flyte-od3iwm33/raw/0991ba2310db3416e9ad85aba218c0d9/prev_checkpoint" args: "--resolver" args: "flytekit.core.python_auto_container.default_task_resolver" args: "--" args: "task-module" args: "armada_test_agent.example" args: "task-name" args: "say_hello" resources { } }>
get called resource_meta=ExampleMetadata(job_id='temp')
```

If it doesn't appear that your agent is running you can debug what might be going on by passing the `-v` verbose flag to `pyflyte`.

## Existing Flyte Agent Tasks

You can also run a existing Flyte Agent tasks in your Python interpreter to test the agent locally.

![](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/bigquery_task.png)


```{note}

In some cases, you will need to store credentials in your local environment when testing locally.
For example, you need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable when running BigQuery tasks to test the BigQuery agent.

```

## Databricks example
To test the Databricks agent, copy the following code to a file called `databricks_task.py`, modifying as needed.

```python
@task(task_config=Databricks(...))
def hello_spark(partitions: int) -> float:
    print("Starting Spark with Partitions: {}".format(partitions))

    n = 100000 * partitions
    sess = flytekit.current_context().spark_session
    count = (
        sess.sparkContext.parallelize(range(1, n + 1), partitions).map(f).reduce(add)
    )
    pi_val = 4.0 * count / n
    print("Pi val is :{}".format(pi_val))
    return pi_val
```

To execute the Spark task on the agent, you must configure the `raw-output-data-prefix` with a remote path.
This configuration ensures that flytekit transfers the input data to the blob storage and allows the Spark job running on Databricks to access the input data directly from the designated bucket.

```{note}
The Spark task will run locally if the `raw-output-data-prefix` is not set.
```

```bash
pyflyte run --raw-output-data-prefix s3://my-s3-bucket/databricks databricks_task.py hello_spark
```

![](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/spark_task.png)