.. _plugins_extend:

###############
Extending Flyte
###############

The core of Flyte is a container execution engine, where you can write one or more tasks and compose them together to
form a data dependency DAG, called a ``workflow``. If your work involves writing simple Python or Java tasks that can
either perform operations on their own or call out to :ref:`Supported external services <external_service_backend_plugins>`,
then there's *no need to extend Flyte*.

====================
Define a Custom Type
====================

Flyte, just like a programming language, has a core type-system, which can be extended by adding user-defined data types.
For example, Flyte supports adding support for a dataframe type from a new library, a custom user data structure, or a
grouping of images in a specific encoding.

Flytekit natively supports structured data like :py:func:`~dataclasses.dataclass` using JSON as the
representation format (see :ref:`Using Custom Python Objects <sphx_glr_auto_core_type_system_custom_objects.py>`).

Flytekit allows users to extend Flyte's type system and implement types in Python that are not representable as JSON documents. The user has to implement a :py:class:`~flytekit.extend.TypeTransformer`
class to enable the translation of type from user type to Flyte-understood type.

As an example, instead of using :py:class:`pandas.DataFrame` directly, you may want to use
`Pandera <https://pandera.readthedocs.io/en/stable/>`__ to perform validation of an input or output dataframe
(see :ref:`Basic Schema Example <sphx_glr_auto_integrations_flytekit_plugins_pandera_examples_basic_schema_example.py>`).

To extend the type system, refer to :std:ref:`advanced_custom_types`.

=====================
Add a New Task Plugin
=====================

Often you want to interact with services like:

- Databases (e.g., Postgres, MySQL, etc.)
- DataWarehouses (e.g., Snowflake, BigQuery, Redshift etc.)
- Computation (e.g., AWS EMR, Databricks etc.)

You might want this interaction to be available as a template for the open-source community or in your organization. This
can be done by creating a task plugin, which makes it possible to reuse the task's underlying functionality within Flyte
workflows.

If you want users to write code simply using the :py:func:`~flytekit.task` decorator, but want to provide the
capability of running the function as a spark job or a sagemaker training job, then you can extend Flyte's task system.

.. code-block:: python

    @task(task_config=MyContainerExecutionTask(
        plugin_specific_config_a=...,
        plugin_specific_config_b=...,
        ...
    ))
    def foo(...) -> ...:
        ...

Alternatively, you can provide an interface like this:

.. code-block:: python

    query_task = SnowflakeTask(
        query="Select * from x where x.time < {{.inputs.time}}",
        inputs=kwtypes(time=datetime),
        output_schema_type=pandas.DataFrame,
    )

    @workflow
    def my_wf(t: datetime) -> ...:
        df = query_task(time=t)
        return process(df=df)

There are two options when writing a new task plugin: you can write a task plugin as an extension in Flytekit or you
can go deeper and write a plugin in the Flyte backend.

Flytekit-Only Task Plugin
=========================

Flytekit is designed to be extremely extensible. You can add new task-types that are useful only for your use-case.
Flyte does come with the capability of extending the backend, but that is only required if you want the capability to be
extended to all users of Flyte, or there is a cost/visibility benefit of doing so.

Writing your own Flytekit plugin is simple and is typically where you want to start when enabling custom task functionality.

.. list-table::
   :widths: 50 50
   :header-rows: 1

   * - Pros
     - Cons
   * - Simple to write — implement in Python. Flyte will treat it like a container execution and blindly pass
       control to the plugin.
     - Limited ways of providing additional visibility in progress, external links, etc.
   * - Simple to publish: ``flytekitplugins`` can be published as independent libraries and they follow a simple API.
     - Has to be implemented in every language as these are SDK-side plugins only.
   * - Simple to perform testing: test locally in flytekit.
     - In case of side-effects, it could lead to resource leaks. For example, if the plugin runs a BigQuery job,
       it is possible that the plugin may crash after running the job and Flyte cannot guarantee that the BigQuery job
       will be successfully terminated.
   * -
     - Potentially expensive: in cases where the plugin runs a remote job, running a new pod for every task execution
       causes severe strain on Kubernetes and the task itself uses almost no CPUs. Also because of its stateful nature,
       using spot-instances is not trivial.
   * -
     - A bug fix to the runtime needs a new library version of the plugin.
   * -
     - Not trivial to implement resource controls, like throttling, resource pooling, etc.

