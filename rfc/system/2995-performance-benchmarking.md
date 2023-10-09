# Performance Metrics and Benchmarking

**Authors:**

- @hamersaw

## 1 Executive Summary

In designing a performance benchmarking framework for Flyte it is important to clearly define the goal. A naive suggestion may be to "make workflows execute faster", but that could be easier to attain by horizontally or vertically scaling k8s clusters in most scenarios rather than improve Flyte performance. We argue that directing effort at making workflow executions faster is the wrong approach. Rather, the primary focus should be to reduce the overhead that Flyte imposes on workflow evaluations. In pursuing this goal, we then achieve better efficiency and faster workflow evaluations.

To these ends, we first attempt to quantify this overhead by presenting an approach for a lower-bound estimation of node and workflow runtime overhead. We then dive into orchestration scoped metrics to identify what exactly within Flyte incurs the overhead by proposing integration of telemetry solution to provide a fine-grained time-series breakdown of workflow evaluations. With these robust metrics, we design an experimental setup which provides diversity in k8s / Flyte environments and facilitates benchmarking problematic Flyte workflows. This process, and the specific components thereof, are intended to be easily understandable, objective, and reproducible to ensure utility for both dynamic performance analysis in production environments and static benchmarking of feature proposals.

We should preface this discussion by highlighting that the scope of this proposal is meant to be aggressive. It is very likely this integration, in the scopes that are implemented, will happen in multiple tiers and should add benefit with each step.

## 2 Motivation

Performance evaluation is paramount to understanding and improving platform efficiency. However, thoughtless, or otherwise ill-suited metrics / benchmarks can hinder progress which risks unintuitive production behavior (ex. Flyte is much slower than advertised) and wasted development cycles mitigating bottlenecks that are unimportant or non-existent. Therefore, this process needs to be well designed, ensuring metrics and results are objective, specific, and reproducible. Within the Flyte ecosystem, exposing such a collection of information is motivated under two scopes, specifically (1) providing users with objective, actionable metrics to assess and explain Flyte performance and (2) identify infrastructure-level bottlenecks to reduce Flyte-imposed overhead on workflow executions.

To address to former motivation, users are frequently interested in performance evaluations of Flyte. This often includes direct questions like "What is the overhead of Flyte for each workflow?" and "What is Flyte doing to incur this overhead?". These are very fair questions and are important in understanding the viability of Flyte within their ecosystem. However, currently these answers are not straight-forward. They require a significant amount of esoteric knowledge into the Flyte platform. Metrics often relate to very in-depth Flyte architectural and algorithmic design, and exposed through prometheus are poor indicators of single workflow performance. Other solutions require parsing logs, etc - which is far from user-friendly. We must simplify this process.

The later motivation, namely identifying system bottlenecks, is important to explain the overhead imposed by Flyte; and in doing so, to prioritize and quantify the returns on work mitigating these inefficiencies. Generally, our intuition is that Flyte is I/O bound, where things like event reporting, blobstore operations, state persistence, and external API calls are responsible for the lions share of overhead during workflow evaluations. However, it is impossible to quantify the most significant issues without objective evaluation. Additionally, when we propose a change, or develop a new feature, we need a rigorous benchmarking framework to definitively say approach "A" is faster than approach "B" and quantify the trade-offs thereof.

## 3 Proposed Implementation

### Metric Definitions

Providing actionable metrics within Flyte is challenging given the inherit disjoint between runtime and orchestration performance. This may be easiest explained when framed in a single task execution. During task execution, Flyte is still operating in the background periodically checking state, use eventing to report task updates, etc. However, these orchestration operations do not contribute the Flyte-imposed overhead on the overall runtime of that task, but do still represent orchestration cost in executing that task. This means that analysis over orchestration metrics in regards to overall runtime performance may tell a story, but it can not tell the whole story.

