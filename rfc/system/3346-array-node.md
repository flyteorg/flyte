# [RFC] Introduce ArrayNode for Functionally Complete MapTask

**Authors:**

- @hamersaw

## 1 Executive Summary

We propose a new Flyte Node type, namely ArrayNode, to replace the existing `map_task` backend plugin implementation. Flyte offers `map_task` as a way of parallelizing homogeneous operations, with many users enjoying the efficient evaluation and ergonomic API. However, by implementing this as a backend plugin the evaluation is inherently excluded from much of the core Flyte logic. This has evolved into functionally incomplete subtask executions. The proposed solution will add support for all existing features when executing subtasks in addition to extending support for mapping over all plugins and Flyte Node types. The main challenge is maintaining the performance improvements while supporting the additional feature-set.

## 2 Motivation

The force behind this proposal is to support a functionally complete and correct `map_task` implementation. This is important to minimize unintended behavior (ex. `map_task` does not work as expected) and in doing so, support complex workflow logic with the right abstractions.

Currently `map_task` is implemented as a separate backend plugin, namely `k8s_array`. Consequently, this bypasses a significant amount of the DAG execution login in FlytePropeller, inherently excluding functionality thereof. A brief, non-comprehensive, outline of missing functionality is listed below:
- Only supports mapping over K8s tasks (ex. python tasks, containers tasks, and pod tasks)
- Does not support cache serialize or cache overwriting for subtask executions
- Does not support intra-task checkpointing
- Subtasks are not recoverable during workflow recovery, requiring re-execution
- Handling subtask failures to abort other running subtasks is not handled correctly
- Subtasks may only be defined with a single input value

The proposed solution should fix these aforementioned issues and hopefully stymie additional problems.

## 3 Proposed Implementation

We propose to introduce a new Flyte Node type, namely ArrayNode. The ArrayNode will contain a definition for the mappable subnode, operating on a collection of input / output lists, and integrating into the existing Flyte DAG just like any other Node type. This means a `map_task` will be executed entirely within FlytePropeller, leveraging the existing Node traversal rather than encapsulating as a separate plugin. This is the "secret sauce" in supporting the missing functionality.

The main challenge in this implementation is maintaining support for the specific efficiencies that the current approach employes. The `map_task` is more than syntactic sugar, to achieve better performance it:
- Does not copy individual list input values for each subtask, rather flytekit is informed of the specific input index that it will read directly. This reduces the I/O costs of scaling `map_task`.
- Performs asynchronous batched cache lookups, minimizing the cost of datacatalog RPCs
- Stores subtask state using a bitarray rather than full resolution plugin metadata. This minimization reduces the information on each subtask, allowing for more effective scaling.
- Uses a single `TaskExecutionEvent` to report subtask status', layering the information as a collection of `ExternalResource`. This significantly reduces the cost of eventing to FlyteAdmin by reducing the frequency of messages.

The implementation will require updates in multiple repositories, most notably flyteidl, flytepropeller, flytekit, and flyteconsole.

#### flyteidl
We will need to add a new ArrayNode protobuf message in flyteidl. This can be very similar to the existing ArrayJob message with the addition of a subtask Node definition, we expect something similar to:

    message ArrayNode {
        Node  subtask     = 1;
        int64 parallelism = 2;
        int64 size        = 3;
        oneof success_criteria {
            int64 min_successes     = 4;
            float min_success_ratio = 5;
        }
    }

#### flytepropeller

Adding a Node type in FlytePropeller requires a myriad of borderline boilerplate code. This includes registration time validation to ensure the Node input / output values match with upstream / downstream dependencies and general validation of the Node definition, defining a `NodeStateWriter` / `NodeStateReader` for state persistence, and other support.

The most challenging component is defining the execution of the ArrayNode. We envision this can be logically partitioned into a number of stages:
1. Mock a subtask DAG: FlytePropellers Node evaluation depends on traversing a DAG. In the case of ArrayNode this DAG is a simple fanout of the `map_task` definition. This mocked DAG will then be used to iterate over the `map_task` execution using the exact same code that all other Flyte Nodes progress through.
2. Override `NodeExecutionContext` components in the case of a `TaskNode` subtask: A high priority item is ensuring the same level of performance, and perhaps more importantly compatibility, as the existing `map_task` implementation. To these ends, we need to override the `eventRecorder` so that each subtask execution does not send individual events but rather appends an `ExternalResource` to a single `TaskExecutionEvent`, we need to override the `TaskStateWriter` and `TaskStateReader` so that subtask state persistence is minimal, among others. This will be difficult, but will ensure a backwards compatible and performant evaluation. In the case where this is not possible, we can choose to proceed without this level of emulation, resulting in a less performant and backwards incompatible solution.
3. Execute subtasks using the `RecursiveNodeHandler` call using our mocked and overridden entities. This is a common entry point for all Flyte Node evaluation and ensures good software engineering practices and identical support for Node evaluation in subtasks as all other Flyte Nodes.
 
There are (at least) two very important things to note with this implementation:

