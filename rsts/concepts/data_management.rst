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

Flytekit TypeTransformers make it possible to use these objects as if they are available locally - providing a feeling of persistent file handles. But, the Flyte only deals with
the references.

Thus, primitive data types and references to large objects is called the Metadata - `Meta input` or `meta output` and the actual large object is called the `Raw data`.
A unique property of this separation is that, all `meta values` are read by FlytePropeller the engine and available on FlyteConsole or the CLI from the control plane.
The `Raw` data is not read by any of the Flyte components and hence it is possible to store it in a completely separate blob storage or alternate stores, which Flyte control plane components
do not have access to, as long as users's container/tasks can access this data.

Raw Data Prefix
~~~~~~~~~~~~~~~~
Every task can read/write its own data-files, but if users decide to use Flytefile, or any natively supported types like pandas.DataFrame, then Flyte will automatically offload and download the
data from configured object store paths. These paths are completely customizable per `LaunchPlan` or per `Execution`.

- To configure the default Raw output path (prefix in an object store like S3/GCS), you can configure it during registration as shown in :std:ref:`flytectl_register_files`. The argument ``--outputLocationPrefix`` allows you to set
the destination directory for all the raw data produced. Flyte will create randomized folders in this path to store the data.

- To override the RawOutput path (prefix in an object store like S3/GCS), users can specify an alternate location when invoking a Flyte execution as shown in this screenshot of the LaunchForm in FlyteConsole

.. image:: 

Metadata
~~~~~~~~~

Data Movement
==============

Between Flytepropeller and Tasks
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_data_movement.png


Between Tasks
~~~~~~~~~~~~~~
.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_data_transfer.png
