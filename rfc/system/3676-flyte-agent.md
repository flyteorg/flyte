# [RFC] Introduce Flyte Agent to Mitigate Pod Cold-Start Overhead During Task Execution

**Authors:**

- @hamersaw

## 1 Executive Summary

In the most basic execution strategy, Flyte uses individual k8s Pods to execute workflow tasks. This means that task durations include inherit overhead while k8s schedules, launches, and monitors the Pod. For short-lived tasks (ex. order of milliseconds) this overhead (ex. 1s - 2s) may be a bottleneck. This document proposes the Flyte Agent, which is an individual application which polls Flyte for work, invoking `pyflyte` directly to execute tasks within a "hot" environment, and reporting task completion immediately. This approach effectively mitigates k8s overhead, but may be deployed within a k8s cluster to maintain many of the advantages. To integrate with the existing framework, we present a lightweight Flyte plugin to enable this execution model within the existing framework.

## 2 Motivation

Flyte executes tasks in isolated k8s Pods. This provides many inherit advantages including fine-grained resource configuration, control over dependencies, workflow isolation, etc. However, cold-starting a k8s Pod for each task execution imposes unavoidable hot-path overhead in k8s resource creation, scheduling, and state tracking among other operations. This overhead is typically quite small (on the order of 2 seconds per task) and, at the scale Flyte is typically deployed, is amortized by parallelizing task executions.

While Flyte excels at orchestrating these relatively long-lived executions, the intrinsic overhead of singular k8s Pod execution is not well suited for low-latency workflows. In these scenarios, task executions are often measurable in milliseconds, and the overhead of cold-starting k8s Pods may be responsible for a large portion of the execution duration. Fortunately, the variance in task requirements (ex. resource requirements, dependencies, etc) tend to correlate with the execution durations -- meaning short-lived tasks are often quite similar. We hope to exploit this characteristic to more effectively schedule low-latency workflows on Flyte by keeping k8s Pods "hot", so that a single instance may execute multiple tasks, mitigating the cold-start overhead with a new k8s Pod for each individual task execution.

## 3 Proposed Implementation

This document proposes a scheme where a collection of Flyte Agents are deployed providing a "hot" environment for task execution, mitigating the order of seconds overhead in managing individual k8s Pods per task. Because this paradigm is orthogonal to the current approach, there are a number of challenges in seamlessly integrating with Flyte, and these challenges are compounded with each step towards massive scale and optimized performance. Consequently, this solution is submitted as a lightweight, non-intrusive scheme to support low-latency workflows within Flyte and should be treated as a PoC rather than an end solution. A key implementation component will be profiling scale and performance improvements in overhead reduction.

### Solution Architecture
Existing solutions for similar problems may rely on distributed, persistent task queues. While, for massive scale especially, this is a proven solution -- it involves significant distributed architecture, further complicating Flyte integration and the deployment model with additional dependencies. Our design is much more lightweight, leveraging existing Flyte components and constructs to emulate persistence, state consistency, Specifically we introduce two components, namely the Flyte Agent and Flyte Agent Plugin.

**Flyte Agent**: The Flyte Agent is a Rust application that executes python tasks within the same application process to ensure minimal overhead. The PyO3 library enables Rust to initialize Python environments to execute arbitrary code, in this scenario we envision the `pyflyte execute ...` command that Flyte currently uses to execute tasks in k8s Pods.

The Agent uses a long-poll pull-based gRPC request to both retrieve and report the state of task executions. This effectively minimizes the latency of both operations as FlytePropeller, during initial task launch or status check, will immediately have access to an open connection from the appropriate Agent. Additionally, Agent initiated communication enables fast reporting of task completion.

**Flyte Agent Plugin**: The Flyte Agent Plugin is, quite obviously, a Flyte plugin which bridges Flytes native scheduling logic and Flyte Agent execution. By implementing this as a separate plugin, we can enable / disable easily and alternatively, set as the default for specific Flyte task types (ex. "python-task").

The plugin contains a passive gRPC service to handle the pull-based connections from each Flyte Agent. As part of the gRPC request handling, the plugin maintains connection pools of agents based on either the agent ID, for retrieving the next task, or the task ID for reporting task status. Internally, this will be architected using a sequence of golang channels / callback functions to emulate a task queue.

