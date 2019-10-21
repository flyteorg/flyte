.. _contributor-extending:

Extending Flyte
===============

.. toctree::
    :maxdepth: 2
    :name: extendingtoc
    :caption: Extending and customizing Flyte


Custom Tasks
##############
Writing logic for your own task types is the most common way to extend Flyte.  In fact, Flyte ships with several extensions by default.  These are tasks like the Qubole-run Hive queries or K8s-run Spark tasks, which were critical to Lyft's internal deployment of Flyte but aren't part of Flyte's core.

Extending the IDL
********************
Writing your own task will likely start with adding your own IDL message, which will look something like this :idl:`protos/flyteidl/plugins/sidecar.proto` or :idl:`protos/flyteidl/plugins/qubole.proto`.  Your custom task's proto message can reference other objects like in the Qubole case, but ultimately it has to be one message class.

An instance of this message class will be included alongside the rest of the ``TaskTemplate`` (in the ``custom`` field) for a given task.  Your message should include all the additional information that the execution layer of Flyte will need to execute the task.  That is, you don't need to worry about the container image, cpu/memory resources, input/output types, etc. since that is all covered in the normal task definition.  You only need to worry about the custom information for your task.  Technically, if your custom task doesn't need any additional information whatsoever, you can skip this step.


Extending the SDK
********************
The next step is to write a task handler on the SDK side.  That is, now that we have the definition of what your custom task will need, we need a way for users to write that task in Python, and then transform those tasks into task specifications containing that Protobuf message.  Continuing with the above examples, we can look at how the SDK bits are built for the sidecar task and the Qubole Hive task.

Broadly, the steps are:

#. Define a task type.  Concretely, this is just the string here :idl:`protos/flyteidl/core/tasks.proto#L92`.  As mentioned above, technically you do not need an IDL if for some reason your custom task has no additional information whatsoever.  But even in that case, you'll need a new task type string here.  This is the key that the execution plane will reference to decide how to run your task.
#. Create a class for your custom task that wraps the base task class (``flytekit.common.tasks.sdk_runnable.SdkRunnableTask`` or just ``flytekit.common.tasks.task.SdkTask`` if a decorator is not required).
#. Optionally, create a decorator to make it simple to define user code.

Qubole Hive Example
---------------------
The Hive task is slightly more complicated in that it produces a futures file, but the basic steps are the same.

#. First write a class that subclasses either the ``SdkRunnableTask`` or ``SdkTask`` like so :kit:`flytekit/common/tasks/hive_task.py#L27`
#. Override the :kit:`execute method <flytekit/common/tasks/hive_task.py#L179>` to have the behavior that you want to see.  In this case, we're running the user code, and then compiling the output of the user's code into our futures file.
   * Also, an instance of the custom Protobuf message defined in the IDL should be created (if you need it), and added here.
#. Create a decorator for your task like this :kit:`flytekit/sdk/tasks.py#L623`.

Ultimately, FlyteKit is a large wrapper around the Flyte IDL objects.  That is, its primary function is to translate user Python code into the Flyte component protobufs, sometimes in multiple stages like in the Hive example, that the Flyte engine and control plane can then understand.


Extending Propeller
********************

Flyte plugins extend the Flyte execution plane with custom behavior for special tasks.  When Flyte Propeller begins to run a task, it will look at the task type that you've defined and invoke the appropriate plugin.  Spark tasks and Qubole Hive tasks are examples of tasks that are run by plugins.

Structure of Plugins
-----------------------

At a high level, a Flyte Propeller plugin is something that satisfies the ``Executor`` interface specified in :plugins:`go/tasks/v1/types/task.go`.  The plugin's initialization code should register itself against the aforementioned task type string.  When Propeller comes across a task with this type, the plugin will be invoked.  Be cognizant that a plugin runs as a singleton in the engine.

One of the important objects to understand is the ``TaskContext``.  :plugins:`This interface <go/tasks/v1/types/task_context.go>` will be an object created by Propeller, and supplied to your plugin code.  Most importantly, the ``GetCustomState()`` function returns a custom struct that is integral to the cycle of your task's execution.

You supply the initial value of this custom struct as the output of your ``StartTask`` call.  On each call of the check loop thereafter, you get the version of the custom state that you returned before.  Since this is the only state that is stored in a durable store (etcd), it should be your source of truth.

.. note::
   Keep in mind that Flyte Propeller can restart at any time, which means your plugin can restart at any time.  This custom state is the only state that your plugin can rely on.

Note that this custom state is different than the custom IDL object that you previously defined.  The IDL message should be thought of as data describing the task itself whereas this customstate should be thought of as a way to keep track of state during execution of your task.

.. note::
   Note that while the ``CustomState`` object returned by the ``GetCustomState()`` function is a ``map[string]interface{}``, those interface values are not directly convertible to your Golang custom state objects.  That is, they need to be first marshaled from JSON into bytes, and then unmarshaled from bytes back into your object, like so: :plugins:`go/tasks/v1/qubole/qubole_work.go#L187`.

Task Initialization
^^^^^^^^^^^^^^^^^^^^
The ``StartTask`` function is only called once and will be called with the task template containing the custom IDL struct if you chose to create one.  You are only given this task template on this one ``StartTask`` call, so be sure your plugin code retrieves all the information from it that's necessary to complete the task's execution.  For the Qubole Hive plugin for example, the queries to be run are copied from the custom IDL object into the custom state object.

Task Updates
^^^^^^^^^^^^^^
This is the function that will be called periodically by Flyte Propeller, and is responsible for deciding how your custom task is progressing.  Note that while the task template is there in the function signature, it is not actually used and will always be nil.  This was an unfortunate optimization that had to be made to save on S3 access times.

Please refer to the generated documentation for a brief discussion of the other methods of the interface.


Background Updates
^^^^^^^^^^^^^^^^^^^
Often you'll need something to monitor the progress of your execution.  To this end, some plugins (like our Qubole Hive plugin and the waitable task plugin) make use of an `AutoRefreshCache <https://godoc.org/github.com/lyft/flytestdlib/utils#AutoRefreshCache>`__, to which you can specify a sync function.  This sync function will be periodically run on all objects in your cache.  See the cache's documentation for more information.

