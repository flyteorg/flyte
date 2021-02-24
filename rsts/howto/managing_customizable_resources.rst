.. _howto-managing-customizable-resources:

########################################################################################
How do I configure my Flyte deployment to have specialized behavior per project/domain?
########################################################################################

As the complexity of your user base grows, you may find yourself tweaking resource assignments based on specific projects, domains and workflows. This document walks through how and in what ways you can configure your Flyte deployment.


***************************
Configurable Resource Types
***************************

Flyte allows these custom settings along the following combination of dimensions

- domain
- project and domain
- project, domain, and name (must be either a workflow name or a launch plan name)

Please see the :ref:`divedeep-projects` document for more information on projects and domains. Along these dimensions, the following settings are configurable. Note that not all three of the combinations above are valid for each of these settings.

- Defaults for task resource requests and limits (when not specified by the author of the task).
- Settings for project-namespaced cluster resource configuration that feeds into Admin's cluster resource manager.
- Execution queues that are used for Dynamic Tasks. Read more about execution queues here, but effectively they're meant to be used with constructs like AWS Batch.
- Determining how workflow executions get assigned to clusters in a multi-cluster Flyte deployment.

The proto definition is the definitive source of which
`matchable attributes <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto>`_
can be customized.

Each of the four above settings are discussed below. Eventually all of these customizations will be overridable using
:std:ref:`flytectl`. Until then, flyte-cli command line options can be used to modify frequent use-cases, and barring
that we show examples using curl.


Task Resources
==============

This includes setting default value for task resource requests and limits for the following resources:

- cpu
- gpu
- memory
- storage

In the absence of an override the global
`default values <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L124,L134>`__
in the flyteadmin config are used.

The override values from the database are assigned at execution time.

To update individual project-domain attributes, use the following as an example:

.. prompt:: bash

    curl --request PUT 'https://flyte.company.net/api/v1/project_domain_attributes/projectname/staging' \
        --header 'Content-Type: application/json' --data-raw \
        '{"attributes":{"matchingAttributes":{"taskResourceAttributes":{"defaults":{"cpu": "1000", "memory": "5000Gi"}, "limits": {"cpu": "4000"}}}}'



Cluster Resources
=================

These are free-form key-value pairs which are used when filling in the templates that Admin feeds into its cluster manager. The keys represent templatized variables in `clusterresource template yaml <https://github.com/flyteorg/flyteadmin/tree/master/sampleresourcetemplates>`__ and the values are what you want to see filled in.

In the absence of custom override values, templateData from the `flyteadmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L154,L159>`__ is used as a default.

Note that these settings can only take on domain, or a project and domain specificity. Since Flyte has not tied in the notion of a workflow or a launch plan to any Kubernetes constructs, specifying a workflow or launch plan name doesn't make any sense.

Running the following, will make it so that when Admin fills in cluster resource templates, the K8s namespace ``flyteexamples-development`` will have a resource quota of 1000 CPU cores and 5TB of memory.

.. prompt:: bash

    flyte-cli -h localhost:30081 -p flyteexamples -d development update-cluster-resource-attributes  \
    --attributes projectQuotaCpu 1000 --attributes projectQuotaMemory 5000Gi


These values will in turn be used to fill in the template fields, for example:

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml

from the base of this repository for the ``flyteexamples-development`` namespace and that namespace only.
For other namespaces, the `platform defaults <https://github.com/flyteorg/flyte/blob/c9b9fad428e32255b6839e3244ca8f09d57536ae/kustomize/base/single_cluster/headless/config/admin/cluster_resources.yaml>`__ would still be applied.

=======

    flyte-cli -h localhost:30081 -p flyteexamples -d development update-cluster-resource-attributes  \
    --attributes projectQuotaCpu 1000 --attributes projectQuotaMemory 5000Gi


These values will in turn be used to fill in the template fields, for example:

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml

from the base of this repository for the ``flyteexamples-development`` namespace and that namespace only.
For other namespaces, the `platform defaults <https://github.com/flyteorg/flyte/blob/c9b9fad428e32255b6839e3244ca8f09d57536ae/kustomize/base/single_cluster/headless/config/admin/cluster_resources.yaml>`__ would still be applied.

.. note::

    The template values, e.g. ``projectQuotaCpu`` or ``projectQuotaMemory`` are freeform strings. You must ensure that
    they match the template placeholders in your `template file <https://github.com/flyteorg/flyte/blob/master/kustomize/base/single_cluster/headless/config/clusterresource-templates/ab_project-resource-quota.yaml>`__
    for your changes to take effect.

Execution Queues
================

Execution queues are use to determine where tasks yielded by a dynamic :py:func:`flytekit:flytekit.maptask` run.

Execution queues themselves are currently defined in the
`flyteadmin config <https://github.com/flyteorg/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`__.

The **attributes** associated with an execution queue must match the **tags** for workflow executions. The tags are associated with configurable resources
stored in the Admin database.

.. prompt:: bash

    flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-queue-attributes  \
    --tags critical --tags gpu_intensive

You can view existing attributes for which tags can be assigned by visting `http://localhost:30081/api/v1/matchable_attributes?resource_type=3 <http://localhost:30081/api/v1/matchable_attributes?resource_type=3>`__.

Execution Cluster Label
=======================

This allows forcing a matching execution to always execute on a specific kubernetes cluster.

You can set this using flyte-cli:

.. prompt:: bash

   flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-cluster-label --value mycluster


*********
Hierarchy
*********

Increasing specifity defines how matchable resource attributes get applied. The available configurations, in order of decreasing specifity are:


#. Domain, project, workflow name and launch plan.

#. Domain, project and workflow name

#. Domain and project

#. Domain

Default values for all and per-domain attributes may be specified in the flyteadmin config as documented above.


Example
=======

Let's say that our database includes the following

+------------+--------------+----------+-------------+-----------+
| Domain     | Project      | Workflow | Launch Plan | Tags      |
+============+==============+==========+=============+===========+
| production | widgetmodels |          |             | critical  |
+------------+--------------+----------+-------------+-----------+
| production | widgetmodels | Demand   |             | supply    |
+------------+--------------+----------+-------------+-----------+

Any inbound CreateExecution requests with **[Domain: Production, Project: widgetmodels, Workflow: Demand]** for any launch plan would have a tag value of "supply".
Any inbound CreateExecution requests with **[Domain: Production, Project: widgetmodels]** for any workflow other than Demand and for any launch plan would have a tag value of "critical".

All other inbound CreateExecution requests would use the default values specified in the flyteadmin config (if any).

*********
Debugging
*********

Use the `get <https://github.com/flyteorg/flyteidl/blob/ba13965bcfbf7e7bfce40664800aaf1f2a1088a1/protos/flyteidl/service/admin.proto#L395>`__ endpoint
to see if overrides exist for a specific resource.

E.g. `https://example.com/api/v1/project_domain_attributes/widgetmodels/production?resource_type=2 <https://example.com/api/v1/project_domain_attributes/widgetmodels/production?resource_type=2>`__

To get the global state of the world, use the list all endpoint, e.g. `https://example.com/api/v1/matchable_attributes?resource_type=2 <https://example.com/api/v1/matchable_attributes?resource_type=2>`__.

The resource type enum (int) is defined in the `proto <https://github.com/flyteorg/flyteidl/blob/ba13965bcfbf7e7bfce40664800aaf1f2a1088a1/protos/flyteidl/admin/matchable_resource.proto#L8,L20>`__.