**Cache lookups will happen at the TaskHandler level:** This is in parity with the existing `map_task` approach, where cache lookups are batched and performed asynchronously. The purpose for this change is two-fold (1) it ensures no code duplication is necessary to support all caching functionality (ex. cache serialize and cache overwrite) and (2) it ensure compatibility with the unreleased "fast-cache" work, which should improve evaluation of cached Nodes significantly.

**Resolving subtask Node input values is not clearly defined:** Currently, FlytePropeller sets environment variables to inform flytekit which input index the subtask should read for a collection of data values. This ensure performance in that each subtask does not require separate blobstore storage. The main problem is that FlytePropeller needs a single input value to compute the cache key. There are a few potential solutions that we will have to evaluate for viability:
- Creating a flyteidl `Literal` which supports indexing in to a List, so a subtask Node input will be index N of the Literal in the blobstore.
- Storing separate inputs in the blobstore. This is very inefficient and should be used as a last resort.
- Some other fancy solution we hack together in the heat of the moment.

#### flytekit

flytekit will require 3 known updates, but further analysis here will defer to those with more complete context:
1. Compiling `map_task` definitions into new ArrayNode rather than a TaskNode using the `k8s_array` backend plugin.
2. Adding support for multiple input values.
3. Related to 2, it is unclear what the best solution for input value resolving in FlytePropeller is which will have consequences on how data is passed through to flytekit.

#### flyteconsole

The current flyteconsole support for `map_task` (ie. subtasks that are TaskNode) should be backwards compatible with this proposed ArrayNode implementation. However, the additional support for mapping over different Flyte Node types at scale will require some thought and design for an intuitive UI. This information will be reported with the same parent/child relationship that is currently used for subworkflows / dynamic tasks / etc, but these visualizations are not designed to scale into thousands of subnode executions.

## 4 Metrics & Dashboards

If this is implemented correctly, the performance should be very similar to existing `map_tasks` (the only penalty being cache lookups). However, subtasks with different Flyte Node types (ex. subworkflows, dynamic tasks, etc) will be unable to benefit from the performance improvements of execution TaskNodes. This is because using `ExternalResource` field in the `TaskExecutionEvent` is not robust enough to describe the execution, additionally the subtasks will require more state that is available in the current bitarray. For this reason, the performance of executing subtasks (both TaskNode and non-TaskNode) should be collected and compared for a better understanding of the limitations of this solution.

## 5 Drawbacks

The only currently known drawback of this proposal is adding cache lookups for subtasks as a synchronous call. The existing implementation, implementing as a backend plugin, uses a separate service to batch cache lookups that are performed asynchronously. This is marginally more efficient but does introduce it's own problems in complexity. Additionally, while the performance of cache lookups may be slightly degraded, the introduction of ArrayNodes means that caching will be functionality complete, specifically cache serialization and cache overwriting will be supported over `map_task` subtasks.

## 6 Alternatives

The obvious alternative is continuing maintenance on the `k8s_array` backend plugin. This has proved unmaintainable in both bug fixes and support for new features, requiring a significant amount of duplicated logic because code reuse between FlytePropeller and the `k8s_array` plugin is very difficult.

## 7 Potential Impact and Dependencies

First, this implementation should be fully backwards compatible with some minor caveats. We will need to keep the existing `k8s_array` plugin around to execute legacy code as the existing flyteidl definition for these tasks will will refer to the `k8s_array` backend plugin. Conversion between these legacy definitions and the new ArrayNode implementation will not be supported because there may be fields without a 1 to 1 mapping. Rather, the flytekit update will compile all new `map_task` declarations into using the ArrayNode, so conversion of existing `map_task` definitions will require recompiling and registering the task.

Currently, the UI supports displaying `map_task` executions using the `ExternalResource` collection defined in each `TaskExecutionEvent`. This will continue working as expected. However, the UI will need to be updated with support for mapping over other Flyte Node types. The metadata for these executions will be very similar to the parent/child relationships between subworkflows and dynamic tasks.

## 8 Unresolved questions

The two questions exposed during this draft are "what is the overhead of subtask executions?" and "is this overhead acceptable?". The current `map_task` implementation has performance efficiencies that we are trying to mimic, but this is very difficult to do with the level of additional functionality. We believe the additional overhead will be minimal, but it will be important to first work on a minimum viable product to better understand the performance implications and determine early if this is a worthy solution.

Otherwise, the implementation details that are subject are the difficulty of mocking components of the `TaskExecutionContext` to ensure TaskNode subtasks are evaluated efficiently and the correct solution in resolving input values for subtask evaluations. 

## 9 Conclusion

In this RFC we outline a proposal for introducing ArrayNodes to achieve functionally complete map tasks. Our solution ensures `map_task` evaluations flow through the core FlytePropeller logic. This results in support for all existing Flyte features in addition to enabling mapping over all Flyte backend plugins and Node types. We combat the challenges in emulating existing performance improvements by mocking many of the Flyte core execution components to emulate the existing subtask execution. This includes efficiently eventing using `ExternalResources` to minimize event frequency and using a bitarray to store subtask status' to minimize metadata overhead. We propose to begin with a minimum viable product to evaluate this solution in the trade-off between performance and completeness.
