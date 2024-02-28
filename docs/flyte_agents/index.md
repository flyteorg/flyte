---
# override the toc-determined page navigation order
prev-page: getting_started/extending_flyte
prev-page-title: Extending Flyte
---

(flyte_agents_guide)=
# Flyte agents

Flyte agents are long-running, stateless services that receive execution requests via gRPC and initiate jobs with appropriate external or internal services. They enable two key workflows: asynchronously launching jobs on hosted platforms (e.g. Databricks or Snowflake) and calling external synchronous services, such as access control, data retrieval, and model inferencing.

Each agent service is a Kubernetes deployment that receives gRPC requests from FlytePropeller when users trigger a particular type of task (for example, the BigQuery agent handles BigQuery tasks). The agent service then initiates a job with the appropriate service. Since agents can be spawned in process, they allow for running all services locally as long as the connection secrets are available. Moreover, agents use a protobuf interface, thus can be implemented in any language, enabling flexibility, reuse of existing libraries, and simpler testing.

You can create different agent services that host different agents, e.g., a production and a development agent service:

:::{figure} https://i.ibb.co/vXhBDjP/Screen-Shot-2023-05-29-at-2-54-14-PM.png
:alt: Agent Service
:class: with-shadow
:::

(using_agents_in_tasks)=
## Using agents in tasks

If you need to connect to an external service in your workflow, we recommend using the corresponding agent rather than a web API plugin. Agents are designed to be scalable and can handle large workloads efficiently, and decrease load on FlytePropeller, since they run outside of it. You can also test agents locally without having to change the Flyte backend configuration, streamlining development.

For a list of agents you can use in your tasks and example usage for each, see the [Integrations](https://docs.flyte.org/en/latest/flytesnacks/integrations.html#agents) documentation.

## Table of contents

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Developing agents <developing_agents>`
  - If the agent you need doesn't exist, follow these steps to create it.
* - {doc}`Testing agents locally <testing_agents_locally>`
  - Whether using an existing agent or developing a new one, you can test the agent locally without needing to configure your Flyte deployment.
* - {doc}`Enabling agents in your Flyte deployment <enabling_agents_in_your_flyte_deployment>`
  - Once you have tested an agent locally and want to use it in production, you must configure your Flyte deployment for the agent.
```

```{toctree}
:maxdepth: -1
:hidden:

developing_agents
testing_agents_locally
enabling_agents_in_your_flyte_deployment
```
