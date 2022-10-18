# executive summary (TODO)
from a high level - performance improvements are an iterative process - these things are a science, not an art
    without proper benchmarking there is no way to determine what needs to be improved, or if an improvement actually helped
    users can report:
        "on this workflow i'm seeing an X% overhead from Flyte - this is irregular"
    and it opens up opportunities to objectively quantify development improvements:
        "we reduced operation X by N% and that reduced flyte overhead by M%"
        "we reduced the number of times we performed X, and that improved workflow execution speeds by up to N%"

From a high-level, performance improvements are an iterative process which may be better explained as a science than an art. They require constant monitoring in diverse workloads and environments. The goal of this work is to expose objective metrics which simplify the understanding of Flyte performance. On top of these metrics we must design

# motivation
Performance evaluation is paramount to understanding and improving platform efficiency. However, thoughtless, or otherwise ill-suited metrics / benchmarks can hinder progress which risks unintuitive production behavior (ex. Flyte is much slower than advertised) and wasted development cycles mitigating bottlenecks that are unimportant or non-existant. Therefore, this process needs to be well designed, ensuring metrics and results are objective, specific, and reproducible. Within the Flyte ecosystem, exposing such a collection of information is motivated under two scopes, specifically (1) providing users with objective, actionable metrics to assess and explain Flyte performance and (2) identify infratructure-level bottlenecks to reduce Flyte-imposed overhead on workflow executions.

To address to former motivation, users are frequently interested in performance evalutions of Flyte. This often includes direct questions like "What is the overhead of Flyte for each workflow?" and "What is Flyte doing to impose this overhead?". These are very fair questions and are important in understanding the viability of Flyte within their ecosystem. However, currently these answers are not straight-forward. They require a significant amount of esoteric knowledge into the Flyte platform. Metrics often relate to very in-depth Flyte architectural and algorithmic design, and exposed through prometheus are poor indicators of single workflow performance. Other solutions require parsing logs, etc - which is far from user-friendly. We must simplify this process.

The later motivation, namely identifying system bottlenecks, is important to explain the overhead imposed by Flyte; and in doing so, to prioritize and quantify the returns on work mitigating these inefficiencies. Generally, our intuition is that Flyte is I/O bound, where things like event reporting, state persistence, and external API calls are responsible for the lions share of overhead during workflow evaluations. However, it is impossible to quantify the most significant issues without objective evaluation. Additionally, when we propose a change, or develop a new feature, we need a rigerous benchmarking framework to definitively say approach "A" is faster than approach "B" and quantify the tradeoffs thereof.

# proposed implementation (TODO)
## metrics
Providing actionable metrics within Flyte is challenging given the inherit disjoint between workflow execution performance and infrastructure performance. This may be easiest explained when framed in a single task execution. All Flyte infrastucture is still operating in the background during the specified execution, so any work done during a task execution (ex. periodically checking state, use eventing to report task updates, etc) does not contribute the Flyte-imposed overhead on the execution of that task. This means that analysis over infrastructure-level metrics in regards to overall workfow evaluation performance may tell a story, but it can not tell the whole story.

For this reason, we believe Flyte metrics should be naturally partitioned into workflow execution and infrastructure-level scopes. Workflow execution refers to user perceived performance, for example how long did it take for the workflow to execute? how long for each node? This should also quantify the overhead imposed by Flyte, k8s, and other operating frameworks. Basically, this level of analysis focuses on comparing what percentage of workflow and node execution time is spent in user-code and what is spent on infrastructure management. Alternatively, infrstructure-level metrics help explain what Flyte is doing that imposes the overhead. This may inlcude the cost of event reporting from FlytePropeller to FlyteAdmin, the latency of etcd updates to persist state in the FlyteWorkflow CRD, etc. As previously mentioned, these metric scopes are correlated. For example, reducing the latency of etcd updates will improve workflow execution performance, but the direct effects are less understood and will certainly vary by workflow definition -- reducing the cost of event reporting by half will not reduce workflow execution duration by half, nor will it reduce the Flyte overhead incurred during event reporting by half (because some events are sent during active node / task executions).
            
#### workflow execution metrics
The workflow execution scope is meant to capture user-perceived performance
    this is easily summarizable by displaying what percentage of workflow execution duration contributes to user-code execution, Flyte management, and external systems.
    in theory, this sounds simple enough, but the complexities of Flyte can make this challenging in certain scenarios.

The metrics related to the workflow execution scope are meant to capture user-perceived performance. This is summarizable by breaking down workflow execution durations into user-code execution, Flyte management, and external system overhead. In theory, this explanation sounds very simple, but the complexities of Flyte can make this challenging in certain scenarios.

perhaps it is most important to begin by defining what we mean by overhead
    within any node execution Flyte requires some pre-processing and post-processing for execution
        these operations may include wrangling input data from multiple upstream nodes, updating node phases both in etcd and using FlyteAdmin eventing, etc
    additionally, k8s (or other external frameworks) require various housekeeping operations to ensure job execution
        in k8s Pod scheduling / creation, pulling images, starting containers, etc
    basically, all nodes within Flyte only spend a portion of that time actually executing user code, the rest may be attributed to overhead in some respect

    we have provided a graphic to help explain the different components here - TODO explain
        TODO - describe graphic

Perhaps, the best place to start is by defining what we mean by overhead. Within any node execution Flyte requires different pre-processing and post-processing operations to ensure cohesion with other nodes. These may inlcude wrangling input data from multiple upstream nodes, using events and etcd writes to update node phases, etc. Additionally, k8s (and other external systems) require various housekeeping operations to ensure job execution. For example, creating / scheduling Pods and metadata maintenance thereof, pulling container images, managing container runtimes, and so on. Basically, all nodes within Flyte spend a portion of their execution time executing user node, the rest, in some respect, may be attributed to overhead.

