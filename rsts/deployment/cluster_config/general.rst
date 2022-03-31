.. _deployment-cluster-config-general:

Configuring Your Flyte Deployment
----------------------------------

***************************
Configurable Resource Types
***************************

Many platform specifications such as task resource defaults, project namespace Kubernetes quota, and more can all be
assigned using default values or custom overrides. Defaults are specified in the FlyteAdmin config and
overrides for specific projects can be registered with the FlyteAdmin service.

Specifically, flyte allows these custom settings along increasing levels of specificity:

- domain
- project and domain
- project, domain, and workflow name
- project, domain, workflow name and launch plan name

Please see the :ref:`concepts/control_plane:control plane` document to get to know about projects and domains.
Along these dimensions, the following settings are configurable:

Task Resources
==============
Configuring task resources includes setting default values for unspecified task requests and limits.
Task resources also include limits which specify the maximum value either a task request or limit can have.

- cpu
- gpu
- memory
- storage
- `ephemeral storage <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#local-ephemeral-storage>`__

In the absence of an override, the global
`default values <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L35,L43>`__
in the FlyteAdmin config are used.

The override values from the database are assigned at execution, rather than registration time.

To update individual project-domain attributes, first define a ``tra.yaml`` file with your overrides:

.. code-block:: yaml

    defaults:
        cpu: "1"
        memory: 150Mi
    limits:
        cpu: "2"
        memory: 450Mi
    project: flyteexamples
    domain: development

Then, use the following command for your reference.

.. prompt:: bash

    flytectl update task-resource-attribute --attrFile tra.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_task-resource-attribute>` to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes, similarly you can run:

.. prompt:: bash

    flytectl get task-resource-attribute -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

You can view all custom task-resource-attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=0>`` and substitute the protocol and host appropriately.

Cluster Resources
=================
These are free-form key-value pairs used when filling the templates that the admin feeds into the process to sync kubernetes resources, known as cluster manager.

The keys represent templatized variables in `cluster resource template YAML <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L737,L760>`__ and the values are what you want to see filled in.


In the absence of custom override values, ``templateData`` from the `FlyteAdmin config <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__
is used as a default. Note, these defaults are specified by domain but will be applied to every project-domain namespace combination.

.. note::
    The above-referenced settings can only take on domain, project, and domain specificity.
    Since Flyte has not tied in the notion of a workflow or a launch plan to any Kubernetes constructs, specifying a workflow or launch plan name doesn't make any sense.
    This is a departure from the usual hierarchy for customizable resources.

Define an attributes file like the the following in ``cra.yaml``

.. code-block:: yaml

    attributes:
        projectQuotaCpu: "1000"
        projectQuotaMemory: 5TB
    domain: development
    project: flyteexamples

Then run the below command to ensure that when the admin fills in cluster resource templates, the Kubernetes namespace ``flyteexamples-development`` will have a resource quota of 1000 CPU cores and 5TB of memory.

.. prompt:: bash

   flytectl update cluster-resource-attribute --attrFile cra.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_cluster-resource-attribute>` to learn more about the command and its supported flag(s).


To fetch and verify the individual project-domain attributes, similarly you can run:

.. prompt:: bash

    flytectl get cluster-resource-attribute -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

The above-updated values will, in turn, be used to fill in the template fields for the flyteexamples-development namespace.

For other namespaces, the `platform defaults <https://github.com/flyteorg/flyte/blob/1e3d515550cb338c2edb3919d79c6fa1f0da5a19/charts/flyte-core/values.yaml#L719,L734>`__ apply.

.. note::
    The template values, e.g. ``projectQuotaCpu`` or ``projectQuotaMemory`` are freeform strings.
    You must ensure that they match the template placeholders in your `template file <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml>`__
    for your changes to take effect and custom values to be substituted.

You can view all custom cluster-resource-attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=1>`` and substitute the protocol and host appropriately.

Execution Cluster Label
=======================
This allows forcing a matching execution to consistently execute on a specific Kubernetes cluster for multi-cluster Flyte deployment set-up.

Define an attributes file like so in `ec.yaml`:

.. code-block:: yaml

    value: mycluster
    domain: development
    project: flyteexamples

Then run the below command to ensure that admin places executions in the flyteexamples project and development domain onto ``mycluster``.

.. prompt:: bash

   flytectl update execution-cluster-label --attrFile ec.yaml

Refer to the :ref:`docs <flytectl:flytectl_update_execution-cluster-label>` to learn more about the command and its supported flag(s).

To fetch and verify the individual project-domain attributes, similarly you can run:

.. prompt:: bash

    flytectl get execution-cluster-label -p flyteexamples -d development

Refer to the :ref:`docs <flytectl:flytectl_get_task-resource-attribute>` to learn more about the command and its supported flag(s).

You can view all custom execution cluster attributes by visiting ``protocol://<host/api/v1/matchable_attributes?resource_type=3>`` and substitute the protocol and host appropriately.

Execution Queues
================
Execution queues themselves are currently defined in the
`flyteadmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`__.
These are used for execution placement for constructs like AWS Batch.

The **attributes** associated with an execution queue must match the **tags** for workflow executions. The tags are associated with configurable resources
stored in the admin database.

.. prompt:: bash

    flytectl update execution-queue-attribute

    Refer to the :ref:`docs <flytectl:flytectl_update_execution-queue-attribute>` to learn more about the command and its supported flag(s).

You can view existing attributes for which tags can be assigned by visiting ``protocol://<host>/api/v1/matchable_attributes?resource_type=2`` and substitute the protocol and host appropriately.

*********
Hierarchy
*********
Increasing specificity defines how matchable resource attributes get applied. The available configurations, in order of decreasing specifity, are:

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