For this reason, we believe Flyte metrics should be naturally partitioned into runtime and orchestration scopes. Runtime refers to wall-clock execution time. This should be broken down to quantify the estimated overhead imposed by Flyte, k8s, and other operating frameworks and present these values as ratios with the overall runtime. This makes it trivial to identify scenarios where user-code contributes a relatively small portion of the workflow / node / task runtime. Alternatively, orchestration metrics help explain what Flyte is doing to imposes the overhead. This may include latencies for Flyte event reporting, etcd state persistence, blobstore read / write operations, etc. As previously mentioned, these metric scopes are correlated. For example, reducing the latency of etcd updates will improve runtime performance, but the direct effects are less understood and will certainly vary by workflow definition -- reducing the cost of event reporting by half will not reduce runtime by half, nor will it reduce the Flyte overhead incurred during event reporting by half (because some events are sent during active node / task executions).
            
**Runtime Metrics:** These metrics breakdown the time spent within workflow executions into user-code and various overhead estimates. In theory, this explanation sounds very simple, but the complexities of Flyte can make this challenging in many scenarios.

We have provided a small image below to help describe the myriad scenarios within this definition. The image depicts the timeline-view of node runtimes for a workflow where `node-0` is a dynamic task which sparks two additional nodes, namely `node-0-0` and `node-0-1`; and `node-1` and `node-2` have no dependencies on `node-0` but `node-2` is dependent on `node-1`. Within each node, runtime is broken down by red boxes denoting Flyte overhead, yellow boxes denoting k8s overhead, green boxes denoting user-code execution, and blue boxes denoting subnode execution. These runtimes, and the exposed overheads, are entirely artificial and some have been inflated, others shrunk, compared to real-world expectations for the purposes of easing this definition. 

<p align="center" width="100%">
    <img width="60%" src="https://drive.google.com/uc?export=view&id=1Dl_xgoVBl1wXZjiUhTMBT8s0iIghgyPd" alt="node-runtime"> 
</p>

Perhaps, the best place to start is by defining what we mean by overhead. Within any node execution Flyte performs a variety or orchestration operations to ensure cohesion within the framework. These may include wrangling input data from multiple upstream nodes, using events and etcd writes to update node phases, etc. As outlined in the node runtime image these are typically pre and post processing operations on node executions. Additionally, k8s (and other external systems) require various housekeeping operations to ensure job execution. For example, creating / scheduling Pods and metadata maintenance thereof, pulling container images, managing container runtimes, and so on. This overhead anchors user-code runtimes and shown in the image. Basically, all nodes within Flyte spend a portion of their execution time executing user node, the rest, in some respect, may be attributed to overhead.

It is important to highlight that this overhead differs significantly between node types. For example, executing a `python-task` creates a k8s Pod and then periodically tracks it's status. The overhead here is clear, for example `node-1` or `node-2` have overhead for all pre-processing and post-processing operations. However, this becomes more difficult if the Pod fails after some time. Flyte will create a new retry attempt, but does the original Pod execution count as overhead? This complexity increases when analyzing dynamic tasks, which use a k8s Pod to dynamically compile a Flyte DAG and then proceed to execute that DAG as depicted in the blue box on `node-0` where it's children `node-0-0` and `node-0-1` are executing, or launchplans, which start an entirely separate workflow. Given the extreme complexity, we must reiterate that this overhead is provided only as an estimate and will likely reflect a lower-bound.