how do we define overhead at the workflow level?
    node-level overhead can be calculated via above ... TODO

    but, there is additional cost - all upstream nodes are completed and Flyte has not yet begun processing the node
        in this scenario overhead should be attributed based on the notification system (subqueue), etc - if k8s notified on change, it's on Flyte, otherwise it's on k8s

    to further complicate this
    flyte often evaluates tasks in parallel - effectively ammortizing any overhead
        for example, if from some time T1 to T2, task A is incurring front-end overhead (ex. copying inputs from upstream nodes) but task B is actively executing - is this overhead?
        we think yes - just because the overhead is ammortized does not mean it is not effecting workflow execution - it could delay downstream tasks

The goal of this scope is to provide a simple, easily understandable value to quickly understand the efficiency of a workflow execution. For this we need to aggregate information to compute an overall workflow-level overhead. The aforementioned node-level overhead computation is a great start, but it is only a component. There are additional considerations, perhaps the most significant is the time between when Flyte processes a node after all of it's upstream node dependencies have completed. In this situation, Flyte processes each workflow both every N seconds and when notified that a node has changed. If the later has high latency it could take seconds for Flyte to process a schedulable node after all of it's upstream node dependencys have succeeded. An additional complexity is the inherit parallelization of workflow executions. If for some time range T1 to T2, task A is actively executing user-code but task B is incurring overhead does this contribute to the overall workflow overhead?

we propose to define overhead of a workflow as an aggregate of overhead at each individual node
    so in our example above the overal runtime, flyte overhead, and external system overhead is aggregated at the node level
        and percentages of overhead are calculated from those values - may be unintiutive, but seems the most accurate

We propose to define workflow overhead as an aggregate of the overhead at each individual node. ... TODO


how are we going to collect this information?
    use events
    send k8s information (ex. pod duration, container duration, etc)
        other frameworks may not be so straight-forward
    aggregate this on FlytePropeller side?
        currently that's how we do acceptance / queued / etc time

#### infrastructure metrics
uses prometheus ...
    very cumbersome to track an individual workflow
        and the use of quantiles result in some degree of inaccuracy
    ill-suited for performance benchmarking

- use opentelemetry to track admin / propellers evaluation of workflow
    use traces to analyze performance on a per-workflow basis
    the goal here is to identify expensive operations - our intuition is that there are heavily I/O centric:
        admin event reporting - workflow / node / task
        ectd updates to persist state in FlyteWorkflow CRD
        blobstore operations (ex. wrangling node inputs, copying subworkflow / dynamic task outputs)

    the trace is hierarchical:
        workflow-evaluation
            streak-round
                processing node N
                    NodeEvent
                        TaskEvent
                WorkflowEvent
                updating CRD in etcd
            streak-round
                processing node N
                    blobstore copy
            streak-round
                processing node N
                    NodeEvent
            ...

        similar to the current fold_logs.py output
            but rather than being parsed from the logs it is based on actual function invocations - meaning much more accurate

    and emit event logs for when workflow / node / task phases are updated

to begin this integration is only necessary in FlytePropeller 
    but if it is very successful and it is advantageous -> perhaps FlyteAdmin / DataCatalog

## performance benchmarking
we have defined an objective set of metrics to ...
    we need to define an experiment setup ensuring:
        (1) accurate performance measurements
        (2) coverage over a diverse portfolio
        (3) reproducible results
    this can be partitioned into three distinct stages:
        (1) provisioning benchmarking infrastructure
        (2) evaluating over a diverse collection of workflow definitions
        (3) aggregating benchmark results

1. provision benchmarking infrastructure
- dynamically spin-up k8s cluster / have lock on existing cluster
    we need to run a single workflow at a time to ensure experiments are accurate
- deploy flyte (single binary?)

2. run a diverse collection of problematic workflows - the goal is 15 - 20 workflows
    these are designed to highlight known bottlenecks within Flyte
        requires some intuition - but we envision these will changeover time
            should still be capped around 25 maximum test cases

    - large fanout map tasks
    - nested dynamic workflows
    - chained cache hits
    - parallelized subworkflows (overhead of `start_node` / `end_node`)
    - etc

3. collect metrics
- flyte overhead (container runtime vs node runtime vs workflow runtime)
- breakdowns of bottleneck operations (ie. counts / durations)
    - flyteadmin eventing
    - etcd updating performance
    - propeller roundes / streaks
    - etc

# metrics and dashboard
literally, this entire thing is designing metrics.

depends on the level of transparency users prefer
    could potentially include overhead estimates in the UI
    could provide links to the open-telemetry traces in the UI

# drawbacks
- metric values are often relative to a specific environment - we can't cover every single case
    - but this is still a good bit of improvement over the current state

# alternative approaches
prometheus metrics
    focus on aggregate metrics (ex. 99th quantile, etc) - not very queryable at the workflow level
flytepropeller fold_logs.py script
    parses logs to compute traces - not very accurate / accessible

# unresolved questions
- what is the most efficient avenue for provisioning infrastructure?
    - does it make sense to ad-hoc startup / teardown clusters?
    - should we use single binary?
- is tracing the best solution for breaking down performance on a per-workflow basis?
    - how difficult is the integration / collection framework setup?
    - what is the level of queryability?
- do we need to capture cpu, memory, network i/o, etc?
    - none of the infrastructure components have been resource bound (ie. CPU, RAM)
        - but FlytePropeller is heavily cached ...

# conclusion
