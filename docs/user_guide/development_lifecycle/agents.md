---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(extend-agent-service)=

# Agents

```{eval-rst}
.. tags:: Extensibility, Contribute, Intermediate
```

:::{note}
This is an experimental feature, which is subject to change the API in the future.
:::

## What is an agent?

In Flyte, an agent is a long-running stateless service that can be used to execute tasks. It reduces the overhead of creating a pod for each task.
In addition, it's easy to scale up and down the agent service based on the workload. Agent services are designed to be language-agnostic.
For now, we only support Python agent, but we may support other languages in the future.

Agent is designed to run a specific type of task. For example, you can create a BigQuery agent to run BigQuery task. Therefore, if you create a new type of task, you can
either run the task in the pod, or you can create a new agent to run it. You can determine how the task will be executed in the FlytePropeller configMap.

Key goals of the agent service include:

- Support for communication with external services: The focus is on enabling agents that seamlessly interact with external services.
- Independent testing and private deployment: Agents can be tested independently and deployed privately, providing flexibility and control over development.
- Flyte Agent usage in local development: Users, especially in `flytekit` and `unionml`, can leverage backend agents for local development, streamlining the development process.
- Language-agnostic: Agents can be authored in any programming language, allowing users to work with their preferred language and tools.
- Scalability: Agents are designed to be scalable, ensuring they can handle large-scale workloads effectively.
- Simple API: Agents offer a straightforward API, making integration and usage straightforward for developers.

## Why do we need an agent service?

Without agents, people need to implement a backend plugin in the propeller. The backend plugin is responsible for
creating a CRD and submitting a http request to the external service. However, it increases the complexity of flytepropeller, and
it's hard to maintain the backend plugin. For example, if we want to add a new plugin, we need to update and compile
flytepropeller, and it's also hard to test. In addition, the backend plugin is running in flytepropeller itself, so it
increases the load of the flytepropeller engine.

Furthermore, implementing backend plugins can be challenging, particularly for data scientists and ML engineers who may lack proficiency in
Golang. Additionally, managing performance requirements, maintenance, and development can be burdensome.
To address these issues, we introduced the "Agent Service" in Flyte. This system enables rapid plugin
development while decoupling them from the core flytepropeller engine.

## Overview

The Flyte agent service is a Python-based agent registry powered by a gRPC server. It allows users and flytepropeller
to send gRPC requests to the registry for executing jobs such as BigQuery and Databricks. Each Agent service is a Kubernetes
deployment. You can create two different Agent services hosting different Agents. For example, you can create one production
agent service and one development agent service.

:::{figure} https://i.ibb.co/vXhBDjP/Screen-Shot-2023-05-29-at-2-54-14-PM.png
:alt: Agent Service
:class: with-shadow
:::

## How to register a new agent

### Flytekit interface specification

To register a new agent, you can extend the `AgentBase` class in the flytekit backend module. Implementing the following three methods is necessary, and it's important to ensure that all calls are idempotent:

- `create`: This method is used to initiate a new task. Users have the flexibility to use gRPC, REST, or an SDK to create a task.
- `get`: This method allows retrieving the job Resource (jobID or output literal) associated with the task, such as a BigQuery Job ID or Databricks task ID.
- `delete`: Invoking this method will send a request to delete the corresponding job.

```python
from flytekit.extend.backend.base_agent import AgentBase, AgentRegistry
from dataclasses import dataclass
import requests

@dataclass
class Metadata:
    # you can add any metadata you want, propeller will pass the metadata to the agent to get the job status.
    # For example, you can add the job_id to the metadata, and the agent will use the job_id to get the job status.
    # You could also add the s3 file path, and the agent can check if the file exists.
    job_id: str

class CustomAgent(AgentBase):
    def __init__(self, task_type: str):
        # Each agent should have a unique task type. Agent service will use the task type to find the corresponding agent.
        self._task_type = task_type

    def create(
        self,
        context: grpc.ServicerContext,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
    ) -> TaskCreateResponse:
        # 1. Submit the task to the external service (BigQuery, DataBricks, etc.)
        # 2. Create a task metadata such as jobID.
        # 3. Return the task metadata, and keep in mind that the metadata should be serialized to bytes.
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

Here is an example of [BigQuery Agent](https://github.com/flyteorg/flytekit/blob/9977aac26242ebbede8e00d476c2fbc59ac5487a/plugins/flytekit-bigquery/flytekitplugins/bigquery/agent.py#L35) implementation.

### How to test the agent

Agent can be tested locally without running backend server. It makes the development of the agent easier.

The task inherited from AsyncAgentExecutorMixin can be executed locally, allowing flytekit to mimic the propeller's behavior to call the agent.
In some cases, you should store credentials in your local environment when testing locally.
For example, you need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable when testing the BigQuery task.
After setting up the CREDENTIALS, you can run the task locally. Flytekit will automatically call the agent to create, get, or delete the task.

```python
bigquery_doge_coin = BigQueryTask(
    name=f"bigquery.doge_coin",
    inputs=kwtypes(version=int),
    query_template="SELECT * FROM `bigquery-public-data.crypto_dogecoin.transactions` WHERE version = @version LIMIT 10;",
    output_structured_dataset_type=StructuredDataset,
    task_config=BigQueryConfig(ProjectID="flyte-test-340607")
)
```

Task above task as an example, you can run the task locally and test agent with the following command:

```bash
pyflyte run wf.py bigquery_doge_coin --version 10
```

### Build a new image

The following is a sample Dockerfile for building an image for a flyte agent.

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

### Update FlyteAgent

1. Update the FlyteAgent deployment's [image](https://github.com/flyteorg/flyte/blob/c049865cba017ad826405c7145cd3eccbc553232/charts/flyteagent/templates/agent/deployment.yaml#L26)
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
    # By default, all the request will be sent to the default agent.
    defaultAgent:
      endpoint: "dns:///flyteagent.flyte.svc.cluster.local:8000"
      insecure: true
      timeouts:
        GetTask: 200ms
      defaultTimeout: 50ms
    agents:
      custom_agent:
        endpoint: "dns:///custom-flyteagent.flyte.svc.cluster.local:8000"
        insecure: false
        defaultServiceConfig: '{"loadBalancingConfig": [{"round_robin":{}}]}'
        timeouts:
          GetTask: 100ms
        defaultTimeout: 20ms
    agentForTaskTypes:
      # It will override the default agent for custom_task, which means propeller will send the request to this agent.
      - custom_task: custom_agent
```

3. Restart the FlytePropeller

```
kubectl rollout restart deployment flytepropeller -n flyte
```
