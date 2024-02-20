---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(creating_an_agent)=
# Creating an agent

The Flyte agent framework enables rapid agent development, since agents are decoupled from the core FlytePropeller engine. Rather than building a complete gRPC service from scratch, you can implement an agent as a Python class, easing development. Agents can be tested independently and deployed privately, making maintenance easier and giving you more flexibility and control over development.

If you need to create a new type of task, we recommend creating a new agent to run it rather than running the task in a pod. After testing the new agent, you can update your FlytePropeller configMap to specify the type of task that the agent should run.

There are two types of agents: **async** and **sync**.
* **Async agents** enable long-running jobs that execute on an external platform over time. They communicate with external services that have asynchronous APIs that support `create`, `get`, and `delete` operations. The vast majority of agents are async agents.
* **Sync agents** enable request/response services that return immediate outputs (e.g. calling an internal API to fetch data or communicating with the OpenAI API).

```{note}

While agents can be written in any programming language, we currently only support Python agents. We may support other languages in the future.

```

## Async agent interface specification

To create a new async agent, extend the `AgentBase` class in the `flytekit.backend` module and implement `create`, `get`, and `delete` methods. All calls must be idempotent.

- `create`: This method is used to initiate a new job. Users have the flexibility to use gRPC, REST, or an SDK to create a job.
- `get`: This method retrieves the job resource (jobID or output literal) associated with the task, such as a BigQuery job ID or Databricks task ID.
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

class CustomAsyncAgent(AsyncAgentBase):
    def __init__(self, task_type: str):
        # Each agent should have a unique task type.
        # The Flyte agent service will use the task type
        # to find the corresponding agent.
        self._task_type = task_type

    def create(
        self,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
        **kwargs,
    ) -> TaskCreateResponse:
        # 1. Submit the task to the external service (BigQuery, DataBricks, etc.)
        # 2. Create metadata for the task, such as jobID.
        # 3. Return the metadata, serialized to bytes.
        res = requests.post(url, json=data)
        return CreateTaskResponse(resource_meta=json.dumps(asdict(Metadata(job_id=str(res.job_id)))).encode("utf-8"))

    def get(self, resource_meta: bytes, **kwargs) -> TaskGetResponse:
        # 1. Deserialize the metadata.
        # 2. Use the metadata to get the job status.
        # 3. Return the job status.
        metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
        res = requests.get(url, json={"job_id": metadata.job_id})
        return GetTaskResponse(resource=Resource(state=res.state)

    def delete(self, resource_meta: bytes, **kwargs) -> TaskDeleteResponse:
        # 1. Deserialize the metadata.
        # 2. Use the metadata to delete the job.
        metadata = Metadata(**json.loads(resource_meta.decode("utf-8")))
        requests.delete(url, json={"job_id": metadata.job_id})
        return DeleteTaskResponse()

# To register the custom agent
AgentRegistry.register(CustomAsyncAgent())
```

For an example implementation, see the [BigQuery agent](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-bigquery/flytekitplugins/bigquery/agent.py#L43).
