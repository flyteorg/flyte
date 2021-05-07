.. _plugins_extend:

###########################
When & How to Extend Flyte
###########################

.. toctree::
    :maxdepth: -1
    :caption: Extending Flyte
    :hidden:
 
    extend_flyte_introduction
    auto/core/extend_flyte/custom_task_plugin
    auto/core/extend_flyte/run_custom_types
    extend_flyte_backend_plugins

.. NOTE:: These docs are still work in progress. Please read through and if you have any questions don't shy away from either filing a github issue or ping us in the Slack channel. The community loves plugins and would love to help you in any way.

The Core of Flyte is a container execution engine, where you can write one or more tasks and string them together to form a data dependency DAG - called a ``workflow``.
If your work involves writing simple python or java tasks that can either perform operations on their own or can call out to external services - then there is **NO NEED to extend FLYTE**.

But, in that case you can almost do everything using python / java or a container - So why should you even have to extend Flyte?

=================
But First - Why?
=================

Case 1: I want to use my special Types - e.g. my own DataFrame format
==========================================================================
Flyte, just like a programming language has a core type-system, but just like most languages, this type system can be extended by allowing users to add ``User defined Data types``.
A User defined data type can be something that Flyte does not really understand, but is extremely useful for a users specific needs. For example it can be a custom user structure or a grouping of images in a specific encoding.

Flytekit natively supports handling of structured data like User defined structures like DataClasses using JSON as the representation format. An example of this is available in FlyteCookbook - :std:doc:`auto_core_intermediate/custom_objects`.

For types that are not simply representable as JSON documents, Flytekit allows users to extends Flyte's type system and implement these types in Python. The user has to essentially implement a :py:class:`flytekit.extend.TypeTransformer` class to enable translation of the type from Users type to flyte understood types. As an example,
instead of using :py:class:`pandas.DataFrame` directly, you may want to use `Pandera <https://pandera.readthedocs.io/en/stable/>`_ to perform validation of an input or output dataframe. an example can be found `here <https://github.com/flyteorg/flytekit/blob/master/plugins/tests/pandera/test_wf.py#L9>`_.

To extend the type system in flytekit refer to an illustrative example found at - :std:ref:`advanced_custom_types`.


Case 2: Add a new Task Type - Flyte capability
===============================================
So often times you want to interact with a service like,

    - a Database (Postgres, MySQL, etc)
    - a DataWarehouse like (Snowflake, BigQuery, Redshift etc)
    - a computation platform like (AWS EMR, Databricks etc)

and you want this to be available like a template for all other users - open source or within your organization. This can be done by creating a task plugin.
A Task-plugin makes it possible for you or other users to use your idea natively within Flyte as this capability was built into the flyte platform.

Thus for example, if you want users to write code simply using the ``@task`` decorator, but you want to provide a capability of running the function as a spark job or a sagemaker training job - then you can extend Flyte's task system - we will refer to this as the plugin and it could be possible to do the following

.. code-block:: python

    @task(task_config=MyContainerExecutionTask(
        plugin_specific_config_a=...,
        plugin_specific_config_b=...,
        ...
    ))
    def foo(...) -> ...:
        ...


OR provide an interface like this

.. code-block:: python

    query_task = SnowflakeQuery(query="Select * from x where x.time < {{.inputs.time}}", inputs=(time=datetime), results=pandas.DataFrame)

    @workflow
    def my_wf(t: datetime) -> ...:
    df = query_task(time=t)
    return process(df=df)



===========================================================
I want to write a Task Plugin or add a new TaskType
===========================================================

Interestingly there are 2 options here. You can write a task plugin simply as an extension in flytekit, or you can go deeper and write a Plugin in the Flyte backend itself.

Flytekit only plugin
======================
An illustrative example of writing a flytekit plugin can be found at - :std:ref:`advanced_custom_task_plugin`. Flytekit plugins are simple to write and should invariably be
the first place you start at. Here

**Pros**

#. Simple to write, just implement in python. Flyte will treat it like a container execution and blindly pass control to the plugin
#. Simple to publish - flytekitplugins can be published as independent libraries and they follow a simple api.
#. Simple to perform testing - just test locally in flytekit

**Cons**

#. Limited ways of providing additional visibility in progress, or external links etc
#. Has to be implemented again in every language as these are SDK side plugins only
#. In case of side-effects, potentially of causing resource leaks. For example if the plugins runs a BigQuery Job, it is possible that the plugin may crash after running the Job and Flyte cannot guarantee that the BigQuery job wil be successfully terminated.
#. Potentially expensive - In cases where the plugin just runs a remote job - e.g how Airflow does, then running a new pod for every task execution causes severe strain on k8s and the task itself uses almost no CPUs. Also because of stateful natute, using spot-instances is not trivial.
#. A bug fix to the runtime, needs a new library version of the plugin
#. Not trivial to implement resource controls - e.g. throttling, resource pooling etc

Backend Plugin
===============

Doc on how to writed a backend plugins is coming soon. A backend plugin essentially makes it possible for users to write extensions for FlytePropeller (Flytes scheduling engine). This enables complete control on the visualization and availability of the plugin.

**Pros**

#. Service oriented way of deploying new plugins - strong contracts. Maintainers can deploy new versions of the backend plugin, fix bugs, without needing the users to upgrade Libraries etc
#. Drastically cheaper and more efficient to execute. FlytePropeller is written in Golang and uses an event loop model. Each process of FlytePropeller can execute 1000's of tasks concurrently.
#. Flyte will guarantee resource cleanup
#. Flyteconsole plugins (capability coming soon) can be added to customize visualization and progress tracking of the execution
#. Resource controls and backpressure management is available
#. Implement once, use in any SDK or language

**Cons**

#. Need to be implemented in golang
#. Needs a FlytePropeller build - *currently*
#. Need to implement contract in some spec language like protobf, openAPI etc
#. Development cycle can be much slower than flytekit only plugins


===============================================
How do I decide which path to take?
===============================================

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/extend_flyte_flowchart.png
    :alt: Ok you want to add a plugin, but which type? Follow the flowchart and then select the right next steps.


Use the conclusion of the flow-chart to refer to the right doc
================================================================

- :ref:`advanced_custom_task_plugin`
- :ref:`extend-plugin-flyte-backend`
