.. _divedeep-data-management:

#################################
Understand How Flyte Handles Data
#################################

.. tags:: Basic, Glossary, Design

Types of Data
=============

In Flyte, data is categorized into metadata and raw data to optimize data handling and improve performance and security.

* **Metadata**: Small values, like integers and strings, are treated as "stack parameters" (passed by value). This metadata is globally accessible to Flyte components (FlytePropeller, FlyteAdmin, and other running pods/jobs). Each entry is limited to 10MB and is passed directly between tasks. On top of that, metadata allow in-memory computations for branches, partial outputs, and composition of multiple outputs as input for other tasks.

* **Raw data**: Larger data, such as files and dataframes, are treated as "heap parameters" (passed by reference). Flyte stores raw data in an object store (e.g., S3), uploading it on first use and passing only a reference thereafter. Tasks can then access this data via Flyte’s automated download or streaming, enabling efficient access to large datasets without needing to transfer full copies.

*Source code reference for auto-offloading value sizes limitation:*

.. raw:: html

   <a href="https://github.com/flyteorg/flyte/blob/6c4f8dbfc6d23a0cd7bf81480856e9ae1dfa1b27/flytepropeller/pkg/controller/config/config.go#L184-L192">View source code on GitHub</a>

Data Flow and Security
~~~~~~~~~~~~~~~~~~~~~~

Flyte’s data separation avoids bottlenecks and security risks:

* **Metadata** remains within Flyte’s control plane, making it accessible through the Flyte Console or CLI.
* **Raw Data** is accessible only by tasks, stored securely in an external blob store, preventing Flyte’s control plane from directly handling large data files.

Moreover, a unique property of this separation is that all meta values are read by FlytePropeller engine and available on the FlyteConsole or CLI from the control plane.

Example
~~~~~~~

Consider a basic Flyte task:

.. code-block:: python

    @task
    def my_task(m: int, n: str, o: FlyteFile) -> pd.DataFrame:
       ...


In this task, ``m``, ``n``, and ``o`` are inputs: ``m`` (int) and ``n`` (str) are simple types, while ``o`` is a large, arbitrarily sized file.
Flyte treats each differently:

* Metadata: Small values like ``m`` and ``n`` are inlined within Flyte’s metadata and passed directly between tasks.
* Raw data: Objects like ``o`` and the output pd.DataFrame are offloaded to an object store (e.g., S3), with only references retained in metadata.

Flytekit TypeTransformers make it possible to use complex objects as if they are available locally, just like persistent filehandles. However, the Flyte backend only deals with the references.

Raw Data Prefix
~~~~~~~~~~~~~~~

Every task can read/write its own data files. If ``FlyteFile`` or any natively supported type like ``pandas.DataFrame`` is used, Flyte will automatically offload and download
data from the configured object-store paths. These paths are completely customizable per `LaunchPlan` or `Execution`.

* The default Rawoutput path (prefix in an object store like S3/GCS) can be configured during registration as shown in :std:ref:`flytectl_register_files`.
  The argument ``--outputLocationPrefix`` allows us to set the destination directory for all the raw data produced. Flyte will create randomized folders in this path to store the data.
* To override the ``RawOutput`` path (prefix in an object store like S3/GCS),
    you can specify an alternate location when invoking a Flyte execution, as shown in the following screenshot of the LaunchForm in FlyteConsole:

  .. image:: https://raw.githubusercontent.com/flyteorg/static-resources/9cb3d56d7f3b88622749b41ff7ad2d3ebce92726/flyte/concepts/data_movement/launch_raw_output.png

* In the sandbox, the default Rawoutput-prefix is configured to be the root of the local bucket.
    Hence Flyte will write all the raw data (reference types like blob, file, df/schema/parquet, etc.) under a path defined by the execution.


``LiteralType`` & Literals
~~~~~~~~~~~~~~~~~~~~~~~~~~

SERIALIZATION TIME
^^^^^^^^^^^^^^^^^^

When a task is declared with inputs and outputs, Flyte extracts the interface of the task and converts it to an internal representation called a :std:ref:`ref_flyteidl.core.typedinterface`.
For each variable, a corresponding :std:ref:`ref_flyteidl.core.literaltype` is created.

For example, the following Python function's interface is transformed as follows:

.. code-block:: python

    @task
    def my_task(a: int, b: str) -> FlyteFile:
        """
        Description of my function

        :param a: My input integer
        :param b: My input string
        :return: My output file
        """
        ...

