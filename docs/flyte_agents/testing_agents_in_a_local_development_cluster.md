---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(testing_agents_in_a_local_development_cluster)=
# Testing agents in a local development cluster

The Flyte agent service runs in a separate deployment instead of inside FlytePropeller. To test an agent server in a local development cluster, you must run both the single binary and agent server at the same time, allowing FlytePropeller to send requests to your local agent server.

## Backend plugin vs agent service architecture

To understand why you must run both the single binary and agent server at the same time, it is helpful to compare the backend plugin architecture to the agent service architecture.

### Backend plugin architecture

In this architecture, FlytePropeller sends requests through the SDK:

![image.png](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/plugin_life_cycle.png)

### Agent service architecture

With the agent service framework:
1. Flyteplugins send gRPC requests to the agent server.
2. The agent server sends requests through the SDK and returns the query data.

![image.png](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/async_agent_life_cycle.png)

## Configuring the agent service in development mode

1. Start the demo cluster in dev mode:
```bash
flytectl demo start --dev
```

2. Start the agent grpc server:
```bash
pyflyte serve agent
```

3. (Optional) Update the timeout configuration for each request to the agent deployment in the single binary YAML file:
```bash
cd flyte
vim ./flyte-single-binary-local.yaml
```
```yaml
plugins:
  # Registered Task Types
  agent-service:
    defaultAgent:
      endpoint: "localhost:8000" # your grpc agent server port
      insecure: true
      timeouts:
        # CreateTask, GetTask and DeleteTask are for async agents.
        # ExecuteTaskSync is for sync agents.
        CreateTask: 5s
        GetTask: 5s
        DeleteTask: 5s
        ExecuteTaskSync: 10s
      defaultTimeout: 10s
```

4. Start the Flyte server with the single binary config file:
```bash
make compile
./flyte start --config ./flyte-single-binary-local.yaml
```

5. Set up your secrets:
In the development environment, you can set up your secrets on your local machine by adding secrets to `/etc/secrets/SECRET_NAME`. 

Since your agent server is running locally rather than within Kubernetes, it can retrieve the secret from your local file system.

6. Test your agent task:
```bash
pyflyte run --remote agent_workflow.py agent_task
```

:::{note}
You must build an image that includes the plugin for the task and specify its config with the [`--image` flag](https://docs.flyte.org/en/latest/api/flytekit/pyflyte.html#cmdoption-pyflyte-run-i) when running `pyflyte run` or in an {ref}`ImageSpec <imagespec>` definition in your workflow file.
:::
