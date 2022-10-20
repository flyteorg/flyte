# Performance Metrics and Benchmarking

**Authors:**

- @hamersaw

## 1 Executive Summary

In designing a performance benchmarking framework for Flyte it is important to clearly define the goal. A naive suggestion may be to "make workflows execute faster", but that could be easier to attain by horizontally or vertically scaling k8s clusters in most scenarios rather than improve Flyte performance. We argue that directing effort at making workflow executions faster is the wrong approach. Rather, the primary focus should be to reduce the overhead that Flyte imposes on workflow evaluations. In purusing this goal, we then achieve better efficiency and faster workflow evaluations.

To these ends, we first attempt to quantify this overhead by presenting an approach for a lower-bound estimation of node and workflow level overhead. We then dive into infrastructure scopped metrics to identify what exactly within Flyte incurs the overhead by proposing integration of a distributing tracing solution to provide a fine-grained time-series breakdown of workflow evaluations. With these robust metrics, we design an experimental setup which provides diversity in k8s / Flyte environments and faciitates benchmarking problematic Flyte workflows. This process, and the specific components thereof, are intended to be easily understandable, objective, and reproducible to ensure utility for both dynamic performance analysis in production environments and static benchmarking of feature proposals.

We should preface this discussion by highlighting that the scope of this proposal is meant to be aggressive. It is very likely this integration, in the scopes that are implemented, will happen in multiple tiers and should add benefit with each step.

## 2 Motivation

Performance evaluation is paramount to understanding and improving platform efficiency. However, thoughtless, or otherwise ill-suited metrics / benchmarks can hinder progress which risks unintuitive production behavior (ex. Flyte is much slower than advertised) and wasted development cycles mitigating bottlenecks that are unimportant or non-existant. Therefore, this process needs to be well designed, ensuring metrics and results are objective, specific, and reproducible. Within the Flyte ecosystem, exposing such a collection of information is motivated under two scopes, specifically (1) providing users with objective, actionable metrics to assess and explain Flyte performance and (2) identify infratructure-level bottlenecks to reduce Flyte-imposed overhead on workflow executions.

To address to former motivation, users are frequently interested in performance evalutions of Flyte. This often includes direct questions like "What is the overhead of Flyte for each workflow?" and "What is Flyte doing to incur this overhead?". These are very fair questions and are important in understanding the viability of Flyte within their ecosystem. However, currently these answers are not straight-forward. They require a significant amount of esoteric knowledge into the Flyte platform. Metrics often relate to very in-depth Flyte architectural and algorithmic design, and exposed through prometheus are poor indicators of single workflow performance. Other solutions require parsing logs, etc - which is far from user-friendly. We must simplify this process.

The later motivation, namely identifying system bottlenecks, is important to explain the overhead imposed by Flyte; and in doing so, to prioritize and quantify the returns on work mitigating these inefficiencies. Generally, our intuition is that Flyte is I/O bound, where things like event reporting, blobstore operations, state persistence, and external API calls are responsible for the lions share of overhead during workflow evaluations. However, it is impossible to quantify the most significant issues without objective evaluation. Additionally, when we propose a change, or develop a new feature, we need a rigerous benchmarking framework to definitively say approach "A" is faster than approach "B" and quantify the tradeoffs thereof.

## 3 Proposed Implementation

### Metric Definitions

Providing actionable metrics within Flyte is challenging given the inherit disjoint between workflow execution performance and infrastructure performance. This may be easiest explained when framed in a single task execution. During task execution, all Flyte infrastucture is still operating in the background, so any work done during a task execution (ex. periodically checking state, use eventing to report task updates, etc) does not contribute the Flyte-imposed overhead on the execution of that task, but it does still represent infrastructural cost in executing that task. This means that analysis over infrastructure-level metrics in regards to overall workfow evaluation performance may tell a story, but it can not tell the whole story.

