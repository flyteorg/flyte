.. _divedeep-catalog:

What is Data Catalog?
=====================

.. tags:: Advanced, Design

`DataCatalog <https://github.com/flyteorg/datacatalog>`__ is a service to index parameterized, strongly-typed data artifacts across revisions. It allows clients to query artifacts based on meta information and tags.


How Flyte Memoizes Task Executions on Data Catalog
--------------------------------------------------

Flyte `memoizes task executions` by creating artifacts in DataCatalog and associating meta information regarding the execution with the artifact. Let's walk through what happens when a task execution is cached on DataCatalog.

Every task instance is represented as a DataSet:

.. code-block:: javascript

    Dataset {
       project: Flyte project the task was registered in
       domain: Flyte domain for the task execution
       name: flyte_task-<taskName>
       version: <cache_version>-<hash(input params)>-<hash(output params)>
    }

Every task execution is represented as an Artifact in the Dataset above:

.. code-block:: javascript

    Artifact {
       id: uuid
       Metadata: [executionName, executionVersion]
       ArtifactData: [List of ArtifactData]
    }


    ArtifactData {
       Name: <output-name>
       value: <offloaded storage location of the literal>
    }

To retrieve the Artifact, tag the Artifact with a hash of the input values for the memoized task execution:

.. code-block:: javascript

    ArtifactTag {
       Name: flyte_cached-<unique hash of the input values>
    }

When caching an execution, FlytePropeller will:

1. Create a dataset for the task.
2. Create an artifact that represents the execution, along with the artifact data that represents the execution output.
3. Tag the artifact with a unique hash of the input values.

To ensure that the task execution is memoized, Flyte Propeller will:

1. Compute the tag by computing the hash of the input.
2. Check if a tagged artifact exists with that hash.

   - If it exists, we have a cache hit and the Propeller can skip the task execution.
   - If an artifact is not associated with the tag, Propeller needs to run the task.
