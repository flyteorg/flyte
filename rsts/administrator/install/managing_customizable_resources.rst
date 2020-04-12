.. _managing_customizable_resources:

##################################
Configuring customizable resources
##################################

As the complexity of your user base grows, you may find yourself tweaking resource assignments based on specific projects, domains and workflows. This document walks through how and in what ways you can configure your Flyte deployment.


***************************
Configurable Resource Types
***************************

Flyte allows these custom settings along the following combination of dimensions

- domain
- project and domain
- project, domain, and name (must be either a workflow name or a launch plan name)

Please see the :ref:`concepts` document for more information on projects and domains. Along these dimensions, the following settings are configurable. Note that not all three of the combinations above are valid for each of these settings.

- Defaults for task resource requests and limits (when not specified by the author of the task).
- Settings for the cluster resource configuration that feeds into Admin's cluster resource manager.
- Execution queues that are used for Dynamic Tasks. Read more about execution queues here, but effectively they're meant to be used with constructs like AWS Batch.
- Determining how workflow executions get assigned to clusters in a multi=cluster Flyte deployment.

The proto definition is the definitive source of which
`matchable attributes <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto>`_
can be customized.

Each of the four above settings are discussed below.  Also, since the flyte-cli tool does not yet hit these endpoints, we are including some sample ``curl`` commands for administrators to reference.


Task Resources
==============

This includes setting default value for task resource requests and limits for the following resources:

- cpu
- gpu
- memory
- storage

In the absence of an override the global
`default values <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L124,L134>`__
in the flyteadmin config are used.

The override values from the database are assigned at execution time.


Cluster Resources
=================

These are free-form key-value pairs which are used when filling in the templates that Admin feeds into its cluster manager. The keys represent templatized variables in `clusterresource template yaml <https://github.com/lyft/flyteadmin/tree/master/sampleresourcetemplates>`__ and the values are what you want to see filled in.

In the absence of custom override values, templateData from the `flyteadmin config <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L154,L159>`__ is used as a default.

Note that these settings can only take on domain, or a project and domain specificity. Project & domain together in Flyte form Kubernetes namespaces. Since Flyte has not tied in the notion of a workflow or a launch plan to any Kubernetes constructs, specifying a workflow or launch plan name doesn't make any sense.


Command
-------
Running the following, will make it so that when Admin fills in cluster resource templates, the K8s namespace ``projectname-staging`` will have a resource quota of 1000 CPU cores and 5TB of memory.

.. code-block:: console

    curl --request PUT 'https://flyte.company.net/api/v1/project_domain_attributes/projectname/staging' --header 'Content-Type: application/json' --data-raw '{"attributes":{"matchingAttributes":{"clusterResourceAttributes":{"attributes":{"projectQuotaCpu": "1000", "projectQuotaMemory": "5000Gi"}}}}}'


Execution Queues
================

Execution queues are use to determine where dynamic tasks run.

Execution queues themselves are currently defined in the
`flyteadmin config <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`__.

The **attributes** associated with an execution queue must match the **tags** for workflow executions. The tags are associated with configurable resources
stored in the admin database.


Execution Cluster Label
=======================

This allows forcing a matching execution to always execute on a specific kubernetes cluster.


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
Any inbound CreateExecution requests with **[Domain: Production, Project: widgetmodels]** for any workflow other than Deman and for any launch plan would have a tag value of "critical".

All other inbound CreateExecution requests would use the default values specified in the flyteadmin config (if any).

*********
Debugging
*********

Use the `get <https://github.com/lyft/flyteidl/blob/ba13965bcfbf7e7bfce40664800aaf1f2a1088a1/protos/flyteidl/service/admin.proto#L395>`_ endpoint
to see if overrides exist for a specific resource.

E.g. `https://example.com/api/v1/project_domain_attributes/widgetmodels/production?resource_type=2 <https://example.com/api/v1/project_domain_attributes/widgetmodels/production?resource_type=2>`_

To get the global state of the world, use the list all endpoint, e.g. `https://example.com/api/v1/matchable_attributes?resource_type=2 <https://example.com/api/v1/matchable_attributes?resource_type=2>`_.

The resource type enum (int) is defined in the `proto <https://github.com/lyft/flyteidl/blob/ba13965bcfbf7e7bfce40664800aaf1f2a1088a1/protos/flyteidl/admin/matchable_resource.proto#L8,L20>`_.