.. code-block::

    interface {
    inputs {
      variables {
        key: "a"
        value {
          type {
            simple: INTEGER
          }
          description: "My input Integer"
        }
      }
      variables {
        key: "b"
        value {
          type {
            simple: STRING
          }
          description: "My input string"
        }
      }
    }
    outputs {
      variables {
        key: "o0"
        value {
          type {
            blob {
            }
          }
          description: "My output File"
        }
      }
    }
  }


RUNTIME
^^^^^^^

At runtime, data passes through Flyte using :std:ref:`ref_flyteidl.core.literal` where the values are set.
For files, the corresponding ``Literal`` is called ``LiteralBlob`` (:std:ref:`ref_flyteidl.core.blob`) which is a binary large object.
Many different objects can be mapped to the underlying `Blob` or `Struct` types. For example, an image is a Blob, a ``pandas.DataFrame`` is a Blob of type parquet, etc.

Data Movement
=============

Flyte is primarily a **DataFlow Engine**. It enables movement of data and provides an abstraction to enable movement of data between different languages.

One implementation of Flyte is the current workflow engine.

The workflow engine is responsible for moving data from a previous task to the next task. As explained previously, Flyte only deals with Metadata and not the actual Raw data.
The illustration below explains how data flows from engine to the task and how that is transferred between tasks. The medium to transfer the data can change, and will change in the future.
We could use fast metadata stores to speed up data movement or exploit locality.

Between Flytepropeller and Tasks
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/9cb3d56d7f3b88622749b41ff7ad2d3ebce92726/flyte/concepts/data_movement/flyte_data_movement.png


Between Tasks
~~~~~~~~~~~~~

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/9cb3d56d7f3b88622749b41ff7ad2d3ebce92726/flyte/concepts/data_movement/flyte_data_transfer.png


Bringing in Your Own Datastores for Raw Data
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Flytekit has a pluggable data persistence layer.
This is driven by PROTOCOL.
For example, it is theoretically possible to use S3 ``s3://`` for metadata and GCS ``gcs://`` for raw data. It is also possible to create your own protocol ``my_fs://``, to change how data is stored and accessed.
But for Metadata, the data should be accessible to Flyte control plane.

Data persistence is also pluggable. By default, it supports all major blob stores and uses an interface defined in Flytestdlib.

Practical Example
~~~~~~~~~~~~~~~~~

Let's consider a simple example where we have some tasks that needs to operate huge dataframes.

The first task reads a file from the object store, shuffles the data, saves to local disk, and passes the path to the next task.

.. code-block:: python

    @task()
    def task_remove_column(input_file: FlyteFile, column_name: str) -> FlyteFile:
        """
        Reads the input file as a DataFrame, removes a specified column, and outputs it as a new file.
        """
        input_file.download()
        df = pd.read_csv(input_file.path)

        # remove column
        if column_name in df.columns:
            df = df.drop(columns=[column_name])

        output_file_path = "data_finished.csv"
        df.to_csv(output_file_path, index=False)

        return FlyteFile(output_file_path)
       ...

The second task reads the file from the previous task, removes a column, saves to local disk, and returns the path.

.. code-block:: python

    @task()
    def task_remove_column(input_file: FlyteFile, column_name: str) -> FlyteFile:
        """
        Reads the input file as a DataFrame, removes a specified column, and outputs it as a new file.
        """
        input_file.download()
        df = pd.read_csv(input_file.path)

        # remove column
        if column_name in df.columns:
            df = df.drop(columns=[column_name])

        output_file_path = "data_finished.csv"
        df.to_csv(output_file_path, index=False)

        return FlyteFile(output_file_path)
       ...

And here is the workflow:

.. code-block:: python

    @workflow
    def wf() -> FlyteFile:
        existed_file = FlyteFile("s3://custom-bucket/data.csv")
        shuffled_file = task_read_and_shuffle_file(input_file=existed_file)
        result_file = task_remove_column(input_file=shuffled_file, column_name="County")
        return result_file
       ...

This example shows how to access an existing file in a MinIO bucket from the Flyte Sandbox and pass it between tasks with ``FlyteFile``.
When a workflow outputs a local file as a ``FlyteFile``, Flyte automatically uploads it to MinIO and provides an S3 URL for downstream tasks, no manual uploads needed. Take a look at the following:

First task output metadata:

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/9cb3d56d7f3b88622749b41ff7ad2d3ebce92726/flyte/concepts/data_movement/flyte_data_movement_example_output.png

Second task input metadata:

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/9cb3d56d7f3b88622749b41ff7ad2d3ebce92726/flyte/concepts/data_movement/flyte_data_movement_example_input.png
