.. _deployment-customizable-resources:

##################################
Adding New Customizable Resources
##################################

For background on customizable resources, see the :ref:`deployment-cluster-config-general`.
As a quick refresher, custom resources allow you to manage configurations for :ref:`specific combinations <multicluster-setup>` of user projects,
domains and workflows that override the default values. 

Examples of such resources include execution clusters, task resource
defaults, and :std:ref:`more <flyteidl:protos/docs/admin/admin:matchableresource>`.


Example
-------

Let's say you want to inject a default priority annotation for your workflows. Perhaps you start off with a model where everything has a default priority but soon you realize it makes sense that workflows in your production domain should take higher priority than those in your development domain.

Now, one of your users team require critical workflows to have a higher priority than other production workflows. Here's how you could go about building a customizable priority designation.

Flyte IDL
^^^^^^^^^
Introduce a new :std:ref:`matchable resource <protos/docs/admin/admin:matchableresource>`, that includes a unique enum value and proto message definition. 

For example:

::      

   enum MatchableResource {
     ...
     WORKFLOW_PRIORITY = 10;
   }

   message WorkflowPriorityAttribute {
     int priority = 1;
   }

   message MatchingAttributes {
     oneof target {
       ...
       WorkflowPriorityAttribute WorkflowPriority = 11;
     }
   }


See the changes in this `file <https://github.com/flyteorg/flyteidl/commit/b1767697705621a3fddcb332617a5304beba5bec#diff-d3c1945436aba8f7a76755d75d18e671>`__ for an example of what is required.


Flyte admin
^^^^^^^^^^^

Once your IDL changes are released, update the logic of flyteadmin to `fetch <https://github.com/flyteorg/flyteadmin/commit/60b4c876ea105d4c79e3cad7d56fde6b9c208bcd#diff-510e72225172f518850fe582149ff320R122-R128>`__ your new matchable priority resource and use it while creating executions or in relevant use cases.

For example:

::      

   
   resource, err := s.resourceManager.GetResource(ctx, managerInterfaces.ResourceRequest{
       Domain:       domain,
       Project:      project, // optional
       Workflow:     workflow, // optional, must include project when specifying workflow
       LaunchPlan:   launchPlan, // optional, must include project + workflow when specifying launch plan
       ResourceType: admin.MatchableResource_WORKFLOW_PRIORITY,
   })

   if err != nil {
       return err
   }

   if resource != nil && resource.Attributes != nil && resource.Attributes.GetWorkflowPriority() != nil {
        priorityValue := resource.Attributes.GetWorkflowPriority().GetPriority()
        // do something with the priority here
   }


Flytekit
^^^^^^^^
For convenience, add a flyte-cli wrapper for updating your new attribute. See `this PR <https://github.com/flyteorg/flytekit/pull/174>`__ for the entire set of changes required.

That's it! You now have a new matchable attribute that you can configure as the needs of your users evolve.