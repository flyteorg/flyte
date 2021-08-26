.. great_expectations::

Great Expectations
==================

**Great Expectations** is a Python-based open-source library for validating, documenting, and profiling your data. 
It helps to maintain data quality and improve communication about data between teams.

The goodness of data validation in Great Expectations can be integrated with Flyte to validate the data moving in and out of 
the pipeline entities you may have defined in Flyte. This helps establish stricter boundaries around your data to 
ensure that everything is as you expected and data will not crash your pipelines anymore unexpectedly!

How to Define Your Integration
------------------------------

Great Expectations supports native execution of expectations against various `Datasources 
<https://docs.greatexpectations.io/docs/reference/datasources>`__, 
such as Pandas dataframes, Spark dataframes, and SQL databases via SQLAlchemy.

We're supporting two Flyte types that should suit Great Expectations' ``Datasources``:

- :py:class:`flytekit.types.file.FlyteFile`: FlyteFile represents an automatic persistence object in Flyte. 
  It can represent files in remote storage and Flyte will transparently materialize them in every task execution.
- :py:class:`flytekit.types.schema.FlyteSchema`: FlyteSchema supports tabular data, which the plugin will convert into a parquet file 
  and validate the data using Great Expectations.

.. note::
  Flyte types are added because, in Great Expectations, we have the privilege to give a non-string (Pandas/Spark DataFrame) when using a 
  :py:class:`RuntimeDataConnector <greatexpectations:great_expectations.datasource.data_connector.runtime_data_connector.RuntimeDataConnector>` 
  but not when using an 
  :py:class:`InferredAssetFilesystemDataConnector <great_expectations.datasource.data_connector.inferred_asset_filesystem_data_connector.InferredAssetFilesystemDataConnector>` 
  or a 
  :py:class:`ConfiguredAssetFilesystemDataConnector <great_expectations.datasource.data_connector.configured_asset_filesystem_data_connector.ConfiguredAssetFilesystemDataConnector>`. 
  For the latter case, with the integration of Flyte types, we can give a Pandas/Spark DataFrame or a remote URI as the dataset.

The datasources can be well-integrated with the plugin using the following two modes:

- **Flyte Task**: A Flyte task defines the task prototype that one could use within a task or a workflow to validate data using 
  Great Expectations.
- **Flyte Type**: A Flyte type helps attach the ``GreatExpectationsType`` to any dataset. 
  Under the hood, ``GreatExpectationsType`` can be assumed as a combination of Great Expectations and Flyte types where every bit of 
  data is validated against the expectations, much like the OpenAPI Spec or the gRPC validator.

You can see some nice examples in the Python code files. 

.. validation_failure::

Data Validation Failure
^^^^^^^^^^^^^^^^^^^^^^^

If the data validation fails, the plugin will raise a Great Expectations' 
:py:obj:`ValidationError <great_expectations.exceptions.ValidationError>`.

For example, this is how the error message looks like:

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

- **data_source**: Data source, in general, is the "name" we use in the Great Expectations config file. 
  When combined with the data to be validated, the data source helps Great Expectations ascertain the type of data. 
  Moreover, data source assists in building batches out of data (for validation). 
- **expectation_suite**: Defines the data validation.
- **data_connector**: Tells how the data batches have to be identified.

Optional Parameters
^^^^^^^^^^^^^^^^^^^

- **context_root_dir**: Sets the path of the great expectations config directory. 
- **checkpoint_params**: Optional :py:class:`greatexpectations:great_expectations.checkpoint.checkpoint.SimpleCheckpoint` class parameters.
- **batchrequest_config**: Additional batch request configuration parameters.
  
  - data_connector_query: Query to request a data batch
  - runtime_parameters: Parameters to be sent at run-time
  - batch_identifiers: Batch identifiers
  - batch_spec_passthrough: Reader method if your file doesn’t have an extension
- **data_asset_name**: name of the data asset (to be used for ``RuntimeBatchRequest``)
- **local_file_path**: Helpful to download the given dataset to the user-given path

.. note::
  You may always want to mention the **context_root_dir** parameter, as providing a path means no harm! 
  Moreover, **local_file_path** is essential when using ``FlyteFile`` and ``FlyteSchema``.

Plugin Installation
-------------------

To use the Great Expectations <> Flyte plugin, run the following command:

.. prompt:: bash $

    pip install flytekitplugins-great_expectations

.. note:: 
    Make sure to run the workflows in the "flytekit_plugins" directory, both locally and within the sandbox.

