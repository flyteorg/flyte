# [RFC Template] Simplify retry behavior and unify it between plugins

**Authors:**

- [@fellhorn](https://github.com/fellhorn)
- [@hamersaw](https://github.com/hamersaw)
- [@fg91](https://github.com/fg91)

## 1 Executive Summary

Flyte implements a retry mechanism to make workflows robust against failure. This retry mechanism has two different budgets, one for which the user defines the maximum number of retries in the `@task` decorator and one for system failures, which is defined on the platform side.

> We distinguish between Flytekit/system-level exceptions and user exceptions. For instance, if Flytekit encounters an issue while uploading outputs, it is considered a system exception. On the other hand, if a user raises a ValueError due to an unexpected input in the task code, it is classified as a user exception. [(Source)](https://docs.flyte.org/projects/flytekit/en/latest/design/authoring.html#exception-handling)

Especially when it comes to interruptions/node terminations, the details of the retry behavior (which budget a retry counts against and how many retry possibilities are remaining) are intransparent and difficult to understand. The behavior is unfortunately also not consistent between plugins or even within the Pod plugin.

The goal of this RFC is to make the behavior easy to understand and consistent between plugins.

## 2 Motivation

For regular `PythonFunctionTasks` the interruptible behavior (in case of a preemption) is as follows:

The Pod plugin tries to [demistify failures](https://github.com/flyteorg/flyteplugins/blob/dfdf6f95aef7bebff160d6660f5c62f5832c39e4/go/tasks/plugins/k8s/pod/plugin.go#L182) and in case a preemption is detected, [returns a system retriable failure](https://github.com/flyteorg/flyteplugins/blob/dfdf6f95aef7bebff160d6660f5c62f5832c39e4/go/tasks/pluginmachinery/flytek8s/pod_helper.go#L628). 

* For retries number 1 to SystemRetryBudget-1 it runs on a Spot instance. Failed attempts are shown to the user as follows:

<img src="https://github.com/flyteorg/flyte/assets/26092524/8ef59400-066f-434b-ae81-bb61116328e8" width="400" >

* The last retry [happens on a non-interruptible machine to ensure completion of the task](https://github.com/flyteorg/flytepropeller/blob/a3c6e91f19c19601a957b29891437112868845de/pkg/controller/nodes/node_exec_context.go#L213)

(Unintuitively, in case of a `SIGKILL` received during node termination, the failure [is counted towards the user defined retry budget](https://github.com/flyteorg/flyteplugins/blob/dfdf6f95aef7bebff160d6660f5c62f5832c39e4/go/tasks/pluginmachinery/flytek8s/pod_helper.go#L645) in contrast to a preemption.)

When using for instance the kubeflow operators plugin (e.g. Pytorch Task) all preemptions are counted towards the [user retry budget](https://github.com/flyteorg/flyteplugins/blob/master/go/tasks/plugins/k8s/kfoperators/common/common_operator.go#L79). The last attempt is not performed on a non-preemptible node.

Preempted attempts are shown as follows in the console:
<img src="https://github.com/flyteorg/flyte/assets/26092524/4e5beb9d-d247-4ff8-bf9a-2e6b4fa7f73d" width="400" >

The incoherent behavior is intransparent and counterintuitive for platform users. As one can see in the previous screenshots the user can't distinguish which retry budget an attempt counted against. We as platform engineers have been approached multiple times with questions such as: *"Why did my task retry only x times when I specified y times."*

Several users were also surprised to learn that the `retries` parameter in the `@task` decorator has no effect on how many times a task can get preempted - at least in the case of python tasks (not in the case of e.g. pytorch tasks). For longer trainings, a larger number of preemptions can be acceptable to a user. Some of our users remarked they would like to have control over the maximum number of preemptions.

* We would like Flyte to handle interruptible tasks the same, no matter which plugin is responsible.
* We would like users to be able to control how often a task can get preempted.
* We would like the retry mechanism and budgets to be easier to understand for platform users.


## 3 Proposed Implementation

We propose to introduce a new simplified retry behavior (which will have to be activated in the platform configuration and which will **not** be the default behavior in order to not make a breaking change).

This simplified retry behavior does not distinguish between user and system failures but considers a single maximum number of retries (defined by the user in the task decorator) regardless of the cause.

This will require a surprisingly small amount of changes to Flyte:

We propose to:

* Add a configuration flag in FlytePropeller for counting system and user retries the same.
* Apply the new behavior in the [`isEligibleForRetry`](https://github.com/flyteorg/flytepropeller/blob/294f47a18c9b11892f2d1a3573019e85257c3b19/pkg/controller/nodes/executor.go#L440) function where, in contrast to the current logic, if the flag is set, we check if the number of attempts + the number of system failures is larger than the max number of retries set in the task decorator. For the last retries, non-spot instances are used.
    * To not introduce an additional platform config flag, the existing flag `interruptibleFailureThreshold` is reused to decide when to switch to a regular instance. With the existing retry behaviour, this number is a positive integer that is the number of retries that are interruptible. For example, 4 (system failure) retries are allowed where 3 of them are interruptible. With the new retry behaviour, a negative number such as e.g. `-2` means that the last to retries are non-interruptible.
* Add configuration for default number of retries in FlytePropeller. Currently the default maximum number of attempts is hardcoded. This should be exposed in the configuration.


In addition to a much easier to understand behavior, all plugins will automatically count preemptions the same way and users will have control over the maximum number of preemptions acceptable for their workflow.

## 4 Metrics & Dashboards

_NA_

## 5 Drawbacks

Currently the behavior is intransparent and confusing for platform users. We as platform engineers have to look deep into the code to understand which budget a certain failure counts against. Users cannot increase the number of accepted preemptions when their training runs longer. Different plugins handle preemptions in different ways.
We should not leave these quirks untackled.

## 6 Alternatives

### 6.1 Count preemptions towards user retry budget

If we didn't want to try to make the retry behavior easier to understand but only wanted to 1) remove the incoherence between plugins when it comes to preemption handling and 2) give users the ability the control the number of accepted preemptions, we could simply always count preemptions towards the user retry budget.

In this case, [here](https://github.com/flyteorg/flyteplugins/blob/dfdf6f95aef7bebff160d6660f5c62f5832c39e4/go/tasks/pluginmachinery/flytek8s/pod_helper.go#L628), we would return a `PhaseInfoRetryableFailure` instead of a `PhaseInfoSystemRetryableFailure`.

* This would mean that most users wouldn't need to know about the system retry budget and they could adapt the number of allowed preemptions to the length of their training.
* This would also mean that the Pod plugin and e.g. the kubeflow plugin would deal with preemptions the same way.
* If preemptions counted against the user retry budget, we would have to switch to non-preemptible instances for the last retry of the user retry budget [instead of the system retry budget](https://github.com/flyteorg/flytepropeller/blob/a3c6e91f19c19601a957b29891437112868845de/pkg/controller/nodes/node_exec_context.go#L213).

The downside of this approach is that the general retry behavior remains as complex and difficult to understand as it is now.

### 6.2 Demistify failure of Pods belonging to plugins/CRDs

We discussed ways to give plugins access to the status of not only the custom resource they create (e.g. `PyTorchJob`) but also to the statuses of the pods created from the custom resource by the respective third-party operator (e.g. the kubeflow training operator). (This might require changes to the plugin interface, see details [here](https://github.com/flyteorg/flyte/discussions/3793).)

This approach would allow plugins to demistify failures of tasks by looking at the underlying pods which, in contrast to the custom resource, do contain hints of preemptions. This way, we could unify the retry behavior upon preemptions of the different plugins. However, wouldn't make the retry behavior easier to understand and users still wouldn't have control over the number of preemptions which is why we opted to not continue these discussions either.





## 7 Potential Impact and Dependencies

By having to activate the new retry behavior in the platform configuration, we will not break existing workflows.


## 9 Conclusion

With the proposed new retry behavior

* Flyte could handle interruptible tasks the same, no matter which plugin is responsible.
* users would be able to control how often a task can get preempted.
* the retry mechanism and budgets would be much be easier to understand for platform users.
