.. _presto-task-type:

=============
Presto Tasks
=============

########
NOTE
########

Please do not use this yet, as the plugin is still WIP for public use, the page will be updated when it is ready to use.

########
Presto
########
Presto is a query engine, similar to Hive, that is compatible with a lot of different data stores using a pluggable connector model. At a high level, Presto tends to run a lot faster and more efficiently than Hive. To read more about Presto, refer to `Presto official documentation`_

#####################
Query configurations
#####################


The following are various input configurations that can be passed to a Presto query

* Text statement: Presto query specification
* flytekit.common.types.schema.Schema output_schema: Schema that represents that data queried from Presto
* Text routing_group: The routing group that a Presto query should be sent to for the given environment
* Text catalog: The catalog to set for the given Presto query
* Text schema: The schema to set for the given Presto query
* dict[Text,flytekit.common.types.base_sdk_types.FlyteSdkType] task_inputs: Optional inputs to the Presto task
* bool discoverable: <check_with_luis>
* Text discovery_version: String describing the version for task discovery purposes
* int retries: Number of retries to attempt
* datetime.timedelta timeout: Specify the time out


#######
Usage
#######

The following is an example of a simple Presto query which returns 10 rows from airport sessions


.. code-block:: python
   :caption: Presto query example

   schema = Types.Schema([("a", Types.String), ("b", Types.Integer)])
   presto_task = SdkPrestoTask(
   task_inputs=inputs(ds=Types.String, rg=Types.String),
   statement="SELECT * FROM hive.city.fact_airport_sessions WHERE ds = '{{ .Inputs.ds}}' LIMIT 10",
   output_schema=schema,
   routing_group="{{ .Inputs.rg }}",
   catalog="hive",
   schema="city",
   )

   @workflow_class()
   class PrestoWorkflow(object):
    ds = Input(Types.String, required=True, help="Test string with no default")
    # routing_group = Input(Types.String, required=True, help="Test string with no default")
    p_task = presto_task(ds=ds, rg='etl')
    output_a = Output(p_task.outputs.results, sdk_type=schema)

This is a simple presto task which queries

`Flyte Workflow Demo - Presto Workflow`_ is an example Presto workflow for the above query


.. _Flyte Workflow Demo - Presto Workflow: https://github.com/lyft/flytekit/blob/master/tests/flytekit/common/workflows/presto.py
.. _Presto official documentation: https://prestodb.io/docs/current/

