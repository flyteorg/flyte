.. _deployment-configuration-monitoring:

Monitoring
----------

.. tags:: Infrastructure, Advanced

.. tip:: The Flyte core team publishes and maintains Grafana dashboards built using Prometheus data sources, which can be found `here <https://grafana.com/grafana/dashboards?search=flyte>`__.

Metrics for Executions
======================

Flyte-provided Metrics
~~~~~~~~~~~~~~~~~~~~~~

Whenever you run a workflow, Flyte platform automatically emits high-level metrics. These metrics follow a consistent schema and aim to provide visibility into aspects of the platform which might otherwise be opaque.
These metrics help users diagnose whether an issue is inherent to the platform or one's own task or workflow implementation.

We will be adding to this set of metrics as we implement the capabilities to pull more data from the system, so keep checking back for new stats!

At a high level, workflow execution goes through the following discrete steps:

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/monitoring/flyte_wf_timeline.svg?sanitize=true

===================================  ==================================================================================================================================
                       Description of main events for workflow execution
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------
               Events                                                              Description
===================================  ==================================================================================================================================
Acceptance                           Measures the time consumed from receiving a service call to creating an Execution (Unknown) and moving to QUEUED.
Transition Latency                   Measures the latency between two consecutive node executions; the time spent in Flyte engine.
Queuing Latency                      Measures the latency between the node moving to QUEUED and the handler reporting the executable moving to RUNNING state.
Task Execution                       Actual time spent executing the user code.
Repeat steps 2-4 for every task
Transition Latency                   See #2
Completion Latency                   Measures the time consumed by a workflow moving from SUCCEEDING/FAILING state to TERMINAL state.
===================================  ==================================================================================================================================


==========================================================  ===========  ===============================================================================================================================================================
                    Flyte Stats Schema
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
                    Prefix                                     Type                                           Description
==========================================================  ===========  ===============================================================================================================================================================
``propeller.all.workflow.acceptance-latency-ms``            Timer (ms)   Measures the time consumed from receiving a service call to creating an Execution (Unknown) and moving to QUEUED.
``propeller.all.node.queueing-latency-ms``                  Timer (ms)   Measures the latency between the node moving to QUEUED and the handler reporting the executable moving to RUNNING state.
``propeller.all.node.transition-latency-ms``                Timer (ms)   Measures the latency between two consecutive node executions; the time spent in Flyte engine.
``propeller.all.workflow.completion-latency-ms``            Timer (ms)   Measures the time consumed by a workflow moving from SUCCEEDING/FAILING state to TERMINAL state.
``propeller.all.node.success-duration-ms``                  Timer (ms)   Actual time spent executing user code (when the node ends with SUCCESS state).
``propeller.all.node.success-duration-ms-count``            Counter      The number of times a node success has been reported.
``propeller.all.node.failure-duration-ms``                  Timer (ms)   Actual time spent executing user code (when the node ends with FAILURE state).
``propeller.all.node.failure-duration-ms-count``            Counter      The number of times a node failure has been reported.

==========================================================  ===========  ===============================================================================================================================================================

All the above stats are automatically tagged with the following fields for further scoping. This includes user-produced stats.
Users can also provide additional tags (or override tags) for custom stats.

.. _task_stats_tags:

===============  =================================================================================
                     Flyte Stats Tags
--------------------------------------------------------------------------------------------------
      Tag                                                 Description
===============  =================================================================================
wf               Name of the workflow that was executing when this metric was emitted.
                 ``{{project}}:{{domain}}:{{workflow_name}}``
===============  =================================================================================

User Stats With Flyte
~~~~~~~~~~~~~~~~~~~~~~

The workflow parameters object that the SDK injects into various tasks has a ``statsd`` handle that users should call
to emit stats of their workflows not captured by the default metrics. The usual caveats around cardinality apply, of course.

.. todo: Reference to Flytekit task stats

Users are encouraged to avoid creating their own stats handlers.
If not done correctly, these can pollute the general namespace and accidentally interfere with the production stats of live services, causing pages and wreaking havoc.
If you're using any libraries that emit stats, it's best to turn them off if possible.


Use Published Dashboards to Monitor Flyte Deployment
====================================================

Flyte Backend is written in Golang and exposes stats using Prometheus. The stats are labeled with workflow, task, project & domain, wherever appropriate.

Both ``flyteadmin`` and ``flytepropeller`` are instrumented to expose metrics. To visualize these metrics, Flyte provides three Grafana dashboards, each with a different focus:

- **User-facing dashboards**: Dashboards that can be used to triage/investigate/observe performance and characteristics of workflows and tasks.
  The user-facing dashboard is published under ID `13980 <https://grafana.com/grafana/dashboards/13980>`__ in the Grafana marketplace.

- **System Dashboards**: Dashboards that are useful for the system maintainer to investigate the status and performance of their Flyte deployments. These are further divided into:
        - `DataPlane/FlytePropeller <https://grafana.com/grafana/dashboards/13979>`__: execution engine status and performance.
        - `ControlPlane/Flyteadmin <https://grafana.com/grafana/dashboards/13981>`__: API-level monitoring.

The corresponding JSON files for each dashboard are also located at ``deployment/stats/prometheus``.

.. note::

    The dashboards are basic dashboards and do not include all the metrics exposed by Flyte.
    Feel free to use the scripts provided `here <https://github.com/flyteorg/flyte/tree/master/stats>`__ to improve and -hopefully- contribute the improved dashboards.

How to use the dashboards
~~~~~~~~~~~~~~~~~~~~~~~~~

1. We recommend installing and configuring the Prometheus operator as described in `their docs <https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/user-guides/getting-started.md>`__.
This is especially true if you plan to use the Service Monitors provided by the `flyte-core <https://github.com/flyteorg/flyte/blob/master/charts/flyte-core/templates/propeller/service-monitor.yaml>`__ Helm chart.

2. Enable the Prometheus instance to use Service Monitors in the namespace where Flyte is running, configuring the following keys in the ``prometheus`` resource:

.. code-block:: yaml

   spec:
    serviceMonitorSelector: {}
    serviceMonitorNamespaceSelector: {}

.. note::

   The above example configuration lets Prometheus use any ``ServiceMonitor`` in any namespace in the cluster. Adjust the configuration to reduce the scope if needed.

3. Once you have installed and configured the Prometheus operator, enable the Service Monitors in the Helm chart by configuring the following keys in your ``values`` file:

.. code-block:: yaml

   flyteadmin:
     serviceMonitor:
       enabled: true
   
   flytepropeller:
     serviceMonitor:
       enabled: true

.. note::

   By default, the ``ServiceMonitor`` is configured with a ``scrapeTimeout`` of 30s and ``interval`` of 60s. You can customize these values if needed.

With the above configuration in place you should be able to import the dashboards in your Grafana instance.

