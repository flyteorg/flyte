.. _divedeep-data-management:

########################################
Understand how Flyte handles your data
########################################

Types of Data
==============

Let us consider a simple code-sample for a python task as follows.

.. code-block::

   @task
   def my_task(m: int, n: str, o: FlyteFile) -> pd.DataFrame:
      ...

In the above sample, `m, n, o` are inputs to the task. All of them from Flyte's point of view are `data`.
`m` of type `int` and `n` of type `str` are simple primitive types, while `o` is actually an arbitrarily sized File.
The difference is in how Flyte stores and passes each of these data items.
For every task that receives and input - Flyte passes it an `Inputs Metadata` object. Which contains all the primitive or simple scalar values inlined, but
complex, large objects are offloaded and the `Metadata` simply stores a reference to this object. In the above example `m,n` are inlined while
`o` and the output `pd.DataFrame` are offloaded to an object store and reference is captured in the metadata.

Flytekit TypeTransformers make it possible to use these objects as if they are available locally - providing a feeling of persistent file handles. But, Flyte backend only deals with
the references.

Thus, primitive data types and references to large objects is called the Metadata - `Meta input` or `meta output` and the actual large object is called the `Raw data`.
A unique property of this separation is that, all `meta values` are read by FlytePropeller the engine and available on FlyteConsole or the CLI from the control plane.
The `Raw` data is not read by any of the Flyte components and hence it is possible to store it in a completely separate blob storage or alternate stores, which Flyte control plane components
do not have access to, as long as users's container/tasks can access this data.

Raw Data Prefix
~~~~~~~~~~~~~~~~
Every task can read/write its own data-files, but if users decide to use Flytefile, or any natively supported types like pandas.DataFrame, then Flyte will automatically offload and download the
data from configured object store paths. These paths are completely customizable per `LaunchPlan` or per `Execution`.

- To configure the default Raw output path (prefix in an object store like S3/GCS), you can configure it during registration as shown in :std:ref:`flytectl_register_files`. The argument ``--outputLocationPrefix`` allows you to set the destination directory for all the raw data produced. Flyte will create randomized folders in this path to store the data.
- To override the RawOutput path (prefix in an object store like S3/GCS), users can specify an alternate location when invoking a Flyte execution as shown in this screenshot of the LaunchForm in FlyteConsole

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/launch_raw_output.png

Metadata
~~~~~~~~~
Metadata in Flyte is critical to allow passing of data between tasks. It allows for performing in-memory computations for branches or passing partial outputs from a task to another task or composing outputs from multiple tasks into one input for a task.
Thus metadata is restrictred. Each individual metaoutput / input cannot be larger than 1MB. Thus if you have `List[int]` then they cannot be larger than 1MB alongwith any other input entities. In cases when you desire to pass large lists or strings between tasks, it is better to use a file abstraction.

LiteralType & Literals
~~~~~~~~~~~~~~~~~~~~~~~
**Serialization time**
When a task is declared with inputs and outputs, Flyte extracts the interface for this task and converts it to an internal representation called a :std:ref:`ref_flyteidl.core.typedinterface`.
For each variable a corresponding :std:ref:`ref_flyteidl.core.literaltype` is created. Thus as an example the following python function's interface is transformed as follows,

.. code-block:: python

    @task
    def my_task(a: int, b: str) -> FlyteFile:
        """
        Description of my function
        :param a: My input Integer
        :param b: My input string
        :return: My output File
        """
        ...


The converted representation is as follows

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


**Runtime**
At runtime, the data is passed through Flyte using :std:ref:`ref_flyteidl.core.literal`  and the values are set. For Files, the corresponding Literal is called LiteralBlob - :std:ref:`ref_flyteidl.core.blob`, which stands for a
binary large object. Many different objects can be mapped to the underlying `Blob` or `Struct` types. For example and Image is a Blob, as pandas.DataFrame is a Blob of type parquet etc.


Data Movement
==============
Flyte is first and foremost a DataFlow Engine. It enables movement of data and provides an abstraction to enable movement of data between different languages. One implementation of Flyte is the current workflow engine.
the workflow Engine is responsible to move data from a previous task to the next task. As explained above, it only deals with the Metadata and not the actual Raw data.
The illustration below explains how data flows from the engine to the task and how that is transferred between tasks. The medium to transfer the data can change, and will change in the future. We could use faster metadata stores to speed up data  movement or exploit locality.

Between Flytepropeller and Tasks
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_data_movement.png


Between Tasks
~~~~~~~~~~~~~~

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_data_transfer.png


Bringing in your own Datastores for RawData
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Flytekit has a pluggable data persistence layer as explained in :std:ref:`data.extend:extend data persistence layer`. This is driven by the protocol.
For example it is theoretically possible to use S3 ``s3://`` for metadata and GCS ``gcs://`` for raw data. It is also possible to create your own protocol ``my_fs://`` to change how data is stored and accessed.
But, for Metadata, the data should be accessible to Flyte control plane. This is also pluggable and by default supports all major blob stores and uses an interface defined in flytestdlib `here <https://pkg.go.dev/github.com/flyteorg/flytestdlib/storage>`_.