For this reason, we believe Flyte metrics should be naturally partitioned into workflow execution and infrastructure-level scopes. Workflow execution refers to user perceived performance, for example how long did it take for the workflow to execute? how long for each node? This should also quantify the overhead imposed by Flyte, k8s, and other operating frameworks. Basically, this level of analysis focuses on comparing what percentage of workflow and node execution time is spent in user-code and what is spent on infrastructure management. Alternatively, infrstructure-level metrics help explain what Flyte is doing that imposes the overhead. This may inlcude the cost of event reporting from FlytePropeller to FlyteAdmin, the latency of etcd updates to persist state in the FlyteWorkflow CRD, etc. As previously mentioned, these metric scopes are correlated. For example, reducing the latency of etcd updates will improve workflow execution performance, but the direct effects are less understood and will certainly vary by workflow definition -- reducing the cost of event reporting by half will not reduce workflow execution duration by half, nor will it reduce the Flyte overhead incurred during event reporting by half (because some events are sent during active node / task executions).
            
**Workflow Execution Metrics:** The metrics related to the workflow execution scope are meant to capture user-perceived performance. This is summarizable by breaking down workflow execution durations into user-code execution, Flyte management, and external system overhead. In theory, this explanation sounds very simple, but the complexities of Flyte can make this challenging in certain scenarios.

we have provided a graphic to help explain the different components here - TODO provide graphic

Perhaps, the best place to start is by defining what we mean by overhead. Within any node execution Flyte requires different pre-processing and post-processing operations to ensure cohesion with other nodes. These may inlcude wrangling input data from multiple upstream nodes, using events and etcd writes to update node phases, etc. Additionally, k8s (and other external systems) require various housekeeping operations to ensure job execution. For example, creating / scheduling Pods and metadata maintenance thereof, pulling container images, managing container runtimes, and so on. Basically, all nodes within Flyte spend a portion of their execution time executing user node, the rest, in some respect, may be attributed to overhead.

TODO - will need to be implemented on a per-node basis 
    out example for executing k8s pods is relativley simple
    what about the overhead of dynamic tasks with launch a pod to compile a Flyte DAG and then execute the DAG?
    what about launchplans which start a separate FlyteWorkflow

    the best we can do has to be good enough - but disclaimers that this represents an estimate

The goal of this scope is to provide a simple, easily understandable value to quickly understand the efficiency of a workflow execution. For this we need to aggregate information to compute an overall workflow-level overhead. The aforementioned node-level overhead computation is a great start, but it is only a component. There are additional considerations, perhaps the most significant is the time between when Flyte processes a node after all of it's upstream node dependencies have completed. In this situation, Flyte processes each workflow both every N seconds and when notified that a node has changed. If the later has high latency it could take seconds for Flyte to process a schedulable node after all of it's upstream node dependencys have succeeded. This is depicted in TODO. An additional complexity is the inherit parallelization of workflow executions. If for some time range T1 to T2, task A is actively executing user-code but task B is incurring overhead does this contribute to the overall workflow overhead? This is highlighted at TODO.

In consideration of these complexities we propose to define workflow overhead as an aggregate of the overhead at each individual node. So in our example, the workflow overhead would be computed by adding the execution runtime, Flyte management overhead, and external system overhead components of each individual node and computing the overal workflow overhead as ratios of these summed values. This seems to be the most honest and accurate portrayal.

TODO - and there are many corner cases - max parallelsim is not a bad thing

Collecting and correctly reporting this information encompasses it's own challenges. Fortunately, Flyte already incorporates a robust eventing system used to report workflow, node, and task execution information which is then incorporated into the UI. The plan is to compute / collect this information within FlytePropeller and include it in event messages. This additional information includes scheduling overhead, k8s pod metadata (ex. pod durations, container durations, etc), etc. FlyteAdmin will then aggregate the metric timestamps to provide a cohesive view of workflow and node overhead. Specifically, we extend the FlyteIDL event protos and FlyteAdmin models as such... TODO

TODO - layout FlyteIDL and FlyteAdmin model updates

