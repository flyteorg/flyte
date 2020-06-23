.. _on-failuire-policy:

What is it
==========

The default behavior for when a node fails in a workflow is to immediately abort the entire workflow. The reasoning behind this thinking
is to avoid wasting resources since the workflow will end up failing anyway. There are certain cases however, when it's desired for the
workflow to carry on executing the branches it can execute. 

For example when the remaining tasks are marked as :ref:`cacheable <features-task_cache>`. 
Once the failure has been fixed and the workflow is relaunched, cached tasks will be bypassed quickly.

How to use it
-------------

Use on_failure attribute on workflow_class.

.. code:: python

   from flytekit.models.core.workflow import WorkflowMetadata

   @workflow_class(on_failure=WorkflowMetadata.OnFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE)
   class RunToCompletionWF(object):
     pass

Available values in the policy:

.. code:: python

    class OnFailurePolicy(object):
        """
        Defines the execution behavior of the workflow when a failure is detected.
        Attributes:
            FAIL_IMMEDIATELY                        Instructs the system to fail as soon as a node fails in the
                                                    workflow. It'll automatically abort all currently running nodes and
                                                    clean up resources before finally marking the workflow executions as failed.
            FAIL_AFTER_EXECUTABLE_NODES_COMPLETE    Instructs the system to make as much progress as it can. The system
                                                    will not alter the dependencies of the execution graph so any node 
                                                    that depend on the failed node will not be run. Other nodes that will
                                                    be executed to completion before cleaning up resources and marking
                                                    the workflow execution as failed.
        """

        FAIL_IMMEDIATELY = _core_workflow.WorkflowMetadata.FAIL_IMMEDIATELY        
        FAIL_AFTER_EXECUTABLE_NODES_COMPLETE = _core_workflow.WorkflowMetadata.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
