.. _howto-monitoring:

######################################
How do I monitor my Flyte deployment?
######################################

.. tip:: The flyte core team publishes a maintains Grafana dashboards built using prometheus data source and can be found `here <https://grafana.com/grafana/dashboards?search=flyte>`_.

Flyte Backend is  written in Golang and exposes stats using Prometheus. The Stats themselves are labeled with the Workflow, Task, Project & Domain whereever appropriate.

The dashboards are divided into primarily 2 types

- User facing dashboards. These are dashboards that a user can use to triage/investigate/observe performance and characterisitics for their Workflows and tasks.
  The User facing dashboard is published under Grafana marketplace ID `13980 <https://grafana.com/grafana/dashboards/13980>`_

- System Dashboards. These dashboards are useful for the system maintainer to maintain their Flyte deployments. These are further divided into
        - DataPlane/FlytePropeller dashboards published @ `13979 <https://grafana.com/grafana/dashboards/13979>`_
        - ControlPlane/Flyteadmin dashboards published @ `13981 <https://grafana.com/grafana/dashboards/13981>`_

These are basic dashboards and do no include all the metrics that are exposed by Flyte. You can contribute to the dashboards and help us improve them - by referring to the build scripts `here <https://github.com/flyteorg/flyte/tree/master/stats>`_.
