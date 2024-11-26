(deployment-agent-setup)=

# Agent Setup

```{tags} Agent, Integration, Data, Advanced
```

To set configure your Flyte deployment for agents, see the documentation below.

:::{note}
If you are using a managed deployment of Flyte, you will need to contact your deployment administrator to configure agents in your deployment.
:::

```{list-table}
:header-rows: 0
:widths: 20 30

* - {ref}`Airflow Agent <deployment-agent-setup-airflow>`
  - Configuring your Flyte deployment for the Airflow agent
* - {ref}`ChatGPT Agent <deployment-agent-setup-chatgpt>`
  - Configuring your Flyte deployment for the ChatGPT agent.
* - {ref}`Databricks Agent <deployment-agent-setup-databricks>`
  - Configuring your Flyte deployment for the Databricks agent.
* - {ref}`Google BigQuery Agent <deployment-agent-setup-bigquery>`
  - Configuring your Flyte deployment for the BigQuery agent.
* - {ref}`MMCloud Agent <deployment-agent-setup-mmcloud>`
  - Configuring your Flyte deployment for the MMCloud agent.
* - {ref}`SageMaker Inference <deployment-agent-setup-sagemaker-inference>`
  - Deploy models and create, as well as trigger inference endpoints on SageMaker.
* - {ref}`Sensor Agent <deployment-agent-setup-sensor>`
  - Configuring your Flyte deployment for the sensor agent.
* - {ref}`Snowflake Agent <deployment-agent-setup-snowflake>`
  - Configuring your Flyte deployment for the SnowFlake agent.
* - {ref}`OpenAI Batch <deployment-agent-setup-openai-batch>`
  - Submit requests to OpenAI GPT models for asynchronous batch processing.
```

```{toctree}
:maxdepth: 1
:name: Agent setup
:hidden:

airflow
chatgpt
databricks
bigquery
mmcloud
sagemaker_inference
sensor
snowflake
openai_batch
```
