---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst
---

(testing_agents_in_the_development_environment)=
# Testing agents in the development environment

## How agent plugin service works?
Before testing agents in the development environment, let's talk about how it works.

Compare to the bigquery plugin example here's how the bigquery agent service work.

Bigquery plugin

![image.png](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/plugin_life_cycle.png)

The life cycle will be
1. FlytePropeller sends request through SDK.

Bigquery agent service

![image.png](https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/concepts/agents/async_agent_life_cycle.png)

The life cycle will be
1. FlytePlugins send grpc request to the agent server.
2. Agent server sends request through SDK and return the query data.

You can find that it is a little bit slower than the plugin example above.

But it is far more easier to be implemented and still faster than writing the service in a pod task.

## How to setup Agent Service in Development Mode?
Let's develop the powerful agent service!

1. Use the dev mode through flytectl.
```bash
flytectl demo start --dev
```

2. Start the agent grpc server.
```bash
pyflyte serve agent
```

3. Set the config in the yaml file (Bigquery for example)
```bash
cd flyte
vim ./flyte-single-binary-local.yaml
```

```yaml
tasks:
  task-plugins:
    enabled-plugins:
      - agent-service
      - container
      - sidecar
      - K8S-ARRAY
    default-for-task-types:
      - bigquery_query_job_task: agent-service
      - container: container
      - container_array: K8S-ARRAY
```
```yaml
plugins:
  # Registered Task Types
  agent-service:
    defaultAgent:
      endpoint: "dns:///localhost:8000" # your grpc agent server port
      insecure: true
      timeouts:
        GetTask: 10s
      defaultTimeout: 10s
```

4. Start the Flyte server with config yaml file
```bash
make compile
./flyte start --config ./flyte-single-binary-local.yaml
```

5. Set up your secrets
In the development environment, you can set up your secrets in local host.
- Add secrets in `/etc/secrets/SECRET_NAME`. 

Since your agent server is on your localhost, it can retrieve the secret locally.

6. Test your agent task
```bash
pyflyte run --remote agent_workflow.py agent_task
```

:::{note}
Please ensure you've built an image and specified it by `--image` flag or use ImageSpec to include the plugin for the task.
:::
