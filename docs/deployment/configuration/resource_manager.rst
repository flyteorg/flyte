.. _deployment-configuration-resource-manager:

#####################
Flyte ResourceManager
#####################

**Flyte ResourceManager** is a configurable component that helps track resource utilization of tasks that run on Flyte and allows plugins to manage resource allocations independently. Default deployments are configured with the ResourceManager disabled, which means plugins rely on each independent platform to manage resource utilization. See below for the default ResourceManager configuration:

.. code-block:: yaml

    resourcemanager:
      type: noop

When using a plugin that connects to a platform with a robust resource scheduling mechanism, like the K8s plugin, we recommend leaving the default ``flyteresourcemanager`` configuration in place. However, with web API plugins (for example), the rate at which Flyte sends requests may overwhelm a service, and we recommend changing the ``resourcemanager`` configuration.

The ResourceManager provides a task-type-specific pooling system for Flyte tasks. Optionally, plugin writers can request resource allocation in their tasks.

A plugin defines a collection of resource pools using its configuration. Flyte uses tokens as a placeholder to represent a unit of resource.

How Flyte plugins request resources
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Flyte plugins register the desired resource and resource quota with the **ResourceRegistrar** when setting up FlytePropeller. When a plugin is invoked, FlytePropeller provides a proxy for the plugin. This proxy facilitates the plugin's view of the resource pool by controlling operations to allocate and deallocate resources.

Once the setup is complete, FlytePropeller builds a ResourceManager based on the previously requested resource registration. Based on the plugin implementation's logic, resources are allocated and deallocated.

During runtime, the ResourceManager:

#. Allocates tokens to the plugin.
#. Releases tokens once the task is completed.

In this manner, Flyte plugins intelligently throttle resource usage during parallel execution of nodes.

.. note ::

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

Plugin resource allocation
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

When a Flyte task execution needs to send a request to an external service, the plugin claims a unit of the corresponding resource using a **ResourceName**, which is a unique token and a fully qualified resource request (typically an integer). The task execution generates this unique token and registers the token with the ResourceManager by calling the ResourceManager’s ``AllocateResource`` function. If the resource pool has sufficient capacity to fulfill the request, then the requested resources are allocated, and the plugin proceeds further.

When the status changes to **"AllocationGranted"**, the execution sends out the request for those resources.

The granted token is recorded in a token pool which corresponds to the resource that is managed by the ResourceManager.

Plugin resource deallocation
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
When the request is completed, the plugin asks the ResourceManager to release the token by calling the ResourceManager's ``ReleaseResource()`` function, which eliminates the token from the token pool.

Example
^^^^^^^^

Flyte has a built-in `Qubole <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/plugins/plugins.html#qubolehivejob>`__ plugin which allows Flyte tasks to send Hive commands to Qubole. In the plugin, a single Qubole cluster is considered a resource, and sending a single Hive command to a Qubole cluster consumes a token of the corresponding resource.
The resource is allocated when the status is **“AllocationGranted”**. The Qubole plugin calls:

.. code-block:: go

   status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)

In our example scenario, the placeholder values are replaced with the following:

.. code-block:: go

   status, err := AllocateResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0", ResourceConstraintsSpec{})

The resource is deallocated when the Hive command completes its execution and the corresponding token is released. The plugin calls:

.. code-block:: go

   status, err := AllocateResource(ctx, <cluster name>, <token string>, <constraint spec>)

In our example scenario, the placeholder values are replaced with the following:

.. code-block:: go

   err := ReleaseResource(ctx, "default_cluster", "flkgiwd13-akjdoe-0")

See below for an example interface that shows allocation and deallocation of resources:

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

Configuring ResourceManager to force runtime quota allocation constraints
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Runtime quota allocation constraints can be achieved using ResourceConstraintsSpec. It is a contact that a plugin can specify at different project and namespace levels.

For example, you can set ResourceConstraintsSpec to ``nil`` objects, which means there would be no allocation constraints at the respective project and namespace level. When ResourceConstraintsSpec specifies ``nil`` ProjectScopeResourceConstraint, and a non-nil NamespaceScopeResourceConstraint, it suggests no constraints specified at any project or namespace level.
