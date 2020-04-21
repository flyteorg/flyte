.. _presto-task-type:

=============
Presto Tasks
=============

########
Presto
########
Presto is a query engine, similar to Hive, that is compatible with a lot of different data stores using a pluggable
connector model. At a high level, Presto tends to run a lot faster and more efficiently than Hive. To read more about
Presto, refer to `Presto official documentation`_

######################
Presto task parameters
######################

The following are various configurations that can be set for a Presto task

* ``dict[Text,flytekit.common.types.base_sdk_types.FlyteSdkType] task_inputs``: Optional inputs to the Presto task
* ``Text statement``: The templated Presto statement to execute.
* ``flytekit.common.types.schema.Schema output_schema``: Schema representation of the resulting data that was queried from Presto
* ``Text routing_group``: The routing group that a Presto query should be sent to for the given environment
* ``Text catalog``: The Presto catalog for the given query
* ``Text schema``: The Presto schema for the given query

#######
Usage
#######

The following is a simple example of a Presto Flyte task. Note that unlike other tasks, there is no annotation that
denotes this as a Presto task. Instead, you  use the `SdkPrestoTask` class directly.

.. code-block:: python
   :caption: Simple Presto task example

   from __future__ import absolute_import

   from flytekit.sdk.tasks import inputs
   from flytekit.sdk.types import Types
   from flytekit.sdk.workflow import workflow_class, Input, Output
   from flytekit.common.tasks.presto_task import SdkPrestoTask

   schema = Types.Schema([("a", Types.String), ("b", Types.Integer)])

   presto_task = SdkPrestoTask(
       task_inputs=inputs(ds=Types.String, rg=Types.String),
       statement="SELECT * FROM hive.foo.bar WHERE ds = '{{ .Inputs.ds}}' LIMIT 10",
       output_schema=schema,
       routing_group="{{ .Inputs.rg }}",
       # catalog="hive",
       # schema="foo",
   )


   @workflow_class()
   class PrestoWorkflow(object):
       ds = Input(Types.String, required=True, help="Test string with no default")
       # routing_group = Input(Types.String, required=True, help="Test string with no default")

       p_task = presto_task(ds=ds, rg='etl')

       output_a = Output(p_task.outputs.results, sdk_type=schema)


This is another example usage of the Presto task, where each task is generated dynamically:

.. code-block:: python
   :caption: Dynamic Presto task example

   from __future__ import absolute_import

   from flytekit.sdk.tasks import inputs, outputs, dynamic_task
   from flytekit.sdk.types import Types
   from flytekit.sdk.workflow import workflow_class, Input, Output
   from flytekit.common.tasks.presto_task import SdkPrestoTask

   schema = Types.Schema([("session_id", Types.String), ("num_rides_completed", Types.Integer)])

   query_template = """
       SELECT
         session_id, num_rides_completed
       FROM
         hive.city.fact_airport_sessions
       WHERE
         ds = '{{ .Inputs.ds}}'
       LIMIT 10
   """

   presto_task = SdkPrestoTask(
       task_inputs=inputs(ds=Types.String, rg=Types.String),
       statement=query_template,
       output_schema=schema,
       routing_group="{{ .Inputs.rg }}",
       # catalog="hive",
       # schema="city",
   )


   @outputs(presto_results=[schema])
   @dynamic_task
   def multiple_presto_queries(wf_params, presto_results):
       temp = []
       for ds in ('2020-02-20', '2020-02-21', '2020-02-22'):
           x = presto_task(ds=ds, rg='etl')
           yield x
           temp.append(x.outputs.results)

       presto_results.set(temp)


   @workflow_class()
   class PrestoWorkflow(object):
       ds = Input(Types.String, required=True, help="Test string with no default")
       # routing_group = Input(Types.String, required=True, help="Test string with no default")

       p_task = presto_task(ds=ds, rg='etl')
       presto_dynamic = multiple_presto_queries()

       output_a = Output(p_task.outputs.results, sdk_type=schema)
       output_m = Output(presto_dynamic.outputs.presto_results, sdk_type=[schema])




.. _Presto official documentation: https://prestodb.io/docs/current/
