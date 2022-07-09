"""
FlyteOperator Example
---------------------

This tutorial will walk you through constructing an Airflow DAG using the ``FlyteOperator``.

.. note::
    The provider source code can be found in the `airflow-provider-flyte <https://github.com/flyteorg/airflow-provider-flyte>`__ repository.

Watch a demo of this provider below! It demonstrates an example of pulling NYC taxi data from S3, uploading it to CrateDB,
and building an XGBoost model. The extract and load steps are handled by Airflow and the machine learning task is offloaded to Flyte
using the Flyte Airflow Provider.

..  youtube:: F2JyA0O2U4M

The Airflow DAG demonstrated in the video is available `here <https://github.com/flyteorg/airflow-provider-flyte/blob/master/demo/dags/nyc_taxi.py>`__.

Environment Setup
=================

**AIRFLOW**

Astronomer's CLI is the fastest and easiest way to set up Airflow.

Download the `Astro CLI <https://github.com/astronomer/astro-cli>`__ and then initialize a new astro project.

.. code-block::

    mkdir flyte-astro-project
    cd flyte-astro-project
    astro dev init

The directory structure of ``flyte-astro-project`` would look as follows:

.. code-block::

    .
    ├── Dockerfile
    ├── README.md
    ├── airflow_settings.yaml
    ├── dags
    │   ├── example-dag-advanced.py
    │   └── example-dag-basic.py
    ├── include
    ├── packages.txt
    ├── plugins
    ├── requirements.txt
    └── tests
        └── dags
            └── test_dag_integrity.py

**FLYTE**

The `getting started tutorial <https://docs.flyte.org/en/latest/getting_started/index.html>`__ should help you with setting up Flyte.

Create an Airflow Connection
============================

Hit ``http://localhost:8080/``, give the credentials (default username and password: ``admin``), navigate to ``Connections`` and create a
Flyte connection.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/integrations/airflow/airflow_connection.png
    :alt: Airflow Connection

Click `Save` in the end.

.. note::
    Use external IP as the Flyte ``Host``. You can :std:ref:`deploy <deployment>` Flyte on an on-prem machine or on cloud.

Register Flyte Code
===================

At the Flyte end, we'll train an XGBoost model on Pima Indians Diabetes Dataset.
The source code is available `here <https://github.com/flyteorg/flytesnacks/blob/master/cookbook/case_studies/ml_training/pima_diabetes/diabetes.py>`__.

Register the example on the Flyte backend before proceeding with running the Airflow DAG.

- Configure flytectl config at ``~/.flyte/config.yaml`` to point to the relevant endpoint.

  .. code-block:: yaml

    admin:
      endpoint: dns:///<your-endpoint>
      insecure: true # Set to false to enable TLS/SSL connection.
      authType: Pkce # authType: Pkce # if using authentication or just drop this.

- Clone the `flytesnacks repository <https://github.com/flyteorg/flytesnacks>`__ and go into the ``cookbook/case_studies`` directory.

  .. code-block::

    git clone https://github.com/flyteorg/flytesnacks
    cd flytesnacks/cookbook/case_studies

- Serialize the workflow.

  .. code-block::

    pyflyte --pkgs ml_training.pima_diabetes package --image "ghcr.io/flyteorg/flytecookbook:pima_diabetes-latest" -f

- Register the workflow.

  .. code-block::

    flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz --version v1
"""

# %%
# Create an Airflow DAG
# =====================
# Place the following file under the ``dags/`` directory. You can name it ``example_dag_flyte.py``.
from datetime import datetime, timedelta

from airflow import DAG

from flyte_provider.operators.flyte import FlyteOperator
from flyte_provider.sensors.flyte import FlyteSensor

with DAG(
   dag_id="example_flyte",
   schedule_interval=None,
   start_date=datetime(2022, 1, 1),
   dagrun_timeout=timedelta(minutes=60),
   catchup=False,
) as dag:
   task = FlyteOperator(
       task_id="diabetes_predictions",
       flyte_conn_id="flyte_conn",
       project="flytesnacks",
       domain="development",
       launchplan_name="ml_training.pima_diabetes.diabetes.diabetes_xgboost_model",
       inputs={"test_split_ratio": 0.66, "seed": 5},
   )

   sensor = FlyteSensor(
       task_id="sensor",
       execution_name=task.output,
       project="flytesnacks",
       domain="development",
       flyte_conn_id="flyte_conn",
   )

   task >> sensor

# %%
# Also, add ``airflow-provider-flyte`` package to ``requirements.txt`` under the astro project.

# %%
# Run the Workflow
# ================
# - Run the command ``astro dev start``.
# - Trigger the Airflow DAG by clicking the "Trigger DAG" button on the Airflow UI.
# - Verify if Flyte execution got triggered on the FlyteConsole by going to ``http://<path>`` and navigating to the workflow page.

# %%
# That's about it! With the Flyte Airflow Provider, you get to reap the benefits of Flyte, a full-fledged machine learning orchestration service,
# as an extension to Airflow.
# For more example DAGs, refer to `this folder <https://github.com/flyteorg/airflow-provider-flyte/tree/master/flyte_provider/example_dags>`__.
