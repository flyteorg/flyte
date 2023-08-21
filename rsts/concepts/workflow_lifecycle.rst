.. _workflow-lifecycle:

#################################################################
Understand the Lifecycle of a Flyte Workflow
#################################################################

.. tags:: Basic, Design

Let's understand how Flyte's plugin machinery works and how information flows from one component to another in Flyte.

Under the hood, Flyte relies on a primitive called “Plugins”. Every task that you run on Flyte is powered by a plugin. Some of these plugins are native and guaranteed by Flyte system. These native plugins, for example, run your Flyte tasks inside a k8s pod. There are three native plugins, namely, ``Container``, ``K8sPod``, and ``Sql``.

Moreover, there are plugins that are actual extensions; they create additional infrastructure and communicate with SaaS on your behalf. Examples include :ref:`Spark <spark_task>`, :ref:`AWS Athena <AWS Athena>`, etc.

A plugin requires code to live in multiple locations.

1. Some parts of plugins logic resides in Flytekit's SDK. This let users define tasks. You can find this logic in Flytekit’s Python (https://github.com/flyteorg/flytekit/tree/master/plugins). Think of this as a client for an RPC service or a web service

2. Another big chunk of plugins logic lives in
   `Flyteplugins <https://github.com/flyteorg/flyteplugins>`__. This is a library that gets loaded into `FlytePropeller <https://github.com/flyteorg/flytepropeller>`__.
   FlytePropeller (a Kubernetes operator) loads Flyteplugins upon starting. 
   FlytePropeller is aware of the plugins and their dependency on task execution.
   However, FlytePropeller is unaware of how these plugins are executed.

------------

To better Illustrate how things work, lets take for example the “Spark”
plugin and understand what is the sequence of steps that take place for
it to work.

The Spark plugin lets a user define a task that has access to a Spark Session.
In the background Flyte will provide all the needed infrastructure such that by the time the declared task needs to run, all needed Spark infrastructure is ready and running.

1. User codes in python a task that uses Spark (See code below)

.. code:: python

   @task(
       task_config=Spark(
           spark_conf={
               "spark.driver.memory": "1000M",
               "spark.executor.instances": "2",
               "spark.driver.cores": "1",
           }
       )
   )
   def hello_spark(i: int) -> float:
    ...
    ...

As mentioned earlier some part of plugin logic lives on the SDK. In this
case think of ``Spark`` data class here as a placeholder for all the
Spark settings that we need our plugin to know. We need to pass this
data accross multiple places. This is the config that Flyte operator (Flytepropeller)
will need in order to build the needed spark cluster. ``Spark`` class also tells
Flytekit’s SDK that this task will run as a ``PysparkFunctionTask``
because ``task_config`` points to a ``Spark`` object instance, this is
clearly illustrated `in spark plugin registration step run in the
background <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-spark/flytekitplugins/spark/task.py#L129>`__

2. Once the user has finished writing needed Workflows. A packaging step
   is needed before user can run the workflows. This packaging step
   transforms workflows and tasks we described in python into a Protobuf
   representation. This protobuf representation is used by Flyte accross its multiple codebases. For
   further details on the protobuf representation check `FlyteIdl
   repository <https://github.com/flyteorg/flyteidl>`__ . Package step is carried out by the sdk tooling you are using.

This serialization step will transform our ``hello_spark`` task into a
protobuf representation. It will also transform other tasks, workflows
and launch plans to a protobuf representation.

Our ``hello_spark`` protobuf representation will look as below. A Task
is serialized as a
`TaskTemplate <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L102>`__
as defined in ``FlyteIDL``.

::

   Id:  Task, "example.example.hello_spark" 
   Type: "Spark"
   Metadata: 
       runtime: 
           type: FLYTE_SDK
           version: 1.0.3
           flavor: python
           
   interface:
       inputs:
           i : 
             type : simple:Integer
             description: "i"
       outputs: 
          o0: 
             type: FLOAT
             description: o0
   custom:
       executorpath: "/opt/venv/bin/python3"
       mainApplicationFile: /opt/venv/bin/entrypoint.py
       sparkConf: 
          spark.driver.cores: 1
          spark.executor.instances: 2
          spark.driver.memory: 1000M
       

   Container:
       image: "hello_world:1"
       args: 
        [
          "pyflyte-execute"
          "--inputs"
          "{{.input}}"
          "--output-prefix"
          "{{.outputPrefix}}"
          "--raw-output-data-prefix"
          "{{.rawOutputDataPrefix}}"
          "--checkpoint-path"
          "{{.checkpointOutputPrefix}}"
          "--prev-checkpoint"
          "{{.prevCheckpointPrefix}}"
          "--resolver"
          "flytekit.core.python_auto_container.default_task_resolver"
          "--"
          "task-module"
          "example.example"
          "task-name"
          "hello_spark"
        ]

This representation is generated within Flytekit. Esentially the SDK is
generating the instructions that Flyte’s kubernetes operator needs to
know in order to run this task at a later stage.

The ``Type`` field is really important as we will see later this will be
used by Flytepropeller (Kubernetes Operator) to know “how” to execute
this task.

``Interface`` contains information about what are the inputs and outputs
of our task. Flyte uses this interface to check if tasks are composible.

``Custom`` is a collection of arbitrary Key/Values, think of it as a
Json dict that any plugin can define as it wishes. In this case the
Spark plugin expects all its particular settings in this field i.e:
Spark workers, driver memory etc.

`Container <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L152>`__
is part of Flyte’s IDL primitives. Esentially any Flyte task is ran as
either three primitives a ``Container`` a ``K8sPod`` or ``Sql``. Every
task contains a ``Target`` which has to be either of these. In this
particular case, our Spark cluster is a ``Container`` target. A
``Container`` specifies all the needed parameters you would in a K8s
ContainerSpec i.e: What docker image to run, what is the command that
will be ran, args etc.