Durability of task executions is maintained using the existing `pluginStateBytes` construct within the FlyteWorkflow CRD. To begin, this may store the owning agent ID, the last successfully sync (for liveness), and the current plugin phase. There are a number of challenges in providing exactly once execution semantics (if desired) during failures between gRPC requests and FlyteWorkflow CR persistence. We discuss a 2-phase commit lie protocol to combat these issues below.

### Protocol Deep Dive
We purposefully introduced the components with a high-level, abstract description as many of the protocol intricacies complicate this view. However, it is important to explicitly outline implementation specifics.

**Task Definition**: Currently, Flyte uses a simple `pyflyte execute ...` command within k8s Pods to describe the task execution. For an initial solution, given the python entrypoint from PyO3 in Rust, this is a natural task definition. Future work may support more verbose task configurations with the potential for agents to have different execution environments beyond python.

**Queue Topics**: The notion of a "queue topic" (also referred to as "topic") is used to link Flyte task definitions with Flyte Agent instances that are qualified to execute them. Within the Flyte Agent configuration, this can be as simple as including a `topic` configuration option. On the flytekit side, we can introduce an additional field into the task decorator, for example:

`@task(topic="test-1")`

It will be implicit that introduction of the `topic` field will make `resources`, `task_config`, and other fields moot. If this is not desired, we can introduce a new flytekit decorator, with something like `@fasttask` to note that this task should be executed on an agent

While the initial implementation will simply map task topics and agents one-to-one and use this mapping to schedule correctly, more advanced topic strategies would enable more precise provisioning. For example, both agents and task could have a list of topics that all must be satisfied. Alternatively, once agents could handle multiple tasks or if agents had a notion of which node they were deployed on, an agent would only accept task executions for a single topic at a time, effectively providing workload isolation at the topic level. Using the topic field and a "user" or "team" would provide secure isolation as the defined granularity.

**Long-Poll Pull-Based Communication**:
The communicative mechanism between the Flyte Agent and the Flyte Agent Plugin is a long-poll gRPC request initiated by the Agent. This was chosen for a few reasons: 
1. FlytePropeller always has an open connection to the agent when either scheduling a new task or retrieving a running tasks status, minimizing overhead.
2. The Agent can immediately notify FlytePropeller when a task completes allowing FlytePropeller to reenqueue the workflow and very quickly process downstream nodes.
3. The number of agents can scale without FlytePropeller updates. New agents are initialized with a random ID, if the Agent fails then the task fails as well so an Agent restart is treated the same a failing over to a new Agent.
4. The Flyte Agent Plugin does not need to maintain an internal cache of task status'. This requires some configurable LRU cache limit similar to the existing web API plugin and consequently out-of-core implementation.

To begin, the messaging protocol expects that each Agent will only execute a single task, and is unable to execute tasks concurrently. When it completes the task, it will wait for the Flyte Agent Plugin to verify state consistency and then the Agent will request the next task. We expect the gRPC messages to look something like:

    message GetTaskRequest {
        // agent ID of the requesting agent so Flyte Agent Plugin knows which agent the task is assigned to
        string agent_id = 1;

        // topic that this agent is requesting tasks for
        string topic = 2;
    }

    message GetTaskResponse {
        // task ID of the task to be executed
        string task_id = 1;

        // the python command that requires execution
        string python_command = 2;
    }

    message ReportTaskStatusRequest {
        // agent ID to verify that this agent is the one the Flyte Agent Plugin expects to be reporting the task state
        string agent_id = 1;

        // task ID of the task that is being executed
        string task_id = 2;

        // the current phase of the task
        uint32 phase = 3;
    }

    message ReportTaskStatusResponse {
        // terminate notes that the agent should complete this task execution and request a new task
        // this is useful for reporting task finalization or aborts
        bool terminate = 1;
    }

Component failures, for both Agent and Flyte Agent Plugin, are handled by using a configurable grace period and tracking the last successful state sync. If there has not been a successful sync in the configured threshold it will failover. In case of the Flyte Agent Plugin, it will attempt to retry the task, in case of the Agent, it will attempt to retrieve a new task.

