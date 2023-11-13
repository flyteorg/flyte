.. _deployment-configuration-customizable-resources:

#################################
Adding New Customizable Resources
#################################

.. tags:: Infrastructure, Advanced

As a quick refresher, custom resources allow you to manage configurations for specific combinations of user projects, domains and workflows that override default values.
Examples of such resources include execution clusters, task resource defaults, and :std:ref:`more <ref_flyteidl.admin.MatchableResource>`.

.. note::
    For background on customizable resources, refer to :ref:`deployment-configuration-general`.

In a :ref:`multi-cluster setup <deployment-deployment-multicluster>`, an example one could think of is setting routing rules to send certain workflows to specific clusters, which demands setting up custom resources.

Here's how you could go about building a customizable priority designation.

Example
-------

Let's say you want to inject a default priority annotation for your workflows.
Perhaps you start off with a model where everything has a default priority but soon you realize it makes sense that workflows in your production domain should take higher priority than those in your development domain.

Now, one of your user teams requires critical workflows to have a higher priority than other production workflows.

Here's how you could do that.

Flyte IDL
^^^^^^^^^

Introduce a new :std:ref:`matchable resource <ref_flyteidl.admin.MatchableResource>` that includes a unique enum value and proto message definition.

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


FlyteAdmin
^^^^^^^^^^

Once your IDL changes are released, update the logic of FlyteAdmin to `fetch <https://github.com/flyteorg/flyteadmin/commit/60b4c876ea105d4c79e3cad7d56fde6b9c208bcd#diff-510e72225172f518850fe582149ff320R122-R128>`__ your new matchable priority resource and use it while creating executions or in relevant use cases.

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

For convenience, add a FlyteCTL wrapper to update the new attributes. Refer to `this PR <https://github.com/flyteorg/flytectl/pull/65>`__ for the entire set of changes required.

That's it! You now have a new matchable attribute to configure as the needs of your users evolve.

Flyte ResourceManager
---------------------

**Flyte ResourceManager** is a configurable component that allows plugins to manage resource allocations independently. It helps track resource utilization of tasks that run on Flyte. The default deployments are configured as ``noop``, which indicates that the ResourceManager provided by Flyte is disabled and plugins rely on each independent platform to manage resource utilization. In situations like the K8s plugin, where the platform has a robust mechanism to manage resource scheduling, this may work well. However, in a scenario like a simple web API plugin, the rate at which Flyte sends requests may overwhelm a service and benefit from additional resource management.

The below attribute is configurable within FlytePropeller, which can be disabled with:

.. code-block:: yaml

    resourcemanager:
      type: noop

The ResourceManager provides a task-type-specific pooling system for Flyte tasks. Optionally, plugin writers can request resource allocation in their tasks.

A plugin defines a collection of resource pools using its configuration. Flyte uses tokens as a placeholder to represent a unit of resource.

How does a Flyte plugin request for resources?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The Flyte plugin registers the resource and the desired quota of every resource with the **ResourceRegistrar** when setting up FlytePropeller. When a plugin is invoked, FlytePropeller provides a proxy for the plugin. This proxy facilitates the plugin's view of the resource pool by controlling operations to allocate and deallocate resources.

.. dropdown:: :fa:`info-circle` Enabling Redis instance
   :title: text-muted
   :animate: fade-in-slide-down

   The ResourceManager can use a Redis instance as an external store to track and manage resource pool allocation. By default, it is disabled, and can be enabled with:

   .. code-block:: yaml
       
       resourcemanager:
          type: redis
          resourceMaxQuota: 100
          redis:
            hostPaths:
              - foo
            hostKey: bar
            maxRetries: 0

Once the setup is complete, FlytePropeller builds a ResourceManager based on the previously requested resource registration. Based on the plugin implementation's logic, resources are allocated and deallocated. 

During runtime, the ResourceManager:

#. Allocates tokens to the plugin.
#. Releases tokens once the task is completed.

How are resources allocated?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

When a Flyte task execution needs to send a request to an external service, the plugin claims a unit of the corresponding resource. This is done using a **ResourceName**, which is a unique token and a fully qualified resource request (which is typically an integer). The execution generates this unique token and registers this token with the ResourceManager by calling the ResourceManager’s **"AllocateResource function"**. If the resource pool has sufficient capacity to fulfil your request, then the resources requested are allocated, and the plugin proceeds further.

When the status is **"AllocationGranted"**, the execution moves forward and sends out the request for those resources.

The granted token is recorded in a token pool which corresponds to the resource that is managed by the ResourceManager.

How are resources deallocated?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
When the request is completed, the plugin asks the ResourceManager to release the token by calling the ReleaseResource() function present in the ResourceManager. Upon calling the function, the token is eliminated from the token pool. 
In this manner, Flyte plugins intelligently throttle resource usage during parallel execution of nodes. 

Example
^^^^^^^^
Let's take an example to understand resource allocation and deallocation when a plugin requests resources. 

Flyte has a built-in `Qubole <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/plugins/plugins.html#qubolehivejob>`__ plugin. This plugin allows Flyte tasks to send Hive commands to Qubole. In the plugin, a single Qubole cluster is considered a resource, and sending a single Hive command to a Qubole cluster consumes a token of the corresponding resource. 
The resource is allocated when the status is **“AllocationGranted”**. Qubole plugin calls:

.. code-block:: go
   
   status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)

Wherein the placeholders are occupied by:

.. code-block:: go
   
   status, err := AllocateResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0", ResourceConstraintsSpec{})

The resource is deallocated when the Hive command completes its execution and the corresponding token is released. The plugin calls:

.. code-block:: go
   
   status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)

Wherein the placeholders are occupied by:

.. code-block:: go

   err := ReleaseResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0")

Below is an example interface that shows allocation and deallocation of resources.

.. code-block:: go

    type ResourceManager interface {
    GetID() string
    // During execution, the plugin calls AllocateResource() to register a token in the token pool associated with a resource
    // If it is granted an allocation, the token is recorded in the token pool until the same plugin releases it.
    // When calling AllocateResource, the plugin has to specify a ResourceConstraintsSpec that contains resource capping constraints at different project and namespace levels.
    // The ResourceConstraint pointers in ResourceConstraintsSpec can be set to nil to not have a constraint at that level 
    AllocateResource(ctx context.Context, namespace ResourceNamespace, allocationToken string, constraintsSpec ResourceConstraintsSpec) (AllocationStatus, error)
    // During execution, after an outstanding request is completed, the plugin uses ReleaseResource() to release the allocation of the token from the token pool. This way, it redeems the quota taken by the token
    ReleaseResource(ctx context.Context, namespace ResourceNamespace, allocationToken string) error
    }

How can you force ResourceManager to force runtime quota allocation constraints?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Runtime quota allocation constraints can be achieved using ResourceConstraintsSpec. It is a contact that a plugin can specify at different project and namespace levels.

Let's take an example to understand it.

You can set ResourceConstraintsSpec to ``nil`` objects, which means there would be no allocation constraints at the respective project and namespace level. When ResourceConstraintsSpec specifies ``nil`` ProjectScopeResourceConstraint, and a non-nil NamespaceScopeResourceConstraint, it suggests no constraints specified at any project or namespace level.
