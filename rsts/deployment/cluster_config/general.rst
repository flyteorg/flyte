.. _deployment-cluster-config-general:

Configuring Custom K8s Resources
----------------------------------

***************************
Configurable Resource Types
***************************

Many platform specifications such as task resource defaults, project namespace Kubernetes quota, and more can be
assigned using default values or custom overrides. Defaults are specified in the FlyteAdmin config and
overrides for specific projects are registered with the FlyteAdmin service.

You can customize these settings along increasing levels of specificity with Flyte:

- domain
- project and domain
- project, domain, and workflow name
- project, domain, workflow name and launch plan name

See :ref:`concepts/control_plane:control plane` to understand projects and domains.
Along these dimensions, the following settings are configurable:

Task Resources
==============
Configuring task resources includes setting default values for unspecified task requests and limits.
Task resources also include limits which specify the maximum value that a task request or a limit can have.

- cpu
- gpu
- memory
- storage
- `ephemeral storage <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#local-ephemeral-storage>`__

In the absence of an override, the global
`default values <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L35,L43>`__
in the FlyteAdmin config are used.

The override values from the database are assigned at execution, rather than registration time.

To customize resources for project-domain attributes, define a ``tra.yaml`` file with your overrides:

.. code-block:: yaml

    defaults:
        cpu: "1"
        memory: 150Mi
    limits:
        cpu: "2"
        memory: 450Mi
    project: flyteexamples
    domain: development

Update the task resource attributes for a project-domain combination:

.. prompt:: bash

    flytectl update task-resource-attribute --attrFile tra.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_task-resource-attribute>` to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes:

.. prompt:: bash

    flytectl get task-resource-attribute -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

You can view all custom task-resource-attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=0>`` and substitute the protocol and host appropriately.

Cluster Resources
=================
These are free-form key-value pairs used when filling the templates that the admin feeds into the cluster manager; the process that syncs Kubernetes resources.

The keys represent templatized variables in `cluster resource template <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L737,L760>`__ and the values are what you want to see filled in.

In the absence of custom override values, you can use ``templateData`` from the `FlyteAdmin config <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__ as a default. Flyte specifies these defaults by domain and applies them to every project-domain namespace combination.

.. note::
    The settings above can be specified on domain, and project-and-domain.
    Since Flyte hasn't tied the notion of a workflow or a launch plan to any Kubernetes construct, specifying a workflow or launch plan name doesn't make sense.
    This is a departure from the usual hierarchy for customizable resources.

Define an attributes file, ``cra.yaml``:

.. code-block:: yaml

    attributes:
        projectQuotaCpu: "1000"
        projectQuotaMemory: 5TB
    domain: development
    project: flyteexamples

To ensure that the overrides reflect in the Kubernetes namespace ``flyteexamples-development`` (that is, the namespace has a resource quota of 1000 CPU cores and 5TB of memory) when the admin fills in cluster resource templates:

.. prompt:: bash

   flytectl update cluster-resource-attribute --attrFile cra.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_cluster-resource-attribute>` to learn more about the command and its supported flag(s).


To fetch and verify the individual project-domain attributes:

.. prompt:: bash

    flytectl get cluster-resource-attribute -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

Flyte uses these updated values to fill the template fields for the flyteexamples-development namespace.

For other namespaces, the `platform defaults <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__ apply.

.. note::
    The template values, for example, ``projectQuotaCpu`` or ``projectQuotaMemory`` are free-form strings.
    Ensure that they match the template placeholders in your `template file <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml>`__
    for your changes to take effect and custom values to be substituted.

You can view all custom cluster-resource-attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=1>`` and substitute the protocol and host appropriately.

Execution Cluster Label
=======================
This allows forcing a matching execution to consistently execute on a specific Kubernetes cluster for multi-cluster Flyte deployment set-up.

Define an attributes file in `ec.yaml`:

.. code-block:: yaml

    value: mycluster
    domain: development
    project: flyteexamples

Ensure that admin places executions in the flyteexamples project and development domain onto ``mycluster``:

