.. _deployment-configuration-customizable-resources:

#################################################################
Customizing project, domain, and workflow resources with flytectl
#################################################################

For critical projects and workflows, you can use the :ref:`flytectl update <flytectl:flytectl_update_task-resource-attribute>` command to configure
settings for task, cluster, and workflow execution resources, set matching executions to execute on specific clusters, set execution queue attributes, and :ref:`more <ref_flyteidl.admin.MatchableResource>`
that differ from the default values set for your global Flyte installation. These customizable settings are created, updated, and deleted via the API and stored in the FlyteAdmin database.

In code, these settings are sometimes called `matchable attributes` or `matchable resources`, because we use a hierarchy for matching the customizations to applicable Flyte inventory and executions.

*******************************
Configuring existing resources
*******************************


About the resource hierarchy
============================

Many platform specifications set in the FlyteAdmin config are applied to every project and domain. Although these values are customizable as part of your helm installation, they are still applied to every user project and domain combination.

You can choose to customize these settings along increasing levels of specificity with Flyte:

- Domain
- Project and Domain
- Project, Domain, and Workflow name
- Project, Domain, Workflow name and LaunchPlan name

See :ref:`control-plane` to understand projects and domains.
The following section will show you how to configure the settings along
these dimensions.

Task resources
==============

As a system administrator you may want to define default task resource requests and limits across your Flyte deployment. This can be set globally in the FlyteAdmin `config <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L520-L531>`__
in `task_resource_defaults`.

**Default** values get injected as the task requests and limits when a task definition omits a specific :py:class:`resource <flytekit:flytekit.Resources>`.

**Limit** values are only used as validation. Neither a task request nor limit can exceed the limit for a resource type.

Configuring task resources
--------------------------
Available resources for configuration include:

- CPU
- GPU
- Memory
- `Ephemeral Storage <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#local-ephemeral-storage>`__

In the absence of a customization, the global
`default values <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L520-L531>`__
in `task_resource_defaults` are used.

The customized values from the database are assigned at execution, rather than registration time.

Customizing task resource configuration
---------------------------------------

To customize resources for project-domain attributes using `flytectl`, define a ``tra.yaml`` file with your customizations:

.. code-block:: yaml

    project: flyteexamples
    domain: development
    defaults:
        cpu: "1"
        memory: 150Mi
    limits:
        cpu: "2"
        memory: 450Mi

Update the task resource attributes for a project-domain combination:

.. prompt:: bash $

    flytectl update task-resource-attribute --attrFile tra.yaml

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_update_task-resource-attribute>` to
   learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes:

.. prompt:: bash $

    flytectl get task-resource-attribute -p flyteexamples -d development

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn
   more about the command and its supported flag(s).

You can view all custom task-resource-attributes by visiting
``protocol://<host/api/v1/matchable_attributes?resource_type=0>`` and substitute
the protocol and host appropriately.

Cluster resources
=================

Cluster resources are how you configure Kubernetes namespace attributes that are applied at execution time. This includes per-namespace resource quota, patching the default service account with a bounded IAM role, or attaching    `imagePullSecrets` to the default service account for accessing a private container registry


Configuring cluster resources
-----------------------------
The format of all these parameters are free-form key-value pairs used for populating the Kubernetes object templates consumed by the cluster resource controller. The cluster resource controller ensures these fully rendered object templates are applied as Kubernetes resources for each execution namespace.

The keys represent templatized variables in the
`cluster resource template <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L737,L760>`__
and the values are what you want to see filled in.

In the absence of custom customized values, your Flyte installation will use ``templateData`` from the
`FlyteAdmin config <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__
as the per-domain defaults. Flyte specifies these defaults by domain and applies them to every
project-domain namespace combination.


Customizing cluster resource configuration
------------------------------------------
.. note::
    The cluster resource template values can be specified on domain, and project-and-domain.
    Since Flyte execution namespaces are never on a per-workflow or a launch plan basis, specifying a workflow or launch plan level customization is non-actionable.
    This is a departure from the usual hierarchy for customizable resources.


Define an attributes file, ``cra.yaml``:

.. code-block:: yaml

    domain: development
    project: flyteexamples
    attributes:
        projectQuotaCpu: "1000"
        projectQuotaMemory: 5Ti

To ensure that the customizations reflect in the Kubernetes namespace
``flyteexamples-development`` (that is, the namespace has a resource quota of
1000 CPU cores and 5TB of memory) when the admin fills in cluster resource
templates:

.. prompt:: bash $

   flytectl update cluster-resource-attribute --attrFile cra.yaml

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_update_cluster-resource-attribute>`
   to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes:

.. prompt:: bash $

    flytectl get cluster-resource-attribute -p flyteexamples -d development

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to
   learn more about the command and its supported flag(s).

Flyte uses these updated values to fill the template fields for the
``flyteexamples-development`` namespace.

For other namespaces, the
`platform defaults <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__
apply.

.. note::
    The template values, for example, ``projectQuotaCpu`` or ``projectQuotaMemory`` are free-form strings.
    Ensure that they match the template placeholders in your `template file <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml>`__
    for your changes to take effect and custom values to be substituted.