Since each node type requires a unique overhead computation it makes sense to define this on each [NodeHandler](https://github.com/flyteorg/flytepropeller/blob/26ad85757c57cda41bf23a2b054d49eccaa8145d/pkg/controller/nodes/handler/iface.go#L10) individually. This will likely require additional metadata with the [NodeState](https://github.com/flyteorg/flytepropeller/blob/master/pkg/controller/nodes/handler/state.go) for each type, for example temporary timestamps that may or may not be persisted to FlyteAdmin through eventing. Overhead estimates may be computed and reported as part of processing terminal phases.

Given overhead estimates for each node, we can aggregate this information to compute an overall workflow overhead estimate. Again, this computation is not well defined, naively we could aggregate all node overhead estimates, but that omits the time Flyte orchestration spends between executing a node after all of it's upstream dependencies have completed. We outline some of the scenarios where this delta can be large below:

- High Latency k8s Watch: FlytePropeller opens a watch on k8s Pods, which means that every time a Pod status updates FlytePropeller is informed. This enables Flyte to immediately detect and process Pod completions rather than waiting periodically. If this watch API has a large latency it could be seconds before Flyte is able to schedule downstream nodes. This is highlighted in the execution of `node-2` after `node-1` completes, there is about half a second where neither node is executing which should be attributed to Flyte overhead.
- Max Parallelism: Flyte workflows can restrict the number of concurrent node executions. This is useful as a defense mechanism, but does mean that a scheduleable node is held back. In the case it is unclear whether this should count as Flyte overhead or not.

In consideration of these complexities we propose to define the workflow overhead estimate as an aggregate of the overhead estimate and scheduling latency at each individual node. This seems to be the most honest and accurate portrayal.

Collecting and correctly reporting this information encompasses it's own challenges. Fortunately, Flyte already incorporates a robust eventing system used to report workflow, node, and task execution information which is then incorporated into the UI. The plan is to compute / collect this information within FlytePropeller and include it in event messages. Specifically, we propose to extend the existing [FlyteIDL event protos](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/event/event.proto) with an `overhead estimation` value which can then be set on the FlyteAdmin [NodeExecution](https://github.com/flyteorg/flyteadmin/blob/master/pkg/repositories/models/node_execution.go#L16) and [Execution (workflow)](https://github.com/flyteorg/flyteadmin/blob/c80bd7c1a608cfd76f316b6cd4a0ceb9a707592a/pkg/repositories/models/execution.go#L20) models. This is the less intrusive change, alternatively we could report individual timestamps and store them in the models to compute overhead on the FlyteAdmin side. This would additionally improve the precision of Flyte runtime tracking in the UI. The correct approach should be debated.

**Orchestration Metrics:** These metrics provide insight into what Flyte is doing to incur overhead. As previously mentioned, these values do not directly correlate with workflow runtimes because Flyte performs orchestration operations during node / task executions.

We propose to integrate a telemetry library within FlytePropeller to emit distributed traces of orchestration operations for each workflow execution. Telemetry traces involve defining a hierarchical collection of spans, where each span relates to single operation with a starting and ending timestamp. Spans typically relate to function invocations, network calls, database reads / writes, etc. This paradigm fits very well with how FlytePropeller orchestrates workflow executions. For example, it uses a queuing system to periodically check various workflow state in _rounds_. Within each round it executes multiple _streaks_ which can include a single change followed by state persistence. Each _streak_ may involve blobstore reads / writes, k8s Pod operations, etc.

An example of using telemetry to define a FlyteWorkflow trace is depicted below. This is presented at a high level, depicting blue boxes for propeller rounds, green boxes for evaluation streaks, yellow boxes for etcd updates on state persistence, and red boxes for eventing, phase updates, blobstore read / writes, etc. This is not meant as an exact replica of the visualization, rather to provide deeper understanding of what is possible. For example, we can see when propeller evaluated a workflow, and within that evaluation how much time was spent performing various operations. This is very important when dissecting performance on a per-workflow basis. 

<p align="center" width="100%">
    <img width="60%" src="https://drive.google.com/uc?export=view&id=1Jlk71Y0JVv2b1vcbCzI2jbPHjgY4o4bo" alt="workflow-trace"> 
</p>

This will involve an additional dependency in FlytePropeller and likely some boilerplate code in FlyteIDL to abstract initialization (if this will be used in other repositories). Many telemetry libraries require simple context decorating to define each span. This can either be manually implemented for fine-grained control or automatically included in each exported function. It is likely the former solution (ie. manually defining) will ensure more well-defined insight into performance without unnecessarily bloating metrics.

To begin, this integration is only necessary within FlytePropeller. However, if we find this is useful it may be worth revisiting integration into FlyteAdmin and DataCatalog as well.

### Reproducible Performance Benchmarking

We have defined a diverse collection of objective metrics which cover both the runtime and orchestration scopes. These serve as a platform to easily analyze infrastructure performance and relate the impact on workflow execution durations. Now we need to define an experimental setup that ensures accurate performance measurements, coverage over a diverse portfolio, and the ability to yield reproducible results. We propose to partition this work into four distinct stages:

1. Provision Benchmarking Infrastructure
2. Deploy Flyte Components
3. Execute Workflow Portfolio
4. Aggregate Benchmark Results

We envision this process will be highly script-driven, standing on the shoulders of existing tooling. We can break each individual stage into a script (or multiple scripts) so that extensions (or reductions) in the breadth of analysis can use a mix-and-match solution. For example, executing benchmarks on existing clusters / Flyte deployments or evaluating additional workload-specific workflow portfolios like spark or ray tasks benchmarks. Initially, this process will be manually driven, but as the approach matures scripting ensures the ability to transition to automated github actions, etc.

**1. Provision Benchmarking Infrastructure:** The infrastructure, and specifically the k8s cluster, is an important component in the performance analysis of Flyte. Different providers and distributions can have significant effect on performance as there can be large variance on particular metrics, for example blobstore read / write performance between solutions and etcd imposes different CRD size limits depending on the k8s distribution. Simply, within the law of diminishing returns, the more environments that can be involved in testing, the more accurate benchmarks will be.
    
We specifically separate infrastructure provisioning because the goal is to support both ad-hoc testing in diverse environments and benchmarks using existing deployments. For the former, we want the ability to dynamically initialize k8s clusters. These should support the major cloud vendors (ie. EKS, GKE, etc) in addition to on-prem solutions (ex. k3s). This space has robust tooling, like terraform, that we suspect can be heavily leaned on to automate this process.

**2. Deploy Flyte Components:** Supporting automated Flyte deployments can be challenging given the extent and depth of k8s configuration. To promote an smooth benchmarking environment we further complicate this process by requiring support for Flyte specific configuration, where the performance of one option should be easily comparable to another. Additionally, we must support benchmarking over local Flyte branches which often contain unreleased features.

Flyte is a microservice architecture, where the framework is partitioned into many individually scalable parts (ex. FlyteConsole, FlyteAdmin, FlytePropeller, etc). Currently deployment is partitioned into two separate scopes, namely single binary and full deployment. The single binary approach compiles all of the Flyte components into, you may have guessed it, a single binary. This is advantageous for quickly deploying Flyte, easing PoC deployments and development / testing iterations. In the scope of benchmarking we need to determine whether this can accurately represent a production-grade environment. If it can, this greatly simplifies the deployment process. If not, we will need to fall back to a full deployment which involves managing multiple Flyte components where complexities are increased with each additional modified component.

Fortunately the base [Flyte repository](https://github.com/flyteorg/flyte) contains [helm charts](https://github.com/flyteorg/flyte/tree/master/charts) for each release. These include both single binary (for the demo cluster) as well as cloud specific default deployments (ex. EKS, GKE, etc). It should be relatively easy to leverage these charts, with some minor modifications (ex. component images for updates, configuration updates) to help drive automated Flyte deployments.

**3. Execute Workflow Portfolio:** We need to be precise about the workflows we use to benchmark Flyte, with the goal to succinctly highlight known bottlenecks. Admittedly, this requires some esoteric knowledge and as performance bottlenecks are mitigating this will likely be an iterative process. While our initial goal is to define 10 - 15 workflows, which span functionality but are easily understandable, it is likely this process may evolve to break into multiple workflow portfolios. For example, one collection which tests Ray or Spark task performance specifically, etc. To begin, below we outline a few ideas:

- Short Running Tasks: Flyte excels at abstracting the costs associated with cloud execution and amortizing them over parallel execution of complex tasks. In scenarios where tasks are short-lived, Flyte overhead will be relatively large and can be responsible for large amount of workflow runtime. It is important to better understand these limitations to help Flyte adapt and support performant short running operations.
- Chained Cache Hits: Cache hits should be zero-cost but operationally require network I/O, database lookups, and blobstore read / writes. These contribute to unintuitive performance and should be minimized.
- Large Workflows: FlytePropeller is designed as a k8s operator and consequently uses a k8s CRD to track workflow status. Within k8s, CRDs are stored in etcd, a KV store in which performance degradations are well documented as the value sizes increase. Understanding the implications of CRD size compared to workflow runtimes can help inform logical workflow structure.
- Large Fan-Out Map Tasks: Map tasks are designed to perform a single operation on a large collection of homogeneous data. Their implementation within Flyte goes further than syntactic sugar; rather, they reduce metadata maintenance to support larger scale than dynamic tasks, etc. To support increasing scales it is important to mitigate issues in executing over large fan outs.
- Nested Dynamic Workflows and Launchplans: Dynamics require k8s to execute a Pod which compiles a Flyte DAG that is subsequently execution. Similarly, launchplans spawn a separate CRD (and FlyteWorkflow) which is then tracked between workflows. Both of these features are very powerful in supporting developmental use-cases - however this does come at a cost.
- Parallelized Subworkflows: Many users logically partition code-bases into workflows which are then used as subworkflows. Within each of these calls, Flyte internally appends a "start" and "end" node to anchor these in the parent workflow which incurs additional costs in data movements, etc.

The process run a single workflow at a time to ensure accurate benchmark metrics. The goal of this work is not to load-test a deployment, but it could be easily adapted in the future. Fortunately, Flyte tooling already includes `flytectl` which can be leveraged to first ensure the workflows are registered in the deployment and then start and monitor each execution to ensure they are serially executed.

**4. Aggregate Benchmark Results** To analyze and compare benchmarking results we need to aggregate performance metrics in a concise document. This process involves querying for workflow / node overheads and gathering summary statistics. These operations should be able to use existing tooling and just rely on parsing and formatting results. Initially this output will be a textual "benchmarking results" document with breakdowns for each individual workflow, an example of YAML format is below:

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

This format is common in our domain, but we could conceivably support JSON, etc in the future. Additionally, it allows script-driven comparisons between evaluations when it is useful to quantify the advantages of one approach over another. For example, it would be relatively simple to provide a script with mimics the YAML hierarchy with delta values such as below:

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

Like literally, this entire thing is designing metrics ... see `"$everything_else"`.

Creating a separate dashboard specifically for performance benchmarks would be useful but is far too involved to offset the costs. However, integration of specific metric components into FlyteConsole would be both possible and helpful. For example, displaying the overhead estimate for each workflow / node in the UI and, if telemetry data is enabled, linking to the tracing agent from the workflow view makes in-depth performance analysis available to end users.

## 5 Drawbacks

The considerations for this proposal are rooted in esoteric knowledge of the Flyte infrastructure to ensure robust and flexible design. This means we can support diversity in environments and workflow definitions. That being said, the main concern is going down the wrong path. Metric reporting often bloats code-bases, where function logic is often prefixed and postfixed by some boilerplate context setting. If this solution proves to be incorrect, removing this should be relatively easy, but remains a non-negligible cost.

## 6 Alternatives

Currently, Flyte emits a collection of metrics through prometheus. This is very powerful in defining SLOs and macro tracking a Flyte deployment, but it is very cumbersome to analyze performance of an individual workflow. This is because prometheus metrics (1) are meant to have bounded label values, so using workflow id results in massive memory utilization and (2) values are reported in quantiles, so individual values are lost and a time-series based analysis is impossible. For these reasons, using existing prometheus metrics to inform performance benchmarking will result in inaccurate and imprecise results.

The [FlytePropeller repository](https://github.com/flyteorg/flytepropeller) contains a script called [fold_logs.py](https://github.com/flyteorg/flytepropeller/blob/master/script/fold-logs.py). This script parses FlytePropeller logs and outputs a hierarchical time-series breakdown of Flytes management of an individual workflow. This output is probably very close to the telemetry data we expect to produce. However, this is based on parsing log messages which ensures that the results will be inaccurate and it is difficult to quantify operations unless specific "start X" and "stop X" logs are recorded. An example output of this script is provided below:

    hamersaw@ragnarok:~/development/flytepropeller$ ./script/fold-logs.py ~/flyte.log fd5d4ee88f9dc4436a76
    Timestamp   Line    Duration    Hierarchical Log Layout
    ----------------------------------------------------------------------------------------------------
    18:08:38    282     14.0s       1 Workflow
    18:08:38    283     0.0s            1.1 Processing
    18:08:38    284     0.0s                1.1.1 StreakRound(Ready)
    18:08:38    289     0.0s                1.1.2 StreakRound(Running)
    18:08:38    292     0.0s                    1.1.2.1 UpdateNodePhase(start-node,NotYetStarted,Succeeded)
    18:08:38    298     0.0s                1.1.3 StreakRound(Running)
    18:08:38    299     0.0s                    1.1.3.1 UpdateNodePhase(n0,NotYetStarted,Queued)
    18:08:38    304     0.0s                1.1.4 StreakRound(Running)
    18:08:38    319     0.0s                    1.1.4.1 UpdateNodePhase(n0,Queued,Running)
    18:08:38    323     0.0s                1.1.5 StreakRound(Running)
    18:08:38    332     0.0s                1.1.6 StreakRound(Running)
    18:08:38    342     0.0s            1.2 Processing
    18:08:38    343     0.0s                1.2.1 StreakRound(Running)
    18:08:39    352     0.0s            1.3 Processing
    18:08:39    353     0.0s                1.3.1 StreakRound(Running)
    18:08:39    361     0.0s                1.3.2 StreakRound(Running)
    18:08:39    371     0.0s            1.4 Processing
    18:08:39    372     0.0s                1.4.1 StreakRound(Running)
    18:08:40    382     0.0s            1.5 Processing
    18:08:40    383     0.0s                1.5.1 StreakRound(Running)
    18:08:50    397     0.0s            1.6 Processing
    18:08:50    398     0.0s                1.6.1 StreakRound(Running)
    18:08:52    407     0.0s            1.7 Processing
    18:08:52    408     0.0s                1.7.1 StreakRound(Running)
    18:08:52    415     0.0s                    1.7.1.1 UpdateNodePhase(n0,Running,Succeeding)
    18:08:52    419     0.0s                1.7.2 StreakRound(Running)
    18:08:52    426     0.0s                1.7.3 StreakRound(Running)
    18:08:52    428     0.0s                    1.7.3.1 UpdateNodePhase(end-node,NotYetStarted,Queued)
    18:08:52    433     0.0s                1.7.4 StreakRound(Running)
    18:08:52    436     0.0s                    1.7.4.1 UpdateNodePhase(end-node,Queued,Succeeded)
    18:08:52    442     0.0s                1.7.5 StreakRound(Running)
    18:08:52    445     0.0s                1.7.6 StreakRound(Succeeding)
    18:08:52    453     0.0s            1.8 Processing

## 7 Potential Impact and Dependencies

Including a telemetry library is going to require additional complexity in Flyte setup. Similar to the existing support for prometheus, this is not going to be required for a Flyte deployment. Rather, it can be turned on / of on demand. Issues in collecting long-running traces at scale are a concern, so at least to begin, integrating always on fine-grained performance analysis in a production environment is likely in-advisable

## 8 Unresolved questions

- Do we compute overhead on the FlytePropeller or FlyteAdmin side? If FlytePropeller means we compute overhead on terminal phases and report a single `overhead-estimate` value in the event. Alternatively, FlyteAdmin requires additional complexity for storing a collection of timestamps and logic to compute overhead thereof.

- What is the most efficient solution for provisioning infrastructure and deploying Flyte? The proposed solution to use terraform and helm seems like a logical approach, but there may be unforeseen issues.

- Is telemetry the best solution for breaking down performance on a per-workflow basis? Specifically, our concerns are the difficulty of integration / collection framework setup and the levels of queryability.

- Do we need to capture CPU, memory, network I/O, etc? Currently, there is no indication that Flyte components are resource bound. This is especially true when comparing to the workloads they orchestrate. However, as this framework progresses it may be important to track this utilization to better inform deployments.

## 9 Conclusion

In this proposal we have defined runtime and orchestration metric definitions to provide robust insight into (1) overhead estimates incurred by executing workflows with Flyte and (2) time-series representations of Flytes orchestration operations, respectively. These metrics will help end-users better understand Flyte performance in production environment and drive deeper performance analysis and proposals. We then outlined a script-driven experimental benchmarking setup which facilitates performance analysis with diversity in k8s and Flyte environments. Specific experiments manifest as a portfolio of Flyte workflows designed with esoteric knowledge to cover a breadth of performance issues. This process is important to help identify system-wide performance bottlenecks and enable quantifying feature improvements. We expect this aggressive solution can be refined and implemented multiple stages and will incrementally provide additional utility.