**Infrastructure Metrics:** The goal of infrastructure metrics is to understand what Flyte is doing to incur overhead. As previously discussed, these values will not directly correlate with overall workflow performance because the infrastructure is continually monitoring and evaluating workflows (including nodes and task) which they are running. So even when user code is being executed there are background processes which send events, persist phases, and perform various blobstore operations.

Currently, FlytePropeller emits a collection of metrics using a prometheus exporter. These may be scoped using labels to partition by workflow ID, node ID, etc. However, this solution still lacks the time-series granularity to precisely disassemble workflow evaluations. For example, we can know that there were 10 events sent and what the distribution of event latencies are, but will not know if the p99 event latency was the first event sent or the last. This is important in debugging system performance. 

We propose to integrate a distributed tracing solution to track and monitor workflow evaluations within FlytePropeller. These systems work by emitting traces for operations performed on a specific entity. This fits well with FlytePropeller, where each trace is responsible for an individual workflow. Additionally, the heirarchical nature of trace spans make following FlytePropeller rounds, streaks, and the breadth-first traversal of nodes visualizable and easy to dissect. Within each node evaluation we can inspect admin event reporting, etcd updates, and blobstore operations. An example trace is depicted below:

    TODO - graphic of trace
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

To begin, this integration is only necessary within FlytePropeller. However, if we find this is useful it may be worth revisiting integration into FlyteAdmin and DataCatalog as well.

### Reproducible Performance Benchmarking

We have defined a diverse collection of objective metrics which cover both the workflow execution and infrastructure scopes. These serve as a platform to easily analyze infrastructure level performance and relate the impact on workflow executions. Now we need to define an experimental setup that ensures accurate performance measurements, coverage over a diverse portfolio, and the ability to yield reproducible results. We propose to partition this work into four distinct stages:

1. **Provision Benchmarking Infrastructure:** TODO
2. **Deploy Flyte Components:** TODO
3. **Execute Workflow Portfolio:** TODO
4. **Aggregate Benchmark Results:** TODO

We envision this process will be highly script-driven, standing on the shoulders of existing tooling. We can break each individual stage into a script (or multiple scripts) so that extensions (or reductions) in the breadth of analysis can use a mix-and-match solution. For example, executing benchmarks on existing clusters / Flyte deployments or evaluating additional workload-specific workflow portfolios like spark or ray tasks benchmarks. Initially, this process will be manually driven, but as the approach matures scripting ensures the ability to transition to automated github actions, etc.

**Provision Benchmarking Infrastructure:** The infrastructure, and specifically the k8s cluster, is an important component in the performance analysis of Flyte. Different providers and distributions can have signficant effect on performance as there can be large variance on particular metrics, for example blobstore read / write performance between solutions and etcd imposes different CRD size limits depending on the k8s distribution. Simply, within the law of diminishing returns, the more environments that can be involved in testing, the more accurate benchmarks will be.
    
We specifically separate infrastructure provisionsing because the goal is to support both ad-hoc testing in diverse environments and benchmarks using existing deployments. For the former, we want the ability to dynamically initialize k8s clusters. These should support the major cloud vedors (ie. EKS, GKE, etc) in addition to on-prem solutions (ex. k3s). This space has robust tooling, like terraform, that we suspect can be heavily leaned on to automate this process.

**Deploy Flyte Components:** Supporting automated Flyte deployments can be challenging given the extent and depth of k8s configuration. To promote an smooth benchmarking environment we further complicate this process by requiring support for Flyte specific configuration, where the performance of one option should be easily comparable to another. Additionally, we must support benchmarking over local Flyte branches which often contain unreleased features.

