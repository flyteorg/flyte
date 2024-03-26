# [RFC] Runtime Workflow Control Using Signals and Gate Nodes

**Authors:**

- @hamersaw
- @kumare3

## 1 Executive Summary

This RFC describes an approach to integrate signal-driven gates into the Flyte workflow. This scheme provides dynamic, runtime control of workflow executions, where gated nodes require manual interaction to progress and succeed. We envision gate conditionals to include manual signals, argument passing, and sleep durations.

## 2 Motivation

The ability to dynamically control workflow execution at runtime is driven by variety of use cases:
1. A user may manually validate or modify data from task A prior to consuming it in task B. This may include validating the correctness of data before an expensive operation or labelling previously unlabelled data for a model training scenario.
2. A user may pass a parameter to an existing workflow task that must be manually specified or is unknown before workflow execution. For example, specifying an interest rate for expense calculations or operating over a file whose path changes frequently.
3. A task may be delayed for a dynamically controlled duration.

## 3 Proposed Implementation

Implementation of this functionality will require changes in multiple Flyte repositories spanning UI, UX, and backend.

### flytekit API

Additional API components must be designed and implemented in flytekit to facilitate construction of gate nodes with the workflow DAG and evaluation of manual signals. From a high-level, the outlined requirements for this functionality include:
- Clear definition of workflow DAG dependencies: Tasks which consume input relying on a signal are not able to progress until the signal is received. All other workflow tasks should be able to progress.
- Support for multiple signal types: Signalling should allow users to pass boolean, string, etc values to a running workflow.
- Signalling interfaces for multiple sources: Manually designated signalling requires a non-negligible time cost. Although perhaps less powerful, signals may be provided on a time delay or programmatically sent.

Below we provide a few approaches in current discussion for the flytekit API. It is important to note that these have no affect on the backend implementation, rather solutions should focus on usability.

