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

The keys represent templatized variables in `cluster resource template YAML <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L737,L760>`__ and the values are what you want to see filled in.

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


*********
Hierarchy
*********
Increasing specificity defines how matchable resource attributes get applied. The available configurations, in order of decreasing specifity are:

#. Domain, project, workflow name, and launch plan.

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
