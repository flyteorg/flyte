---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(creating_agent)=
# Creating an agent

The Flyte agent framework enables rapid agent development, since agents are decoupled from the core FlytePropeller engine. Agents can be written in Python, easing development for data scientists. Agents can be tested independently and deployed privately, making maintenance easier and giving you more flexibility and control over development.

If you need to create a new type of task, we recommend creating a new agent to run it rather than running the task in a pod. After testing the new agent, you can update your FlytePropeller configMap to specify the type of task that the agent should run.

:::{note}

While agents can be written in any programming language, we currently only support Python agents. We may support other languages in the future.

:::

## Flytekit interface specification

To create a new agent, extend the `AgentBase` class in the `flytekit.backend` module and implement `create`, `get`, and `delete` methods. All calls must be idempotent.

- `create`: This method is used to initiate a new task. Users have the flexibility to use gRPC, REST, or an SDK to create a task.
- `get`: This method allows retrieving the job Resource (jobID or output literal) associated with the task, such as a BigQuery Job ID or Databricks task ID.
- `delete`: Invoking this method will send a request to delete the corresponding job.

```python
from flytekit.extend.backend.base_agent import AgentBase, AgentRegistry
from dataclasses import dataclass
import requests

@dataclass
class Metadata:
    # FlytePropeller will pass the metadata specified in this class to the agent.
    # For example, if you add job_id to the metadata, the agent will use the job_id to get the job status.
    # If you add s3 file path, the agent will check if the file exists.
    job_id: str

class CustomAgent(AgentBase):
    def __init__(self, task_type: str):
        # Each agent should have a unique task type.
        # The Flyte agent service will use the task type
        # to find the corresponding agent.
        self._task_type = task_type

    def create(
        self,
        context: grpc.ServicerContext,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
    ) -> TaskCreateResponse:
        # 1. Submit the task to the external service (BigQuery, DataBricks, etc.)
        # 2. Create metadata for the task, such as jobID.
        # 3. Return the metadata, serialized to bytes.
        res = requests.post(url, json=data)
        return CreateTaskResponse(resource_meta=json.dumps(asdict(Metadata(job_id=str(res.job_id)))).encode("utf-8"))

    def get(self, context: grpc.ServicerContext, resource_meta: bytes) -> TaskGetResponse:
        # 1. Deserialize the metadata.
        # 2. Use the metadata to get the job status.
        # 3. Return the job status.
        metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
        res = requests.get(url, json={"job_id": metadata.job_id})
        return GetTaskResponse(resource=Resource(state=res.state)

    def delete(self, context: grpc.ServicerContext, resource_meta: bytes) -> TaskDeleteResponse:
        # 1. Deserialize the metadata.
        # 2. Use the metadata to delete the job.
        # 3. If failed to delete the job, add the error message to the grpc context.
        #   context.set_code(grpc.StatusCode.INTERNAL)
        #   context.set_details(f"failed to create task with error {e}")
        try:
            metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
            requests.delete(url, json={"job_id": metadata.job_id})
        except Exception as e:
            logger.error(f"failed to delete task with error {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"failed to delete task with error {e}")
        return DeleteTaskResponse()

# To register the custom agent
AgentRegistry.register(CustomAgent())
```

For an example implementation, see the [BigQuery Agent](https://github.com/flyteorg/flytekit/blob/9977aac26242ebbede8e00d476c2fbc59ac5487a/plugins/flytekit-bigquery/flytekitplugins/bigquery/agent.py#L35).
