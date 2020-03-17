.. _managing_customizable_resources:

##################################
Configuring customizable resources
##################################

As the complexity of your user-base grows, you may find yourself tweaking resource assignments based on specific projects, domains and workflows.
This document walks through how to use MatchableResource attributes to customize your workflow execution environment.

Flyteadmin allows for overrides of task resource request and limit defaults, kubernetes cluster resource configuration,
dynamic task execution queues and specifying executions on specific kubernetes clusters. These can all be overriden for specific combinations
of domain; domain and project; domain, project and workflow (name); and domain, project, workflow (name), and launch plan.

***************************
Configurable Resource Types
***************************

The proto definition is the definitive source of
`matchable attributes <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto>`_
which can be customized. See below for a detailed explanation

Task Resources
==============

This includes setting default value for task resource requests and limits for the following resources:

- cpu

- gpu

- memory

- storage

In the absence of an override the global
`default values <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L124,L134>`_
in the flyteadmin config are used.

The override values from the database are assigned at execution time.


Cluster Resources
=================

These are free-form key-value pairs which are used when creating project-domain based resources on Flyte kubernetes clusters.

The keys represent templatized variables in `clusterresource template yaml <https://github.com/lyft/flyteadmin/tree/master/sampleresourcetemplates>`_

In the absence of custom override values, templateData from the `flyteadmin config <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L154,L159>`_ is used as a default.


Execution Queues
================

Execution queues are use to determine where dynamic tasks run.

Execution queues themselves are currently defined in the
`flyteadmin config <https://github.com/lyft/flyteadmin/blob/6a64f00315f8ffeb0472ae96cbc2031b338c5840/flyteadmin_config.yaml#L97,L106>`_.

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