**Approach 1: Use signals to trigger gates in an async / await API**:
Async / await APIs are popular in many languages including [python](https://docs.python.org/3/library/asyncio.html) and [rust](https://rust-lang.github.io/async-book/). This seems to be a natural fit for the Flyte ecosystem, where tasks are asynchronously evaluated within a workflow and resources are passed between using a future / promise paradigm. This approach aims to extend this scheme by defining signals as a top-level flytekit object. Calling `flyte.await()` on a signal object returns the resident value. Below is a brief example of what a workflow using signals within an async / await API might look like:

    BoolSignal = Signal[bool]

    def calclulate_expense(v: int) -> int:
        return v * 2

    def pay(expense: int) -> bool:
        ...
        return true

    @workflow
    def my_wf(a: int, approve: BoolSignal) -> bool:
        e = create_node(calculate_expense, v=a)
        s = flyte.await(approve, timeout=10d)
        p = pay(expense: e)
        e >> s >> p

        c = flyte.sleep(10d)

        return

In this scheme gate nodes are clearly defined by requiring an explicitly defined signal object, so users are aware a signal is necessary to progress. However, it may be difficult to infer task dependencies as tasks and signals are not obviously linked.

**Approach 2: Define gates as operations over variables**
In flytekit the workflow DAG is built by linking together task executions using input / output variables (internally implemented as futures / promises). This approach aims to use this paradigm to gate resources and require signalling to release them. Therefore, flytekit can seamlessly define task dependencies. This requires the introduction of (at least) 3 new functions:

    # requires a signal to approve the resource 'value' (boolean gate)
    flyte.approve(value: t): T

    # releases the resource 'value' after 'duration' has elapsed (time gate)
    flyte.sleep(value: t, duration: int): T

    # retrieve the 'value' send by a signal (value gate)
    flyte.wait(): T

An example workflow is presented below:

    def calclulate_expense(v: int) -> int:
        return v * 2

    def pay(expense: int) -> bool:
        ...
        return true

    # approve example
    @workflow
    def my_approve_wf(a: int) -> bool
        e = calculate_expense(v=a)
        ae = flyte.approve(e)
        return pay(expense=ae)
       
    # wait example
    @workflow
    def my_wait_wf(a: int) -> bool
        b = flyte.wait()
        e = calculate_expense(v=b)
        return pay(expense=e)

    # sleep example
    @workflow
    def my_sleep_wf(a: int) -> bool
        sa = flyte.sleep(duration="2d", a)
        e = calculate_expense(v=sa)
        return pay(expense=e)

This solutions clearly articulates task dependencies by defining gate nodes using one of the 3 functions. However, it abstracts the notion of manual signalling from the workflow code. Instead, users are expected to understand that each of the functions may block indefinitely.

**Approach N: The future**
In weighing the pros / cons of the aforementioned approaches we acknowledge there may be something better. We are very open to additional proposals.

### Backend Architecture and Control Flow
![](https://i.imgur.com/vRy0uPi.png)

FlyteAdmin (potentially with the help of datacatalog?) requires a new service provisioned for maintaining a repository of signals. This can be as simple as a bloom-filter with a backing structured database table.

Users are able to manually provide signals using the UI / flyteconsole by clicking a simple button (with potential input value) or sending a request to a gRPC / RESTful interface.

FlytePropeller requires a new node type, namely GateNode, which contains a timeout value, to ensure forgotten workflows fail gracefully rather than run indefinitely, and a conditional to determine when the gate node succeeds. Initially, this conditional will be either an external signal (returning the passed value) or a sleep duration. In the case of a signal conditional, during iterative FlyteWorkflow CRD evaluations the FlytePropeller instances contacts FlyteAdmin to check if the signal exists. Is it does, FlytePropeller sets the coinciding output variables and progresses; if not, FlytePropeller waits until the next evaluation round and checks again.

## 4 Metrics & Dashboards

Implementation of this API may introduce bottlenecks as the number of tasks waiting for a signal scale. To identify and mitigate these issues it may be useful to track (1) the number of workflows / tasks which are waiting on a signal and (2) the number and latency of repetitive checks for signal existence from flytepropeller to flyteadmin. These metrics will likely be emitted from both flytepropeller and flyteadmin.

## 5 Drawbacks

The proposed functionality will not affect evaluation of existing workflows in anyway. Users need to explicitly define the usage of gate nodes requiring manual signalling. Therefore, the risk is quite small. The only perceived risk is that this feature may become **too popular** and we are forced to implement system-level optimizations to cope with the increased load

## 6 Alternatives

Currently, the discussion on alternatives is quite small. This should evolve as we sift out non-solutions in the python API.

The backend implementation is more clearly defined and should not change depending on the API decisions. The proposed architecture aligns well with the "roles" of each Flyte component. Unless there is significant opposition exploration of alternatives is unnecessary.

## 7 Potential Impact and Dependencies

In its entirety, this functionality will require changes spanning a collection of Flyte repositories. It will certainly touch flyteadmin, flytepropeller, flytekit, and flyteconsole. Potentially requiring changes in datacatalog and flyteidl depending on the implementation specifics. This will obviously require a collaborative effort, as expertise to implement the varying components is spread throughout the team. The features are strictly new and should be able to be iteratively implemented and tested within each component.

As far as introducing additional attack vectors. Most of the communication will be intra-cluster (e.x., checking for a signal), but those that are not will require proper authentication and authorization to mitigate malicious activity. We should be able to leverage the existing infrastructure to this end.

## 8 Unresolved questions

1. What does the flytekit API look like?
This is simultaneously the most important and most unresolved question. It may require additional mock iterations and discussion.

2. What is the structure of a signal? 
The first priority here is to abstract signal ID definitions from the flytekit API, which may be difficult to ensure uniqueness while adequately representing what the signal applies to. The first solution is to keep the signal ID as a free-form string (e.x., "file-timestamp-uuid-exists", "file-timestamp-uuid-validated", etc). An alternative approach is to require a structured signal, potentially based on parameter values. Both approaches may (or may not) associate the signal with a specific workflow, task, or execution ID. Including these options restricts the application of a signal, allowing users to signal a single gate node or multiple instances.

3. Can we integrate with egress events?
Without a notification system excessive signalling requirements may result in frequent timeouts. If users are alerted that a workflow is waiting for a signal it increases the chances for a timely response.

## 9 Conclusion

In this document we have proposed a solution to provide dynamic, runtime control of workflows using signals and gate nodes. The variety of use-cases extend the application of Flyte to better satisfy the needs of our community.  The front-end approaches outline potential APIs to clearly promote this functionality and the backend architecture can scale to cope with integration into large deployments.