User Container vs. Pre-built Container Task Plugin
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

A Flytekit-only task plugin can be a :ref:`user container <user_container>` or :ref:`pre-built container <prebuilt_container>` task plugin.

.. list-table::
   :widths: 10 50 50
   :header-rows: 1

   * -
     - User Container
     - Pre-built Container
   * - Serialization
     - At serialization time, a Docker container image is required. The assumption is that this Docker image has the task code.
     - The Docker container image is hardcoded at serialization time into the task definition by the author of that task plugin.
   * - Serialization
     - The serialized task contains instructions to the container on how to reconstitute the task.
     - Serialized task should contain all the information needed to run that task instance (but not necessarily to reconstitute it).
   * - Run-time
     - When Flyte runs the task, the container is launched, and the user-given instructions recreate a Python object representing the task.
     - When Flyte runs the task, the container is launched. The container should have an executor built into it that knows how to execute the task.
   * - Run-time
     - The task object that gets serialized at compile-time is recreated using the user's code at run time.
     - The task object that gets serialized at compile-time does not exist at run time.
   * - Run-time
     - At platform-run-time, the user-decorated function is executed.
     - At platform-run-time, there is no user function, and the executor is responsible for producing outputs, given the inputs to the task.

Backend Plugin
==============

:std:ref:`Writing a Backend plugin <extend-plugin-flyte-backend>` makes it possible for users to write extensions for
FlytePropeller - Flyte's scheduling engine. This enables complete control of the visualization and availability
of the plugin.

.. list-table::
   :widths: 50 50
   :header-rows: 1

   * - Pros
     - Cons
   * - Service oriented way of deploying new plugins - strong contracts. Maintainers can deploy new versions of the backend plugin, fix bugs, without needing the users to upgrade libraries, etc.
     - Need to be implemented in GOlang.
   * - Drastically cheaper and more efficient to execute. FlytePropeller is written in GOlang and uses an event loop model. Each process of FlytePropeller can execute thousands of tasks concurrently.
     - Needs a FlytePropeller build (*currently*).
   * - Flyte guarantees resource cleanup.
     - Need to implement contract in a spec language like protobuf, OpenAPI, etc.
   * - Flyteconsole plugins (capability coming soon!) can be added to customize visualization and progress tracking of the execution.
     - Development cycle can be much slower than flytekit-only plugins.
   * - Resource controls and backpressure management is available.
     -
   * - Implement once, use in any SDK or language!
     -

=======
Summary
=======

.. mermaid::

    flowchart LR
        U{Use Case}
        F([Python Flytekit Plugin])
        B([Golang<br>Backend Plugin])

        subgraph WFTP[Writing Flytekit Task Plugins]
        UCP([User Container Plugin])
        PCP([Pre-built Container Plugin])
        end

        subgraph WBE[Writing Backend Extensions]
        K8S([K8s Plugin])
        WP([WebAPI Plugin])
        CP([Complex Plugin])
        end

        subgraph WCFT[Writing Custom Flyte Types]
        T([Flytekit<br>Type Transformer])
        end

        U -- Light-weight<br>Extensions --> F
        U -- Performant<br>Multi-language<br>Extensions --> B
        U -- Specialized<br>Domain-specific Types --> T
        F -- Require<br>user-defined<br>container --> UCP
        F -- Provide<br>prebuilt<br>container --> PCP
        B --> K8S
        B --> WP
        B --> CP

        style WCFT fill:#eee,stroke:#aaa
        style WFTP fill:#eee,stroke:#aaa
        style WBE fill:#eee,stroke:#aaa
        style U fill:#fff2b2,stroke:#333
        style B fill:#EAD1DC,stroke:#333
        style K8S fill:#EAD1DC,stroke:#333
        style WP fill:#EAD1DC,stroke:#333
        style CP fill:#EAD1DC,stroke:#333

Use the flow-chart above to point you to one of these examples:
