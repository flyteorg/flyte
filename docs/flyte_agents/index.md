---
# override the toc-determined page navigation order
prev-page: getting_started/extending_flyte
prev-page-title: Extending Flyte
---

(flyte_agents)=
# Flyte agents

In Flyte, an agent is a long-running, stateless service powered by a gRPC server. Each agent service is a Kubernetes deployment that receives gRPC requests from FlytePropeller when users trigger a particular type of task (for example, the BigQuery agent handles BigQuery tasks). The agent service then initiates a job with the appropriate external service. You can create different agent services that host different agents, e.g., a production and a development agent service.

:::{figure} https://i.ibb.co/vXhBDjP/Screen-Shot-2023-05-29-at-2-54-14-PM.png
:alt: Agent Service
:class: with-shadow
:::

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Using agents in tasks <using_agents_in_tasks>`
  - Learn how to use agents to efficiently communicate with external services in your tasks.
* - {doc}`Creating an agent <creating_an_agent>`
  - If the agent you need doesn't exist, follow these steps to create it.
* - {doc}`Testing agents locally <testing_agents_locally>`
  - Whether using an existing agent or creating a new one, you can test it locally without needing to configure or run the backend server.
* - {doc}`Configuring your Flyte deployment for agents <running_a_workflow_locally>`
  - Once you have tested an agent locally and want to use it in production, you must configure your Flyte deployment for the agent.
```

```{toctree}
:maxdepth: -1
:hidden:

using_agents_in_tasks
creating_an_agent
testing_agents_locally
configuring_your_flyte_deployment_for_agents
```