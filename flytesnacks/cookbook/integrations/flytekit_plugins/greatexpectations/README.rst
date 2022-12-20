.. _great_expectations:

Great Expectations
==================

.. tags:: Integration, Data, DataFrame, Intermediate

**Great Expectations** is a Python-based open-source library for validating, documenting, and profiling your data.
It helps maintain data quality and improve communication about data between teams.

The goodness of data validation in Great Expectations can be integrated with Flyte to validate the data moving in and out of
the pipeline entities you may have defined in Flyte. This helps establish stricter boundaries around your data to
ensure that everything is as you expected and hence, data will not crash your pipelines anymore unexpectedly!

How to Define the Integration
-----------------------------

Great Expectations supports native execution of expectations against various `Datasources <https://docs.greatexpectations.io/docs/terms/datasource/>`__,
such as Pandas dataframes, Spark dataframes, and SQL databases via SQLAlchemy.

We're supporting two Flyte types that should suit Great Expectations' ``Datasources``:

- :py:class:`flytekit.types.file.FlyteFile`: ``FlyteFile`` represents an automatic persistence object in Flyte.
  It can represent files in remote storage and Flyte transparently materializes them in every task execution.
- :py:class:`flytekit.types.structured.StructuredDataset`: ``StructuredDataset`` supports pandas dataframes, which the plugin will convert into a parquet file and validate the data using Great Expectations.

.. note::
  Flyte types are added because, in Great Expectations, we have the privilege to give a non-string (Pandas/Spark DataFrame) when using a
  :py:class:`RuntimeDataConnector <greatexpectations:great_expectations.datasource.data_connector.runtime_data_connector.RuntimeDataConnector>`
  but not when using an
  :py:class:`InferredAssetFilesystemDataConnector <greatexpectations:great_expectations.datasource.data_connector.inferred_asset_filesystem_data_connector.InferredAssetFilesystemDataConnector>`
  or a
  :py:class:`ConfiguredAssetFilesystemDataConnector <greatexpectations:great_expectations.datasource.data_connector.configured_asset_filesystem_data_connector.ConfiguredAssetFilesystemDataConnector>`.
  For the latter case, with the integration of Flyte types, we can give a Pandas/Spark DataFrame or a remote URI as the dataset.

The datasources can be well-integrated with the plugin using the following two modes:

- **Flyte Task**: A Flyte task defines the task prototype that one could use within a task or a workflow to validate data using Great Expectations.
- **Flyte Type**: A Flyte type helps attach the ``GreatExpectationsType`` to any dataset.
  Under the hood, ``GreatExpectationsType`` can be assumed as a combination of Great Expectations and Flyte types where every data is validated against the expectations, much like the OpenAPI Spec or the gRPC validator.

.. _validation_failure:

Data Validation Failure
^^^^^^^^^^^^^^^^^^^^^^^

If the data validation fails, the plugin will raise a 
:py:class:`GreatExpectationsValidationError <greatexpectations:great_expectations.exceptions.GreatExpectationsValidationError>`.

For example, this is how the error message looks on the Flyte UI:

.. code-block:: bash

  Traceback (most recent call last):
  ...
  great_expectations.marshmallow__shade.exceptions.ValidationError: Validation failed!
  COLUMN          FAILED EXPECTATION
  passenger_count -> expect_column_min_to_be_between
  passenger_count -> expect_column_mean_to_be_between
  passenger_count -> expect_column_quantile_values_to_be_between
  passenger_count -> expect_column_values_to_be_in_set
  passenger_count -> expect_column_proportion_of_unique_values_to_be_between
  trip_distance -> expect_column_max_to_be_between
  trip_distance -> expect_column_mean_to_be_between
  trip_distance -> expect_column_median_to_be_between
  trip_distance -> expect_column_quantile_values_to_be_between
  trip_distance -> expect_column_proportion_of_unique_values_to_be_between
  rate_code_id -> expect_column_max_to_be_between
  rate_code_id -> expect_column_mean_to_be_between
  rate_code_id -> expect_column_proportion_of_unique_values_to_be_between

Plugin Parameters
-----------------

- **datasource_name**: Data source, in general, is the "name" we use in the Great Expectations config file.
  A Datasource brings together a way of interacting with data (like a database or Spark cluster) and some specific data
  (like a CSV file, or a database table). Moreover, data source assists in building batches out of data (for validation).
- **expectation_suite_name**: Defines the data validation.
- **data_connector_name**: Tells how the data batches have to be identified.

Optional Parameters
^^^^^^^^^^^^^^^^^^^

- **context_root_dir**: Sets the path of the great expectations config directory.
- **checkpoint_params**: Optional :py:class:`greatexpectations:great_expectations.checkpoint.checkpoint.SimpleCheckpoint` class parameters.
- **batch_request_config**: Additional batch request configuration parameters.

  - data_connector_query: Query to request a data batch
  - runtime_parameters: Parameters to be sent at run-time
  - batch_identifiers: Batch identifiers
  - batch_spec_passthrough: Reader method if your file doesn't have an extension
- **data_asset_name**: Name of the data asset (to be used for ``RuntimeBatchRequest``)
- **local_file_path**: Helpful to download the given dataset to the user-given path

.. note::
  You may always want to mention the **context_root_dir** parameter, as providing a path means no harm!
  Moreover, **local_file_path** is essential when using ``FlyteFile`` and ``FlyteSchema``.

Plugin Installation
-------------------

To use the Great Expectations Flyte plugin, run the following command:

.. prompt:: bash $

    pip install flytekitplugins-great_expectations

.. note::
    Make sure to run workflows from the "flytekit_plugins" directory.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://blog.flyte.org/data-quality-enforcement-using-great-expectations-and-flyte
       :type: url
       :text: Blog Post
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    An article detailing Great Expectations and Flyte integration.

.. toctree::
    :maxdepth: -1
    :caption: Contents
    :hidden:

    Blog Post <https://blog.flyte.org/data-quality-enforcement-using-great-expectations-and-flyte>
