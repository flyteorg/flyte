.. _hive-task-type:

=============
Hive Tasks
=============

Hive tasks are an example of dynamic tasks. That is, they are a two-step task where the workflow container is first run to produce the queries, which are later executed using a Flyte plugin. This means that the text of the queries as well as the number of queries can be dynamic.

See the Hive tasks discussion in the generated API documentation for more information.


#######
Usage
#######

***********************
Basic Query Execution
***********************

The following is an example of a simple Hive query.

.. code-block:: python
   :caption: Simple Hive query example 

   @qubole_hive_task
   def generate_simple_queries(wf_params):
       q1 = "SELECT 1"
       q2 = "SELECT 'two'"
       return [q1, q2]

This is a pretty simple query.  Your queries will be run on Qubole, but nothing will happen with the output.


*******************************
Query with Schema Integration
*******************************

A more common and powerful pattern is to integrate querying along with the Flyte ``Schema`` type.

.. code-block:: python
   :caption: Hive query example with Schema integration 

   @outputs(hive_results=[Types.Schema()])
   @qubole_hive_task(tags=['mytag'], cluster='flyte')
   def generate_queries(wf_params, hive_results):
       q1 = "SELECT 1"
       q2 = "SELECT 'two'"
       schema_1, formatted_query_1 = Schema.create_from_hive_query(select_query=q1)
       schema_2, formatted_query_2 = Schema.create_from_hive_query(select_query=q2)

       hive_results.set([schema_1, schema_2])
       return [formatted_query_1, formatted_query_2]
      

This does a couple things.

* Your queries will be amended by the SDK before they are executed.  That is, instead of ``SELECT 1``, the actual query that will be run will be something like ::

   CREATE TEMPORARY TABLE 1757c8c0d7a149b79f2c202c2c78b378_tmp AS SELECT 1;
   CREATE EXTERNAL TABLE 1757c8c0d7a149b79f2c202c2c78b378 LIKE 1757c8c0d7a149b79f2c202c2c78b378_tmp STORED AS PARQUET;
   ALTER TABLE 1757c8c0d7a149b79f2c202c2c78b378 SET LOCATION
      \'s3://my-s3-bucket/ec/b78e1502cef04d5db8bef64a2226f707/\';
   INSERT OVERWRITE TABLE 1757c8c0d7a149b79f2c202c2c78b378
   SELECT *
   FROM 1757c8c0d7a149b79f2c202c2c78b378_tmp;
   DROP TABLE 1757c8c0d7a149b79f2c202c2c78b378;

  When a user's query runs, it's first selected into a temporary table, and then copied from the temporary table into the permanent external table.  The external table is then dropped, which doesn't actually delete the just-queried data, but rather alleviates pressure on the Hive metastore.

* The task's output will have been bound a priori to the location that the Qubole Hive query will end up writing to, so that the rest of Flyte (downstream tasks for instance) will know about them.

###########################
Miscellaneous
###########################

****************************
Hive Execution Environment
****************************
Qubole is currently the primary execution engine for Hive queries, though it doesn't have to be.  In fact, the ``qubole_hive_task`` decorator in the SDK is a refinement on the broader ``hive_task`` decorator.

