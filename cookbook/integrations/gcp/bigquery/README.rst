BigQuery
========

.. tags:: GCP, Data, Integration, Advanced

Flyte backend can be connected with BigQuery service. Once enabled, it can allow you to query a BigQuery table.
This section will provide how to use the BigQuery Plugin using flytekit python.

Installation
------------

To use the flytekit bigquery plugin simply run the following:

.. prompt:: bash

    pip install flytekitplugins-bigquery

No Need of a dockerfile
------------------------
This plugin is purely a spec. Since SQL is completely portable, there is no need to build a Docker container.


Configuring the backend to get bigquery working
-----------------------------------------------
- BigQuery plugins are `enabled in flytepropeller's config <https://docs.flyte.org/en/latest/deployment/plugin_setup/gcp/bigquery.html#deployment-plugin-setup-gcp-bigquery>`_
