.. _divedeep-execution-timeline:

########################################
Timeline of a workflow execution
########################################

.. image:: https://raw.githubusercontent.com/lyft/flyte/assets/img/flyte_wf_timeline.svg?sanitize=true


The illustration above refers to a simple workflow, with 2 nodes N1 & N2. This can be represented as follows,

.. mermaid::

    graph LR;
         Start --> N1;
         N1 --> N2;
         N2 --> End;


Acceptance
===========
Every Workflow starts in the ``Acceptance`` phase. Acceptance refers to the time it takes for the workflow to start execution in FlytePropeller from the time the FlyteAdmin service receives a request.
Usually in this section the K8s queuing latency is the highest component. But overall Acceptance latency of <5s is desirable.

**This is one of the latencies that the core team is working on optimizing**

Transition
===========
Transition latency refers to the time between success node executions, i.e between N1 & N2. For the first node this latency encapsulate executing the ``Start node`` as well. For the last node to End node, this encapsulates executing ``End node`` as well.
the latency involves, time it takes to,
#. Gather outputs for a node after the node completes
#. Send an observation event to FlyteAdmin. Failing to do so is regarded as an error and will be tried untill successful or we exhaust system max retries.
#. Persist data to Kubernetes
#. Receive the persisted object back from Kubernetes (as this process is eventually consistent using informer caches)
#. Gather inputs for a node before the node starts
#. Send a queued event for the next node to Flyteadmin for observability

**This is one of the latencies that the core team is working on optimizing**

Queuing Latency
================
This is the time it takes for Kubernetes to start the pod, or the other service to start the job or an HTTP throttle to be met or any rate-limiting to be overcome. This
is usually tied to available resources and quota and is out of control for Flyte

Completion Latency
===================
Time it takes to mark the workflow as complete and accumulate outputs of a workflow after the last node has completed.

**This is one of the latencies that the core team is working on optimizing**

Overview of various latencies in flytepropeller
=================================================

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