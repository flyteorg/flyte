.. _container-task-type:

Container Task
==============
This is one of the low-level task types. It belongs to task category "SingleStepTask", and is defined using a container image and task interface (TypedInterface). TypedInterface consists of an input map and an output map. Python Task is one implementation of this type in Flyte Python SDK. The task name and inputs are passed to the container when it is started. 

ContainerTask definition also includes RuntimeMetadata which points to the specific implementation of the "Container Task"(FlyteSDK or other).

Python Task
-----------
Python tasks are implemented on top of container tasks. It is a python function that needs to specify all its inputs using flyte input annotation @inputs and outputs using flyte output annotation @outputs. The function is not supposed to return a value but set all the outputs before exiting. The user container built during the build step is used for registration of this container task. 


#######
Usage
#######

The following is an example of a simple Python task.

.. code-block:: python
   :caption: Python task example with inputs and outputs
   
   @inputs(value_to_print=Types.Integer)
   @outputs(out=Types.Integer)
   @python_task(cache_version='1')
   def add_one_and_print(workflow_parameters, value_to_print, out):
       added = value_to_print + 1
       workflow_parameters.logging.info("My printed value: {}".format(added))
       out.set(added)

This does a couple things.

* SDK uses the current container image to register this task
* SDK registers the container with special container entrypoint that knows how to run this task.