.. prompt:: bash

   flytectl update execution-cluster-label --attrFile ec.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_execution-cluster-label>` to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes:

.. prompt:: bash

    flytectl get execution-cluster-label -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

You can view all custom execution cluster attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=3>`` and substitute the protocol and host appropriately.

Execution Queues
================
Execution queues are defined in
`flyteadmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`__.
These are used for execution placement for constructs like AWS Batch.

The **attributes** associated with an execution queue must match the **tags** for workflow executions. The tags associated with configurable resources are stored in the admin database.

.. prompt:: bash

    flytectl update execution-queue-attribute

Refer to the :ref:`docs <flytectl:flytectl_update_execution-queue-attribute>` to learn more about the command and its supported flag(s).

You can view existing attributes for which tags can be assigned by visiting ``protocol://<host>/api/v1/matchable_attributes?resource_type=2`` and substitute the protocol and host appropriately.

Workflow Execution Config
=========================
This helps with overriding the config used for workflows execution which includes `security context <https://docs.flyte.org/projects/flyteidl/en/latest/protos/docs/core/core.html#securitycontext>`__, `annotations or labels <https://docs.flyte.org/projects/cookbook/en/latest/auto/core/containerization/workflow_labels_annotations.html#sphx-glr-auto-core-containerization-workflow-labels-annotations-py>`__  etc
in the `Worflow execution config <https://github.com/flyteorg/flyteidl/blob/master/gen/pb-go/flyteidl/service/flyteadmin/model_admin_workflow_execution_config.go#L14-L23>`__.
And these can be defined at two levels of project-domain or project-domain-workflow

.. prompt:: bash

    flytectl update workflow-execution-config

Refer to the :ref:`docs <flytectl:flytectl_update_workflow-execution-config>` to learn more about the command and its supported flag(s).

Configuring Service Roles
=========================
You can configure service roles along 3 levels:

#. Project + domain defaults (every execution launched in this project/domain uses this service account)

#. Launch plan default (every invocation of this launch plan uses this service account)

#. Execution time override (overrides at invocation for a specific execution only)

*********
Hierarchy
*********
Increasing specificity defines how matchable resource attributes get applied. The available configurations, in order of decreasing specifity are:

#. Domain, project, workflow name, and launch plan

#. Domain, project, and workflow name

#. Domain and project

#. Domain

Default values for all and per-domain attributes may be specified in the FlyteAdmin config as documented in the :std:ref:`deployment-customizable-resources`.

Example
=======
If the database includes the following:

+------------+--------------+----------+-------------+-----------+
| Domain     | Project      | Workflow | Launch Plan | Tags      |
+============+==============+==========+=============+===========+
| production | widgetmodels |          |             | critical  |
+------------+--------------+----------+-------------+-----------+
| production | widgetmodels | Demand   |             | supply    |
+------------+--------------+----------+-------------+-----------+

Any inbound ``CreateExecution`` requests with **[Domain: Production, Project: widgetmodels, Workflow: Demand]** for any launch plan will have a tag value of "supply".

Any inbound ``CreateExecution`` requests with **[Domain: Production, Project: widgetmodels]** for any workflow other than ``Demand`` and any launch plan will have a tag value "critical".

All other inbound CreateExecution requests will use the default values specified in the FlyteAdmin config (if any).


Configuring K8s Pod
--------------------

There are two approaches to applying the K8s Pod configuration. The **recommended** method is to use Flyte's default PodTemplate scheme. You can do this by creating K8s PodTemplate resource/s that serves as the base configuration for all the task Pods that Flyte initializes. This solution ensures completeness regarding support configuration options and maintainability as new features are added to K8s. 

The legacy technique is to set configuration options in Flyte's K8s plugin configuration. 

.. note ::

    These two approaches can be used simultaneously, where the K8s plugin configuration will override the default PodTemplate values.

*******************************
Using Default K8s PodTemplates
*******************************
`PodTemplate <https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates>`__ is a K8s native resource used to define a K8s Pod. It contains all the fields in the PodSpec, in addition to ObjectMeta to control resource-specific metadata such as Labels or Annotations. They are commonly applied in Deployments, ReplicaSets, etc to define the managed Pod configuration of the resources. Within Flyte, you can leverage this resource to configure Pods created as part of Flyte's task execution. It ensures complete control over Pod configuration, supporting all options available through the resource and ensuring maintainability in future versions.

To initialize a default PodTemplate in Flyte:

#. **Set the 'default-pod-template-name' in FlytePropeller:** This `option <https://docs.flyte.org/en/latest/deployment/cluster_config/flytepropeller_config.html#default-pod-template-name-string>`__ initializes a K8s informer internally to track system PodTemplate updates (creates, updates, etc) so that FlytePropeller is aware of the latest PodTemplate definitions in the K8s environment.
 
#. **Create a PodTemplate resource:** Flyte recognizes PodTemplate definitions with the ``default-pod-template-name`` at two granularities. 

   #. A system-wide configuration can be created in the same namespace that FlytePropeller is running in (typically `flyte`). 

   #. PodTemplates can be applied from the same namespace that the Pod will be created in. FlytePropeller always favours the PodTemplate with the more specific namespace. For example, a Pod created in the ``flytesnacks-development`` namespace will first look for a PodTemplate from the ``flytesnacks-development`` namespace. If that PodTemplate doesn't exist, it will look for a PodTemplate in the same namespace that FlytePropeller is running in (in our example, ``flyte``), and if that doesn't exist, it will begin configuration with an empty PodTemplate.

Flyte configuration supports all the fields available in the PodTemplate resource. It is important to note the PodTemplate definitions need to contain the 'containers' field within the PodSpec. This is required by K8s to be a valid PodTemplate. Currently, Flyte overrides these values during configuration (so the containers are never initialized) with the platform-specific definitions it requires. Hence, a simple `noop` container is enough to satisfy K8s validation requirements.

*********************************
Flyte's K8s Plugin Configuration
*********************************
The FlytePlugins repository defines `configuration <https://github.com/flyteorg/flyteplugins/blob/902b902fcf487f30ebb5dbeee3bb14e17eb0ec21/go/tasks/pluginmachinery/flytek8s/config/config.go#L67-L162>`__ for the Flyte K8s Plugin. They contain a variety of common options for Pod configuration which are applied when constructing a Pod. Typically, these options map one-to-one with K8s Pod fields. This makes it difficult to maintain configuration options as K8s versions change and fields are added/deprecated.

********
Example
********
To better understand how Flyte constructs task execution Pods based on the default PodTemplate and K8s plugin configuration options, let's take an example. 
If you have the default PodTemplate defined in the ``flyte`` namespace (where FlytePropeller instance is running), then it is applied to all Pods that Flyte creates, unless a **more specific** PodTemplate is defined in the namespace where you start the Pod.

An example PodTemplate is shown:

.. code-block:: yaml
    
    apiVersion: v1
    kind: PodTemplate
    metadata:
      name: flyte-template
      namespace: flyte
    template:
      metadata:
        labels:
          - foo
        annotations:
          - foo: initial-value
          - bar: initial-value
      spec:
        containers:
          - name: noop
           image: docker.io/rwgrim/docker-noop
         hostNetwork: false

In addition, the K8s plugin configuration in FlytePropeller defines the default Pod Labels, Annotations, and enables the host networking.

.. code-block:: yaml
    
    plugins:
       k8s:
        default-labels:
          - bar
        default-annotations:
          - foo: overridden-value
          - baz: non-overridden-value
        enable-host-networking-pod: true

To construct a Pod, FlytePropeller initializes a Pod definition using the default PodTemplate. This definition is applied to the K8s plugin configuration values, and any task-specific configuration is overlaid. During the process, when lists are merged, values are appended and when maps are merged, the values are overridden. 
The resultant Pod using the above default PodTemplate and K8s Plugin configuration is shown:

.. code-block:: yaml

    apiVersion: v1
    kind: Pod
    metadata:
      name: example-pod
      namespace: flytesnacks-development
      labels:
        - foo // maintained initial value
        - bar // value appended by k8s plugin configuration
      annotations:
        - foo: overridden-value // value overridden by k8s plugin configuration
        - bar: initial-value // maintained initial value
        - baz: non-overridden-value // value added by k8s plugin configuration
    spec:
      containers:
         // omitted Flyte-specific overridden containers
       hostNetwork: true // overridden by the k8s plugin configuration

The last step in constructing a Pod is to apply any task-specific configuration. These options follow the same rules as merging the default PodTemplate and K8s Plugin configuration (that is, list appends and map overrides). Task-specific options are intentionally robust to provide fine-grained control over task execution in diverse use-cases. Therefore, exploration is beyond this scope and has therefore been omitted from this documentation.
