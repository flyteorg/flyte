.. _dynamic-task-type:

Dynamic Tasks
=============

Flyte offers many task types to support a varity of use-cases that are all based on statically defining what 
those tasks do beforehand. Dynamic Task type is the only exception to this rule where the system knows very little
about what the nature of the workload that will be executed until execution time.

A Dynamic task is executed as follows:
  - A generator step; This step runs like a :ref:`Container Task <container-task-type>`. The expected outcome of this step is a :ref:`dynamic job spec<concepts-dynamic-spec>`.
  - Execute the dynamic job spec; The spec can contain any of the flyte supported workflow nodes. One or more of these nodes might contain other
    dynamic tasks. In which case, it'll get recursively executed.
  - Assemble final outputs of the task.

Some of the potential use cases:
  - Launching arbitrary workflows/tasks:
    You might build a workflow where one of its tasks chooses and executes a launchplan based on values passed as inputs to the task.
  - Dynamically creating a workflow
    If the desired set of nodes/tasks that need to be executed is controlled by inputs, you can use dynamic tasks to build
    a workflow spec at execution time and yield that for it to be executed.   

.. code-block:: python
   :caption: Dynamic task example
   
    @inputs(tasks_count=Types.Integer)
    @outputs(out=[Types.Integer])
    @dynamic_task(cache_version='1')
    def generate_others(workflow_parameters, tasks_count, out):
      res = []
      # Launch an arbitrary number of tasks based on the input tasks_count
      for i in range(0, tasks_count):
        task = other_task(index=i)
        yield task
        # Define how to assemble the final result
        res.append(task.outputs.other_task_out)

      # Define how to set the final result of the task
      out.set(res)

A few notes about how this task will run:
  - This code will be executed only once. Based on the way outputs are declared, :ref:`Bindings <language-bindings>` will
    be created to instruct the system on how to assemble the final outputs after all yielded tasks are executed.
  - Manipulating Outputs of the yielded tasks is not supported. Think of this step as a `Map Concept`_. If a simple reduce is
    required, it'll have to happen as a separate task that consumes the assembled outputs here.
  - There is a restriction on the size of individual tasks outputs as well as on the final output of this task. If large outputs
    are desired, consider using Blob types.

.. _Map Concept: https://en.wikipedia.org/wiki/MapReduce