It is important for the reader to note that Flyte expects to run in a
container that has an entrypoint called ``pyflyte-execute``. This
entrypoint is provided when you ``pip install flytekit``. This
entrypoint and flytekit is what provides a lot of the plumbing logic
inside Flyte. For example It is this entrypoint what automagically
deserializes parquet dataframes an injects them to our task’s functions
if need be.

It should be clear to the reader that a lot of parameters are surrounded
by ``{}`` these are template variables that are to be rendered at
execution time.

What is important from this representation is that it contains all the
information that Flyte’s operator needs to know to execute this task: It
is a ``"Spark"`` task, it has a function signature (inputs and outputs),
it tells what docker image to run, and finally, it tells what spark
settings are needed for the cluster.

For more information on why this task contains these fields check
``TaskTemplate`` in `FlyteIDL
repository <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L102>`__.
I strongly advice you to take a look at the data structures in this file
as they provide good insight in the interfaces used all accross Flyte’s
codebases.

3. Once user has packaged workflows and tasks then a registration step
   is needed. During registration Flyte adds these protocolbuffer files to its
   database, esentially making these tasks and workflows runnable for
   the user. Registration is done via `Flytectl <https://github.com/flyteorg/flytectl>` __

4. At somepoint a Flyte user will trigger a Workflow run. The workflow
   run will start running the defined DAG. Eventually our Spark task
   will need to run,. This is where the second step of a plugin kicks
   in. Flytepropeller (Kubernetes Operator) will realize that this is a
   Task of type ``Spark`` and it will handle it differently.

   -  FlytePropeller knows a task is of type Spark, because our ``TaskTemplate`` defined it so ``Type: Spark``
      
   -  Flyte has a ``PluginRegistry`` which has a dictionary from ``Task Type`` to ``Plugin Handlers``.
   
   -  At run time Flytepropeller will run our task, Flytepropeller will figure out it is a Spark task, and then call the method ``BuildResource`` in Spark's plugin implementation. ``BuildResource`` is a method that each plugin has to implement.
   
   -  `Plugin <https://github.com/flyteorg/flyteplugins/blob/master/go/tasks/pluginmachinery/k8s/plugin.go#L80>`__ is a Golang interface providing an important method ``BuildResource``
   
   -  Spark has its own Plugin defined `here in Flyteplugins repo <https://github.com/flyteorg/flyteplugins/blob/master/go/tasks/plugins/k8s/spark/spark.go>`__

Inside Spark’s
`BuildResource <https://github.com/flyteorg/flyteplugins/blob/master/go/tasks/plugins/k8s/spark/spark.go#L65>`__
method is where magic happens. At task runtime:

   -  Flytepropeller will call ``BuildResource`` method. This method will ask for the ``Custom`` field, tasks flagged as ``type=Spark`` will have a dictionary containing all sort of Spark settings.

   -  Using these settings Flytepropeller will use Spark’s K8s Operator to spawn a spark cluster on the go and run a Spark app (Our python task).

   -  The spark app will run a pod with ``pyflyte-execute`` as entrypoint. All the inputs and outputs rendered to what they need to be i.e: paths to the actual data inputs instead of ``{{input}}``

   -  For more information on Spark’s K8s operator see : `SparkApplicationSpec <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/master/docs/api-docs.md#sparkapplicationspec>`__

5. A pod with entrypoint to ``pyflyte-execute`` execute starts running (Spark App).


   -  ``pyflyte-execute`` provides all the plumbing magic that is needed. In this particular case, It will create a SparkSession and injects it somewhere so that it is ready for when the user defined python’s code starts running. Be aware that this is part of the SDK code (Flytekit).

   -  ``pyflyte-execute`` points to `execute_task_cmd <https://github.com/flyteorg/flytekit/blob/master/flytekit/bin/entrypoint.py#L445>`__.

   This entrypoint does a lot of things:
   
   -  Resolves the function that the user wants to run. i.e: where is the needed package where this function lives? . this is what ``"flytekit.core.python_auto_container.default_task_resolver"`` does
   
   -  Downloads needed inputs and do a transformation if need be. I.e: is this a Dataframe? if so we need to transform it into a Pandas DF from parquet.
   
   -  Calls `dispatch_execute <https://github.com/flyteorg/flytekit/blob/771aa8a72fbc3ded437b6ff8498404767fc438db/flytekit/core/base_task.py#L449>`__ . This trigger the execution of our spark task.
   
   -  `PysparkFunctionTask <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-spark/flytekitplugins/spark/task.py#L78>`__. defines what gets run just before the user's task code gets executed. It esentially creatse a spark session and then run the user function (The actual code we want to run!).

------------

Recap
-----

-  Flyte requires coordination between multiple pieces of code. In this
   case the SDK and FlytePropeller (K8s operator)
- `Flyte IDL (Interface Language Definition) <https://github.com/flyteorg/flyteidl>`__  provides some primitives
   for services to talk with each other. Flyte uses Procolbuffer
   representations of these primitives
-  Three important primitives are : ``Container``, ``K8sPod``, ``Sql``.
   At the end of the day all tasks boil down to one of those three.
-  github.com/flyteorg/FlytePlugins repository contains all code for plugins:
   Spark, AWS Athena, BigQuery…
-  Flyte entrypoints are the ones carrying out the heavy lifting: making
   sure that inputs are downloaded and/or transformed as needed.
-  When running workflows on Flyte, if we want to use Flyte underlying plumbing then
   we should include Flyte entrypoints: either Jflyte or Flytekit.