Flyte is a microservice architecture, where the framework is partitioned into many individually scalable parts (ex. FlyteConsole, FlyteAdmin, FlytePropeller, etc). Currently deployment is partitioned into two separate scopes, namely single binary and full deployment. The single binary approach compiles all of the Flyte components into, you may have guessed it, a single binary. This is advantageous for quickly deploying Flyte, easing PoC deployments and development / testing iterations. In the scope of benchmarking we need to determine whether this can accurately represent a production-grade environment. If it can, this greatly simplifies the deployment process. If not, we will need to fallback to a full deployment which involves managing multiple Flyte components where complexities are increased with each additional modified component.

Fortunately the base [Flyte repository](https://github.com/flyteorg/flyte) contains [helm charts](https://github.com/flyteorg/flyte/tree/master/charts) for each release. These include both single binary (for the demo cluster) as well as cloud specific default deployments (ex. EKS, GKE, etc). It should be relatively easy to leverage these charts, with some minor modifications (ex. component images for updates, configuration updates) to help drive automated Flyte deployments.

**Execute Workflow Portfolio:** We need to be precise about the workflows we use to benchmark Flyte, with the goal to succinctly highlight known bottlenecks. Admittedly, this requires some esoteric knowledge and as performance bottlenecks are mitigating this will likely be an iterative process. While our initial goal is to define 10 - 15 workflows, which span functionality but are easily understandable, it is likely this process may evolve to break into multiple workflow portfolios. For example, one collection which tests Ray or Spark task performance specifically, etc. To begin, below we outline a few ideas:

- Short Running Tasks: Flyte excels at abstracting the costs associated with cloud execution and ammortizing them over parallel execution of complex tasks. In scenarios where tasks are short-lived, Flyte overhead will be relatively large and can be responible for large amount of workflow runtime. It is important to better understand these limitations to help Flyte adapt and support performant short running operations.
- Chained Cache Hits: Cache hits should be zero-cost but operationally require network I/O, database lookups, and blobstore read / writes. These contribute to unintuitive performance and should be minimized.
- Large Workflows: FlytePropeller is designed as a k8s operater and consequently uses a k8s CRD to track workflow status. Within k8s, CRDs are stored in etcd, a KV store in which performance degredations are well documented as the value sizes increase. Understanding the implications of CRD size compared to workflow runtimes can help inform logical workflow structure.
- Large Fan-Out Map Tasks: Map tasks are designed to perform a single operation on a large collection of homogeneous data. Their implementation within Flyte goes further than syntatic sugar; rather, they reduce metadata maintenance to support larger scale than dynamic tasks, etc. To support increasing scales it is important to mitigate issues in executing over large fan outs.
- Nested Dynamic Workflows and Launchplans: Dynamics require k8s to execute a Pod which compiles a Flyte DAG that is subsequently execution. Similarly, launchplans spawn a separate CRD (and FlyteWorkflow) which is then tracked between workflows. Both of these features are very powerful in supporting developmental use-cases - however this does come at a cost.
- Parallelized Subworkflows: Many users logically partition code-bases into workflows which are then used as subworkflows. Within each of these calls, Flyte internally appends a "start" and "end" node to anchor these in the parent workflow which incurs additional costs in data movements, etc.

The process run a single workflow at a time to ensure accurate benchmark metrics. The goal of this work is not to load-test a deployment, but it could be easily adapted in the future. Fortunately, Flyte tooling already includes `flytectl` which can be leveraged to first ensure the workflows are registered in the deployment and then start and monitor each execution to ensure they are serially executed.

**Aggregate Benchmark Results** To analyze and compare benchmarking results we need to aggregate performance metrics in a concise document. This process involves querying for workflow / node overheads and gathering summary statistics. These operations should be able to use existing tooling and just rely on parsing and formatting results. Intially this output will be a textual "benchmarking results" document with breakdowns for each indvidual workflow, an example of YAML format is below:

    flyte.benchmark.chained-cache:
      duration: 1m53s
      flyteOverhead: 14%
      nodes:
      - n1:
        duration: 1m2s
        flyteOverhead: 6%
      - n2:
        ...
      events:
        workflow:
          count: 5
          latency:
            p50: 21ms
            p90: 30ms
            p99: 35ms
        node:
          count: 18
          latency:
            ...
        task:
          ...
      blobstore
        writes:
          count: 20
          latency:
            p50: 98ms
            ...
        reads:
          count: 40
          latency:
            p50: 23ms
            ...
      propeller:
        rounds:
          ...
        streaks:
          ...

This format is common in our domain, but we could conceivably support JSON, etc in the future. Additionally, it allows script-driven comparrisons between evaluations when it is useful to quanitfy the advantages of one approach over another. For example, it would be relatively simple to provide a script with mimics the YAML hierarchy with delta values such as below:

    flyte.benchmark.chained-cache:
      duration: -10s
      flyteOverhead: -2.4%
      nodes:
      - n1:
        duration: -2s
        flyteOverhead: -0.1%
      - n2:
        ...
      events:
        workflow:
          count: +0
          latency:
            p50: +0.1ms
            p90: +0.2ms
            p99: -0.1ms
        node:
          count: -2
          latency:
            ...

## 4 Metrics & Dashboards

Like literally, this entire thing is designing metrics ... see "$everything_else"

Creating a separate dashboard specifically for performance benchmarks would be useful but is far too involved to offset the costs. However, integration of specific metric components into FlyteConsole would be both possible and helpful. For example, displaying the overhead estimate for each workflow / node in the UI and, if distributed tracing is enabled, linking to the tracing agent from the workflow view makes in-depth performance analysis available to end users.

## 5 Drawbacks

The considerations for this proposal are rooted in esoteric knowledge of the Flyte infrastructure to ensure robust and flexible design. This means we can support diversity in environments and workflow definitions. That being said, the main concern is going down the wrong path. Metric reporting often bloats code-bases, where function logic is often prefixed and postfixed by some boilerplate context setting. If this solution prooves to be unuseful, removing this should be relatively easy, but remains a non-negligible cost.

## 6 Alternatives

Currently, Flyte emits a collection of metrics through prometheus. This is very powerful in defining SLOs and macro tracking a Flyte deployment, but it is very cumbersome to analyze performance of an individual workflow. This is because prometheus metrics (1) are meant to have bounded label values, so using workflow id results in massive memory utilization and (2) values are reported in quantiles, so individual values are lost and a time-series based analysis is impossible. For these reasons, using existing prometheus metrics to inform performance benchmarking will result in inaccurate and unprecise results.

The [FlytePropeller repository](https://github.com/flyteorg/flytepropeller) contains a script called [fold_logs.py](https://github.com/flyteorg/flytepropeller/blob/master/script/fold-logs.py). This script parses FlytePropeller logs and outputs a hierarchical time-series breakdown of Flytes management of an individual workflow. This output is probably very close to the distributed trace we expect to produce. However, this is based on parsing log messages which ensures that the results will be inaccurate and it is difficult to quantify operations unless specific "start X" and "stop X" logs are recorded. An example output of this script is provided below:

    TODO - include script output

## 7 Potential Impact and Dependencies

Including a distributed tracing solution is going to require additional complexity in Flyte setup. Similar to the existing support for prometheus, this is not going to be requried for a Flyte deployment. Rather, it can be turned on / of on demand. Issues in collecting long-running traces at scale are a concern, so at least to begin, integrating always on fine-grainded performance analysis in a production environment is likely in-advisable

## 8 Unresolved questions

- What is the most efficient solution for provisioning infrastructure and deploying Flyte? The proposed solution to use terraform and helm seems like a logical approach, but there may be unforeseen issues.

- Is distributed tracing the best solution for breaking down performance on a per-workflow basis? Specifically, our concerns are the difficulty of integration / collection framework setup and the levels of queryability.

- Do we need to capture CPU, memory, network I/O, etc? Currently, there is no indication that Flyte components are resource bound. This is especially true when comparing to the workloads they orchestrate. However, as this framework progresses it may be important to track this utilization to better inform deployments.

## 9 Conclusion

TODO
