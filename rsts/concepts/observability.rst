.. _divedeep-observability:

Metrics for your executions
===========================

.. tip:: Refer to the :ref:`deployment-cluster-config-monitoring` to see how to use prebuilt dashboards published to Grafana Marketplace. The following section explains some other metrics that are very important.

Flyte-Provided Metrics
~~~~~~~~~~~~~~~~~~~~~~~
Whenever you run a workflow, Flyte Platform automatically emits high-level metrics. These metrics follow a consistent schema and
aim to provide visibility into aspects of the Platform which might otherwise be opaque.  These metrics will help users diagnose whether an issue is inherent to the Platform or to one's own task or workflow implementation. We will be adding to this set of metrics as we implement the capabilities to pull more data from the system, so keep checking back for new stats!

At a highlevel, workflow execution goes through the following discrete steps:

.. image:: https://raw.githubusercontent.com/lyft/flyte/assets/img/flyte_wf_timeline.svg?sanitize=true

===================================  ==================================================================================================================================
                       Description of main events for workflow execution
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------
               Events                                                              Description
===================================  ==================================================================================================================================
Acceptance                           Measures the time between when we receive service call to create an Execution (Unknown) and when it has moved to Queued.
Transition Latency                   Measures the latency between two consecutive node executions, the time spent in Flyte engine.
Queuing Latency                      Measures the latency between the time a node's been queued to the time the handler reported the executable moved to running state.
Task Execution                       Actual time spent executing user code
Repeat steps 2-4 for every task
Transition Latency                   See #2
Completion Latency                   Measures the time between when the WF moved to succeeding/failing state and when it finally moved to a terminal state.
===================================  ==================================================================================================================================


==========================================================  ===========  ===============================================================================================================================================================
                    Flyte Stats Schema
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
                    Prefix                                     Type                                           Description
==========================================================  ===========  ===============================================================================================================================================================
``propeller.all.workflow.acceptance-latency-ms``            Timer (ms)   Measures the time between when we receive service call to create an Execution (Unknown) and when it has moved to Queued.
``propeller.all.node.queueing-latency-ms``                  Timer (ms)   Measures the latency between the time a node's been queued to the time the handler reported the executable moved to running state.
``propeller.all.node.transition-latency-ms``                Timer (ms)   Measures the latency between two consecutive node executions, the time spent in Flyte engine.
``propeller.all.workflow.completion-latency-ms``            Timer (ms)   Measures the time between when the WF moved to succeeding/failing state and when it finally moved to a terminal state.
``propeller.all.node.success-duration-ms``                  Timer (ms)   Actual time spent executing user code (when the node ends with success phase).
``propeller.all.node.success-duration-ms-count``            Counter      Count of how many times a node success was reported.
``propeller.all.node.failure-duration-ms``                  Timer (ms)   Actual time spent executing user code (when the node ends with failure phase).
``propeller.all.node.failure-duration-ms-count``            Counter      Count of how many times a node failure was reported.

==========================================================  ===========  ===============================================================================================================================================================

All the above stats are automatically tagged with the following fields for further scoping.  This includes user-produced stats.  Users
can also provide additional tags (or override tags) for custom stats.

.. _task_stats_tags:

===============  =================================================================================
                     Flyte Stats Tags
--------------------------------------------------------------------------------------------------
      Tag                                                 Description
===============  =================================================================================
wf               This is the name of the workflow that was executing when this metric was emitted.
                 ``{{project}}:{{domain}}:{{workflow_name}}``
===============  =================================================================================

User Stats With Flyte
~~~~~~~~~~~~~~~~~~~~~~
The workflow parameters object that the SDK injects into the various tasks has a statsd handle that users should call
to emit stats related to their workflows not captured by the default metrics. The usual caveats around cardinality apply of course.

.. todo: Reference to flytekit task stats

Users are encouraged to avoid creating their own stats handlers.  These can pollute the general namespace if not done 
correctly, and also can accidentally interfere with production stats of live services, causing pages and wreaking 
havoc in general.  In fact, if you're using any libraries that emit stats, it's best to turn them off if possible.
