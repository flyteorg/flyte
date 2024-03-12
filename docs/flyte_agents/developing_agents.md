---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(developing_agents)=
# Developing agents

The Flyte agent framework enables rapid agent development, since agents are decoupled from the core FlytePropeller engine. Rather than building a complete gRPC service from scratch, you can implement an agent as a Python class, easing development. Agents can be tested independently and deployed privately, making maintenance easier and giving you more flexibility and control over development.

If you need to create a new type of task, we recommend creating a new agent to run it rather than running the task in a pod. After testing the new agent, you can update your FlytePropeller configMap to specify the type of task that the agent should run.

```{note}

We strongly encourage you to contribute your agent to the Flyte community. To do so, follow the steps in "[Contributing to Flyte](https://docs.flyte.org/en/latest/community/contribute.html)" to add your agent to [Flytekit](https://github.com/flyteorg/flytekit/tree/master/plugins) and [create an example](https://docs.flyte.org/en/latest/flytesnacks/contribute.html) of your agent for the [Integrations](https://docs.flyte.org/en/latest/flytesnacks/integrations.html) documentation. If you have any questions, reach out to us on [Slack](https://docs.flyte.org/en/latest/community/contribute.html#).

```

There are two types of agents: **async** and **sync**.
* **Async agents** enable long-running jobs that execute on an external platform over time. They communicate with external services that have asynchronous APIs that support `create`, `get`, and `delete` operations. The vast majority of agents are async agents.
* **Sync agents** enable request/response services that return immediate outputs (e.g. calling an internal API to fetch data or communicating with the OpenAI API).

```{note}

While agents can be written in any programming language, we currently only support Python agents. We may support other languages in the future.

```

## Steps

### 1. Implement required methods

#### Async agent interface specification

To create a new async agent, extend the [`AsyncAgentBase`](https://github.com/flyteorg/flytekit/blob/master/flytekit/extend/backend/base_agent.py#L127) class and implement `create`, `get`, and `delete` methods. These methods must be idempotent.

- `create`: This method is used to initiate a new job. Users have the flexibility to use gRPC, REST, or an SDK to create a job.
- `get`: This method retrieves the job resource (jobID or output literal) associated with the task, such as a BigQuery job ID or Databricks task ID.
- `delete`: Invoking this method will send a request to delete the corresponding job.

```python
from flytekit.extend.backend.base_agent import AsyncAgentBase, AgentRegistry, Resource
from flytekit import StructuredDataset
from dataclasses import dataclass

@dataclass
class BigQueryMetadata(ResourceMeta):
    """
    This is the metadata for the job. For example, the id of the job.
    """
    job_id: str

class BigQueryAgent(AsyncAgentBase):
    def __init__(self):
        super().__init__(task_type_name="bigquery", metadata_type=BigQueryMetadata)

    def create(
        self,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
        **kwargs,
    ) -> BigQueryMetadata:
        # Submit the job to BigQuery here.
        return BigQueryMetadata(job_id=job_id, outputs={"o0": StructuredDataset(uri=result_table_uri))}

    def get(self, resource_meta: BigQueryMetadata, **kwargs) -> Resource:
        # Get the job status from BigQuery.
        return Resource(phase=res.phase)

    def delete(self, resource_meta: BigQueryMetadata, **kwargs):
        # Delete the job from BigQuery.
        ...

# To register the custom agent
AgentRegistry.register(BigQueryAgent())
```

For an example implementation, see the [BigQuery agent](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-bigquery/flytekitplugins/bigquery/agent.py#L43).

#### Sync agent interface specification

To create a new sync agent, extend the [`SyncAgentBase`](https://github.com/flyteorg/flytekit/blob/master/flytekit/extend/backend/base_agent.py#L107) class and implement a `do` method. This method must be idempotent.

- `do`: This method is used to execute the synchronous task, and the worker in Flyte will be blocked until the method returns.

```python
from flytekit.extend.backend.base_agent import SyncAgentBase, AgentRegistry, Resource

class OpenAIAgent(SyncAgentBase):
    def __init__(self):
        super().__init__(task_type_name="openai")

    def do(self, task_template: TaskTemplate, inputs: Optional[LiteralMap], **kwargs) -> Resource:
        # Convert the literal map to python value.
        ctx = FlyteContextManager.current_context()
        python_inputs = TypeEngine.literal_map_to_kwargs(ctx, inputs, literal_types=task_template.interface.inputs)
        # Call the OpenAI API here.
        return Resource(phase=phaseTaskExecution.SUCCEEDED, outputs={"o0": "Hello world!"})

AgentRegistry.register(OpenAIAgent())
```

### 2. Test the agent locally

See {doc}`"Testing agents locally" <testing_agents_locally>` to test your agent locally.

### 3. Build a new Docker image

The following is a sample Dockerfile for building an image for a Flyte agent:

```Dockerfile
FROM python:3.9-slim-buster

MAINTAINER Flyte Team <users@flyte.org>
LABEL org.opencontainers.image.source=https://github.com/flyteorg/flytekit

WORKDIR /root
ENV PYTHONPATH /root

# flytekit will autoload the agent if package is installed.
RUN pip install flytekitplugins-bigquery
CMD pyflyte serve agent --port 8000
```

:::{note}
For flytekit versions `<=v1.10.2`, use `pyflyte serve`.
For flytekit versions `>v1.10.2`, use `pyflyte serve agent`.
:::

### 4. Update FlyteAgent

1. Update the FlyteAgent deployment's [image](https://github.com/flyteorg/flyte/blob/master/charts/flyteagent/templates/agent/deployment.yaml#L35)
2. Update the FlytePropeller configmap.

```YAML
 tasks:
   task-plugins:
     enabled-plugins:
       - agent-service
     default-for-task-types:
       - bigquery_query_job_task: agent-service
       - custom_task: agent-service

 plugins:
   agent-service:
     supportedTaskTypes:
       - bigquery_query_job_task
       - default_task
       - custom_task
     # By default, all requests will be sent to the default agent.
     defaultAgent:
       endpoint: "dns:///flyteagent.flyte.svc.cluster.local:8000"
       insecure: true
       timeouts:
         GetTask: 5s
       defaultTimeout: 10s
     agents:
       custom_agent:
         endpoint: "dns:///custom-flyteagent.flyte.svc.cluster.local:8000"
         insecure: false
         defaultServiceConfig: '{"loadBalancingConfig": [{"round_robin":{}}]}'
         timeouts:
           GetTask: 5s
         defaultTimeout: 10s
     agentForTaskTypes:
       # It will override the default agent for custom_task, which means propeller will send the request to this agent.
       - custom_task: custom_agent
 ```

3. Restart the FlytePropeller

```
kubectl rollout restart deployment flytepropeller -n flyte
```

