"""
#################################
Configuring Your Flyte Deployment
#################################

As the complexity of your user base grows, you may find yourself tweaking resource assignments based on specific projects, domains, and workflows. This document walks through how and in what ways you can configure your Flyte deployment.

.. _config_resource_types:

***************************
Configurable Resource Types
***************************
Flyte allows these custom settings along with the following combination of dimensions:

- domain
- project and domain
- project, domain, and name (must be either the name of a workflow name or a launch plan)

Please see the :doc:`Control Plane <flyte:concepts/control_plane>` document to get to know about projects and domains. 
Along these dimensions, the following settings are configurable. 

.. admonition:: Alert
    Not all three of the combinations mentioned above are valid for each of these settings.

- Defaults are used for task resource requests and limits (when not specified by the author of the task).
- Settings for project-namespaced cluster resource configuration that feeds into admin's cluster resource manager.
- Execution queues that are used for dynamic workflows. Effectively, they're meant to be used with constructs like AWS Batch.

.. note::
    Execution queues are used to determine where tasks yielded by a :py:func:`flytekit:flytekit.dynamic` workflow or ``map task`` run.

- Determining how workflow executions are assigned to clusters in a multi-cluster Flyte deployment.

.. tip::
  The proto definition is the definitive source encapsulating which :ref:`Matchable Resource <protos/docs/admin/admin:matchableresource>` attributes can be customized.

Each of the four above settings is discussed below. 

Task Resources
==============
Configuring task resources includes setting default values for the requests and limits for the following resources:

- cpu
- gpu
- memory
- storage

In the absence of an override, the global
`default values <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L124,L134>`__
in the FlyteAdmin config are used.

The override values from the database are assigned at execution time.

To update individual project-domain attributes, use the following command for your reference.

.. prompt:: bash

    curl --request PUT 'https://flyte.company.net/api/v1/project_domain_attributes/projectname/staging' \
        --header 'Content-Type: application/json' --data-raw \
        '{"attributes":{"matchingAttributes":{"taskResourceAttributes":{"defaults":{"cpu": "1000", "memory": "5000Gi"}, "limits": {"cpu": "4000"}}}}'

.. tip::
    The equivalent ``flytectl`` command is:

    .. prompt:: bash
        flytectl update task-resource-attribute

        Refer to the :ref:`docs <flytectl:flytectl_update_task-resource-attribute>` to learn more about the command and its supported flag(s).

Cluster Resources
=================
These are free-form key-value pairs that are used when filling in the templates that the admin feeds into its cluster manager. The keys represent templatized variables in `cluster resource template YAML <https://github.com/flyteorg/flyteadmin/tree/master/sampleresourcetemplates>`__ and the values are what you want to see filled in.

In the absence of custom override values, ``templateData`` from the `FlyteAdmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L154,L159>`__ is used as a default.

.. note::
    The above-referenced settings can only take on domain, project, and domain specificity. Since Flyte has not tied in the notion of a workflow or a launch plan to any Kubernetes constructs, specifying a workflow or launch plan name doesn't make any sense.

Running the following will ensure that when the admin fills in cluster resource templates, the Kubernetes namespace ``flyteexamples-development`` will have a resource quota of 1000 CPU cores and 5TB of memory.

.. prompt:: bash

    flyte-cli -h localhost:30081 -p flyteexamples -d development update-cluster-resource-attributes  \
    --attributes projectQuotaCpu 1000 --attributes projectQuotaMemory 5000Gi

.. tip::
   The equivalent ``flytectl`` command is:

   .. prompt:: bash
       flytectl update cluster-resource-attribute 
      
   Refer to the :ref:`docs <flytectl:flytectl_update_cluster-resource-attribute>` to learn more about the command and its supported flag(s).

The above-updated values will, in turn, be used to fill in the template fields. 

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml

The values can now be copied from the base of this repository to the ``flyteexamples-development`` namespace only.

For other namespaces, the `platform defaults <https://github.com/flyteorg/flyte/blob/c9b9fad428e32255b6839e3244ca8f09d57536ae/kustomize/base/single_cluster/headless/config/admin/cluster_resources.yaml>`__ apply.

.. note::
    The template values, e.g. ``projectQuotaCpu`` or ``projectQuotaMemory`` are freeform strings. You must ensure that they match the template placeholders in your `template file <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml>`__
    for your changes to take effect.

Execution Queues
================
Execution queues themselves are currently defined in the
`flyteadmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`__.

The **attributes** associated with an execution queue must match the **tags** for workflow executions. The tags are associated with configurable resources
stored in the admin database.

.. prompt:: bash

    flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-queue-attributes  \
    --tags critical --tags gpu_intensive

.. tip::
    The equivalent command in ``flytectl`` is:

    .. prompt:: bash
        flytectl update execution-queue-attribute

    Refer to the :ref:`docs <flytectl:flytectl_update_execution-queue-attribute>` to learn more about the command and its supported flag(s).

You can view existing attributes for which tags can be assigned by visiting ``protocol://<host>/api/v1/matchable_attributes?resource_type=3``.

Execution Cluster Label
=======================
This allows forcing a matching execution to consistently execute on a specific Kubernetes cluster.

You can set this using flyte-cli:

.. prompt:: bash

   flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-cluster-label --value mycluster

.. tip::
   The equivalent command in ``flytectl`` is:

   .. prompt:: bash
      flytectl update execution-cluster-label

   Refer to the :ref:`docs <flytectl:flytectl_update_execution-cluster-label>` to learn more about the command and its supported flag(s).

*********
Hierarchy
*********
Increasing specificity defines how matchable resource attributes get applied. The available configurations, in order of decreasing specifity, are:

#. Domain, project, workflow name, and launch plan.

#. Domain, project, and workflow name

#. Domain and project

#. Domain

Default values for all and per-domain attributes may be specified in the FlyteAdmin config as documented in the :ref:`config_resource_types`.

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

*********
Debugging
*********
To get the matchable resources of :ref:`execution queue attributes <flytectl_get_execution-queue-attribute>`, run the command:

.. prompt:: bash
    flytectl get execution-queue-attribute

.. note::
    Alternatively, you can also hit the URL: ``protocol://<host/api/v1/project_domain_attributes/widgetmodels/production?resource_type=2>``.

To get the global state of the world, list all endpoints. For example, visit ``protocol://<host>/api/v1/matchable_attributes?resource_type=2``.

The resource type enum (int) is defined in the :ref:`Matchable Resource <protos/docs/admin/admin:matchableresource>`.

"""
