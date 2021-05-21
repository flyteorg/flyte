"""
Monitoring
----------

.. tip:: The Flyte core team publishes a maintains Grafana dashboards built using Prometheus data sources and can be found `here <https://grafana.com/grafana/dashboards?search=flyte>`__.

Flyte Backend is  written in Golang and exposes stats using Prometheus. The Stats themselves are labeled with the Workflow, Task, Project & Domain wherever appropriate.

The dashboards are divided into primarily 2 types:

- User facing dashboards. These are dashboards that can be used to triage/investigate/observe performance and characterisitics of Workflows and tasks.
  The User facing dashboard is published under Grafana marketplace ID `13980 <https://grafana.com/grafana/dashboards/13980>`_

- System Dashboards. These dashboards are useful for the system maintainer to maintain their Flyte deployments. These are further divided into
        - DataPlane/FlytePropeller dashboards published @ `13979 <https://grafana.com/grafana/dashboards/13979>`_
        - ControlPlane/Flyteadmin dashboards published @ `13981 <https://grafana.com/grafana/dashboards/13981>`_

These are basic dashboards and do no include all the metrics that are exposed by Flyte. Please help us improve the dashboards by contributing to them. Refer to the build scripts `here <https://github.com/flyteorg/flyte/tree/master/stats>`__.
"""