You can view all custom cluster-resource-attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=1>``
and substitute the protocol and host appropriately.


Workflow execution configuration
================================


Although many execution-time parameters can be overridden at execution time itself, it is helpful to set defaults on a per-project or per-workflow basis. This config includes
- `annotations and labels <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/containerization/workflow_labels_annotations.html#sphx-glr-auto-core-containerization-workflow-labels-annotations-py>`__
etc. in the `Workflow execution config <https://github.com/flyteorg/flyteidl/blob/master/gen/pb-go/flyteidl/service/flyteadmin/model_admin_workflow_execution_config.go#L14-L23>`__.
- `max_parallelism`: Limits maximum number of nodes that can be evaluated for an individual workflow in parallel
- `security context <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/core/core.html#securitycontext>`__: configures the pod identity and auth credentials for task pods at execution time
- `raw_output_data_config`: where offloaded user data is stored
- `interruptible`: whether to use [spot instances](https://docs.flyte.org/en/latest/flytesnacks/examples/productionizing/spot_instances.html#using-spot-preemptible-instances)
- `overwrite_cache`: Allows for all cached values of a workflow and its tasks to be overwritten for a single execution.
- `envs`: Custom environment variables to apply for task pods brought up during execution

Customizing workflow execution configuration
--------------------------------------------

These can be defined at two levels of project-domain or project-domain-workflow:

.. prompt:: bash $

    flytectl update workflow-execution-config

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_update_workflow-execution-config>`
   to learn more about the command and its supported flag(s).

Execution cluster label
=======================

This matchable attributes allows forcing a matching execution to consistently execute on a specific Kubernetes cluster for multi-cluster Flyte deployment set-up. In lieu of an explicit customization, cluster assignment is random.

For setting up a multi-cluster environment, follow :ref:`the guide <deployment-deployment-multicluster>`


Customizing execution cluster label configuration
-------------------------------------------------

Define an attributes file in `ec.yaml`:

.. code-block:: yaml

    value: mycluster
    domain: development
    project: flyteexamples

Ensure that admin places executions in the flyteexamples project and development domain onto ``mycluster``:

.. prompt:: bash $

   flytectl update execution-cluster-label --attrFile ec.yaml

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_update_execution-cluster-label>`
   to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes:

.. prompt:: bash $

    flytectl get execution-cluster-label -p flyteexamples -d development

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to
   learn more about the command and its supported flag(s).

You can view all custom execution cluster attributes by visiting
``protocol://<host/api/v1/matchable_attributes?resource_type=3>`` and substitute
the protocol and host appropriately.

.. _deployment-customizable-resources-execution-queues:

Execution queues
================

Execution queues are defined in
`FlyteAdmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/FlyteAdmin_config.yaml#L97,L106>`__.
These are used for execution placement for constructs like AWS Batch.

The **attributes** associated with an execution queue must match the **tags**
for workflow executions. The tags associated with configurable resources are
stored in the admin database.

Customizing execution queue configuration
-----------------------------------------

.. prompt:: bash $

    flytectl update execution-queue-attribute

.. note::

   Refer to the :ref:`docs <flytectl:flytectl_update_execution-queue-attribute>`
   to learn more about the command and its supported flag(s).

You can view existing attributes for which tags can be assigned by visiting
``protocol://<host>/api/v1/matchable_attributes?resource_type=2`` and substitute
the protocol and host appropriately.


*********************************
Adding new customizable resources
*********************************

.. tags:: Infrastructure, Advanced

As a quick refresher, custom resources allow you to manage configurations for specific combinations of user projects, domains and workflows that customize default values.
Examples of such resources include execution clusters, task resource defaults, and :std:ref:`more <ref_flyteidl.admin.MatchableResource>`.

.. note::
    For background on customizable resources, refer to :ref:`deployment-configuration-general`.

In a :ref:`multi-cluster setup <deployment-deployment-multicluster>`, an example one could think of is setting routing rules to send certain workflows to specific clusters, which demands setting up custom resources.

Here's how you could go about building a customizable priority designation.


Example
=======

Let's say you want to inject a default priority annotation for your workflows.
Perhaps you start off with a model where everything has a default priority but soon you realize it makes sense that workflows in your production domain should take higher priority than those in your development domain.

Now, one of your user teams requires critical workflows to have a higher priority than other production workflows.

Here's how you could do that.

Flyte IDL
---------

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
----------

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
--------

For convenience, add a FlyteCTL wrapper to update the new attributes. Refer to `this PR <https://github.com/flyteorg/flytectl/pull/65>`__ for the entire set of changes required.

That's it! You now have a new matchable attribute to configure as the needs of your users evolve.
<<<<<<< HEAD
=======

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

Flyte has a built-in `Qubole <https://docs.flyte.org/en/latest/protos/docs/plugins/plugins.html#qubolehivejob>`__ plugin. This plugin allows Flyte tasks to send Hive commands to Qubole. In the plugin, a single Qubole cluster is considered a resource, and sending a single Hive command to a Qubole cluster consumes a token of the corresponding resource.
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
>>>>>>> master