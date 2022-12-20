"""
Adding Workflow Labels and Annotations
--------------------------------------

.. tags:: Kubernetes, Intermediate

In Flyte, workflow executions are created as Kubernetes resources. These can be extended with
`labels <https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/>`__ and
`annotations <https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/>`__.

**Labels** and **annotations** are key value pairs which can be used to identify workflows for your own uses.

.. Warning::
   Note that adding labels and annotations to your K8s resources may have side-effects depending on webhook behavior on your execution clusters.

Labels are meant to be used as identifying attributes, whereas annotations are arbitrary, *non-identifying* metadata.

Using labels and annotations is entirely optional. They can be used to categorize and identify workflow executions.

Labels and annotations are optional parameters to launch plan and execution invocations. When an execution
defines labels and/or annotations *and* the launch plan does as well, the execution spec values will be preferred.

Launch Plan Usage Example
#########################

.. code:: python

    from flytekit import Labels, Annotations

    @workflow
    class MyWorkflow(object):
        ...

    my_launch_plan = MyWorkflow.create_launch_plan(
        labels=Labels({"myexecutionlabel": "bar", ...}),
        annotations=Annotations({"region": "SEA", ...}),
        ...
    )

    my_launch_plan.execute(...)

Execution Example
#################

.. code:: python

    from flytekit import Labels, Annotations

    @workflow
    class MyWorkflow(object):
        ...

    my_launch_plan = MyWorkflow.create_launch_plan(...)

    my_launch_plan.execute(
        labels=Labels({"myexecutionlabel": "bar", ...}),
        annotations=Annotations({"region": "SEA", ...}),
        ...
    )


"""
