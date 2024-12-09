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

## Table of contents

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Testing agents locally <testing_agents_in_a_local_python_environment>`
  - Whether using an {ref}`existing agent <flyte_agents>` or developing a new one, you can quickly test the agent in local Python environment without needing to configure your Flyte deployment.
* - {doc}`Enabling agents in your Flyte deployment <enabling_agents_in_your_flyte_deployment>`
  - After you have tested an {ref}`existing agent <flyte_agents>` in a local Python environment, you must configure your Flyte deployment for the agent to use it in production.
* - {doc}`Developing agents <developing_agents>`
  - If the agent you need doesn't exist, follow these steps to create a new agent.
* - {doc}`Testing agents in a local development cluster <testing_agents_in_a_local_development_cluster>`
  - After developing your new agent and testing it in a local Python environment, you can test it in a local development cluster to ensure it works well remotely.
* - {doc}`Deploying agents to the Flyte sandbox <deploying_agents_to_the_flyte_sandbox>`
  - Once you have tested your new agent in a local development cluster and want to use it in production, you should test it in the Flyte sandbox.
* - {doc}`Implementing the agent metadata service <implementing_the_agent_metadata_service>`
  - If you want to develop an agent server in a language other than Python (e.g., Rust or Java), you must implement the agent metadata service in your agent server.
* - {doc}`How secret works in agent <how_secret_works_in_agent>`
  - Explain how secret works in your agent server.
```

```{toctree}
:maxdepth: -1
:hidden:

testing_agents_in_a_local_python_environment
enabling_agents_in_your_flyte_deployment
developing_agents
testing_agents_in_a_local_development_cluster
deploying_agents_to_the_flyte_sandbox
implementing_the_agent_metadata_service
how_secret_works_in_agent
```
