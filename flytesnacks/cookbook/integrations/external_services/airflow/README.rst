Airflow Provider
================

.. tags:: Integration, Intermediate

The ``airflow-provider-flyte`` package provides an operator, a sensor, and a hook that integrates Flyte into Apache Airflow.
``FlyteOperator`` is helpful to trigger a task/workflow in Flyte and ``FlyteSensor`` enables monitoring a Flyte execution status for completion.

The primary use case of this provider is to **scale Airflow for machine learning tasks using Flyte**.
With the Flyte Airflow provider, you can construct your ETL pipelines in Airflow and machine learning pipelines in Flyte
and use the provider to trigger machine learning or Flyte pipelines from within Airflow.

Installation
------------

.. code-block::

    pip install airflow-provider-flyte

All the configuration options for the provider are available in the provider repo's `README <https://github.com/flyteorg/airflow-provider-flyte#readme>`__.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://blog.flyte.org/scale-airflow-for-machine-learning-tasks-with-the-flyte-airflow-provider
       :type: url
       :text: Blog Post
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    An article on how to use the Flyte Airflow provider to trigger Flyte tasks from within Airflow.

.. toctree::
    :maxdepth: -1
    :caption: Contents
    :hidden:

    Blog Post <https://blog.flyte.org/scale-airflow-for-machine-learning-tasks-with-the-flyte-airflow-provider>
