# [RFC] Simplifying task resource assignment

**Authors:**

- @katrogan

## 1 Executive Summary

Task resource allocation in Flyte includes the process of setting *CPU, memory, GPU* and *ephemeral storage* requests and limits for containers running Flyte tasks on [kubernetes](https://docs.flyte.org/projects/cookbook/en/latest/native_backend_plugins.html#native-backend-plugins). These resource selections affect pod scheduling decisions and as such sensible defaults ought to be applied when a user doesn't specify requests or limits. This fallback behavior currently exists in Flyte but has been configured and modified organically such that the default value assignment has grown convoluted and unfortunately error-prone.

## 2 Motivation

As the system control plane, flyteadmin is the authoritative store for defaults and per-project (and other) overrides. However, the process of converting a user-defined task to a Kubernetes object is handled by appropriate plugins. Therefore, it makes sense to have task resource resolution occur at execution time while leveraging admin to store and propagate task default values.

## 3 Proposed Implementation

Background
----------
Kubernetes allows users to specify both [requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/). **Requests** are used to schedule pods onto nodes. **Limits** are hard stops that running containers are not permitted to exceed.

In the context of what a Flyte user can specify, flytekit [task decorators](https://docs.flyte.org/projects/flytekit/en/latest/generated/flytekit.task.html#flytekit-task) permit setting both requests and limits. Furthermore, in their workflow definitions, users can specify node-level overrides which supersede static task definition resource values.

In the Flyte back-end, **default** values can be applied as requests and limits when a user omits them from a task specification. Furthermore, **max** values are used to enforce that either user-specified resource requests or limits do not exceed a configured threshold.

Current behavior
----------------
The user-specified task resources, or node-level overrides when they exist, populate the TaskExecutionContext used at execution time. Flytepropeller [handles reconciling](https://github.com/flyteorg/flytepropeller/blob/a7e6ed2762ac7dea677f0a5ba1ca3556b51c262c/pkg/controller/nodes/task/taskexec_context.go#L187,L207) these user-values with system defaults and enforcing max boundaries. Flyteplugins then uses these [resolved values and incorporates](https://github.com/flyteorg/flyteplugins/blob/ab659fa0d973cad98fdabd1ad4d77def5456c3e7/go/tasks/pluginmachinery/flytek8s/container_helper.go#L164,L178) a k8s plugin specific set of defaults as a further fall-back when building task container definitions. This is in short-convoluted and quickly becomes buggy.

Proposal
--------
The `ExecutionConfig` already stores the [admin-resolved values](https://github.com/flyteorg/flytepropeller/blob/master/pkg/apis/flyteworkflow/v1alpha1/execution_config.go#L28,L30) for task resource requests and limits as part of the Workflow CRD. We should pass these values along to each call to invoke a plugin handler when [building a resource](https://github.com/flyteorg/flyteplugins/blob/bd6c1b60f09907706683863187fe387e7e373c0e/go/tasks/pluginmachinery/k8s/plugin.go#L84). By adding an additional `GetPlatformTaskResourceValues() TaskResourceSpec` method to the [TaskExecutionMetadata](https://github.com/flyteorg/flyteplugins/blob/93b339a71b32b8b43cf0e5cf3cfb17ef3dae0b5c/go/tasks/pluginmachinery/core/exec_metadata.go#L24) we can propagate the values from the `ExecutionConfig` and make them accessible to each plugin handler which can then build corresponding kubernetes resources appropriately

### Default Plugin Behavior
When building a task container, the default container task plugin will use user-specified resource values and verify they do not exceed the platform-supplied **max** values. When a request and/or limit for a specific resource (e.g. CPU, memory, etc) is not user-supplied the platform-supplied **default** values will be used. After this resolution is performed, the plugin handler will finally make sure that no resource requests exceed resource limits, leading to an impossible-to-schedule situation.

### K8s Pod Plugin Behavior
This will behave similarly to the default container plugin behavior. Platform-supplied **max** values will be enforced for *all* containers defined in a k8s pod spec for a task. However, platform-supplied **default** values will only be substituted for primary container tasks (see more [here](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/kubernetes/pod/pod.html#sphx-glr-auto-integrations-kubernetes-pod-pod-py)).



## 5 Drawbacks

As the control plane, admin represents the centralized authority responsible for task resource defaults and limits. Resolving individual task resources in admin provides a centralized implementation whereas the move to handle this in plugins disperses resource assignment. This is necessary to support the rich variety of container-less task types but does complicate debugging where and how resource allocation happens.

Introducing this change as always, has the potential for leading to bugs and conflicts. Migrating resource resolution from different components requires maintaining duplicate code in flyteadmin and flyteplugins across a period of several releases until the flyteadmin code can be safely removed.

## 6 Alternatives

*What are other ways of achieving the same outcome?*

## 7 Potential Impact and Dependencies

### Fallback Behavior
Introducing this revamped task resource resolution should not result in existing executions failing! In this case, we must build with backwards compatibility in mind and not depend on the `GetPlatformTaskResourceValues()` method necessarily being populated with values. 

## 8 Unresolved questions

*What parts of the proposal are still being defined or not covered by this proposal?*

## 9 Conclusion

*Here, we briefly outline why this is the right decision to make at this time and move forward!*
