.. _recipe-1:

############################################################
How do I call a workflow from within another workflow?
############################################################

There are four possible ways of incorporating a workflow into another workflow - you have two choices each along two dimensions.

Workflow vs Launch Plan
  * The workflow can be referenced directly (by using its identifier). Think of this as pass-by-value.
  * The workflow can be referenced by a launch plan (by using the identifier for the launch plan instead). This of this as pass-by-reference.

Static vs Dynamic
  * The workflow can be declared statically inside another workflow. This is called a sub-workflow and the contents of the inner workflow will appear statically inside the definition of the parent workflow.
  * The workflow can be yielded dynamically from within a dynamic task. In this case, the workflow will not show up until execution time.


********
Examples
********

Each combination is documented in the workflows in this folder. Below are some more concrete details. The full output of each workflow from the ``flyte-cli`` command is also linked in each case below.

Statically
===========

Calling a Launch Plan
----------------------

Workflow name: ``StaticLaunchPlanCaller``

This is the node that gets included in the compiled workflow. Note that the version that's pulled in is assumed to be the version that the workflow registration itself ran with on

.. code-block::

        nodes {
          id: "identity-lp-execution"
          metadata {
            name: "identity-lp-execution"
            timeout {
            }
            retries {
            }
          }
          inputs {
            var: "a"
            binding {
              promise {
                node_id: "start-node"
                var: "outer_a"
              }
            }
          }
          workflow_node {
            launchplan_ref {
              resource_type: LAUNCH_PLAN
              project: "flytetester"
              domain: "development"
              name: "cookbook.sample_workflows.formula_1.outer.id_lp"
              version: "7be6342b4d5d95f5e31e6ad89636ad48925643ab"
            }
          }
        }

To see the complete workflow specification ::

    flyte-cli -p flytesnacks -d development get-workflow -u wf:flytesnacks:development:workflows.recipe_1.outer.StaticLaunchPlanCaller:<sha>

:ref:`Full output <st-lp>`


Calling a Sub-Workflow
----------------------

Workflow name: ``StaticSubWorkflowCaller``


To get the workflow ::

    flyte-cli -p flytesnacks -d development get-workflow -u wf:flytesnacks:development:workflows.recipe_1.outer.StaticSubWorkflowCaller:<sha>

.. code-block::

    sub_workflows {
      template {
        id {
          resource_type: WORKFLOW
          project: "flytesnacks"
          domain: "development"
          name: "workflows.recipe_1.inner.IdentityWorkflow"
          version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
        }
        metadata {
        ...

The interesting bit here in the output is the sub-workflow section of the template.  The entire definition of the subworkflow is reproduced in this section.

:ref:`Full output <st-swf>`


Dynamically
===========

Calling a Launch Plan
----------------------

Workflow name: ``DynamicLaunchPlanCaller``

To get the workflow ::

    flyte-cli -p flytesnacks -d development get-workflow -u wf:flytesnacks:development:workflows.recipe_1.outer.DynamicLaunchPlanCaller:<sha>

Note that here there are no subworkflows in the workflow specification, just a task. However this is a ``dynamic_task``, and when executed, it will yield two launch plans which in turn yield their own executions, all of which will appear on the same execution page.

Also, if you look at the execution using ``flyte-cli`` ::

    flyte-cli -p flytesnacks -d development get-execution -u ex:flytesnacks:development:hmi4y7so5j

You should see that it returns in a "Subtasks" section, a new ``flyte-cli`` command that you can run again which will show you the deeper executions. (You'll need to replace the execution name in the command above with yours.)

:ref:`Full output <dyn-lp>`

Calling a Sub-Workflow
----------------------

Workflow name: ``DynamicSubWorkflowCaller``

To get the workflow ::

    flyte-cli -p flytesnacks -d development get-workflow -u wf:flytesnacks:development:sample_workflows.formula_1.outer.DynamicSubWorkflowCaller:4fffbe68d0cb37c8b43e17fa214bbfb4a6ae416e

Even though this workflow eventually calls a sub-workflow, since that happens inside another ``dynamic_task``, the subworkflow is again not present in the workflow specification.

:ref:`Full output <dyn-swf>`
