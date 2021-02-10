.. _howto_labels_annotations:

####################################
How to add Labels and Annotations?
####################################
In Flyte, workflow executions are created as kubernetes resources. These can be extended with
`labels <https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/>`_  and
`annotations <https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/>`_.

**Labels** and **annotations** are key value pairs which can be used to identify workflows for your own uses.
Labels are meant to be used as identifing attributes whereas annotations are arbitrary, *non-identifying* metadata.

Using labels and annotations is entirely optional. They can be used to categorize and identify workflow executions.

Labels and annotations are optional parameters to launch plan and execution invocations. In the case an execution
defines labels and/or annotations *and* the launch plan does as well, the execution spec values will be preferred.

Launch plan usage example
-------------------------

.. code:: python

    from flytekit.models.common import Labels, Annotations

    @workflow
    class MyWorkflow(object):
        ...

    my_launch_plan = MyWorkflow.create_launch_plan(
        labels=Labels({"myexecutionlabel": "bar", ...}),
        annotations=Annotations({"region": "SEA", ...}),
        ...
    )

    my_launch_plan.execute(...)

Execution example
-----------------

.. code:: python

    from flytekit.models.common import Labels, Annotations

    @workflow
    class MyWorkflow(object):
        ...

    my_launch_plan = MyWorkflow.create_launch_plan(...)

    my_launch_plan.execute(
        labels=Labels({"myexecutionlabel": "bar", ...}),
        annotations=Annotations({"region": "SEA", ...}),
        ...
    )