Perhaps the most difficult aspect of this implementation is handling failures between gRPC communication and state persistence in the Flyte Agent Plugin. For example, the Flyte Agent Plugin assigns a task to Flyte Agent A, before persisting state FlytePropeller fails, then when FlytePropeller restarts it assigns the same task to Flyte Agent B and we have two Agents executing the same task. There are (at least) two approaches to handle this (1) allow these minor inconsistencies, task executions should be measurable in milliseconds and the relatively infrequent occurrence of duplicate executions (only on failures) is a non-issue or (2) introduce additional process to ensure exactly once execution semantics. In the latter case, we could add additional phases in the Flyte Agent Plugin (which are stored in `pluginStateBytes`) where assigning the task to a Flyte Agent is done in one step, and then in another step the Flyte Agent Plugin confirms state persistence (through another plugin phase transition) and approves the Agent to begin executing the task. A similar protocol could be used for consistently reporting task completions, where the Agent will not proceed to the next task until Flyte Agent Plugin confirms task completion state persistence. The obvious trade-off is allowing duplicate executions vs. the additional overhead.

### Deployment Model
The proposed model can support a variety of diverse deployment patterns. Presumably, Flyte Agents will be deployed as k8s Pods and therefore the resource availability will be the container requests / limits. Similarly, the python dependencies available for task execution will need to be included in the container. This is all amenable to k8s autoscaling, so essentially there can be multiple autoscaling worker groups that are each identified using queue topics so Flyte can direct task execution to the correct workerpool.

## 4 Metrics & Dashboards

It is important to properly profile a PoC, especially in this instance where the implementation may be testing viability of core algorithms for a larger scale effort. The two areas that are necessary to focus on are:

**Performance**:
- Latencies to start task and the report task completion when compared to the existing k8s Pod per task execution scheme.
- Overhead of using 2-phase commit like protocol to ensure Flyte Agent and Flyte Agent Plugin state consistency.

**Scalability**
- The maximum number of open polling connections to inform scalability of worker pools.

## 5 Drawbacks

The decision to design the Flyte bridge as a separate Flyte plugin ensures this effort is relatively risk averse. Nothing will be embedded into core Flyte logic, so if it fails removal is very easy. Of course, the main drawback of this decision is that, in simplifying this problem we may be overlooking issues which would have otherwise affected the design decisions.

## 6 Alternatives

The motivation for this proposal is reducing task startup / teardown latency. Under this guise, there are two branches of solution (1) reducing k8s Pod scheduling and state reporting latencies and (2) executing outside of k8s Pods. The former may require modifications, and substantial effort, in container runtime improvements / k8s core codebase which is impossible to deploy in controlled cloud environments (ie. EKS, GKE, etc). The later, something like keeping "hot" environments is the crux of this solution.

It may be beneficial to discuss alternatives as they apply to scope as well. Our scheme focus' on being a lightweight PoC to test viability of the Flyte Agent. An alternative may be to increase the scope and target a more scalable solution, where rather than using a singular Flyte Plugin to emulate a persistent task queue we could introduce a distributed architecture. We believe that given the lack of understand about need or interest in low-latency workflows when coupled with the expanse of untested components, the increased effort for a more complex solution would be unwise.

## 7 Potential Impact and Dependencies

Users of this solution will need to install and manager instance(s) of the Flyte Agent to provide "hot" environments.

This will also be one of, if not the, first effort to introduce Rust into the Flyte ecosystem. The interoperability with Python presents a natural paring. This will involve additional effort in constructing manageable CI/CD pipelines for a new `flyteagent` repository and any complexities thereof.

## 8 Unresolved questions

There are a number of unresolved questions that, though solutions may not be available until implementation, should still be presented openly:

1. Is this long-poll pull-based gRPC connection the right communication protocol?
2. How to handle failover in task schedules / completions? Given millisecond task execution times is re-execution acceptable or do we need to introduce the overhead of a 2-phase commit like algorithm to ensure exactly once execution semantics?

## 9 Conclusion

In this document, we proposed a solution to mitigate k8s overhead when executing Flyte tasks. This involves implementing a Flyte Agent, which uses Rust and Py03 to execute `pyflyte` in a "hot" environment. Additionally, it uses pull-based long-poll gRPC connections to ensure minimal overhead during task retrieval and completion reporting. Within the Flyte framework, we integrate through a new plugin, namely the Flyte Agent Plugin, which uses connection pools to manage task scheduling / status reporting. Effectively, this scheme provides a lightweight, minimally intrusive solution to support short-lived tasks with the Flyte execution model.
