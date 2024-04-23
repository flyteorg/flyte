(bigquery_plugin)=

# BigQuery plugin

```{note}

This is a legacy implementation of the BigQuery integration. We recommend using the {ref}`BigQuery agent <bigquery_agent>` instead.

```

## Installation

To install the BigQuery plugin, run the following command:

```{eval-rst}
.. prompt:: bash

    pip install flytekitplugins-bigquery
```

This plugin is purely a spec. Since SQL is completely portable, there is no need to build a Docker container.

## Example usage

For a usage example, see the {ref}`BigQuery example query<bigquery_plugin_example>` page.

## Flyte deployment configuration

BigQuery plugins are [enabled in FlytePropeller's config](https://docs.flyte.org/en/latest/deployment/plugin_setup/gcp/bigquery.html#deployment-plugin-setup-gcp-bigquery).

To run the BigQuery plugin on a Flyte cluster, you must configure it in your Flyte deployment. For more information, see the {ref}`BigQuery plugin setup guide <deployment-plugin-setup-gcp-bigquery>`.

```{toctree}
:maxdepth: -1
:hidden:

bigquery_plugin_example

```
